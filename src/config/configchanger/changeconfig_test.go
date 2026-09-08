package configchanger

import (
	"testing"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func configInt(value int) *int {
	return &value
}

func validBackupConfig() *config.JsonConfig {
	return &config.JsonConfig{
		BackupKeepNewestCount:        configInt(2),
		BackupDailyRetentionDays:     configInt(7),
		BackupWeeklyRetentionWeeks:   configInt(4),
		BackupMonthlyRetentionMonths: configInt(3),
		BackupCleanupIntervalHours:   configInt(24),
	}
}

func TestValidateBackupSettingsAllowsDisabledRetentionRules(t *testing.T) {
	cfg := validBackupConfig()
	cfg.BackupKeepNewestCount = configInt(0)
	cfg.BackupDailyRetentionDays = configInt(0)
	cfg.BackupWeeklyRetentionWeeks = configInt(0)
	cfg.BackupMonthlyRetentionMonths = configInt(0)

	if err := validateBackupSettings(cfg); err != nil {
		t.Fatalf("validateBackupSettings() returned unexpected error: %v", err)
	}
}

func TestValidateBackupSettingsRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*config.JsonConfig)
	}{
		{name: "negative keep newest", mutate: func(cfg *config.JsonConfig) { cfg.BackupKeepNewestCount = configInt(-1) }},
		{name: "negative daily", mutate: func(cfg *config.JsonConfig) { cfg.BackupDailyRetentionDays = configInt(-1) }},
		{name: "negative weekly", mutate: func(cfg *config.JsonConfig) { cfg.BackupWeeklyRetentionWeeks = configInt(-1) }},
		{name: "negative monthly", mutate: func(cfg *config.JsonConfig) { cfg.BackupMonthlyRetentionMonths = configInt(-1) }},
		{name: "zero cleanup interval", mutate: func(cfg *config.JsonConfig) { cfg.BackupCleanupIntervalHours = configInt(0) }},
		{name: "overflowing cleanup interval", mutate: func(cfg *config.JsonConfig) { cfg.BackupCleanupIntervalHours = configInt(2562048) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validBackupConfig()
			tt.mutate(cfg)
			if err := validateBackupSettings(cfg); err == nil {
				t.Fatal("validateBackupSettings() returned nil, want validation error")
			}
		})
	}
}

func TestValidateWorldGenerationSettingsAcceptsPositionalPrefixes(t *testing.T) {
	tests := []config.JsonConfig{
		{WorldID: "Mars2"},
		{WorldID: "Mars2", Difficulty: "Normal"},
		{WorldID: "Mars2", Difficulty: "Normal", StartCondition: "DefaultStart"},
		{WorldID: "Mars2", Difficulty: "Normal", StartCondition: "DefaultStart", StartLocation: "MarsSpawnRoundRobin"},
	}
	for _, cfg := range tests {
		if err := validateWorldGenerationSettings(&cfg); err != nil {
			t.Errorf("validateWorldGenerationSettings(%+v) error = %v", cfg, err)
		}
	}
}

func TestValidateWorldGenerationSettingsRejectsGaps(t *testing.T) {
	tests := []config.JsonConfig{
		{WorldID: "Mars2", StartCondition: "DefaultStart"},
		{WorldID: "Mars2", Difficulty: "Normal", StartLocation: "MarsSpawnRoundRobin"},
		{WorldID: "Mars2", StartLocation: "MarsSpawnRoundRobin"},
	}
	for _, cfg := range tests {
		if err := validateWorldGenerationSettings(&cfg); err == nil {
			t.Errorf("validateWorldGenerationSettings(%+v) error = nil, want positional gap error", cfg)
		}
	}
}
