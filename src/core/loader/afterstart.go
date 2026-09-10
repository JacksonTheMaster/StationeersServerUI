package loader

import (
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/connectivity"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/setup"
)

func AfterStartComplete() {
	config.SetSaveConfig() // Save config after startup through setters
	err := setup.CleanUpOldSSUIFolderFiles()
	if err != nil {
		logger.Core.Error("AfterStartComplete: Failed to clean up old pre-v5.5 UI mod folder files: " + err.Error())
	}
	err = setup.CleanUpOldExecutables()
	if err != nil {
		logger.Core.Error("AfterStartComplete: Failed to clean up old executables: " + err.Error())
	}
	connectivity.RunStartupCheck()
	if config.GetAutoStartServerOnStartup() {
		logger.Core.Info("AutoStartServerOnStartup is enabled, starting server...")
		gamemgr.InternalStartServer()
	}
	// deactivated for now, as we are working on a new way to handle this
	//setup.SetupAutostartScripts()

	time.Sleep(500 * time.Millisecond)
	printStartupMessage()

	if config.GetIsFirstTimeSetup() {
		printFirstTimeSetupMessage()
	}
}
