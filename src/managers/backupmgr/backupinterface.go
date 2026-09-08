package backupmgr

import (
	"sync"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/google/uuid"
)

// GlobalBackupManager is the singleton instance of the backup manager
var GlobalBackupManager *BackupManager

var managerMu sync.RWMutex

// CurrentBackupManager takes a consistent snapshot while config reload swaps instances.
func CurrentBackupManager() *BackupManager {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return GlobalBackupManager
}

// Track all HTTP handlers that need updating when manager changes
var activeHTTPHandlers []*HTTPHandler

// initMutex ensures thread-safe initialization of the global backup manager
var initMutex sync.Mutex

var backupCopiedHandler = struct {
	sync.RWMutex
	handler func(SaveSummary)
}{}

// SetBackupCopiedHandler configures the optional consumer for cheap metadata
// from newly copied safe backups. Passing nil disables the callback.
func SetBackupCopiedHandler(handler func(SaveSummary)) {
	backupCopiedHandler.Lock()
	backupCopiedHandler.handler = handler
	backupCopiedHandler.Unlock()
}

func notifyBackupCopied(summary SaveSummary) {
	backupCopiedHandler.RLock()
	handler := backupCopiedHandler.handler
	backupCopiedHandler.RUnlock()
	if handler != nil {
		handler(summary)
	}
}

// InitGlobalBackupManager initializes the global backup manager instance
func InitGlobalBackupManager(bmconfig BackupConfig) error {
	// Lock to prevent concurrent initialization
	initMutex.Lock()
	defer initMutex.Unlock()

	previous := CurrentBackupManager()
	// Shut down existing manager if it exists
	if previous != nil {
		logger.Backup.Debugf("%s Previous Backup manager found. Shutting it down.", bmconfig.Identifier)
		previous.Shutdown()
	}

	logger.Backup.Debugf("%s Creating a global backup manager with ID %s", bmconfig.Identifier, bmconfig.Identifier)
	manager := NewBackupManager(bmconfig)
	inheritDetectorState(manager, previous)
	managerMu.Lock()
	GlobalBackupManager = manager

	// Update all active HTTP handlers with the new manager
	for _, handler := range activeHTTPHandlers {
		handler.mu.Lock()
		handler.manager = GlobalBackupManager
		handler.mu.Unlock()
	}
	managerMu.Unlock()

	// Start the backup manager in a goroutine to avoid blocking
	go func(m *BackupManager) {
		if err := m.Start(bmconfig.Identifier); err != nil {
			logger.Backup.Warnf("%s Exited: %v", bmconfig.Identifier, err)
		}
	}(manager)

	logger.Backup.Debugf("%s Backup manager reload scheduled", bmconfig.Identifier)
	return nil
}

// RegisterHTTPHandler registers an HTTP handler to be updated when the manager changes
func RegisterHTTPHandler(handler *HTTPHandler) {
	managerMu.Lock()
	defer managerMu.Unlock()
	if GlobalBackupManager != nil {
		handler.mu.Lock()
		handler.manager = GlobalBackupManager
		handler.mu.Unlock()
	}
	activeHTTPHandlers = append(activeHTTPHandlers, handler)
}

// GetBackupConfig returns a properly configured BackupConfig
func GetBackupConfig() BackupConfig {

	id := uuid.New()
	bmIdentifier := "[BM" + id.String()[:6] + "]:"
	return BackupConfig{
		WorldName:     config.GetSaveName(),
		BackupDir:     config.GetConfiguredBackupDir(),
		SafeBackupDir: config.GetConfiguredSafeBackupDir(),
		WaitTime:      defaultWaitTime,
		RetentionPolicy: RetentionPolicy{
			KeepNewestCount: config.GetBackupKeepNewestCount(),
			DailyDays:       config.GetBackupDailyRetentionDays(),
			WeeklyWeeks:     config.GetBackupWeeklyRetentionWeeks(),
			MonthlyMonths:   config.GetBackupMonthlyRetentionMonths(),
			CleanupInterval: config.GetBackupCleanupInterval(),
		},
		Identifier: bmIdentifier,
	}
}

// ReloadBackupManagerFromConfig reloads the global backup manager with the current config. This should be called whenever the config is changed.
func ReloadBackupManagerFromConfig() error {
	// Create a new backupManager config from the global config
	backupConfig := GetBackupConfig()

	// Reinitialize the global backup manager with the new config
	return InitGlobalBackupManager(backupConfig)
}
