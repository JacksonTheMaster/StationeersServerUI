package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// All configuration variables can be found in vars.go
	Version = "5.15.0"
	Branch  = "release"
)

/*
If you read this, you are likely a developer. I sincerly apologize for the way the config works.
While I would love to refactor the config to not write to file then read the file every time a config value is changed,
I have not found the time to do so. So, for now, we save to file, then read the file and rely on whatever the file says. Although this is not ideal, it works for now. Deal with it.
JacksonTheMaster
*/

type JsonConfig struct {
	// reordered in 5.6.4 to simplify the order of the config file.

	// Gameserver Settings
	GameBranch       string `json:"gameBranch"`
	GamePort         string `json:"GamePort"`
	ServerName       string `json:"ServerName"`
	SaveInfo         string `json:"SaveInfo,omitempty"` // deprecated, kept for backwards compatibility
	SaveName         string `json:"SaveName"`           // replaces SaveInfo
	WorldID          string `json:"WorldID"`            // replaces SaveInfo
	ServerMaxPlayers string `json:"ServerMaxPlayers"`
	ServerPassword   string `json:"ServerPassword"`
	ServerAuthSecret string `json:"ServerAuthSecret"`
	AdminPassword    string `json:"AdminPassword"`
	UpdatePort       string `json:"UpdatePort"`
	UPNPEnabled      *bool  `json:"UPNPEnabled"`
	AutoSave         *bool  `json:"AutoSave"`
	SaveInterval     string `json:"SaveInterval"`
	AutoPauseServer  *bool  `json:"AutoPauseServer"`
	LocalIpAddress   string `json:"LocalIpAddress"`
	StartLocalHost   *bool  `json:"StartLocalHost"`
	ServerVisible    *bool  `json:"ServerVisible"`
	UseSteamP2P      *bool  `json:"UseSteamP2P"`
	AdditionalParams string `json:"AdditionalParams"`
	Difficulty       string `json:"Difficulty"`
	StartCondition   string `json:"StartCondition"`
	StartLocation    string `json:"StartLocation"`

	// Logging and debug settings
	Debug                   *bool    `json:"Debug"`
	CreateSSUILogFile       *bool    `json:"CreateSSUILogFile"`
	CreateGameServerLogFile *bool    `json:"CreateGameServerLogFile"`
	LogLevel                int      `json:"LogLevel"`
	SubsystemFilters        []string `json:"subsystemFilters"`
	AdvertiserOverride      string   `json:"AdvertiserOverride"`

	// Authentication Settings
	Users             map[string]string `json:"users"`       // Map of username to hashed password
	AuthEnabled       *bool             `json:"authEnabled"` // Toggle for enabling/disabling auth
	JwtKey            string            `json:"JwtKey"`
	AuthTokenLifetime int               `json:"AuthTokenLifetime"`

	// SSUI Settings
	IsNewTerrainAndSaveSystem *bool  `json:"IsNewTerrainAndSaveSystem"` // Use new terrain and save system
	ExePath                   string `json:"ExePath"`
	LogClutterToConsole       *bool  `json:"LogClutterToConsole"`
	IsSSCMEnabled             *bool  `json:"IsSSCMEnabled"`
	AutoRestartServerTimer    string `json:"AutoRestartServerTimer"`
	AutoRestartCountdown      string `json:"AutoRestartCountdown"`
	IsConsoleEnabled          *bool  `json:"IsConsoleEnabled"`
	IsCLIDashboardEnabled     *bool  `json:"IsCLIDashboardEnabled"`
	LanguageSetting           string `json:"LanguageSetting"`
	AutoStartServerOnStartup  *bool  `json:"AutoStartServerOnStartup"`
	SSUIIdentifier            string `json:"SSUIIdentifier"`
	SSUIWebPort               string `json:"SSUIWebPort"`
	ShowExpertSettings        *bool  `json:"ShowExpertSettings"` // Show Expert Settings tab in web UI

	// Update Settings
	IsUpdateEnabled            *bool `json:"IsUpdateEnabled"`
	AllowPrereleaseUpdates     *bool `json:"AllowPrereleaseUpdates"`
	AllowMajorUpdates          *bool `json:"AllowMajorUpdates"`
	AllowAutoGameServerUpdates *bool `json:"AllowAutoGameServerUpdates"`

	// SLP Modding Settings
	IsStationeersLaunchPadEnabled            *bool `json:"IsStationeersLaunchPadEnabled"`
	IsStationeersLaunchPadAutoUpdatesEnabled *bool `json:"IsStationeersLaunchPadAutoUpdatesEnabled"`

	// Discord Settings
	DiscordToken                      string `json:"discordToken"`
	ControlChannelID                  string `json:"controlChannelID"`
	EventLogChannelID                 string `json:"eventLogChannelID"`
	StatusChannelID                   string `json:"statusChannelID,omitempty"`         // deprecated, migrated to EventLogChannelID
	ConnectionListChannelID           string `json:"connectionListChannelID,omitempty"` // deprecated, migrated to StatusPanelChannelID
	StatusPanelChannelID              string `json:"statusPanelChannelID"`              // replaces ConnectionListChannelID and ServerInfoPanelChannelID
	LogChannelID                      string `json:"logChannelID"`
	SaveChannelID                     string `json:"saveChannelID,omitempty"` // deprecated, merged into EventLogChannelID
	ControlPanelChannelID             string `json:"controlPanelChannelID"`
	DiscordCharBufferSize             int    `json:"DiscordCharBufferSize"`
	BlackListFilePath                 string `json:"blackListFilePath"`
	IsDiscordEnabled                  *bool  `json:"isDiscordEnabled"`
	RotateServerPassword              *bool  `json:"rotateServerPassword"`
	DiscordRestartVoteEnabled         *bool  `json:"discordRestartVoteEnabled"`
	DiscordRestoreVoteEnabled         *bool  `json:"discordRestoreVoteEnabled"`
	DiscordVoteDurationMinutes        *int   `json:"discordVoteDurationMinutes"`
	DiscordRestartVoteThreshold       *int   `json:"discordRestartVoteThreshold"`
	DiscordRestartVoteMinimum         *int   `json:"discordRestartVoteMinimum"`
	DiscordRestartVoteCooldownMinutes *int   `json:"discordRestartVoteCooldownMinutes"`
	DiscordRestoreVoteThreshold       *int   `json:"discordRestoreVoteThreshold"`
	DiscordRestoreVoteMinimum         *int   `json:"discordRestoreVoteMinimum"`
	DiscordRestoreVoteCooldownMinutes *int   `json:"discordRestoreVoteCooldownMinutes"`

	// Backup retention settings. The former backupKeep* / isCleanupEnabled keys
	// are intentionally not migrated so upgrades return to the safe disabled state.
	BackupRetentionEnabled       *bool `json:"backupRetentionEnabled"`
	BackupKeepNewestCount        *int  `json:"backupKeepNewestCount"`
	BackupDailyRetentionDays     *int  `json:"backupDailyRetentionDays"`
	BackupWeeklyRetentionWeeks   *int  `json:"backupWeeklyRetentionWeeks"`
	BackupMonthlyRetentionMonths *int  `json:"backupMonthlyRetentionMonths"`
	BackupCleanupIntervalHours   *int  `json:"backupCleanupIntervalHours"`
}

// LoadConfig loads and initializes the configuration
func LoadConfig() (*JsonConfig, error) {
	ConfigMu.Lock()
	defer ConfigMu.Unlock()

	var jsonConfig JsonConfig
	file, err := os.Open(ConfigPath)
	if err == nil {
		// File exists, proceed to decode it
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&jsonConfig); err != nil {
			return nil, fmt.Errorf("failed to decode config: %v", err)
		}
	} else if os.IsNotExist(err) {
		// File is missing, log it and proceed with defaults (probably first time setup)
		fmt.Println("config file was not found, proceeding with defaults.")
	} else {
		// Other errors (e.g., permissions), fail immediately
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	// Apply configuration
	applyConfig(&jsonConfig)

	return &jsonConfig, nil
}

// applyConfig applies the configuration with JSON -> env -> fallback hierarchy
func applyConfig(cfg *JsonConfig) {
	// Apply values with hierarchy
	DiscordToken = getString(cfg.DiscordToken, "DISCORD_TOKEN", "")
	ControlChannelID = getString(cfg.ControlChannelID, "CONTROL_CHANNEL_ID", "")
	EventLogChannelID = getString(cfg.EventLogChannelID, "GAME_EVENT_LOG_CHANNEL_ID", "")
	StatusPanelChannelID = getString(cfg.StatusPanelChannelID, "STATUS_PANEL_CHANNEL_ID", "")
	LogChannelID = getString(cfg.LogChannelID, "LOG_CHANNEL_ID", "")
	ControlPanelChannelID = getString(cfg.ControlPanelChannelID, "CONTROL_PANEL_CHANNEL_ID", "")
	DiscordCharBufferSize = getInt(cfg.DiscordCharBufferSize, "DISCORD_CHAR_BUFFER_SIZE", 1000)
	BlackListFilePath = getString(cfg.BlackListFilePath, "BLACKLIST_FILE_PATH", "./Blacklist.txt")

	isDiscordEnabledVal := getBool(cfg.IsDiscordEnabled, "IS_DISCORD_ENABLED", false)
	IsDiscordEnabled = isDiscordEnabledVal
	cfg.IsDiscordEnabled = &isDiscordEnabledVal

	rotateServerPasswordVal := getBool(cfg.RotateServerPassword, "ROTATE_SERVER_PASSWORD", false)
	RotateServerPassword = rotateServerPasswordVal
	cfg.RotateServerPassword = &rotateServerPasswordVal

	discordRestartVoteEnabledVal := getBool(cfg.DiscordRestartVoteEnabled, "DISCORD_RESTART_VOTE_ENABLED", false)
	DiscordRestartVoteEnabled = discordRestartVoteEnabledVal
	cfg.DiscordRestartVoteEnabled = &discordRestartVoteEnabledVal
	discordRestoreVoteEnabledVal := getBool(cfg.DiscordRestoreVoteEnabled, "DISCORD_RESTORE_VOTE_ENABLED", false)
	DiscordRestoreVoteEnabled = discordRestoreVoteEnabledVal
	cfg.DiscordRestoreVoteEnabled = &discordRestoreVoteEnabledVal
	DiscordVoteDurationMinutes = getOptionalInt(cfg.DiscordVoteDurationMinutes, "DISCORD_VOTE_DURATION_MINUTES", 5)
	cfg.DiscordVoteDurationMinutes = intPointer(DiscordVoteDurationMinutes)
	DiscordRestartVoteThreshold = getOptionalInt(cfg.DiscordRestartVoteThreshold, "DISCORD_RESTART_VOTE_THRESHOLD", 60)
	cfg.DiscordRestartVoteThreshold = intPointer(DiscordRestartVoteThreshold)
	DiscordRestartVoteMinimum = getOptionalInt(cfg.DiscordRestartVoteMinimum, "DISCORD_RESTART_VOTE_MINIMUM", 1)
	cfg.DiscordRestartVoteMinimum = intPointer(DiscordRestartVoteMinimum)
	DiscordRestartVoteCooldownMinutes = getOptionalInt(cfg.DiscordRestartVoteCooldownMinutes, "DISCORD_RESTART_VOTE_COOLDOWN_MINUTES", 30)
	cfg.DiscordRestartVoteCooldownMinutes = intPointer(DiscordRestartVoteCooldownMinutes)
	DiscordRestoreVoteThreshold = getOptionalInt(cfg.DiscordRestoreVoteThreshold, "DISCORD_RESTORE_VOTE_THRESHOLD", 100)
	cfg.DiscordRestoreVoteThreshold = intPointer(DiscordRestoreVoteThreshold)
	DiscordRestoreVoteMinimum = getOptionalInt(cfg.DiscordRestoreVoteMinimum, "DISCORD_RESTORE_VOTE_MINIMUM", 2)
	cfg.DiscordRestoreVoteMinimum = intPointer(DiscordRestoreVoteMinimum)
	DiscordRestoreVoteCooldownMinutes = getOptionalInt(cfg.DiscordRestoreVoteCooldownMinutes, "DISCORD_RESTORE_VOTE_COOLDOWN_MINUTES", 60)
	cfg.DiscordRestoreVoteCooldownMinutes = intPointer(DiscordRestoreVoteCooldownMinutes)

	backupRetentionEnabledVal := getBool(cfg.BackupRetentionEnabled, "BACKUP_RETENTION_ENABLED", false)
	BackupRetentionEnabled = backupRetentionEnabledVal
	cfg.BackupRetentionEnabled = &backupRetentionEnabledVal

	backupKeepNewestCountVal := getOptionalInt(cfg.BackupKeepNewestCount, "BACKUP_KEEP_NEWEST_COUNT", 2)
	BackupKeepNewestCount = backupKeepNewestCountVal
	cfg.BackupKeepNewestCount = &backupKeepNewestCountVal

	backupDailyRetentionDaysVal := getOptionalInt(cfg.BackupDailyRetentionDays, "BACKUP_DAILY_RETENTION_DAYS", 7)
	BackupDailyRetentionDays = backupDailyRetentionDaysVal
	cfg.BackupDailyRetentionDays = &backupDailyRetentionDaysVal

	backupWeeklyRetentionWeeksVal := getOptionalInt(cfg.BackupWeeklyRetentionWeeks, "BACKUP_WEEKLY_RETENTION_WEEKS", 4)
	BackupWeeklyRetentionWeeks = backupWeeklyRetentionWeeksVal
	cfg.BackupWeeklyRetentionWeeks = &backupWeeklyRetentionWeeksVal

	backupMonthlyRetentionMonthsVal := getOptionalInt(cfg.BackupMonthlyRetentionMonths, "BACKUP_MONTHLY_RETENTION_MONTHS", 3)
	BackupMonthlyRetentionMonths = backupMonthlyRetentionMonthsVal
	cfg.BackupMonthlyRetentionMonths = &backupMonthlyRetentionMonthsVal

	backupCleanupIntervalHoursVal := getOptionalInt(cfg.BackupCleanupIntervalHours, "BACKUP_CLEANUP_INTERVAL_HOURS", 24)
	BackupCleanupInterval = time.Duration(backupCleanupIntervalHoursVal) * time.Hour
	cfg.BackupCleanupIntervalHours = &backupCleanupIntervalHoursVal

	isNewTerrainAndSaveSystemVal := getBool(cfg.IsNewTerrainAndSaveSystem, "ENABLE_DOT_SAVES", true)
	IsNewTerrainAndSaveSystem = isNewTerrainAndSaveSystemVal
	cfg.IsNewTerrainAndSaveSystem = &isNewTerrainAndSaveSystemVal

	GameBranch = getString(strings.ToLower(cfg.GameBranch), "GAME_BRANCH", "public")
	Difficulty = getString(cfg.Difficulty, "DIFFICULTY", "")
	StartCondition = getString(cfg.StartCondition, "START_CONDITION", "")
	StartLocation = getString(cfg.StartLocation, "START_LOCATION", "")
	ServerName = getString(cfg.ServerName, "SERVER_NAME", "Stationeers Server UI")
	SaveInfo = getString(cfg.SaveInfo, "SAVE_INFO", "") // deprecated, kept for backwards compatibility - if set, this gets migrated to SaveName and WorldID and the field is not written back to config.json
	SaveName = getString(cfg.SaveName, "SAVE_NAME", "MyMapName")
	WorldID = getString(cfg.WorldID, "WORLD_ID", "Lunar")
	ServerMaxPlayers = getString(cfg.ServerMaxPlayers, "SERVER_MAX_PLAYERS", "6")
	ServerPassword = getString(cfg.ServerPassword, "SERVER_PASSWORD", "")
	ServerAuthSecret = getString(cfg.ServerAuthSecret, "SERVER_AUTH_SECRET", "")
	AdminPassword = getString(cfg.AdminPassword, "ADMIN_PASSWORD", "")
	GamePort = getString(cfg.GamePort, "GAME_PORT", "27016")
	UpdatePort = getString(cfg.UpdatePort, "UPDATE_PORT", "27015")
	LanguageSetting = getString(cfg.LanguageSetting, "LANGUAGE_SETTING", "en-US")
	SSUIIdentifier = getString(cfg.SSUIIdentifier, "SSUI_IDENTIFIER", "")
	SSUIWebPort = getString(cfg.SSUIWebPort, "SSUI_WEB_PORT", "8443")

	upnpEnabledVal := getBool(cfg.UPNPEnabled, "UPNP_ENABLED", false)
	UPNPEnabled = upnpEnabledVal
	cfg.UPNPEnabled = &upnpEnabledVal

	autoSaveVal := getBool(cfg.AutoSave, "AUTO_SAVE", true)
	AutoSave = autoSaveVal
	cfg.AutoSave = &autoSaveVal

	SaveInterval = getString(cfg.SaveInterval, "SAVE_INTERVAL", "300")

	autoPauseServerVal := getBool(cfg.AutoPauseServer, "AUTO_PAUSE_SERVER", true)
	AutoPauseServer = autoPauseServerVal
	cfg.AutoPauseServer = &autoPauseServerVal

	LocalIpAddress = getString(cfg.LocalIpAddress, "LOCAL_IP_ADDRESS", "0.0.0.0")

	startLocalHostVal := getBool(cfg.StartLocalHost, "START_LOCAL_HOST", true)
	StartLocalHost = startLocalHostVal
	cfg.StartLocalHost = &startLocalHostVal

	serverVisibleVal := getBool(cfg.ServerVisible, "SERVER_VISIBLE", true)
	ServerVisible = serverVisibleVal
	cfg.ServerVisible = &serverVisibleVal

	useSteamP2PVal := getBool(cfg.UseSteamP2P, "USE_STEAM_P2P", false)
	UseSteamP2P = useSteamP2PVal
	cfg.UseSteamP2P = &useSteamP2PVal

	ExePath = getString(cfg.ExePath, "EXE_PATH", getDefaultExePath())
	AdditionalParams = getString(cfg.AdditionalParams, "ADDITIONAL_PARAMS", "")
	Users = getUsers(cfg.Users, "SSUI_USERS", map[string]string{})

	authEnabledVal := getBool(cfg.AuthEnabled, "SSUI_AUTH_ENABLED", false)
	AuthEnabled = authEnabledVal
	cfg.AuthEnabled = &authEnabledVal

	JwtKey = getString(cfg.JwtKey, "SSUI_JWT_KEY", generateJwtKey())
	AuthTokenLifetime = getInt(cfg.AuthTokenLifetime, "SSUI_AUTH_TOKEN_LIFETIME", 1440)

	debugVal := getBool(cfg.Debug, "DEBUG", false)
	IsDebugMode = debugVal
	cfg.Debug = &debugVal

	createSSUILogFileVal := getBool(cfg.CreateSSUILogFile, "CREATE_SSUI_LOGFILE", false)
	CreateSSUILogFile = createSSUILogFileVal
	cfg.CreateSSUILogFile = &createSSUILogFileVal

	createGameServerLogFileVal := getBool(cfg.CreateGameServerLogFile, "CREATE_GAMESERVER_LOGFILE", false)
	CreateGameServerLogFile = createGameServerLogFileVal
	cfg.CreateGameServerLogFile = &createGameServerLogFileVal

	LogLevel = getInt(cfg.LogLevel, "LOG_LEVEL", 20)

	isUpdateEnabledVal := getBool(cfg.IsUpdateEnabled, "IS_UPDATE_ENABLED", true)
	IsUpdateEnabled = isUpdateEnabledVal
	cfg.IsUpdateEnabled = &isUpdateEnabledVal

	allowPrereleaseUpdatesVal := getBool(cfg.AllowPrereleaseUpdates, "ALLOW_PRERELEASE_UPDATES", false)
	AllowPrereleaseUpdates = allowPrereleaseUpdatesVal
	cfg.AllowPrereleaseUpdates = &allowPrereleaseUpdatesVal

	allowMajorUpdatesVal := getBool(cfg.AllowMajorUpdates, "ALLOW_MAJOR_UPDATES", false)
	AllowMajorUpdates = allowMajorUpdatesVal
	cfg.AllowMajorUpdates = &allowMajorUpdatesVal

	allowAutoGameServerUpdatesVal := getBool(cfg.AllowAutoGameServerUpdates, "ALLOW_AUTO_GAME_SERVER_UPDATES", false)
	AllowAutoGameServerUpdates = allowAutoGameServerUpdatesVal
	cfg.AllowAutoGameServerUpdates = &allowAutoGameServerUpdatesVal

	isStationeersLaunchPadEnabledVal := getBool(cfg.IsStationeersLaunchPadEnabled, "IS_SLP_MODDING_ENABLED", false)
	IsStationeersLaunchPadEnabled = isStationeersLaunchPadEnabledVal
	cfg.IsStationeersLaunchPadEnabled = &isStationeersLaunchPadEnabledVal

	isStationeersLaunchPadAutoUpdatesEnabledVal := getBool(cfg.IsStationeersLaunchPadAutoUpdatesEnabled, "IS_SLP_MODDING_AUTO_UPDATES_ENABLED", true)
	IsStationeersLaunchPadAutoUpdatesEnabled = isStationeersLaunchPadAutoUpdatesEnabledVal
	cfg.IsStationeersLaunchPadAutoUpdatesEnabled = &isStationeersLaunchPadAutoUpdatesEnabledVal

	SubsystemFilters = getStringSlice(cfg.SubsystemFilters, "SUBSYSTEM_FILTERS", []string{})
	AutoRestartServerTimer = getString(cfg.AutoRestartServerTimer, "AUTO_RESTART_SERVER_TIMER", "0")
	AutoRestartCountdown = getString(cfg.AutoRestartCountdown, "AUTO_RESTART_COUNTDOWN", "65")
	isSSCMEnabledVal := getBool(cfg.IsSSCMEnabled, "IS_SSCM_ENABLED", true)
	IsSSCMEnabled = isSSCMEnabledVal
	cfg.IsSSCMEnabled = &isSSCMEnabledVal

	isConsoleEnabledVal := getBool(cfg.IsConsoleEnabled, "IS_CONSOLE_ENABLED", true)
	IsConsoleEnabled = isConsoleEnabledVal
	cfg.IsConsoleEnabled = &isConsoleEnabledVal

	isCLIDashboardEnabledVal := getBool(cfg.IsCLIDashboardEnabled, "IS_CLI_DASHBOARD_ENABLED", false)
	IsCLIDashboardEnabled = isCLIDashboardEnabledVal
	cfg.IsCLIDashboardEnabled = &isCLIDashboardEnabledVal

	logClutterToConsoleVal := getBool(cfg.LogClutterToConsole, "LOG_CLUTTER_TO_CONSOLE", false)
	LogClutterToConsole = logClutterToConsoleVal
	cfg.LogClutterToConsole = &logClutterToConsoleVal

	autoStartServerOnStartupVal := getBool(cfg.AutoStartServerOnStartup, "AUTO_START_SERVER_ON_STARTUP", false)
	AutoStartServerOnStartup = autoStartServerOnStartupVal
	cfg.AutoStartServerOnStartup = &autoStartServerOnStartupVal

	showExpertSettingsVal := getBool(cfg.ShowExpertSettings, "SHOW_EXPERT_SETTINGS", false)
	ShowExpertSettings = showExpertSettingsVal
	cfg.ShowExpertSettings = &showExpertSettingsVal

	// START MIGRATIONS AND BACKWARDS COMPATIBILITY

	// Process SaveInfo to maintain backwards compatibility with pre-5.6.6 SaveInfo field (deprecated)
	if SaveInfo != "" {
		parts := strings.Split(SaveInfo, " ")
		if len(parts) > 0 {
			SaveName = parts[0]
			fmt.Println("SaveName: " + SaveName)
		}
		if len(parts) > 1 {
			WorldID = parts[1]
			fmt.Println("WorldID: " + WorldID)
		}
		cfg.SaveInfo = ""
	}

	// Migrate ConnectionListChannelID -> StatusPanelChannelID (pre-5.13.0)
	if cfg.ConnectionListChannelID != "" && StatusPanelChannelID == "" {
		StatusPanelChannelID = cfg.ConnectionListChannelID
		fmt.Println("Migrated ConnectionListChannelID to StatusPanelChannelID: " + StatusPanelChannelID)
		cfg.ConnectionListChannelID = ""
	}

	// Migrate StatusChannelID -> EventLogChannelID (pre-5.14.0)
	if cfg.StatusChannelID != "" && EventLogChannelID == "" {
		EventLogChannelID = cfg.StatusChannelID
		fmt.Println("Migrated StatusChannelID to EventLogChannelID: " + EventLogChannelID)
		cfg.StatusChannelID = ""
	}

	// Migrate SaveChannelID -> EventLogChannelID (pre-5.14.0, save events now go to game event log)
	if cfg.SaveChannelID != "" {
		fmt.Println("SaveChannelID is deprecated and has been removed. Save events now go to the Game Event Log Channel. Dropping SaveChannelID value: " + cfg.SaveChannelID)
	}

	// END MIGRATIONS AND BACKWARDS COMPATIBILITY

	// Enforce branch compatibility (#172): hard error on legacy terrain branches.
	// Only modern branches (public, beta, others new) supported in current SSUI.
	// Legacy preterrain etc require old SSUI ~5.12.x .
	legacyBranches := map[string]bool{
		"preterrain":      true,
		"prerocket":       true,
		"prephasechange":  true,
		"multiplayersafe": true,
	}
	normBranch := strings.ToLower(GameBranch)
	if legacyBranches[normBranch] {
		fmt.Println("❌ FATAL: Legacy terrain/save branch '" + GameBranch + "' is no longer supported.")
		fmt.Println("This version of SSUI has removed legacy terrain system checks and support.")
		fmt.Println("Recommendation: switch your GameBranch to 'public' (or 'beta'), or use an older SSUI release (e.g. 5.12.x series) that retains preterrain backup manager support.")
		fmt.Println("Exiting now.")
		os.Exit(1)
	}

	// New terrain and save system is now always enforced for supported branches.
	IsNewTerrainAndSaveSystem = true

	// Set backup paths for old or new style saves
	if IsNewTerrainAndSaveSystem {
		// use new new style autosave folder
		ConfiguredBackupDir = filepath.Join("./saves/", SaveName, "autosave")
	} else {
		// use old style Backups folder
		ConfiguredBackupDir = filepath.Join("./saves/", SaveName, "Backup")
	}
	// use Safebackups folder either way.
	ConfiguredSafeBackupDir = filepath.Join("./saves/", SaveName, "Safebackups")

	AdvertiserOverride = getString(cfg.AdvertiserOverride, "ADVERTISER_OVERRIDE", "")

	safeSaveConfig()
}

// use safeSaveConfig EXCLUSIVELY though setter functions
// M U S T be called while holding a lock on ConfigMu!
func safeSaveConfig() error {
	cfg := JsonConfig{
		DiscordToken:                             DiscordToken,
		ControlChannelID:                         ControlChannelID,
		EventLogChannelID:                        EventLogChannelID,
		StatusPanelChannelID:                     StatusPanelChannelID,
		LogChannelID:                             LogChannelID,
		ControlPanelChannelID:                    ControlPanelChannelID,
		DiscordCharBufferSize:                    DiscordCharBufferSize,
		BlackListFilePath:                        BlackListFilePath,
		IsDiscordEnabled:                         &IsDiscordEnabled,
		RotateServerPassword:                     &RotateServerPassword,
		DiscordRestartVoteEnabled:                &DiscordRestartVoteEnabled,
		DiscordRestoreVoteEnabled:                &DiscordRestoreVoteEnabled,
		DiscordVoteDurationMinutes:               intPointer(DiscordVoteDurationMinutes),
		DiscordRestartVoteThreshold:              intPointer(DiscordRestartVoteThreshold),
		DiscordRestartVoteMinimum:                intPointer(DiscordRestartVoteMinimum),
		DiscordRestartVoteCooldownMinutes:        intPointer(DiscordRestartVoteCooldownMinutes),
		DiscordRestoreVoteThreshold:              intPointer(DiscordRestoreVoteThreshold),
		DiscordRestoreVoteMinimum:                intPointer(DiscordRestoreVoteMinimum),
		DiscordRestoreVoteCooldownMinutes:        intPointer(DiscordRestoreVoteCooldownMinutes),
		BackupRetentionEnabled:                   &BackupRetentionEnabled,
		BackupKeepNewestCount:                    intPointer(BackupKeepNewestCount),
		BackupDailyRetentionDays:                 intPointer(BackupDailyRetentionDays),
		BackupWeeklyRetentionWeeks:               intPointer(BackupWeeklyRetentionWeeks),
		BackupMonthlyRetentionMonths:             intPointer(BackupMonthlyRetentionMonths),
		BackupCleanupIntervalHours:               intPointer(int(BackupCleanupInterval / time.Hour)),
		IsNewTerrainAndSaveSystem:                &IsNewTerrainAndSaveSystem,
		GameBranch:                               GameBranch,
		Difficulty:                               Difficulty,
		StartCondition:                           StartCondition,
		StartLocation:                            StartLocation,
		ServerName:                               ServerName,
		SaveName:                                 SaveName,
		WorldID:                                  WorldID,
		ServerMaxPlayers:                         ServerMaxPlayers,
		ServerPassword:                           ServerPassword,
		ServerAuthSecret:                         ServerAuthSecret,
		AdminPassword:                            AdminPassword,
		GamePort:                                 GamePort,
		UpdatePort:                               UpdatePort,
		UPNPEnabled:                              &UPNPEnabled,
		AutoSave:                                 &AutoSave,
		SaveInterval:                             SaveInterval,
		AutoPauseServer:                          &AutoPauseServer,
		LocalIpAddress:                           LocalIpAddress,
		StartLocalHost:                           &StartLocalHost,
		ServerVisible:                            &ServerVisible,
		UseSteamP2P:                              &UseSteamP2P,
		ExePath:                                  ExePath,
		AdditionalParams:                         AdditionalParams,
		Users:                                    Users,
		AuthEnabled:                              &AuthEnabled,
		JwtKey:                                   JwtKey,
		AuthTokenLifetime:                        AuthTokenLifetime,
		Debug:                                    &IsDebugMode,
		CreateSSUILogFile:                        &CreateSSUILogFile,
		CreateGameServerLogFile:                  &CreateGameServerLogFile,
		LogLevel:                                 LogLevel,
		LogClutterToConsole:                      &LogClutterToConsole,
		SubsystemFilters:                         SubsystemFilters,
		IsUpdateEnabled:                          &IsUpdateEnabled,
		IsSSCMEnabled:                            &IsSSCMEnabled,
		AutoRestartServerTimer:                   AutoRestartServerTimer,
		AutoRestartCountdown:                     AutoRestartCountdown,
		AllowPrereleaseUpdates:                   &AllowPrereleaseUpdates,
		AllowMajorUpdates:                        &AllowMajorUpdates,
		AllowAutoGameServerUpdates:               &AllowAutoGameServerUpdates,
		IsStationeersLaunchPadEnabled:            &IsStationeersLaunchPadEnabled,
		IsStationeersLaunchPadAutoUpdatesEnabled: &IsStationeersLaunchPadAutoUpdatesEnabled,
		IsConsoleEnabled:                         &IsConsoleEnabled,
		IsCLIDashboardEnabled:                    &IsCLIDashboardEnabled,
		LanguageSetting:                          LanguageSetting,
		AutoStartServerOnStartup:                 &AutoStartServerOnStartup,
		SSUIIdentifier:                           SSUIIdentifier,
		SSUIWebPort:                              SSUIWebPort,
		AdvertiserOverride:                       AdvertiserOverride,
		ShowExpertSettings:                       &ShowExpertSettings,
	}

	file, err := os.Create(ConfigPath)
	if err != nil {
		return fmt.Errorf("error creating config.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("error encoding config.json: %v", err)
	}

	return nil
}

// use SaveConfig EXCLUSIVELY though loader.SaveConfig to trigger a reload afterwards!
// when the config gets updated, changes do not get reflected at runtime UNLESS a backend reload / config reload is triggered
// This can be done via configchanger.SaveConfig
func SaveConfigToFile(cfg *JsonConfig) error {

	ConfigMu.Lock()
	defer ConfigMu.Unlock()

	file, err := os.Create(ConfigPath)
	if err != nil {
		return fmt.Errorf("error creating config.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("error encoding config.json: %v", err)
	}

	return nil
}
