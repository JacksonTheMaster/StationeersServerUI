package configchanger

import (
	"fmt"
	"strconv"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/loader"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

type ValidationError struct {
	Err error
}

func (err *ValidationError) Error() string {
	return err.Err.Error()
}

func SaveConfig(cfg *config.JsonConfig, reloadBackend ...bool) error {
	previous, err := config.ReadConfigFile()
	if err != nil {
		return err
	}
	if err := validateDiscordAdminRole(cfg.DiscordAdminRoleID); err != nil {
		return err
	}
	err = config.SaveConfigToFile(cfg)
	if err != nil {
		logger.Core.Error("Failed to save config: " + err.Error())
		return err
	}
	// Apply runtime changes by default, unless this is part of startup.
	if len(reloadBackend) == 0 || reloadBackend[0] {
		loader.ReloadChangedConfig(previous, cfg)
	}
	return nil
}

// UpdateConfig is the normal runtime write path. Reading, changing and writing
// happen under one config lock so concurrent requests cannot overwrite each
// other with an older copy.
func UpdateConfig(update func(*config.JsonConfig) error) error {
	previous, next, err := config.UpdateConfig(func(cfg *config.JsonConfig) error {
		if err := update(cfg); err != nil {
			return err
		}
		if err := validateConfig(cfg); err != nil {
			return &ValidationError{Err: err}
		}
		return nil
	})
	if err != nil {
		return err
	}
	loader.ReloadChangedConfig(previous, next)
	return nil
}

func validateConfig(cfg *config.JsonConfig) error {
	if err := validateDiscordAdminRole(cfg.DiscordAdminRoleID); err != nil {
		return err
	}
	if err := validateBackupSettings(cfg); err != nil {
		return err
	}
	if err := validateDiscordVoteSettings(cfg); err != nil {
		return err
	}
	return nil
}

func validateDiscordAdminRole(role string) error {
	if role == "" {
		return nil
	}
	for _, digit := range role {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("Discord Admin Role ID must be a numeric Discord role ID, or empty to disable admin actions")
		}
	}
	id, err := strconv.ParseUint(role, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid Discord Admin Role ID")
	}
	return nil
}
