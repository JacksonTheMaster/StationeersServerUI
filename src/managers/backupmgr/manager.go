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

// Initialize prepares the archive inventory. Missing autosave folders are handled
// by the detector; they must not prevent analysis of existing safe backups.
func (m *BackupManager) Initialize(identifier string) <-chan error {
	result := make(chan error, 1)
	if err := m.ctx.Err(); err != nil {
		result <- fmt.Errorf("%s I have to go, the config was likely changed: %s", identifier, err)
	} else if err := checkBackupFolders(m.config); err != nil {
		result <- err
	} else if err := os.MkdirAll(m.config.SafeBackupDir, 0755); err != nil {
		result <- err
	} else {
		result <- loadInventory(m)
	}
	close(result)
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
	initialized := false
	defer func() {
		if !initialized {
			m.lifecycleMu.Lock()
			m.started = false
			m.lifecycleMu.Unlock()
		}
	}()

	// Archives can be analyzed before the game creates autosave/.
	if err := <-m.Initialize(identifier); err != nil {
		return fmt.Errorf("load backup inventory: %w", err)
	}
	if err := m.ctx.Err(); err != nil {
		return err
	}
	m.wg.Add(3)
	go watchBackups(m)
	go handleBackups(m)
	go persistInventory(m)
	initialized = true
	logger.Backup.Infof("%s Backup manager instance started", identifier)

	if config.GetBackupRetentionEnabled() {
		m.wg.Add(1)
		go m.startCleanupRoutine()
	}

	return nil
}

// handleNewBackup processes a newly created backup file
func (m *BackupManager) handleNewBackup(filePath string) error {
	identity, err := identifySave(filePath)
	if err != nil {
		return err
	}
	name, err := backupName(m.config.BackupDir, filePath)
	if err != nil {
		return err
	}
	return handleBackup(m, name, identity)
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
	return getBackupAnalysis(ctx, m, index, "")
}

func getBackupAnalysis(ctx context.Context, m *BackupManager, index int, saveFile string) (SaveAnalysis, error) {
	if err := m.ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	save, err := selectBackup(m, index, saveFile)
	if err != nil {
		return SaveAnalysis{}, fmt.Errorf("failed to get backup files: %w", err)
	}
	name, err := backupName(m.config.SafeBackupDir, save.SaveFile)
	if err != nil {
		return SaveAnalysis{}, err
	}
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(m.ctx, cancel)
	defer stop()
	defer cancel()
	return analyzeBackup(ctx, m, name)
}

// GetBackupFileData retrieves backup file data by index for download/transfer
func (m *BackupManager) GetBackupFileData(index int) (*BackupFileData, error) {
	return getBackupFileData(m, index, "")
}

func getBackupFileData(m *BackupManager, index int, saveFile string) (*BackupFileData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ctx.Err(); err != nil {
		return nil, err
	}

	targetSave, err := selectBackup(m, index, saveFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup files: %w", err)
	}

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
	// Wait for any on-demand scan to observe shutdown, too.
	m.scanGate <- struct{}{}
	<-m.scanGate
	// A restore or download that already holds the operation lock may finish.
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := saveManifest(m); err != nil {
		logger.Backup.Warnf("Could not save backup manifest on shutdown: %v", err)
	}

	logger.Backup.Debug("Backup manager shut down completely")
}

// NewBackupManager creates a new BackupManager instance
func NewBackupManager(cfg BackupConfig) *BackupManager {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.WaitTime <= 0 {
		cfg.WaitTime = defaultWaitTime
	}

	if cfg.RetentionPolicy.CleanupInterval <= 0 {
		cfg.RetentionPolicy.CleanupInterval = time.Hour
	}

	return &BackupManager{
		config:   cfg,
		records:  make(map[string]backupRecord),
		handled:  make(map[string]bool),
		observed: make(map[string]saveObservation),
		pending:  make(map[string]saveIdentity),
		wake:     make(chan struct{}, 1),
		scanGate: make(chan struct{}, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}
