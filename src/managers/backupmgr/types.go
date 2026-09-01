package backupmgr

import (
	"context"
	"sync"
	"time"
)

const (
	defaultWaitTime = 45 * time.Second
)

// BackupConfig holds configuration for backup operations
type BackupConfig struct {
	WorldName       string
	BackupDir       string
	SafeBackupDir   string
	RetentionPolicy RetentionPolicy
	WaitTime        time.Duration
	Identifier      string
}

// RetentionPolicy defines backup retention rules
type RetentionPolicy struct {
	KeepNewestCount int           // Keep newest backups regardless of age
	DailyDays       int           // Keep one representative per calendar day
	WeeklyWeeks     int           // Keep one representative per ISO calendar week
	MonthlyMonths   int           // Keep one representative per calendar month
	CleanupInterval time.Duration // How often to run cleanup
}

type BackupSaveFile struct {
	Index    int
	SaveFile string
	SaveTime time.Time
	Summary  SaveSummary `json:"-"`
}

// BackupFileData contains the backup file bytes and metadata for download/transfer
type BackupFileData struct {
	Data     []byte
	Filename string
	Size     int64
	SaveTime time.Time
}

// BackupManager manages backup operations
type BackupManager struct {
	config   BackupConfig
	mu       sync.Mutex
	watcher  *fsWatcher
	analyzer *SaveAnalyzer
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup // Added for tracking goroutines
}
