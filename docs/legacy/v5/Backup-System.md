!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# Stationeers Server UI - Backup System v3 Documentation

The Backup System has been completely overhauled in version 4.x, introducing "Backup System v2", or, just "Backup system" now. This new system offers improved performance, a cleaner architecture, and enhanced features.
This system is designed to manage game save backups with advanced retention policies and secure storage.


!!! important
    The Backup Manager was overhauled once again in v5, dropping support for the preterrain save trio. If you want to play an older version of Stationeers, use an older version of SSUI. Feel free to ask away on the [SSUI Discord](https://discord.gg/xMPg7A2aDs)

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Components](#core-components)
3. [Backup Process](#backup-process)
4. [Retention Policy](#retention-policy)
5. [API Reference](#api-reference)
6. [Important Notes on Filesystem Support](#important-notes-on-filesystem-support)
7. [Integration Guide](#integration-guide)
8. [Configuration](#configuration)

## Architecture Overview

The Backup System v3 is built with a modular architecture that separates concerns into distinct components:

- **BackupManager**: The central component that coordinates all backup operations.
- **File Watcher**: Monitors backup directories for changes and triggers backup processing.
- **HTTP API**: Provides RESTful endpoints for backup management.

The system handles two types of backup directories:
- **Backup Directory**: Where the game saves backups initially.
- **Safe Backup Directory**: Where copies are stored with retention policies applied.

## Core Components

### BackupManager

The `BackupManager` is the central controller for backup operations. It:

- Initializes the backup system with appropriate configuration
- Starts and manages background monitoring processes
- Coordinates backup and restore operations
- Applies retention policies through the cleanup process

```go
type BackupManager struct {
    config  BackupConfig
    mu      sync.Mutex
    watcher *fsWatcher
    ctx     context. Context
    cancel  context.CancelFunc
}
```

Each BackupManager instance is independent with its own configuration, making it possible to manage multiple game worlds (or games..Am I onto something?!) in the future.

### File Watcher

The system uses `fsnotify` to monitor for new backup files. When a new backup file is detected, it:

1. Validates the file format
2. Waits a configurable delay to ensure the file is fully written
3. Copies the file to the safe backup directory

```go
func (m *BackupManager) watchBackups() {
    // Monitors for new backup files and processes them
}
```

### Retention Policy

_This feature was removed_

## Backup Process

The backup process follows this workflow:

1. Game creates backup files in the primary backup directory
2. File watcher detects new backup files
3. After a delay of 30 seconds, the detected save file is copied to the safe backup directory

## API Reference

The backup system exposes RESTful API endpoints for management:

### Endpoints

#### List Backups
```
GET /api/v2/backups?limit={limit}
```
- **Description**: Returns a list of available backups
- **Parameters**:
  - `limit` (optional): Maximum number of backups to return
- **Response**: JSON array of backup groups with metadata

#### Restore Backup
```
GET /api/v2/backups/restore?index={index}
```
- **Description**: Restores a backup with the specified index
- **Parameters**:
  - `index` (required): Index of the backup to restore
- **Response**: Success/error message

### Authentication

All API endpoints are protected by JWT Cookie authentication. Clients must:

1. Authenticate using `/auth/login`
2. Include the JWT Cookie in subsequent requests

## Important Notes on Filesystem Support

### Unsupported Filesystems
 - **NFS v1/v2 (Network File System) mounts**
 - **SMB/CIFS (Windows file sharing) mounts**

### Possibly working Filesystems
 - **NFS v3/v4 (Network File System) mounts**
 - **ZFS, might be slow with caching & make life harder but works™**

### Recommended Filesystems
 - **any block level storage (iSCSI, Fibre channel, plain old ext4)**

### Technical Limitations
 The system relies on `fsnotify` for filesystem monitoring, which cannot:
 - Force cache refreshes on network filesystems
 - Guarantee consistent event notifications across network storage
 - Ensure real-time synchronization between write operations and filesystem events

### Potential Issues
 When using network filesystems:
 1. New backups might not be immediately detected
 2. File existence checks may return inconsistent results
 3. Users might not see backups that actually exist due to delayed filesystem event propagation

### Recommendation
For reliable backup operations, use **local storage** or **block-level network storage** (iSCSI, Fibre Channel) that presents as a local block device to the operating system. If you require network level storage backups, we recommend using `rsync`.

## Integration Guide

### Using the Global Backup Manager

The system provides a global singleton instance for easy integration:

```go
// Initialize the manager
backupConfig := backupsv2.GetBackupConfig()
err := backupsv2.InitGlobalBackupManager(backupConfig)

// Use the manager throughout your application
backups, err := backupsv2.GlobalBackupManager.ListBackups(10)
```

### Creating Custom Backup Managers

_This is not needed and not used in the current state of this Software._
For more advanced use cases, we can create custom instances too:

```go
config := backupsv2.BackupConfig{
    WorldName:     "MyWorld",
    BackupDir:     "./saves/MyWorld/Backup",
    SafeBackupDir: "./saves/MyWorld/Safebackups",
    WaitTime:      30 * time.Second,
    RetentionPolicy: backupsv2.RetentionPolicy{
        KeepLastN:       10,
        KeepDailyFor:    30 * 24 * time.Hour,
        KeepWeeklyFor:   90 * 24 * time.Hour,
        KeepMonthlyFor:  365 * 24 * time.Hour,
        CleanupInterval: 1 * time.Hour,
    },
}

manager := backupsv2.NewBackupManager(config)
if err := manager.Initialize(); err != nil {
    // Handle error
}
if err := manager.Start(); err != nil {
    // Handle error
    As Start calls the cleanup go routines too, be careful with "Starting" a new Manager. Initializing it is generally fine.
}

// Don't forget to shut down when done
defer manager.Shutdown()
```

## Configuration

The default configuration can be obtained through `GetBackupConfig()`:

```go
func GetBackupConfig() BackupConfig {
    backupDir := filepath.Join("./saves/" + config.WorldName + "/Backup")
    safeBackupDir := filepath.Join("./saves/" + config.WorldName + "/Safebackups")

return BackupConfig{
		WorldName:     config.GetSaveName(),
		BackupDir:     config.GetConfiguredBackupDir(),
		SafeBackupDir: config.GetConfiguredSafeBackupDir(),
		WaitTime:      30 * time.Second, // not sure why we are not using config.BackupWaitTime here, but ill not touch it in this commit (config rework)
		RetentionPolicy: RetentionPolicy{
			KeepLastN:       config.GetBackupKeepLastN(),
			KeepDailyFor:    config.GetBackupKeepDailyFor(),
			KeepWeeklyFor:   config.GetBackupKeepWeeklyFor(),
			KeepMonthlyFor:  config.GetBackupKeepMonthlyFor(),
			CleanupInterval: config.GetBackupCleanupInterval(),
		},
		Identifier: bmIdentifier,
	}
```

Key configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| WorldName | Name of the game world | From global config |
| BackupDir | Primary backup directory | `./saves/{WorldName}/Backup` |
| SafeBackupDir | Safe storage directory | `./saves/{WorldName}/Safebackups` |
| WaitTime | Delay before processing new backups | 30 seconds |

---

This documentation is intended for developers. For end-user documentation, please refer to the main wiki.
