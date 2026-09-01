package backupmgr

import (
	"fmt"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/google/uuid"
)

// GlobalBackupManager is the singleton instance of the backup manager
var GlobalBackupManager *BackupManager

// Track all HTTP handlers that need updating when manager changes
var activeHTTPHandlers []*HTTPHandler

// initMutex ensures thread-safe initialization of the global backup manager
var initMutex sync.Mutex

// InitGlobalBackupManager initializes the global backup manager instance
func InitGlobalBackupManager(bmconfig BackupConfig) error {
	// Lock to prevent concurrent initialization
	initMutex.Lock()
	defer initMutex.Unlock()

	// Shut down existing manager if it exists
	if GlobalBackupManager != nil {
		logger.Backup.Debugf("%s Previous Backup manager found. Shutting it down.", bmconfig.Identifier)
		GlobalBackupManager.Shutdown()
		GlobalBackupManager = nil // Clear the manager to avoid stale references
	}

	logger.Backup.Debugf("%s Creating a global backup manager with ID %s", bmconfig.Identifier, bmconfig.Identifier)
	manager := NewBackupManager(bmconfig)
	GlobalBackupManager = manager

	// Update all active HTTP handlers with the new manager
	for _, handler := range activeHTTPHandlers {
		handler.manager = GlobalBackupManager
	}

	// Do not handle old terrain and save system backups
	if !config.GetIsNewTerrainAndSaveSystem() {
		return fmt.Errorf("the old terrain system and save format are no longer supported by backup manager. Please switch to the new terrain and save system if you wish to continue to use new SSUI features. Alternatively, you can continue to use the old system by using an older version of SSUI (5.8 and below), disabling auto-updates via the config.json file")
	}

	// Start the backup manager in a goroutine to avoid blocking
	go func(m *BackupManager) {
		if err := m.Start(bmconfig.Identifier); err != nil {
			logger.Backup.Warnf("%s Exited: "+err.Error(), bmconfig.Identifier)
		}
	}(manager)

	logger.Backup.Infof("%s Backup manager reloaded successfully", bmconfig.Identifier)
	return nil
}

// RegisterHTTPHandler registers an HTTP handler to be updated when the manager changes
func RegisterHTTPHandler(handler *HTTPHandler) {
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
