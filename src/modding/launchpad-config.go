package modding

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// ToggleSLPAutoUpdates enables or disables the SLP built-in auto-updater
// by setting CheckForUpdate and AutoUpdateOnStart to true/false in
// BepInEx/config/stationeers.launchpad.cfg
//
// If the file doesn't exist → does nothing (returns nil)
// If the file exists but keys are missing → adds them under [Startup]
// Returns true if the file was modified, false if no change/no file, error on failure
func ToggleSLPAutoUpdates(enable bool) (modified bool, err error) {
	const configPath = "BepInEx/config/stationeers.launchpad.cfg"

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Install.Debug("SLP config not found, skipping auto-update flag setup")
		return false, nil
	}

	// Read existing content
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false, fmt.Errorf("failed to read SLP config: %w", err)
	}

	original := string(data)
	content := original

	value := "true"
	if !enable {
		value = "false"
	}

	// Regex to match lines like:   CheckForUpdate = false   (any whitespace, case insensitive)
	reCheck := regexp.MustCompile(`(?mi)^\s*CheckForUpdate\s*=\s*(true|false)\s*(?:#.*)?$`)
	reAuto := regexp.MustCompile(`(?mi)^\s*AutoUpdateOnStart\s*=\s*(true|false)\s*(?:#.*)?$`)

	// Replace existing values
	newContent := reCheck.ReplaceAllString(content, fmt.Sprintf("CheckForUpdate = %s", value))
	newContent = reAuto.ReplaceAllString(newContent, fmt.Sprintf("AutoUpdateOnStart = %s", value))

	modified = (newContent != content)

	// If any key is missing → append them under [Startup] section
	if !reCheck.MatchString(content) || !reAuto.MatchString(content) {
		// Look for [Startup] section
		startupSectionRe := regexp.MustCompile(`(?m)^\[Startup\]\s*$`)

		if startupSectionRe.MatchString(newContent) {
			// Append to existing [Startup] section
			newContent = startupSectionRe.ReplaceAllStringFunc(newContent, func(match string) string {
				return match + fmt.Sprintf("\nCheckForUpdate = %s\nAutoUpdateOnStart = %s", value, value)
			})
		} else {
			// No [Startup] section → add it at the end
			newContent = strings.TrimRight(newContent, "\r\n") + fmt.Sprintf("\n\n[Startup]\nCheckForUpdate = %s\nAutoUpdateOnStart = %s\n", value, value)
		}
		modified = true
	}

	if !modified {
		logger.Install.Debug("SLP auto-update flags already set correctly, no change needed")
		return false, nil
	}

	// Write back
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return false, fmt.Errorf("failed to write updated SLP config: %w", err)
	}

	action := "enabled"
	if !enable {
		action = "disabled"
	}
	logger.Install.Info(fmt.Sprintf("SLP auto-updater %s (CheckForUpdate & AutoUpdateOnStart = %s)", action, value))

	return true, nil
}
