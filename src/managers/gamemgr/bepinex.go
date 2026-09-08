package gamemgr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// BepInEx version: 5.4.23.2 or v5-lts
// SSCM version: 1.0.0
// SetupBepInExEnvironment configures the environment for running the game with BepInEx/Doorstop and SSCM. (or other Mods, technically)
// Returns a map of environment variables and an error if setup fails.
func SetupBepInExEnvironment() ([]string, error) {

	executablePath := config.GetExePath()

	if !config.GetIsSSCMEnabled() {
		logger.Core.Debug("SSCM is disabled, skipping environment setup")
		return nil, nil
	}

	// Validate executable
	if executablePath == "" {
		return nil, fmt.Errorf("invalid executable path: %s", executablePath)
	}

	// Get base directory (current directory)
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}
	logger.Core.Debug(fmt.Sprintf("Using base directory: %s", baseDir))

	// Set up environment variables for Doorstop
	targetAssembly := filepath.Join(baseDir, "BepInEx/core/BepInEx.Preloader.dll")
	logger.Core.Debug(fmt.Sprintf("Target assembly: %s", targetAssembly))

	// Get current environment
	envVars := os.Environ()

	// Add Doorstop environment variables
	envVars = append(envVars, "DOORSTOP_ENABLED=1")
	envVars = append(envVars, fmt.Sprintf("DOORSTOP_TARGET_ASSEMBLY=%s", targetAssembly))

	// Set up LD_LIBRARY_PATH and LD_PRELOAD
	doorstopName := "libdoorstop.so"

	// Update LD_LIBRARY_PATH
	ldLibraryPath := fmt.Sprintf("%s:%s", baseDir, os.Getenv("LD_LIBRARY_PATH"))
	envVars = append(envVars, fmt.Sprintf("LD_LIBRARY_PATH=%s", ldLibraryPath))

	// Update LD_PRELOAD
	currentLdPreload := os.Getenv("LD_PRELOAD")
	if currentLdPreload == "" {
		envVars = append(envVars, fmt.Sprintf("LD_PRELOAD=%s", doorstopName))
	} else {
		envVars = append(envVars, fmt.Sprintf("LD_PRELOAD=%s:%s", doorstopName, currentLdPreload))
	}

	for _, env := range envVars {
		logger.Core.Debug(fmt.Sprintf("Environment variable: %s", env))
	}

	return envVars, nil
}
