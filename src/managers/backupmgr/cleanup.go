package backupmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// Cleanup performs backup cleanup according to retention policy
func (m *BackupManager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ctx.Err(); err != nil {
		return err
	}
	logger.Backup.Debugf("%s Starting backup retention cleanup", m.config.Identifier)

	// Verify and clean the safe backup directory first. If it is unavailable,
	// leave the original autosaves untouched so cleanup cannot remove the only
	// remaining copy.
	if err := m.cleanSafeBackupDir(); err != nil {
		return fmt.Errorf("safe backup dir cleanup failed: %w", err)
	}

	// Clean regular backup dir (keep only recent)
	if err := m.cleanBackupDir(); err != nil {
		return fmt.Errorf("backup dir cleanup failed: %w", err)
	}

	logger.Backup.Debugf("%s Backup retention cleanup complete", m.config.Identifier)
	return nil
}

// cleanBackupDir cleans the regular backup directory
func (m *BackupManager) cleanBackupDir() error {
	files, err := os.ReadDir(m.config.BackupDir)
	if err != nil {
		return err
	}

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Keep only files from last 24 hours

	for _, file := range files {
		if file.IsDir() || !isValidBackupFile(file.Name()) {
			continue
		}

		fullPath := filepath.Join(m.config.BackupDir, file.Name())
		m.stateMu.RLock()
		archived := m.handled[file.Name()]
		record, hasArchive := m.records[file.Name()]
		m.stateMu.RUnlock()
		if !archived || !hasArchive || !analysisReady(record) {
			continue
		}
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			// A cached archive entry is not proof that its file is still intact.
			identity, err := identifySave(filepath.Join(m.config.SafeBackupDir, file.Name()))
			if err != nil {
				return fmt.Errorf("verify archive before source cleanup: %w", err)
			}
			if identity != recordIdentity(record) {
				return fmt.Errorf("archive changed before source cleanup: %s", file.Name())
			}
			if err := os.Remove(fullPath); err != nil {
				logger.Backup.Error("Failed to remove old backup " + fullPath + ": " + err.Error())
			} else {
				logger.Backup.Debugf("%s Removed old autosave %q; analyzed archive verified", m.config.Identifier, file.Name())
			}
		}
	}

	return nil
}

// sameCalendarDay returns true if two times fall on the same calendar day (year + day-of-year).
func sameCalendarDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func startOfCalendarDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func startOfISOWeek(value time.Time) time.Time {
	day := int(value.Weekday())
	if day == 0 {
		day = 7
	}
	return startOfCalendarDay(value).AddDate(0, 0, -(day - 1))
}

func withinDailyWindow(saveTime, now time.Time, days int) bool {
	if days <= 0 {
		return false
	}
	cutoff := startOfCalendarDay(now).AddDate(0, 0, -(days - 1))
	return !saveTime.Before(cutoff)
}

func withinWeeklyWindow(saveTime, now time.Time, weeks int) bool {
	if weeks <= 0 {
		return false
	}
	cutoff := startOfISOWeek(now).AddDate(0, 0, -7*(weeks-1))
	return !saveTime.Before(cutoff)
}

func withinMonthlyWindow(saveTime, now time.Time, months int) bool {
	if months <= 0 {
		return false
	}
	cutoff := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -(months - 1), 0)
	return !saveTime.Before(cutoff)
}

// updateRetentionTrackers updates the daily/weekly/monthly tracker timestamps for a kept backup.
func updateRetentionTrackers(saveTime time.Time, lastKeptDaily, lastKeptWeekly, lastKeptMonthly *time.Time) {
	// Track daily
	if lastKeptDaily.IsZero() || !sameCalendarDay(saveTime, *lastKeptDaily) {
		*lastKeptDaily = saveTime
	}
	// Track weekly
	y1, w1 := saveTime.ISOWeek()
	y2, w2 := lastKeptWeekly.ISOWeek()
	if lastKeptWeekly.IsZero() || y1 != y2 || w1 != w2 {
		*lastKeptWeekly = saveTime
	}
	// Track monthly
	if lastKeptMonthly.IsZero() || saveTime.Month() != lastKeptMonthly.Month() || saveTime.Year() != lastKeptMonthly.Year() {
		*lastKeptMonthly = saveTime
	}
}

// cleanSafeBackupDir cleans the safe backup directory with retention policy
func (m *BackupManager) cleanSafeBackupDir() error {
	// A disconnected mount must not turn retention into deletion of source saves.
	if _, err := os.ReadDir(m.config.SafeBackupDir); err != nil {
		return err
	}
	saves, err := m.getBackupSaveFiles()
	if err != nil {
		return err
	}
	// There are normally only five source saves. Persist their processed names
	// once before any deletions, not one full manifest write per removed archive.
	if m.config.BackupDir != "" {
		sources, err := scanBackupFiles(m.ctx, m.config.BackupDir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			m.stateMu.Lock()
			for path := range sources {
				name, _ := backupName(m.config.BackupDir, path)
				if _, exists := m.records[name]; exists && !m.handled[name] {
					m.handled[name] = true
					m.revision++
				}
			}
			m.stateMu.Unlock()
		}
	}
	if err := saveManifest(m); err != nil {
		return err
	}

	// Sort newest first
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].SaveTime.After(saves[j].SaveTime)
	})

	now := time.Now()
	var (
		lastKeptDaily   time.Time
		lastKeptWeekly  time.Time
		lastKeptMonthly time.Time
	)

	var expired []BackupSaveFile
	knownIndex := 0
	for _, backup := range saves {
		// Never make retention decisions using a placeholder timestamp.
		if !backup.SummaryReady {
			continue
		}
		keepNewest := knownIndex < m.config.RetentionPolicy.KeepNewestCount
		knownIndex++
		// Save timestamps are decoded from Windows FILETIME values as UTC. Apply
		// retention buckets in the server's local calendar so window checks and
		// daily/weekly/monthly grouping use the same day boundaries.
		saveTime := backup.SaveTime.In(now.Location())

		// Always keep the most recent N backups, but also update the retention
		// trackers so the daily/weekly/monthly logic doesn't redundantly keep
		// backups for periods already covered by KeepNewestCount.
		if keepNewest {
			updateRetentionTrackers(saveTime, &lastKeptDaily, &lastKeptWeekly, &lastKeptMonthly)
			continue
		}

		// Keep one backup per calendar day within the configured number of days.
		// Compare full calendar day (year + day-of-year) instead of just day-of-month
		// to avoid incorrectly treating e.g. Jan 15 and Feb 15 as the "same day".
		if withinDailyWindow(saveTime, now, m.config.RetentionPolicy.DailyDays) {
			if lastKeptDaily.IsZero() || !sameCalendarDay(saveTime, lastKeptDaily) {
				updateRetentionTrackers(saveTime, &lastKeptDaily, &lastKeptWeekly, &lastKeptMonthly)
				continue
			}
		}

		// Keep one backup per ISO calendar week within the configured week window.
		if withinWeeklyWindow(saveTime, now, m.config.RetentionPolicy.WeeklyWeeks) {
			year1, week1 := saveTime.ISOWeek()
			year2, week2 := lastKeptWeekly.ISOWeek()
			if lastKeptWeekly.IsZero() || year1 != year2 || week1 != week2 {
				updateRetentionTrackers(saveTime, &lastKeptDaily, &lastKeptWeekly, &lastKeptMonthly)
				continue
			}
		}

		// Keep one backup per calendar month within the configured month window.
		if withinMonthlyWindow(saveTime, now, m.config.RetentionPolicy.MonthlyMonths) {
			if lastKeptMonthly.IsZero() ||
				saveTime.Month() != lastKeptMonthly.Month() ||
				saveTime.Year() != lastKeptMonthly.Year() {
				updateRetentionTrackers(saveTime, &lastKeptDaily, &lastKeptWeekly, &lastKeptMonthly)
				continue
			}
		}

		name := backup.Name
		m.stateMu.RLock()
		hasSource := m.handled[name]
		m.stateMu.RUnlock()
		// A file already missing before cleanup is a repair candidate, not an
		// intentional deletion. Check before recording any retention intent.
		if hasSource {
			if _, err := checkBackupIdentity(m, backup); err != nil {
				return err
			}
		}
		expired = append(expired, backup)
	}

	// Record intentional deletions before removing files, including across a
	// crash between removal and the final manifest write. Only live sources
	// need tombstones; the detector drops them when the game rotates them out.
	m.stateMu.Lock()
	for _, backup := range expired {
		name := backup.Name
		if m.handled[name] && !m.retired[name] {
			m.retired[name] = true
			m.revision++
		}
	}
	m.stateMu.Unlock()
	if err := saveManifest(m); err != nil {
		return err
	}
	for _, backup := range expired {
		if err := deleteBackup(m, backup); err != nil {
			return err
		}
	}

	return saveManifest(m)
}

func deleteBackup(m *BackupManager, saveFile BackupSaveFile) error {
	name, err := checkBackupIdentity(m, saveFile)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(saveFile.Name))); err != nil {
		return err
	}
	m.stateMu.Lock()
	delete(m.records, name)
	m.revision++
	m.stateMu.Unlock()
	logger.Backup.Debugf("%s Removed expired archive %q", m.config.Identifier, name)
	return nil
}

func checkBackupIdentity(m *BackupManager, saveFile BackupSaveFile) (string, error) {
	name := saveFile.Name
	if err := validateBackupName(name); err != nil {
		return "", err
	}
	m.stateMu.RLock()
	record, exists := m.records[name]
	m.stateMu.RUnlock()
	if !exists {
		return "", fmt.Errorf("archive is no longer in the inventory: %s", name)
	}
	identity, err := identifySave(filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(name)))
	if err != nil {
		return "", err
	}
	if identity != recordIdentity(record) {
		return "", fmt.Errorf("archive changed before retention cleanup: %s", name)
	}
	return name, nil
}
