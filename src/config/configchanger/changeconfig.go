package configchanger

import (
	"fmt"
	"strings"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func validateBackupSettings(cfg *config.JsonConfig) error {
	settings := []struct {
		name  string
		value *int
	}{
		{name: "backupKeepNewestCount", value: cfg.BackupKeepNewestCount},
		{name: "backupDailyRetentionDays", value: cfg.BackupDailyRetentionDays},
		{name: "backupWeeklyRetentionWeeks", value: cfg.BackupWeeklyRetentionWeeks},
		{name: "backupMonthlyRetentionMonths", value: cfg.BackupMonthlyRetentionMonths},
	}

	for _, setting := range settings {
		if setting.value != nil && *setting.value < 0 {
			return fmt.Errorf("%s cannot be negative", setting.name)
		}
	}

	if cfg.BackupCleanupIntervalHours != nil {
		maxCleanupIntervalHours := int64((1<<63 - 1) / int64(time.Hour))
		if *cfg.BackupCleanupIntervalHours <= 0 || int64(*cfg.BackupCleanupIntervalHours) > maxCleanupIntervalHours {
			return fmt.Errorf("backupCleanupIntervalHours must produce a positive valid duration")
		}
	}

	return nil
}

func validateDiscordVoteSettings(cfg *config.JsonConfig) error {
	percentageSettings := []struct {
		name  string
		value *int
	}{
		{name: "discordRestartVoteThreshold", value: cfg.DiscordRestartVoteThreshold},
		{name: "discordRestoreVoteThreshold", value: cfg.DiscordRestoreVoteThreshold},
	}
	for _, setting := range percentageSettings {
		if setting.value != nil && (*setting.value < 1 || *setting.value > 100) {
			return fmt.Errorf("%s must be between 1 and 100", setting.name)
		}
	}

	positiveSettings := []struct {
		name  string
		value *int
	}{
		{name: "discordVoteDurationMinutes", value: cfg.DiscordVoteDurationMinutes},
		{name: "discordRestartVoteMinimum", value: cfg.DiscordRestartVoteMinimum},
		{name: "discordRestoreVoteMinimum", value: cfg.DiscordRestoreVoteMinimum},
	}
	for _, setting := range positiveSettings {
		if setting.value != nil && *setting.value < 1 {
			return fmt.Errorf("%s must be greater than zero", setting.name)
		}
	}

	for _, setting := range []struct {
		name  string
		value *int
	}{
		{name: "discordRestartVoteCooldownMinutes", value: cfg.DiscordRestartVoteCooldownMinutes},
		{name: "discordRestoreVoteCooldownMinutes", value: cfg.DiscordRestoreVoteCooldownMinutes},
	} {
		if setting.value != nil && *setting.value < 0 {
			return fmt.Errorf("%s cannot be negative", setting.name)
		}
	}
	return nil
}

func validateWorldGenerationSettings(cfg *config.JsonConfig) error {
	difficulty := strings.TrimSpace(cfg.Difficulty)
	startCondition := strings.TrimSpace(cfg.StartCondition)
	startLocation := strings.TrimSpace(cfg.StartLocation)

	if startCondition != "" && difficulty == "" {
		return fmt.Errorf("Difficulty is required when StartCondition is set because Stationeers parses world-generation values positionally")
	}
	if startLocation != "" && startCondition == "" {
		return fmt.Errorf("StartCondition is required when StartLocation is set because Stationeers parses world-generation values positionally")
	}
	return nil
}
