package backupmgr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// One worker drains new autosaves before doing another old archive. The detector
// keeps polling while a scan runs, and the startup backlog is attempted once.
func handleBackups(m *BackupManager) {
	defer m.wg.Done()
	backlog := pendingAnalyses(m)
	for m.ctx.Err() == nil {
		name, identity := nextAutosave(m)
		if name != "" {
			if err := handleBackup(m, name, identity); err != nil && m.ctx.Err() == nil {
				logger.Backup.Warnf("Backup handling failed, will retry %s: %v", name, err)
			}
			continue
		}
		if len(backlog) > 0 {
			name, backlog = backlog[0], backlog[1:]
			if _, err := analyzeBackup(m.ctx, m, name); err != nil && m.ctx.Err() == nil && !errors.Is(err, os.ErrNotExist) {
				logger.Backup.Warnf("Backup analysis failed for %s (retry on next start): %v", name, err)
			}
			continue
		}
		select {
		case <-m.ctx.Done():
			return
		case <-m.wake:
		}
	}
}

func pendingAnalyses(m *BackupManager) []string {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	backlog := make([]string, 0, len(m.records))
	for name, record := range m.records {
		if !analysisReady(record) {
			backlog = append(backlog, name)
		}
	}
	// Show useful recent statistics first. DDMMYY names do not sort chronologically.
	sort.Slice(backlog, func(i, j int) bool {
		a, b := recordTimestamp(backlog[i], m.records[backlog[i]]), recordTimestamp(backlog[j], m.records[backlog[j]])
		if a.Equal(b) {
			return backlog[i] > backlog[j]
		}
		return a.After(b)
	})
	return backlog
}

func nextAutosave(m *BackupManager) (string, saveIdentity) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()
	var next string
	for name := range m.pending {
		if next == "" || backupTimestamp(name, m.pending[name].modifiedNS).Before(backupTimestamp(next, m.pending[next].modifiedNS)) {
			next = name
		}
	}
	identity := m.pending[next]
	delete(m.pending, next)
	return next, identity
}

func handleBackup(m *BackupManager, name string, expected saveIdentity) error {
	if !filepath.IsLocal(filepath.FromSlash(name)) || !isValidBackupFile(name) {
		return fmt.Errorf("invalid backup name %q", name)
	}
	if err := loadInventory(m); err != nil {
		return err
	}
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case m.scanGate <- struct{}{}:
	}
	defer func() { <-m.scanGate }()
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ctx.Err(); err != nil {
		return err
	}
	m.stateMu.RLock()
	_, exists := m.records[name]
	done := m.handled[name]
	m.stateMu.RUnlock()
	if exists || done {
		return nil
	}
	source := filepath.Join(m.config.BackupDir, filepath.FromSlash(name))
	destination := filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(name))
	// A file imported since startup is never overwritten, even without a manifest entry.
	if identity, err := identifySave(destination); err == nil {
		m.stateMu.Lock()
		m.records[name] = backupRecord{Size: identity.size, ModifiedNS: identity.modifiedNS}
		m.handled[name] = true
		m.revision++
		m.stateMu.Unlock()
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := validateBackupSave(m.ctx, source, expected); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	temp, err := copyBackupToTemp(source, filepath.Dir(destination))
	if err != nil {
		return err
	}
	defer os.Remove(temp)
	if err := checkCopiedSource(source, expected); err != nil {
		return err
	}
	identity, err := identifySave(temp)
	if err != nil {
		return err
	}
	record, err := scanCopiedBackup(m.ctx, temp, identity)
	if err != nil {
		return err
	}
	// Publish only the completed copy. Operations within this manager share m.mu.
	if err := publishBackup(temp, destination); err != nil {
		return err
	}
	m.stateMu.Lock()
	m.records[name] = record
	m.handled[name] = true
	delete(m.observed, name)
	delete(m.pending, name)
	m.revision++
	m.stateMu.Unlock()
	if err := saveManifest(m); err != nil {
		logger.Backup.Warnf("Backup saved, but its manifest could not be written: %v", err)
	}
	if record.ScanError != "" && m.ctx.Err() == nil {
		logger.Backup.Warnf("Backup saved without deep analysis, will retry on next start: %s: %v", name, record.ScanError)
	}
	notifyLatestAnalysis(m, name, record.Analysis.SaveSummary)
	return nil
}

// Finish validating an already-copied archive even during reload. Deep analysis
// can stop and remain pending, but a good copy must still become a safe backup.
func scanCopiedBackup(ctx context.Context, path string, identity saveIdentity) (backupRecord, error) {
	if err := validateBackupSave(context.WithoutCancel(ctx), path, identity); err != nil {
		return backupRecord{}, err
	}
	summary, err := ReadSaveSummary(path)
	if err != nil {
		return backupRecord{}, err
	}
	record := backupRecord{Size: identity.size, ModifiedNS: identity.modifiedNS, SummaryReady: true, Analysis: SaveAnalysis{SaveSummary: summary}}
	analysis, scanErr := AnalyzeSave(ctx, path)
	if scanErr == nil {
		record.Analysis = analysis
		record.ScanVersion = analysisVersion
	} else {
		record.ScanError = shortScanError(scanErr)
	}
	return record, nil
}

func checkCopiedSource(source string, expected saveIdentity) error {
	after, err := identifySave(source)
	// Rotation may unlink the source after we opened it. Validate the completed
	// temporary copy instead of throwing away what may now be the only copy.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if after != expected {
		return fmt.Errorf("autosave changed during copy")
	}
	return nil
}

func copyBackupToTemp(source, directory string) (string, error) {
	input, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer input.Close()
	output, err := os.CreateTemp(directory, ".backup-*.tmp")
	if err != nil {
		return "", err
	}
	name := output.Name()
	_, err = io.Copy(output, input)
	if err == nil {
		err = output.Sync()
	}
	closeErr := output.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

func publishBackup(temp, destination string) error {
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("backup already exists: %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(temp, destination)
}

func shortScanError(err error) string {
	message := err.Error()
	if len(message) > 512 {
		message = message[:512]
	}
	return message
}

func analyzeBackup(ctx context.Context, m *BackupManager, name string) (SaveAnalysis, error) {
	if err := m.ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}
	if err := ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}
	m.stateMu.RLock()
	record, exists := m.records[name]
	m.stateMu.RUnlock()
	if !exists {
		return SaveAnalysis{}, os.ErrNotExist
	}
	if analysisReady(record) {
		return record.Analysis, nil
	}
	select {
	case <-ctx.Done():
		return SaveAnalysis{}, ctx.Err()
	case m.scanGate <- struct{}{}:
	}
	defer func() { <-m.scanGate }()
	if err := m.ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}
	if err := ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}
	m.stateMu.RLock()
	record, exists = m.records[name]
	m.stateMu.RUnlock()
	if !exists {
		return SaveAnalysis{}, os.ErrNotExist
	}
	if analysisReady(record) {
		return record.Analysis, nil
	}
	path := filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(name))
	expected := recordIdentity(record)
	summary, summaryErr := ReadSaveSummary(path)
	if summaryErr == nil {
		record.Analysis.SaveSummary = summary
		record.SummaryReady = true
	}
	scanErr := validateBackupSave(ctx, path, expected)
	if scanErr == nil {
		record.Analysis, scanErr = AnalyzeSave(ctx, path)
	}
	if scanErr == nil {
		record.SummaryReady = true
		record.ScanVersion = analysisVersion
		record.ScanError = ""
	} else {
		record.ScanError = shortScanError(scanErr)
		if summaryErr == nil {
			record.Analysis.SaveSummary = summary
		}
	}
	after, err := identifySave(path)
	if err != nil {
		return SaveAnalysis{}, err
	}
	if after != expected {
		return SaveAnalysis{}, fmt.Errorf("archive changed during analysis")
	}
	m.stateMu.Lock()
	current, exists := m.records[name]
	if exists && recordIdentity(current) == expected {
		m.records[name] = record
		m.revision++
	}
	m.stateMu.Unlock()
	if !exists {
		return SaveAnalysis{}, os.ErrNotExist
	}
	if record.SummaryReady {
		notifyLatestAnalysis(m, name, record.Analysis.SaveSummary)
	}
	return record.Analysis, scanErr
}

func notifyLatestAnalysis(m *BackupManager, name string, summary SaveSummary) {
	if m.ctx.Err() != nil {
		return
	}
	m.stateMu.RLock()
	latest := name
	latestTime := summary.SavedAt
	for candidate, record := range m.records {
		stamp := recordTimestamp(candidate, record)
		if stamp.After(latestTime) || stamp.Equal(latestTime) && candidate > latest {
			latest, latestTime = candidate, stamp
		}
	}
	m.stateMu.RUnlock()
	if latest == name {
		go notifyBackupCopied(summary)
	}
}
