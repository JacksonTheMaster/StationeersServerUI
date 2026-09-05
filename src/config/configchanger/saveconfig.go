package configchanger

import (
	"fmt"
	"strconv"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/loader"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func SaveConfig(cfg *config.JsonConfig, reloadBackend ...bool) error {
	if err := validateDiscordAdminRole(cfg.DiscordAdminRoleID); err != nil {
		return err
	}
	err := config.SaveConfigToFile(cfg)
	if err != nil {
		logger.Core.Error("Failed to save config: " + err.Error())
		return err
	}
	// Call ReloadBackend by default, unless reloadBackend is explicitly false
	if len(reloadBackend) == 0 || reloadBackend[0] {
		loader.ReloadBackend()
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
