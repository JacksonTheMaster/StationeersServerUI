package backupmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"

	"github.com/fsnotify/fsnotify"
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

	// Wait for initialization to complete
	logger.Backup.Debugf("%s is waiting for save folder initialization...", identifier)
	initResult := <-m.Initialize(identifier)
	if initResult != nil {
		return fmt.Errorf("%s failed to initialize backup manager : %w", identifier, initResult)
	}
	logger.Backup.Infof("%s Backup manager instance started", identifier)

	// Start file watcher
	watcher, err := newFsWatcher(m.config.BackupDir, identifier)
	if err != nil {
		return fmt.Errorf("failed to create autosave watcher: %w", err)
	}
	m.watcher = watcher
	go m.watchBackups(identifier)

	if config.GetBackupRetentionEnabled() {
		go m.startCleanupRoutine()
	}

	return nil
}

// watchBackups monitors the backup directory for new files
func (m *BackupManager) watchBackups(identifier string) {
	m.wg.Add(1)
	defer m.wg.Done()

	logger.Backup.Debugf("%s Starting backup file watcher...", identifier)
	defer logger.Backup.Debugf("%s Backup file watcher stopped", identifier)

	for {
		select {
		case <-m.ctx.Done():
			logger.Backup.Debugf("%s WatchBackups stopped due to context cancellation", identifier)
			return
		case event, ok := <-m.watcher.events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				logger.Backup.Infof("%s New backup file detected: %s", identifier, event.Name)
				m.handleNewBackup(event.Name)
			}
		case err, ok := <-m.watcher.errors:
			if !ok {
				return
			}
			logger.Backup.Errorf("%s Backup watcher error: %s", identifier, err.Error())
		}
	}
}

// handleNewBackup processes a newly created backup file
func (m *BackupManager) handleNewBackup(filePath string) {
	if !isValidBackupFile(filepath.Base(filePath)) {
		return
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		time.Sleep(m.config.WaitTime)

		//// save the world into Head save too if SSCM is enabled
		//if config.GetIsSSCMEnabled() && config.GetIsNewTerrainAndSaveSystem() {
		//	commandmgr.WriteCommand("SAVE")
		//	logger.Backup.Debug("HEAD Save triggered via SSCM")
		//} else {
		//	logger.Backup.Debug("HEAD Save NOT refreshed via SSCM")
		//}

		m.mu.Lock()
		defer m.mu.Unlock()

		fileName := filepath.Base(filePath)
		relativePath, err := filepath.Rel(m.config.BackupDir, filePath)
		if err != nil {
			logger.Backup.Error("Error getting relative path for " + filePath + ": " + err.Error())
			return
		}
		dstPath := filepath.Join(m.config.SafeBackupDir, relativePath)

		if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
			logger.Backup.Error("Error creating destination dir for " + dstPath + ": " + err.Error())
			return
		}

		if err := copyFile(filePath, dstPath); err != nil {
			logger.Backup.Error("Error copying backup " + fileName + ": " + err.Error())
			return
		}

		logger.Backup.Debug("Backup successfully copied to safe location: " + dstPath)
	}()
}

// startCleanupRoutine runs periodic backup cleanup
func (m *BackupManager) startCleanupRoutine() {
	m.wg.Add(1)
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

	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
		logger.Backup.Debug("Context canceled for previous backup manager")
	}

	if m.watcher != nil {
		m.watcher.close()
		m.watcher = nil
		logger.Backup.Debug("File watcher closed")
	}
	m.mu.Unlock()

	// Wait for all goroutines to finish
	logger.Backup.Debug("Waiting for background tasks to complete...")
	m.wg.Wait()

	logger.Backup.Debug("Backup manager shut down completely")
}

// NewBackupManager creates a new BackupManager instance
func NewBackupManager(cfg BackupConfig) *BackupManager {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.WaitTime == 0 {
		cfg.WaitTime = defaultWaitTime
	}

	return &BackupManager{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}
