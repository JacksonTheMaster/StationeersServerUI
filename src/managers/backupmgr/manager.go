package backupmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

/*
The BackupManager manages backup operations. Each instance is independent with its own config and context.
Background routines (file watching and cleanup) only start when Start() is called. Multiple instances
can coexist but may conflict if configured with overlapping directories.
*/

// Initialize checks for BackupDir and waits until it exists, then ensures SafeBackupDir exists.
// It returns a channel that signals when initialization is complete or an error occurs.
func (m *BackupManager) Initialize(identifier string) <-chan error {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(chan error, 1)

	go func() {
		defer close(result)
		const timeout = 90 * time.Minute
		const pollInterval = 2500 * time.Millisecond
		deadline := time.Now().Add(timeout)

		// Wait for BackupDir to exist
		for time.Now().Before(deadline) {
			if stat, err := os.Stat(m.config.BackupDir); err == nil {
				if stat.IsDir() {
					// Directory exists, proceed
					logger.Backup.Debugf("%s found backup directory: %s", identifier, m.config.BackupDir)
					break
				}
				result <- fmt.Errorf("%s backup path %s is not a directory", identifier, m.config.BackupDir)
				return
			} else if !os.IsNotExist(err) {
				// An error other than "not exists" occurred
				result <- fmt.Errorf("%s error checking backup directory %s: %v", identifier, m.config.BackupDir, err)
				return
			}

			logger.Backup.Debugf("%s waiting for save folder "+m.config.BackupDir+" to be created by Stationeers...", identifier)
			select {
			case <-m.ctx.Done():
				result <- fmt.Errorf("%s I have to go, the config was likely changed: %s", identifier, m.ctx.Err())
				return
			case <-time.After(pollInterval):
				// Continue polling
			}
		}

		if time.Now().After(deadline) {
			result <- fmt.Errorf("%s timeout waiting for backup directory %s to be created", identifier, m.config.BackupDir)
			return
		}

		// Ensure SafeBackupDir exists, create it if it doesn't
		if err := os.MkdirAll(m.config.SafeBackupDir, os.ModePerm); err != nil {
			result <- fmt.Errorf("%s error creating safe backup directory %s: %v", identifier, m.config.SafeBackupDir, err)
			return
		}
		logger.Backup.Debugf("%s created safebackups at %s", identifier, m.config.SafeBackupDir)

		result <- nil
	}()

	return result
}

// Start begins the backup monitoring and cleanup routines
func (m *BackupManager) Start(identifier string) error {
	m.lifecycleMu.Lock()
	if err := m.ctx.Err(); err != nil {
		m.lifecycleMu.Unlock()
		return err
	}
	if m.started {
		m.lifecycleMu.Unlock()
		return nil
	}
	m.started = true
	m.wg.Add(1)
	m.lifecycleMu.Unlock()
	defer m.wg.Done()

	// Wait for initialization to complete
	logger.Backup.Debugf("%s is waiting for save folder initialization...", identifier)
	initResult := <-m.Initialize(identifier)
	if initResult != nil {
		return fmt.Errorf("%s failed to initialize backup manager : %w", identifier, initResult)
	}
	logger.Backup.Infof("%s Backup manager instance started", identifier)

	// Preserve the old startup baseline; polling handles subsequent new/changed files.
	baseline, err := scanBackupFiles(m.ctx, m.config.BackupDir)
	if err != nil {
		return fmt.Errorf("read autosave baseline: %w", err)
	}
	if err := m.ctx.Err(); err != nil {
		return err
	}
	m.wg.Add(1)
	go m.watchBackups(identifier, baseline)

	if config.GetBackupRetentionEnabled() {
		m.wg.Add(1)
		go m.startCleanupRoutine()
	}

	return nil
}

// handleNewBackup processes a newly created backup file
func (m *BackupManager) handleNewBackup(filePath string) error {
	if !isValidBackupFile(filepath.Base(filePath)) {
		return fmt.Errorf("invalid backup filename")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ctx.Err(); err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	relativePath, err := filepath.Rel(m.config.BackupDir, filePath)
	if err != nil {
		logger.Backup.Error("Error getting relative path for " + filePath + ": " + err.Error())
		return err
	}
	dstPath := filepath.Join(m.config.SafeBackupDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		logger.Backup.Error("Error creating destination dir for " + dstPath + ": " + err.Error())
		return err
	}

	if err := copyFile(filePath, dstPath); err != nil {
		logger.Backup.Error("Error copying backup " + fileName + ": " + err.Error())
		return err
	}

	logger.Backup.Debug("Backup successfully copied to safe location: " + dstPath)
	summary, err := ReadSaveSummary(dstPath)
	if err != nil {
		logger.Backup.Warn("Could not read metadata from copied backup " + dstPath + ": " + err.Error())
		return err
	}
	// Consumers such as Discord may perform network I/O. Do not keep the
	// backup manager locked while notifying them.
	go notifyBackupCopied(summary)
	return nil
}

// startCleanupRoutine runs periodic backup cleanup
func (m *BackupManager) startCleanupRoutine() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.RetentionPolicy.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logger.Backup.Debug("Cleanup routine stopped due to context cancellation")
			return
		case <-ticker.C:
			if err := m.Cleanup(); err != nil {
				logger.Backup.Error("Backup cleanup error: " + err.Error())
			}
		}
	}
}

// ListBackups returns information about available backups
// limit: number of recent backups to return (0 for all)
func (m *BackupManager) ListBackups(limit int) ([]BackupSaveFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	saves, err := m.getBackupSaveFiles()

	// Reverse the saves to have newest first
	for i, j := 0, len(saves)-1; i < j; i, j = i+1, j-1 {
		saves[i], saves[j] = saves[j], saves[i]
	}

	if err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(saves) {
		saves = saves[:limit]
	}

	return saves, nil
}

// AnalyzeBackup returns cached deep metadata for the backup index. Deep scans are
// serialized by default so listing many saves cannot saturate the host.
func (m *BackupManager) AnalyzeBackup(ctx context.Context, index int) (SaveAnalysis, error) {
	m.mu.Lock()
	saves, err := m.getBackupSaveFiles()
	if err != nil {
		m.mu.Unlock()
		return SaveAnalysis{}, fmt.Errorf("failed to get backup files: %w", err)
	}
	if index < 0 || index >= len(saves) {
		m.mu.Unlock()
		return SaveAnalysis{}, fmt.Errorf("backup index %d out of range (0-%d)", index, len(saves)-1)
	}
	path := saves[index].SaveFile
	analyzer := m.analyzer
	m.mu.Unlock()

	if analyzer == nil {
		return SaveAnalysis{}, fmt.Errorf("backup analyzer is not initialized")
	}
	return analyzer.Analyze(ctx, path)
}

// GetBackupFileData retrieves backup file data by index for download/transfer
func (m *BackupManager) GetBackupFileData(index int) (*BackupFileData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	saves, err := m.getBackupSaveFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get backup files: %w", err)
	}

	if index < 0 || index >= len(saves) {
		return nil, fmt.Errorf("backup index %d out of range (0-%d)", index, len(saves)-1)
	}

	targetSave := saves[index]
	filePath := targetSave.SaveFile

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	filename := filepath.Base(filePath)

	return &BackupFileData{
		Data:     data,
		Filename: filename,
		Size:     int64(len(data)),
		SaveTime: targetSave.SaveTime,
	}, nil
}

// Shutdown stops all backup operations
func (m *BackupManager) Shutdown() {
	logger.Backup.Debug("Shutting down previous backup manager...")

	m.lifecycleMu.Lock()
	m.cancel()
	m.lifecycleMu.Unlock()

	// Wait for all goroutines to finish
	logger.Backup.Debug("Waiting for background tasks to complete...")
	m.wg.Wait()

	logger.Backup.Debug("Backup manager shut down completely")
}

// NewBackupManager creates a new BackupManager instance
func NewBackupManager(cfg BackupConfig) *BackupManager {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.WaitTime <= 0 {
		cfg.WaitTime = defaultWaitTime
	}

	return &BackupManager{
		config:   cfg,
		analyzer: NewSaveAnalyzer(defaultAnalysisConcurrency, defaultAnalysisCacheSize),
		ctx:      ctx,
		cancel:   cancel,
	}
}
