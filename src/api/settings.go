package api

// SettingsPatch is the public settings surface used by the web UI. Fields not
// present here remain private implementation details of config.json.
type SettingsPatch struct {
	GameBranch       *string `json:"gameBranch,omitempty"`
	GamePort         *string `json:"gamePort,omitempty"`
	ServerName       *string `json:"serverName,omitempty"`
	SaveName         *string `json:"saveName,omitempty"`
	WorldID          *string `json:"worldId,omitempty"`
	ServerMaxPlayers *string `json:"serverMaxPlayers,omitempty"`
	ServerPassword   *string `json:"serverPassword,omitempty"`
	ServerAuthSecret *string `json:"serverAuthSecret,omitempty"`
	AdminPassword    *string `json:"adminPassword,omitempty"`
	UpdatePort       *string `json:"updatePort,omitempty"`
	UPNPEnabled      *bool   `json:"upnpEnabled,omitempty"`
	AutoSave         *bool   `json:"autoSave,omitempty"`
	SaveInterval     *string `json:"saveInterval,omitempty"`
	AutoPauseServer  *bool   `json:"autoPauseServer,omitempty"`
	LocalIpAddress   *string `json:"localIpAddress,omitempty"`
	StartLocalHost   *bool   `json:"startLocalHost,omitempty"`
	ServerVisible    *bool   `json:"serverVisible,omitempty"`
	UseSteamP2P      *bool   `json:"useSteamP2P,omitempty"`
	AdditionalParams *string `json:"additionalParams,omitempty"`
	Difficulty       *string `json:"difficulty,omitempty"`
	StartCondition   *string `json:"startCondition,omitempty"`
	StartLocation    *string `json:"startLocation,omitempty"`

	Debug                    *bool `json:"debug,omitempty"`
	CreateSSUILogFile        *bool `json:"createSSUILogFile,omitempty"`
	CreateGameServerLogFile  *bool `json:"createGameServerLogFile,omitempty"`
	LogLevel                 *int  `json:"logLevel,omitempty"`
	ConnectivityCheckEnabled *bool `json:"connectivityCheckEnabled,omitempty"`

	ExePath                  *string `json:"exePath,omitempty"`
	LogClutterToConsole      *bool   `json:"logClutterToConsole,omitempty"`
	IsSSCMEnabled            *bool   `json:"sscmEnabled,omitempty"`
	AutoRestartServerTimer   *string `json:"autoRestartServerTimer,omitempty"`
	AutoRestartCountdown     *string `json:"autoRestartCountdown,omitempty"`
	IsConsoleEnabled         *bool   `json:"consoleEnabled,omitempty"`
	LanguageSetting          *string `json:"language,omitempty"`
	AutoStartServerOnStartup *bool   `json:"autoStartServerOnStartup,omitempty"`
	SSUIWebPort              *string `json:"ssuiWebPort,omitempty"`
	ShowExpertSettings       *bool   `json:"showExpertSettings,omitempty"`

	IsUpdateEnabled            *bool `json:"updatesEnabled,omitempty"`
	AllowPrereleaseUpdates     *bool `json:"allowPrereleaseUpdates,omitempty"`
	AllowMajorUpdates          *bool `json:"allowMajorUpdates,omitempty"`
	AllowAutoGameServerUpdates *bool `json:"allowAutoGameServerUpdates,omitempty"`

	IsStationeersLaunchPadAutoUpdatesEnabled *bool `json:"stationeersLaunchPadAutoUpdatesEnabled,omitempty"`

	DiscordToken                      *string `json:"discordToken,omitempty"`
	DiscordAdminRoleID                *string `json:"discordAdminRoleId,omitempty"`
	EventLogChannelID                 *string `json:"eventLogChannelId,omitempty"`
	StatusPanelChannelID              *string `json:"statusPanelChannelId,omitempty"`
	LogChannelID                      *string `json:"logChannelId,omitempty"`
	BlackListFilePath                 *string `json:"blacklistFilePath,omitempty"`
	IsDiscordEnabled                  *bool   `json:"discordEnabled,omitempty"`
	RotateServerPassword              *bool   `json:"rotateServerPassword,omitempty"`
	DiscordRestartVoteEnabled         *bool   `json:"discordRestartVoteEnabled,omitempty"`
	DiscordRestoreVoteEnabled         *bool   `json:"discordRestoreVoteEnabled,omitempty"`
	DiscordVoteDurationMinutes        *int    `json:"discordVoteDurationMinutes,omitempty"`
	DiscordRestartVoteThreshold       *int    `json:"discordRestartVoteThreshold,omitempty"`
	DiscordRestartVoteMinimum         *int    `json:"discordRestartVoteMinimum,omitempty"`
	DiscordRestartVoteCooldownMinutes *int    `json:"discordRestartVoteCooldownMinutes,omitempty"`
	DiscordRestoreVoteThreshold       *int    `json:"discordRestoreVoteThreshold,omitempty"`
	DiscordRestoreVoteMinimum         *int    `json:"discordRestoreVoteMinimum,omitempty"`
	DiscordRestoreVoteCooldownMinutes *int    `json:"discordRestoreVoteCooldownMinutes,omitempty"`

	BackupRetentionEnabled       *bool `json:"backupRetentionEnabled,omitempty"`
	BackupKeepNewestCount        *int  `json:"backupKeepNewestCount,omitempty"`
	BackupDailyRetentionDays     *int  `json:"backupDailyRetentionDays,omitempty"`
	BackupWeeklyRetentionWeeks   *int  `json:"backupWeeklyRetentionWeeks,omitempty"`
	BackupMonthlyRetentionMonths *int  `json:"backupMonthlyRetentionMonths,omitempty"`
	BackupCleanupIntervalHours   *int  `json:"backupCleanupIntervalHours,omitempty"`
}

type Settings struct {
	SettingsPatch
	DiscordTokenConfigured bool `json:"discordTokenConfigured"`
}

type SettingsUpdate struct {
	Message         string   `json:"message"`
	RestartRequired []string `json:"restartRequired,omitempty"`
}

// SetupSettingsPatch is deliberately smaller than the authenticated settings
// API. It is available only while first-time setup is still incomplete.
type SetupSettingsPatch struct {
	GameBranch       *string `json:"gameBranch,omitempty"`
	ServerName       *string `json:"serverName,omitempty"`
	SaveName         *string `json:"saveName,omitempty"`
	WorldID          *string `json:"worldId,omitempty"`
	ServerMaxPlayers *string `json:"serverMaxPlayers,omitempty"`
	ServerPassword   *string `json:"serverPassword,omitempty"`
	GamePort         *string `json:"gamePort,omitempty"`
	UpdatePort       *string `json:"updatePort,omitempty"`
	UPNPEnabled      *bool   `json:"upnpEnabled,omitempty"`
	LocalIpAddress   *string `json:"localIpAddress,omitempty"`
	LanguageSetting  *string `json:"language,omitempty"`
}
