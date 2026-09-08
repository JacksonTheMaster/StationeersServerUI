// loader.go
package loader

import (
	"embed"
	"os"
	"reflect"
	"sync"
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

var discordReload = struct {
	sync.Mutex
	running bool
	pending bool
}{}

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

// ReloadChangedConfig applies a saved config and only restarts the parts that
// keep their own runtime state. Most settings are read through config getters
// and need no subsystem reload at all.
func ReloadChangedConfig(previous, next *config.JsonConfig) {
	ReloadConfig()
	plan := planConfigReload(previous, next)

	if plan.backupManager {
		ReloadBackupManager()
	}
	if plan.localizer {
		ReloadLocalizer()
	}
	if plan.appInfoPoller {
		ReloadAppInfoPoller()
	}
	if plan.discord {
		reloadDiscordBotFunc()
	}
	if plan.sscm {
		ReloadSSCM()
	}
	if plan.slpAutoUpdates {
		EnsureSLPAutoUpdates()
	}
}

type configReloadPlan struct {
	backupManager  bool
	localizer      bool
	appInfoPoller  bool
	discord        bool
	sscm           bool
	slpAutoUpdates bool
}

func planConfigReload(previous, next *config.JsonConfig) configReloadPlan {
	return configReloadPlan{
		backupManager:  backupConfigChanged(previous, next),
		localizer:      previous.LanguageSetting != next.LanguageSetting,
		appInfoPoller:  boolChanged(previous.AllowAutoGameServerUpdates, next.AllowAutoGameServerUpdates),
		discord:        discordConfigChanged(previous, next),
		sscm:           boolChanged(previous.IsSSCMEnabled, next.IsSSCMEnabled),
		slpAutoUpdates: boolChanged(previous.IsStationeersLaunchPadAutoUpdatesEnabled, next.IsStationeersLaunchPadAutoUpdatesEnabled),
	}
}

func backupConfigChanged(previous, next *config.JsonConfig) bool {
	return previous.SaveName != next.SaveName ||
		boolChanged(previous.BackupRetentionEnabled, next.BackupRetentionEnabled) ||
		intChanged(previous.BackupKeepNewestCount, next.BackupKeepNewestCount) ||
		intChanged(previous.BackupDailyRetentionDays, next.BackupDailyRetentionDays) ||
		intChanged(previous.BackupWeeklyRetentionWeeks, next.BackupWeeklyRetentionWeeks) ||
		intChanged(previous.BackupMonthlyRetentionMonths, next.BackupMonthlyRetentionMonths) ||
		intChanged(previous.BackupCleanupIntervalHours, next.BackupCleanupIntervalHours)
}

func discordConfigChanged(previous, next *config.JsonConfig) bool {
	return previous.DiscordToken != next.DiscordToken ||
		previous.DiscordAdminRoleID != next.DiscordAdminRoleID ||
		previous.EventLogChannelID != next.EventLogChannelID ||
		previous.StatusPanelChannelID != next.StatusPanelChannelID ||
		previous.LogChannelID != next.LogChannelID ||
		previous.DiscordCharBufferSize != next.DiscordCharBufferSize ||
		previous.BlackListFilePath != next.BlackListFilePath ||
		boolChanged(previous.IsDiscordEnabled, next.IsDiscordEnabled) ||
		boolChanged(previous.RotateServerPassword, next.RotateServerPassword) ||
		boolChanged(previous.DiscordRestartVoteEnabled, next.DiscordRestartVoteEnabled) ||
		boolChanged(previous.DiscordRestoreVoteEnabled, next.DiscordRestoreVoteEnabled) ||
		intChanged(previous.DiscordVoteDurationMinutes, next.DiscordVoteDurationMinutes) ||
		intChanged(previous.DiscordRestartVoteThreshold, next.DiscordRestartVoteThreshold) ||
		intChanged(previous.DiscordRestartVoteMinimum, next.DiscordRestartVoteMinimum) ||
		intChanged(previous.DiscordRestartVoteCooldownMinutes, next.DiscordRestartVoteCooldownMinutes) ||
		intChanged(previous.DiscordRestoreVoteThreshold, next.DiscordRestoreVoteThreshold) ||
		intChanged(previous.DiscordRestoreVoteMinimum, next.DiscordRestoreVoteMinimum) ||
		intChanged(previous.DiscordRestoreVoteCooldownMinutes, next.DiscordRestoreVoteCooldownMinutes)
}

func boolChanged(previous, next *bool) bool {
	return !reflect.DeepEqual(previous, next)
}

func intChanged(previous, next *int) bool {
	return !reflect.DeepEqual(previous, next)
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
	discordReload.Lock()
	if discordReload.running {
		discordReload.pending = true
		discordReload.Unlock()
		logger.Discord.Debug("Discord bot reload already running; keeping one pending reload")
		return
	}
	discordReload.running = true
	discordReload.Unlock()

	go runDiscordReloads()
	logger.Discord.Info("Discord bot reload requested")
}

func runDiscordReloads() {
	for {
		reloadDiscordBot()

		discordReload.Lock()
		if !discordReload.pending {
			discordReload.running = false
			discordReload.Unlock()
			return
		}
		discordReload.pending = false
		discordReload.Unlock()
	}
}

func reloadDiscordBot() {
	if !config.GetIsDiscordEnabled() {
		discordbot.DisableRuntimeState()
		return
	}

	if config.GetDiscordToken() == "" {
		logger.Discord.Warn("Discord is enabled but no token is configured")
		discordbot.DisableRuntimeState()
		return
	}

	discordbot.InitializeDiscordBot()
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
	// The disabled loop stays local and cheap, but needs to exist so enabling
	// updates at runtime takes effect without restarting SSUI.
	go update.StartUpdateCheckLoop()
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
