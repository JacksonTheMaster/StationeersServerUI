package backupmgr

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrInvalidBackupName = errors.New("invalid backup name")

// Names are slash-separated paths relative to Safebackups on every platform.
// Do not clean input: accepting aliases would give one archive several identities.
func validateBackupName(name string) error {
	if !fs.ValidPath(name) || !filepath.IsLocal(filepath.FromSlash(name)) ||
		strings.ContainsAny(name, "\\:") || !isValidBackupFile(name) ||
		strings.ContainsFunc(name, func(r rune) bool { return r < 32 || r == 127 }) {
		return ErrInvalidBackupName
	}
	return nil
}

func selectBackup(m *BackupManager, name string) (BackupSaveFile, error) {
	if err := validateBackupName(name); err != nil {
		return BackupSaveFile{}, err
	}
	if err := loadInventory(m); err != nil {
		return BackupSaveFile{}, err
	}
	m.stateMu.RLock()
	record, exists := m.records[name]
	m.stateMu.RUnlock()
	if !exists {
		return BackupSaveFile{}, fmt.Errorf("backup %q is no longer available: %w", name, os.ErrNotExist)
	}
	return BackupSaveFile{Name: name, SaveTime: recordTimestamp(name, record), Summary: record.Analysis.SaveSummary, SummaryReady: record.SummaryReady}, nil
}

// CheckBackupAvailable lets callers reject stale selections before stopping the game.
// Restore and download check again when they acquire the operation lock.
func CheckBackupAvailable(m *BackupManager, name string) error {
	_, err := backupFilePath(m, name)
	return err
}

func backupFilePath(m *BackupManager, name string) (string, error) {
	if err := m.ctx.Err(); err != nil {
		return "", err
	}
	if _, err := selectBackup(m, name); err != nil {
		return "", err
	}
	root, err := filepath.EvalSymlinks(m.config.SafeBackupDir)
	if err != nil {
		return "", err
	}
	path, err := filepath.EvalSymlinks(filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(name)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(relative) {
		return "", ErrInvalidBackupName
	}
	identity, err := identifySave(path)
	if err != nil {
		return "", err
	}
	m.stateMu.RLock()
	record, exists := m.records[name]
	m.stateMu.RUnlock()
	if !exists {
		return "", os.ErrNotExist
	}
	if identity != recordIdentity(record) {
		return "", fmt.Errorf("backup %q changed since it was last observed", name)
	}
	return path, nil
}

// Unknown archives are visible immediately; metadata arrives in the background.
func (m *BackupManager) getBackupSaveFiles() ([]BackupSaveFile, error) {
	if err := loadInventory(m); err != nil {
		return nil, err
	}
	m.stateMu.RLock()
	saves := make([]BackupSaveFile, 0, len(m.records))
	for name, record := range m.records {
		savedAt := recordTimestamp(name, record)
		saves = append(saves, BackupSaveFile{Name: name, SaveTime: savedAt, Summary: record.Analysis.SaveSummary, SummaryReady: record.SummaryReady})
	}
	m.stateMu.RUnlock()
	sort.Slice(saves, func(i, j int) bool {
		if saves[i].SaveTime.Equal(saves[j].SaveTime) {
			return saves[i].Name < saves[j].Name
		}
		return saves[i].SaveTime.Before(saves[j].SaveTime)
	})
	return saves, nil
}

func recordTimestamp(name string, record backupRecord) time.Time {
	if record.SummaryReady {
		return record.Analysis.SavedAt
	}
	return backupTimestamp(name, record.ModifiedNS)
}

func backupTimestamp(name string, modifiedNS int64) time.Time {
	stamp := strings.TrimSuffix(filepath.Base(name), "_auto.save")
	if parsed, err := time.ParseInLocation("020106_150405", stamp, time.Local); err == nil {
		return parsed
	}
	return time.Unix(0, modifiedNS)
}
