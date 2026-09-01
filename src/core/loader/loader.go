// loader.go
package loader

import (
	"embed"
	"os"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/advertiser"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/discordbot"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/localization"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/modding"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"
)

var reloadDiscordBotFunc = func() {
	ReloadDiscordBot()
}

// only call this once at startup
func InitBackend() {
	ReloadConfig()
	ReloadSSCM()
	ReloadBackupManager()
	ReloadLocalizer()
	ReloadAppInfoPoller()
	ReloadDiscordBot()
	EnsureSLPAutoUpdates()
	InitDetector()
	StartIsGameServerRunningCheck()
	StartUpdateCheckLoop()
	LoadAdvertiser()
}

// use this to reload backend at runtime
func ReloadBackend() {

	logger.Core.Info("Reloading backend...")
	ReloadConfig()
	ReloadSSCM()
	ReloadBackupManager()
	ReloadLocalizer()
	ReloadAppInfoPoller()
	reloadDiscordBotFunc()
	PrintConfigDetails()
	logger.Core.Info("Backend reload done!")
}

// should ideally not be called standalone, if feasable, call ReloadBackend instead
func ReloadConfig() {
	if _, err := config.LoadConfig(); err != nil {
		logger.Core.Error("Failed to load config: " + err.Error())
		return
	}
	logger.Core.Info("Config loaded successfully")

}

func ReloadSSCM() {
	if config.GetIsSSCMEnabled() {
		setup.InstallSSCM()
	}
}

func ReloadBackupManager() {
	if err := backupmgr.ReloadBackupManagerFromConfig(); err != nil {
		logger.Backup.Error("Failed to reload backup manager: " + err.Error())
		return
	}
}

func ReloadDiscordBot() {
	if !config.GetIsDiscordEnabled() {
		discordbot.DisableRuntimeState()
		if config.DiscordSession != nil {
			logger.Discord.Info("Discord integration disabled, closing existing session")
			config.DiscordSession.Close()
			config.DiscordSession = nil
		}
		return
	}

	if config.GetDiscordToken() == "" {
		logger.Discord.Warn("Discord is enabled but no token is configured")
		return
	}

	go discordbot.InitializeDiscordBot()
	logger.Discord.Info("Discord bot reload requested")
}

// The detector should NOT be reloaded, as it is a singleton. Instead, dynamic changes come in via the custom detections manager.
func InitDetector() {
	detector := detectionmgr.Start()
	detectionmgr.RegisterDefaultHandlers(detector)
	detectionmgr.InitCustomDetectionsManager(detector)
	detectionmgr.SetServerStateHandler(func(eventType detectionmgr.EventType) {
		switch eventType {
		case detectionmgr.EventGameManagerReady:
			gamemgr.SetServerState(gamemgr.ServerStateLoadingMap)
		case detectionmgr.EventServerHosted, detectionmgr.EventSessionStarting:
			gamemgr.SetServerState(gamemgr.ServerStateHostingSession)
		case detectionmgr.EventSessionRegistered, detectionmgr.EventWorldSaved:
			// Completing a world save proves the server is healthy and overrides
			// transitional or uncertain startup state.
			gamemgr.SetServerState(gamemgr.ServerStateRunning)
		}
	})
	go detectionmgr.StreamLogs(detector)
	logger.Detection.Info("Detector loaded successfully")
}

func RestartBackend() {
	update.RestartMySelf()
}

func ReloadLocalizer() {
	localization.ReloadLocalizer()
}

func StartIsGameServerRunningCheck() {
	gamemgr.StartIsGameServerRunningCheck()
}

func ReloadAppInfoPoller() {
	steamcmd.AppInfoPoller()
}

func LoadAdvertiser() {
	if config.GetAdvertiserOverride() != "" {
		logger.Advertiser.Info("Starting server advertiser...")
		advertiser.StartAdvertiser()
	}
}

func StartUpdateCheckLoop() {
	if config.GetIsUpdateEnabled() {
		go update.StartUpdateCheckLoop()
	}
}

// InitBundler initialized the onboard bundled assets for the web UI
func InitVirtFS(v1uiFS embed.FS) {
	config.SetV1UIFS(v1uiFS)
}

func InstallSLP() {
	version, err := modding.InstallSLP()
	if err != nil {
		logger.Install.Error("SLP installation failed: " + err.Error())
		return
	}
	logger.Install.Infof("SLP %s installed successfully", version)
}

func EnsureSLPAutoUpdates() {
	modified, err := modding.ToggleSLPAutoUpdates(config.GetIsStationeersLaunchPadAutoUpdatesEnabled())
	if err != nil {
		logger.Install.Error("Failed to toggle SLP auto-updates: " + err.Error())
		return
	}
	if modified {
		logger.Install.Infof("StationeersLaunchPad auto-updates toggled to %t", config.GetIsStationeersLaunchPadAutoUpdatesEnabled())
	}
}

func SanityCheck() {
	err := runSanityCheck()
	if err != nil {
		logger.Main.Error("Sanity check failed, exiting in 10 secconds: " + err.Error())
		logger.Main.Info("If you want to continue anyway, run SSUI with the --NoSanityCheck flag, but be aware there may be Dragons ahead.")
		logger.Main.Info("This is not recommended nor supported and may cause unexpected behavior, including potential data loss!")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
}
