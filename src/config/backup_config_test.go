package config

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestBackupConfigPreservesExplicitZeroValues(t *testing.T) {
	input := []byte(`{
		"backupKeepNewestCount": 0,
		"backupDailyRetentionDays": 0
	}`)
	var cfg JsonConfig
	if err := json.Unmarshal(input, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.BackupKeepNewestCount == nil || *cfg.BackupKeepNewestCount != 0 {
		t.Fatalf("BackupKeepNewestCount = %v, want explicit zero", cfg.BackupKeepNewestCount)
	}
	if cfg.BackupDailyRetentionDays == nil || *cfg.BackupDailyRetentionDays != 0 {
		t.Fatalf("BackupDailyRetentionDays = %v, want explicit zero", cfg.BackupDailyRetentionDays)
	}
}

func TestBackupConfigDropsLegacyRetentionKeys(t *testing.T) {
	input := []byte(`{
		"isCleanupEnabled": true,
		"backupKeepLastN": 2000,
		"backupKeepDailyFor": 24,
		"backupKeepWeeklyFor": 168,
		"backupKeepMonthlyFor": 730,
		"backupCleanupInterval": 730,
		"backupWaitTime": 30
	}`)
	var cfg JsonConfig
	if err := json.Unmarshal(input, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.BackupRetentionEnabled != nil {
		t.Fatalf("legacy cleanup flag populated the new retention setting: %v", *cfg.BackupRetentionEnabled)
	}

	output, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, legacyKey := range []string{
		"isCleanupEnabled",
		"backupKeepLastN",
		"backupKeepDailyFor",
		"backupKeepWeeklyFor",
		"backupKeepMonthlyFor",
		"backupCleanupInterval",
		"backupWaitTime",
	} {
		if bytes.Contains(output, []byte(`"`+legacyKey+`"`)) {
			t.Fatalf("legacy key %q was written back: %s", legacyKey, output)
		}
	}
}

func TestApplyConfigDisablesUnsafeBackupRetention(t *testing.T) {
	enabled := true
	negative := -1
	zero := 0
	cfg := JsonConfig{
		BackupRetentionEnabled:     &enabled,
		BackupKeepNewestCount:      &negative,
		BackupCleanupIntervalHours: &zero,
	}

	applyConfig(&cfg)

	if BackupRetentionEnabled || *cfg.BackupRetentionEnabled {
		t.Fatal("unsafe backup settings did not disable automatic cleanup")
	}
	if BackupKeepNewestCount != 2 || *cfg.BackupKeepNewestCount != 2 {
		t.Fatalf("BackupKeepNewestCount = %d, want safe default 2", BackupKeepNewestCount)
	}
	if BackupCleanupInterval != 24*time.Hour || *cfg.BackupCleanupIntervalHours != 24 {
		t.Fatalf("BackupCleanupInterval = %s, want safe default 24h", BackupCleanupInterval)
	}
}
