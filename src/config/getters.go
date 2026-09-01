package config

import (
	"time"
)

func GetDiscordToken() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return DiscordToken
}

func GetControlChannelID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ControlChannelID
}

func GetEventLogChannelID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return EventLogChannelID
}

func GetStatusPanelChannelID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return StatusPanelChannelID
}

func GetLogChannelID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LogChannelID
}

func GetControlPanelChannelID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ControlPanelChannelID
}

func GetDiscordCharBufferSize() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return DiscordCharBufferSize
}

func GetBlackListFilePath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BlackListFilePath
}

func GetIsDiscordEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsDiscordEnabled
}

func GetRotateServerPassword() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return RotateServerPassword
}

func GetBackupKeepNewestCount() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupKeepNewestCount
}

func GetBackupRetentionEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupRetentionEnabled
}

func GetBackupDailyRetentionDays() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupDailyRetentionDays
}

func GetBackupWeeklyRetentionWeeks() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupWeeklyRetentionWeeks
}

func GetBackupMonthlyRetentionMonths() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupMonthlyRetentionMonths
}

// GetBackupCleanupInterval returns the cleanup interval in hours.
func GetBackupCleanupInterval() time.Duration {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return BackupCleanupInterval
}

func GetIsNewTerrainAndSaveSystem() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsNewTerrainAndSaveSystem
}

func GetGameBranch() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return GameBranch
}

func GetDifficulty() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return Difficulty
}

func GetStartCondition() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return StartCondition
}

func GetStartLocation() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return StartLocation
}

func GetServerName() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ServerName
}

// special getter for backwards compatibility with SaveInfo
func GetLegacySaveInfo() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	saveinfo := SaveName + ";" + WorldID
	return saveinfo
}

func GetSaveName() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SaveName
}

func GetWorldID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return WorldID
}

func GetServerMaxPlayers() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ServerMaxPlayers
}

func GetServerPassword() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ServerPassword
}

func GetServerAuthSecret() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ServerAuthSecret
}

func GetAdminPassword() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AdminPassword
}

func GetGamePort() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return GamePort
}

func GetUpdatePort() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return UpdatePort
}

func GetUPNPEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return UPNPEnabled
}

func GetAutoSave() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AutoSave
}

func GetSaveInterval() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SaveInterval
}

func GetAutoPauseServer() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AutoPauseServer
}

func GetLocalIpAddress() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LocalIpAddress
}

func GetStartLocalHost() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return StartLocalHost
}

func GetServerVisible() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ServerVisible
}

func GetUseSteamP2P() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return UseSteamP2P
}

func GetExePath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ExePath
}

func GetAdditionalParams() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AdditionalParams
}

func GetUsers() map[string]string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return Users
}

func GetAuthEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AuthEnabled
}

func GetJwtKey() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return JwtKey
}

func GetAuthTokenLifetime() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AuthTokenLifetime
}

func GetIsDebugMode() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsDebugMode
}

func GetCreateSSUILogFile() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return CreateSSUILogFile
}

func GetCreateGameServerLogFile() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return CreateGameServerLogFile
}

func GetLogLevel() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LogLevel
}

func GetLogClutterToConsole() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LogClutterToConsole
}

func GetSubsystemFilters() []string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SubsystemFilters
}

func GetIsUpdateEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsUpdateEnabled
}

func GetIsSSCMEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsSSCMEnabled
}

func GetSSCMFilePath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSCMFilePath
}

func GetSSCMPluginDir() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSCMPluginDir
}

func GetSSCMWebDir() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSCMWebDir
}

func GetAutoRestartServerTimer() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AutoRestartServerTimer
}

func GetAutoRestartCountdown() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AutoRestartCountdown
}

func GetNextAutoRestartTime() time.Time {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return NextAutoRestartTime
}

func GetAllowPrereleaseUpdates() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AllowPrereleaseUpdates
}

func GetAllowMajorUpdates() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AllowMajorUpdates
}

func GetIsConsoleEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsConsoleEnabled
}

func GetIsCLIDashboardEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsCLIDashboardEnabled
}

func GetLanguageSetting() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LanguageSetting
}

func GetAutoStartServerOnStartup() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AutoStartServerOnStartup
}

func GetSSUIIdentifier() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSUIIdentifier
}

func GetSSUIWebPort() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSUIWebPort
}

func GetShowExpertSettings() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ShowExpertSettings
}

// GetIsFirstTimeSetup returns the IsFirstTimeSetup
func GetIsFirstTimeSetup() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsFirstTimeSetup
}

func GetConfigPath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ConfigPath
}

func GetVersion() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return Version
}

func GetBranch() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return Branch
}

func GetTLSCertPath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return TLSCertPath
}

func GetTLSKeyPath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return TLSKeyPath
}

func GetUIModFolder() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return UIModFolder
}

func GetMaxSSEConnections() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return MaxSSEConnections
}

func GetSSEMessageBufferSize() int {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SSEMessageBufferSize
}

func GetLogFolder() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return LogFolder
}

func GetConfiguredBackupDir() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ConfiguredBackupDir
}

func GetConfiguredSafeBackupDir() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ConfiguredSafeBackupDir
}

func GetCustomDetectionsFilePath() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return CustomDetectionsFilePath
}

func GetGameServerAppID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return GameServerAppID
}

func GetCurrentBranchBuildID() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return CurrentBranchBuildID
}

func GetAllowAutoGameServerUpdates() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AllowAutoGameServerUpdates
}

func GetExtractedGameVersion() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return ExtractedGameVersion
}

func GetSkipSteamCMD() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return SkipSteamCMD
}

func GetNoSanityCheck() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return NoSanityCheck
}

func GetIsDockerContainer() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsDockerContainer
}

func GetIsGameServerRunning() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsGameServerRunning
}
func GetAdvertiserOverride() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return AdvertiserOverride
}

func GetStationeersServerPingEndpoint() string {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return StationeersServerPingEndpoint
}

func GetIsStationeersLaunchPadEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsStationeersLaunchPadEnabled
}

func GetIsStationeersLaunchPadAutoUpdatesEnabled() bool {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return IsStationeersLaunchPadAutoUpdatesEnabled
}
