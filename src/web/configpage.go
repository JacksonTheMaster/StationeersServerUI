package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"text/template"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/localization"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func ServeConfigPage(w http.ResponseWriter, r *http.Request) {

	htmlFS, err := fs.Sub(config.V1UIFS, "SSUI/onboard_bundled/ui")
	if err != nil {
		http.Error(w, "Error accessing Virt FS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(htmlFS, "config.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Core.Error("failed to serve config.html")
		return
	}

	// Determine selected attributes for boolean fields
	upnpTrueSelected := ""
	upnpFalseSelected := ""
	if config.GetUPNPEnabled() {
		upnpTrueSelected = "selected"
	} else {
		upnpFalseSelected = "selected"
	}

	discordTrueSelected := ""
	discordFalseSelected := ""
	if config.GetIsDiscordEnabled() {
		discordTrueSelected = "selected"
	} else {
		discordFalseSelected = "selected"
	}

	autoSaveTrueSelected := ""
	autoSaveFalseSelected := ""
	if config.GetAutoSave() {
		autoSaveTrueSelected = "selected"
	} else {
		autoSaveFalseSelected = "selected"
	}

	autoPauseTrueSelected := ""
	autoPauseFalseSelected := ""
	if config.GetAutoPauseServer() {
		autoPauseTrueSelected = "selected"
	} else {
		autoPauseFalseSelected = "selected"
	}

	startLocalTrueSelected := ""
	startLocalFalseSelected := ""
	if config.GetStartLocalHost() {
		startLocalTrueSelected = "selected"
	} else {
		startLocalFalseSelected = "selected"
	}

	serverVisibleTrueSelected := ""
	serverVisibleFalseSelected := ""
	if config.GetServerVisible() {
		serverVisibleTrueSelected = "selected"
	} else {
		serverVisibleFalseSelected = "selected"
	}

	isNewTerrainAndSaveSystemTrueSelected := ""

	if config.GetIsNewTerrainAndSaveSystem() {
		isNewTerrainAndSaveSystemTrueSelected = "selected"
	}

	autoStartServerTrueSelected := ""
	autoStartServerFalseSelected := ""
	if config.GetAutoStartServerOnStartup() {
		autoStartServerTrueSelected = "selected"
	} else {
		autoStartServerFalseSelected = "selected"
	}

	steamP2PTrueSelected := ""
	steamP2PFalseSelected := ""
	if config.GetUseSteamP2P() {
		steamP2PTrueSelected = "selected"
	} else {
		steamP2PFalseSelected = "selected"
	}

	autoGameServerUpdatesTrueSelected := ""
	autoGameServerUpdatesFalseSelected := ""
	if config.GetAllowAutoGameServerUpdates() {
		autoGameServerUpdatesTrueSelected = "selected"
	} else {
		autoGameServerUpdatesFalseSelected = "selected"
	}

	createSSUILogFileTrueSelected := ""
	createSSUILogFileFalseSelected := ""
	if config.GetCreateSSUILogFile() {
		createSSUILogFileTrueSelected = "selected"
	} else {
		createSSUILogFileFalseSelected = "selected"
	}

	createGameServerLogFileTrueSelected := ""
	createGameServerLogFileFalseSelected := ""
	if config.GetCreateGameServerLogFile() {
		createGameServerLogFileTrueSelected = "selected"
	} else {
		createGameServerLogFileFalseSelected = "selected"
	}

	isStationeersLaunchPadEnabled := "false"
	if config.GetIsStationeersLaunchPadEnabled() {
		isStationeersLaunchPadEnabled = "true"
	}

	rotateServerPasswordTrueSelected := ""
	rotateServerPasswordFalseSelected := ""
	if config.GetRotateServerPassword() {
		rotateServerPasswordTrueSelected = "selected"
	} else {
		rotateServerPasswordFalseSelected = "selected"
	}
	discordRestartVoteEnabledTrueSelected, discordRestartVoteEnabledFalseSelected := "", ""
	if config.GetDiscordRestartVoteEnabled() {
		discordRestartVoteEnabledTrueSelected = "selected"
	} else {
		discordRestartVoteEnabledFalseSelected = "selected"
	}
	discordRestoreVoteEnabledTrueSelected, discordRestoreVoteEnabledFalseSelected := "", ""
	if config.GetDiscordRestoreVoteEnabled() {
		discordRestoreVoteEnabledTrueSelected = "selected"
	} else {
		discordRestoreVoteEnabledFalseSelected = "selected"
	}

	// Expert Settings toggle
	showExpertSettingsTrueSelected := ""
	showExpertSettingsFalseSelected := ""
	if config.GetShowExpertSettings() {
		showExpertSettingsTrueSelected = "selected"
	} else {
		showExpertSettingsFalseSelected = "selected"
	}

	// Expert Settings booleans
	debugTrueSelected := ""
	debugFalseSelected := ""
	if config.GetIsDebugMode() {
		debugTrueSelected = "selected"
	} else {
		debugFalseSelected = "selected"
	}

	logClutterToConsoleTrueSelected := ""
	logClutterToConsoleFalseSelected := ""
	if config.GetLogClutterToConsole() {
		logClutterToConsoleTrueSelected = "selected"
	} else {
		logClutterToConsoleFalseSelected = "selected"
	}

	isSSCMEnabledTrueSelected := ""
	isSSCMEnabledFalseSelected := ""
	if config.GetIsSSCMEnabled() {
		isSSCMEnabledTrueSelected = "selected"
	} else {
		isSSCMEnabledFalseSelected = "selected"
	}

	isConsoleEnabledTrueSelected := ""
	isConsoleEnabledFalseSelected := ""
	if config.GetIsConsoleEnabled() {
		isConsoleEnabledTrueSelected = "selected"
	} else {
		isConsoleEnabledFalseSelected = "selected"
	}

	isUpdateEnabledTrueSelected := ""
	isUpdateEnabledFalseSelected := ""
	if config.GetIsUpdateEnabled() {
		isUpdateEnabledTrueSelected = "selected"
	} else {
		isUpdateEnabledFalseSelected = "selected"
	}

	allowPrereleaseUpdatesTrueSelected := ""
	allowPrereleaseUpdatesFalseSelected := ""
	if config.GetAllowPrereleaseUpdates() {
		allowPrereleaseUpdatesTrueSelected = "selected"
	} else {
		allowPrereleaseUpdatesFalseSelected = "selected"
	}

	allowMajorUpdatesTrueSelected := ""
	allowMajorUpdatesFalseSelected := ""
	if config.GetAllowMajorUpdates() {
		allowMajorUpdatesTrueSelected = "selected"
	} else {
		allowMajorUpdatesFalseSelected = "selected"
	}

	authEnabledTrueSelected := ""
	authEnabledFalseSelected := ""
	if config.GetAuthEnabled() {
		authEnabledTrueSelected = "selected"
	} else {
		authEnabledFalseSelected = "selected"
	}

	isStationeersLaunchPadAutoUpdatesEnabledTrueSelected := ""
	isStationeersLaunchPadAutoUpdatesEnabledFalseSelected := ""
	if config.GetIsStationeersLaunchPadAutoUpdatesEnabled() {
		isStationeersLaunchPadAutoUpdatesEnabledTrueSelected = "selected"
	} else {
		isStationeersLaunchPadAutoUpdatesEnabledFalseSelected = "selected"
	}

	backupRetentionEnabledTrueSelected := ""
	backupRetentionEnabledFalseSelected := ""
	if config.GetBackupRetentionEnabled() {
		backupRetentionEnabledTrueSelected = "selected"
	} else {
		backupRetentionEnabledFalseSelected = "selected"
	}

	data := ConfigTemplateData{
		Permissions:         pagePermissions(r),
		CanViewSettings:     pageCan(r, security.PermissionSettingsView) || pageCan(r, security.PermissionSettingsManage),
		CanManageSettings:   pageCan(r, security.PermissionSettingsManage),
		CanManageSLP:        pageCan(r, security.PermissionSLPManage),
		CanManageDetections: pageCan(r, security.PermissionDetectionsManage),

		// Config values
		DiscordToken:                            config.GetDiscordToken(),
		DiscordAdminRoleID:                      config.GetDiscordAdminRoleID(),
		EventLogChannelID:                       config.GetEventLogChannelID(),
		StatusPanelChannelID:                    config.GetStatusPanelChannelID(),
		LogChannelID:                            config.GetLogChannelID(),
		BlackListFilePath:                       config.GetBlackListFilePath(),
		IsDiscordEnabled:                        fmt.Sprintf("%v", config.GetIsDiscordEnabled()),
		IsDiscordEnabledTrueSelected:            discordTrueSelected,
		IsDiscordEnabledFalseSelected:           discordFalseSelected,
		RotateServerPassword:                    fmt.Sprintf("%v", config.GetRotateServerPassword()),
		RotateServerPasswordTrueSelected:        rotateServerPasswordTrueSelected,
		RotateServerPasswordFalseSelected:       rotateServerPasswordFalseSelected,
		DiscordRestartVoteEnabledTrueSelected:   discordRestartVoteEnabledTrueSelected,
		DiscordRestartVoteEnabledFalseSelected:  discordRestartVoteEnabledFalseSelected,
		DiscordRestoreVoteEnabledTrueSelected:   discordRestoreVoteEnabledTrueSelected,
		DiscordRestoreVoteEnabledFalseSelected:  discordRestoreVoteEnabledFalseSelected,
		DiscordVoteDurationMinutes:              strconv.Itoa(config.GetDiscordVoteDurationMinutes()),
		DiscordRestartVoteThreshold:             strconv.Itoa(config.GetDiscordRestartVoteThreshold()),
		DiscordRestartVoteMinimum:               strconv.Itoa(config.GetDiscordRestartVoteMinimum()),
		DiscordRestartVoteCooldownMinutes:       strconv.Itoa(config.GetDiscordRestartVoteCooldownMinutes()),
		DiscordRestoreVoteThreshold:             strconv.Itoa(config.GetDiscordRestoreVoteThreshold()),
		DiscordRestoreVoteMinimum:               strconv.Itoa(config.GetDiscordRestoreVoteMinimum()),
		DiscordRestoreVoteCooldownMinutes:       strconv.Itoa(config.GetDiscordRestoreVoteCooldownMinutes()),
		GameBranch:                              config.GetGameBranch(),
		Difficulty:                              config.GetDifficulty(),
		StartCondition:                          config.GetStartCondition(),
		StartLocation:                           config.GetStartLocation(),
		ServerName:                              config.GetServerName(),
		SaveName:                                config.GetSaveName(),
		WorldID:                                 config.GetWorldID(),
		ServerMaxPlayers:                        config.GetServerMaxPlayers(),
		ServerPassword:                          config.GetServerPassword(),
		ServerAuthSecret:                        config.GetServerAuthSecret(),
		AdminPassword:                           config.GetAdminPassword(),
		GamePort:                                config.GetGamePort(),
		UpdatePort:                              config.GetUpdatePort(),
		UPNPEnabled:                             fmt.Sprintf("%v", config.GetUPNPEnabled()),
		UPNPEnabledTrueSelected:                 upnpTrueSelected,
		UPNPEnabledFalseSelected:                upnpFalseSelected,
		AutoSave:                                fmt.Sprintf("%v", config.GetAutoSave()),
		AutoSaveTrueSelected:                    autoSaveTrueSelected,
		AutoSaveFalseSelected:                   autoSaveFalseSelected,
		SaveInterval:                            config.GetSaveInterval(),
		AutoPauseServer:                         fmt.Sprintf("%v", config.GetAutoPauseServer()),
		AutoPauseServerTrueSelected:             autoPauseTrueSelected,
		AutoPauseServerFalseSelected:            autoPauseFalseSelected,
		LocalIpAddress:                          config.GetLocalIpAddress(),
		StartLocalHost:                          fmt.Sprintf("%v", config.GetStartLocalHost()),
		StartLocalHostTrueSelected:              startLocalTrueSelected,
		StartLocalHostFalseSelected:             startLocalFalseSelected,
		ServerVisible:                           fmt.Sprintf("%v", config.GetServerVisible()),
		ServerVisibleTrueSelected:               serverVisibleTrueSelected,
		ServerVisibleFalseSelected:              serverVisibleFalseSelected,
		UseSteamP2P:                             fmt.Sprintf("%v", config.GetUseSteamP2P()),
		UseSteamP2PTrueSelected:                 steamP2PTrueSelected,
		UseSteamP2PFalseSelected:                steamP2PFalseSelected,
		ExePath:                                 config.GetExePath(),
		AdditionalParams:                        config.GetAdditionalParams(),
		AutoRestartServerTimer:                  config.GetAutoRestartServerTimer(),
		AutoRestartCountdown:                    config.GetAutoRestartCountdown(),
		IsNewTerrainAndSaveSystem:               fmt.Sprintf("%v", config.GetIsNewTerrainAndSaveSystem()),
		IsNewTerrainAndSaveSystemTrueSelected:   isNewTerrainAndSaveSystemTrueSelected,
		AutoStartServerOnStartup:                fmt.Sprintf("%v", config.GetAutoStartServerOnStartup()),
		AutoStartServerOnStartupTrueSelected:    autoStartServerTrueSelected,
		AutoStartServerOnStartupFalseSelected:   autoStartServerFalseSelected,
		AllowAutoGameServerUpdates:              fmt.Sprintf("%v", config.GetAllowAutoGameServerUpdates()),
		AllowAutoGameServerUpdatesTrueSelected:  autoGameServerUpdatesTrueSelected,
		AllowAutoGameServerUpdatesFalseSelected: autoGameServerUpdatesFalseSelected,
		CreateSSUILogFile:                       fmt.Sprintf("%v", config.GetCreateSSUILogFile()),
		CreateSSUILogFileTrueSelected:           createSSUILogFileTrueSelected,
		CreateSSUILogFileFalseSelected:          createSSUILogFileFalseSelected,
		CreateGameServerLogFile:                 fmt.Sprintf("%v", config.GetCreateGameServerLogFile()),
		CreateGameServerLogFileTrueSelected:     createGameServerLogFileTrueSelected,
		CreateGameServerLogFileFalseSelected:    createGameServerLogFileFalseSelected,
		BackupKeepNewestCount:                   strconv.Itoa(config.GetBackupKeepNewestCount()),
		BackupRetentionEnabled:                  fmt.Sprintf("%v", config.GetBackupRetentionEnabled()),
		BackupRetentionEnabledTrueSelected:      backupRetentionEnabledTrueSelected,
		BackupRetentionEnabledFalseSelected:     backupRetentionEnabledFalseSelected,
		BackupDailyRetentionDays:                strconv.Itoa(config.GetBackupDailyRetentionDays()),
		BackupWeeklyRetentionWeeks:              strconv.Itoa(config.GetBackupWeeklyRetentionWeeks()),
		BackupMonthlyRetentionMonths:            strconv.Itoa(config.GetBackupMonthlyRetentionMonths()),
		BackupCleanupIntervalHours:              strconv.Itoa(int(config.GetBackupCleanupInterval().Hours())),

		// Localized UI text
		UIText_ConfigHeadline:            localization.GetString("UIText_ConfigHeadline"),
		UIText_ServerConfig:              localization.GetString("UIText_ServerConfig"),
		UIText_BackToDashboard:           localization.GetString("UIText_BackToDashboard"),
		UIText_DiscordIntegration:        localization.GetString("UIText_DiscordIntegration"),
		UIText_SLPModIntegration:         localization.GetString("UIText_SLPModIntegration"),
		UIText_DetectionManager:          localization.GetString("UIText_DetectionManager"),
		UIText_ConfigurationWizard:       localization.GetString("UIText_ConfigurationWizard"),
		UIText_PleaseSelectSection:       localization.GetString("UIText_PleaseSelectSection"),
		UIText_UseWizardAlternative:      localization.GetString("UIText_UseWizardAlternative"),
		UIText_BasicSettings:             localization.GetString("UIText_BasicSettings"),
		UIText_NetworkSettings:           localization.GetString("UIText_NetworkSettings"),
		UIText_AdvancedSettings:          localization.GetString("UIText_AdvancedSettings"),
		UIText_TerrainSettings:           localization.GetString("UIText_TerrainSettings"),
		UIText_ConfigSaving:              localization.GetString("UIText_ConfigSaving"),
		UIText_ConfigSaveSuccess:         localization.GetString("UIText_ConfigSaveSuccess"),
		UIText_ConfigSaveFailed:          localization.GetString("UIText_ConfigSaveFailed"),
		UIText_SaveConfiguration:         localization.GetString("UIText_SaveConfiguration"),
		UIText_BackupRetention:           localization.GetString("UIText_BackupRetention"),
		UIText_BackupRetentionTitle:      localization.GetString("UIText_BackupRetentionTitle"),
		UIText_BackupRetentionIntro:      localization.GetString("UIText_BackupRetentionIntro"),
		UIText_BackupCleanupEnabled:      localization.GetString("UIText_BackupCleanupEnabled"),
		UIText_BackupCleanupInfo:         localization.GetString("UIText_BackupCleanupInfo"),
		UIText_BackupKeepNewest:          localization.GetString("UIText_BackupKeepNewest"),
		UIText_BackupKeepNewestInfo:      localization.GetString("UIText_BackupKeepNewestInfo"),
		UIText_BackupKeepDaily:           localization.GetString("UIText_BackupKeepDaily"),
		UIText_BackupKeepDailyInfo:       localization.GetString("UIText_BackupKeepDailyInfo"),
		UIText_BackupKeepWeekly:          localization.GetString("UIText_BackupKeepWeekly"),
		UIText_BackupKeepWeeklyInfo:      localization.GetString("UIText_BackupKeepWeeklyInfo"),
		UIText_BackupKeepMonthly:         localization.GetString("UIText_BackupKeepMonthly"),
		UIText_BackupKeepMonthlyInfo:     localization.GetString("UIText_BackupKeepMonthlyInfo"),
		UIText_BackupCleanupInterval:     localization.GetString("UIText_BackupCleanupInterval"),
		UIText_BackupCleanupIntervalInfo: localization.GetString("UIText_BackupCleanupIntervalInfo"),
		UIText_BackupHours:               localization.GetString("UIText_BackupHours"),
		UIText_BackupDays:                localization.GetString("UIText_BackupDays"),
		UIText_BackupWeeks:               localization.GetString("UIText_BackupWeeks"),
		UIText_BackupMonths:              localization.GetString("UIText_BackupMonths"),
		UIText_BackupFiles:               localization.GetString("UIText_BackupFiles"),
		UIText_BackupWarningTitle:        localization.GetString("UIText_BackupWarningTitle"),
		UIText_BackupWarning:             localization.GetString("UIText_BackupWarning"),
		UIText_BackupCopyDelayInfo:       localization.GetString("UIText_BackupCopyDelayInfo"),
		UIText_BasicServerSettings:       localization.GetString("UIText_BasicServerSettings"),

		UIText_ServerName:                     localization.GetString("UIText_ServerName"),
		UIText_ServerNameInfo:                 localization.GetString("UIText_ServerNameInfo"),
		UIText_SaveName:                       localization.GetString("UIText_SaveName"),
		UIText_SaveNameInfo:                   localization.GetString("UIText_SaveNameInfo"),
		UIText_WorldID:                        localization.GetString("UIText_WorldID"),
		UIText_WorldIDInfo:                    localization.GetString("UIText_WorldIDInfo"),
		UIText_MaxPlayers:                     localization.GetString("UIText_MaxPlayers"),
		UIText_MaxPlayersInfo:                 localization.GetString("UIText_MaxPlayersInfo"),
		UIText_ServerPassword:                 localization.GetString("UIText_ServerPassword"),
		UIText_ServerPasswordInfo:             localization.GetString("UIText_ServerPasswordInfo"),
		UIText_AdminPassword:                  localization.GetString("UIText_AdminPassword"),
		UIText_AdminPasswordInfo:              localization.GetString("UIText_AdminPasswordInfo"),
		UIText_AutoSave:                       localization.GetString("UIText_AutoSave"),
		UIText_AutoSaveInfo:                   localization.GetString("UIText_AutoSaveInfo"),
		UIText_SaveInterval:                   localization.GetString("UIText_SaveInterval"),
		UIText_SaveIntervalInfo:               localization.GetString("UIText_SaveIntervalInfo"),
		UIText_AutoPauseServer:                localization.GetString("UIText_AutoPauseServer"),
		UIText_AutoPauseServerInfo:            localization.GetString("UIText_AutoPauseServerInfo"),
		UIText_NetworkConfiguration:           localization.GetString("UIText_NetworkConfiguration"),
		UIText_GamePort:                       localization.GetString("UIText_GamePort"),
		UIText_GamePortInfo:                   localization.GetString("UIText_GamePortInfo"),
		UIText_UpdatePort:                     localization.GetString("UIText_UpdatePort"),
		UIText_UpdatePortInfo:                 localization.GetString("UIText_UpdatePortInfo"),
		UIText_UPNPEnabled:                    localization.GetString("UIText_UPNPEnabled"),
		UIText_UPNPEnabledInfo:                localization.GetString("UIText_UPNPEnabledInfo"),
		UIText_LocalIpAddress:                 localization.GetString("UIText_LocalIpAddress"),
		UIText_LocalIpAddressInfo:             localization.GetString("UIText_LocalIpAddressInfo"),
		UIText_StartLocalHost:                 localization.GetString("UIText_StartLocalHost"),
		UIText_StartLocalHostInfo:             localization.GetString("UIText_StartLocalHostInfo"),
		UIText_ServerVisible:                  localization.GetString("UIText_ServerVisible"),
		UIText_ServerVisibleInfo:              localization.GetString("UIText_ServerVisibleInfo"),
		UIText_UseSteamP2P:                    localization.GetString("UIText_UseSteamP2P"),
		UIText_UseSteamP2PInfo:                localization.GetString("UIText_UseSteamP2PInfo"),
		UIText_AdvertiserTitle:                localization.GetString("UIText_AdvertiserTitle"),
		UIText_AdvertiserSummary:              localization.GetString("UIText_AdvertiserSummary"),
		UIText_AdvertiserConfigure:            localization.GetString("UIText_AdvertiserConfigure"),
		UIText_AdvertiserExplanation:          localization.GetString("UIText_AdvertiserExplanation"),
		UIText_AdvertiserMasterServer:         localization.GetString("UIText_AdvertiserMasterServer"),
		UIText_AdvertiserDisabled:             localization.GetString("UIText_AdvertiserDisabled"),
		UIText_AdvertiserAuto:                 localization.GetString("UIText_AdvertiserAuto"),
		UIText_AdvertiserIPv4:                 localization.GetString("UIText_AdvertiserIPv4"),
		UIText_AdvertiserDNS:                  localization.GetString("UIText_AdvertiserDNS"),
		UIText_AdvertiserAddress:              localization.GetString("UIText_AdvertiserAddress"),
		UIText_AdvertiserRestartWarning:       localization.GetString("UIText_AdvertiserRestartWarning"),
		UIText_AdvertiserSaveRestart:          localization.GetString("UIText_AdvertiserSaveRestart"),
		UIText_AdvertiserRestarting:           localization.GetString("UIText_AdvertiserRestarting"),
		UIText_Cancel:                         localization.GetString("UIText_Cancel"),
		UIText_Close:                          localization.GetString("UIText_Close"),
		UIText_AdvancedConfiguration:          localization.GetString("UIText_AdvancedConfiguration"),
		UIText_TLSCertificateTitle:            localization.GetString("UIText_TLSCertificateTitle"),
		UIText_TLSCertificateSummary:          localization.GetString("UIText_TLSCertificateSummary"),
		UIText_TLSConfigure:                   localization.GetString("UIText_TLSConfigure"),
		UIText_TLSCertificateExplanation:      localization.GetString("UIText_TLSCertificateExplanation"),
		UIText_TLSCertificateFormats:          localization.GetString("UIText_TLSCertificateFormats"),
		UIText_TLSCertificateFile:             localization.GetString("UIText_TLSCertificateFile"),
		UIText_TLSPrivateKeyFile:              localization.GetString("UIText_TLSPrivateKeyFile"),
		UIText_TLSRestartRequired:             localization.GetString("UIText_TLSRestartRequired"),
		UIText_TLSRestartWarning:              localization.GetString("UIText_TLSRestartWarning"),
		UIText_TLSSaveRestart:                 localization.GetString("UIText_TLSSaveRestart"),
		UIText_TLSRestarting:                  localization.GetString("UIText_TLSRestarting"),
		UIText_ServerAuthSecret:               localization.GetString("UIText_ServerAuthSecret"),
		UIText_ServerAuthSecretInfo:           localization.GetString("UIText_ServerAuthSecretInfo"),
		UIText_ServerExePath:                  localization.GetString("UIText_ServerExePath"),
		UIText_ServerExePathInfo:              localization.GetString("UIText_ServerExePathInfo"),
		UIText_ServerExePathInfo2:             localization.GetString("UIText_ServerExePathInfo2"),
		UIText_AdditionalParams:               localization.GetString("UIText_AdditionalParams"),
		UIText_AdditionalParamsInfo:           localization.GetString("UIText_AdditionalParamsInfo"),
		UIText_ShowExpertSettings:             localization.GetString("UIText_ShowExpertSettings"),
		UIText_ShowExpertSettingsInfo:         localization.GetString("UIText_ShowExpertSettingsInfo"),
		UIText_ExpertSettings:                 localization.GetString("UIText_ExpertSettings"),
		UIText_ExpertSettingsWarning:          localization.GetString("UIText_ExpertSettingsWarning"),
		UIText_AutoRestartServerTimer:         localization.GetString("UIText_AutoRestartServerTimer"),
		UIText_AutoRestartServerTimerInfo:     localization.GetString("UIText_AutoRestartServerTimerInfo"),
		UIText_AutoRestartCountdown:           localization.GetString("UIText_AutoRestartCountdown"),
		UIText_AutoRestartCountdownInfo:       localization.GetString("UIText_AutoRestartCountdownInfo"),
		UIText_GameBranch:                     localization.GetString("UIText_GameBranch"),
		UIText_GameBranchInfo:                 localization.GetString("UIText_GameBranchInfo"),
		UIText_TerrainSettingsHeader:          localization.GetString("UIText_TerrainSettingsHeader"),
		UIText_TerrainWarning:                 localization.GetString("UIText_TerrainWarning"),
		UIText_Difficulty:                     localization.GetString("UIText_Difficulty"),
		UIText_DifficultyInfo:                 localization.GetString("UIText_DifficultyInfo"),
		UIText_StartCondition:                 localization.GetString("UIText_StartCondition"),
		UIText_StartConditionInfo:             localization.GetString("UIText_StartConditionInfo"),
		UIText_StartLocation:                  localization.GetString("UIText_StartLocation"),
		UIText_StartLocationInfo:              localization.GetString("UIText_StartLocationInfo"),
		UIText_TerrainSettingsFillHint:        localization.GetString("UIText_TerrainSettingsFillHint"),
		UIText_AutoStartServerOnStartup:       localization.GetString("UIText_AutoStartServerOnStartup"),
		UIText_AutoStartServerOnStartupInfo:   localization.GetString("UIText_AutoStartServerOnStartupInfo"),
		UIText_AllowAutoGameServerUpdates:     localization.GetString("UIText_AllowAutoGameServerUpdates"),
		UIText_AllowAutoGameServerUpdatesInfo: localization.GetString("UIText_AllowAutoGameServerUpdatesInfo"),
		UIText_CreateSSUILogFile:              localization.GetString("UIText_CreateSSUILogFile"),
		UIText_CreateSSUILogFileInfo:          localization.GetString("UIText_CreateSSUILogFileInfo"),
		UIText_CreateGameServerLogFile:        localization.GetString("UIText_CreateGameServerLogFile"),
		UIText_CreateGameServerLogFileInfo:    localization.GetString("UIText_CreateGameServerLogFileInfo"),

		UIText_DiscordIntegrationTitle:    localization.GetString("UIText_DiscordIntegrationTitle"),
		UIText_DiscordBotToken:            localization.GetString("UIText_DiscordBotToken"),
		UIText_DiscordBotTokenInfo:        localization.GetString("UIText_DiscordBotTokenInfo"),
		UIText_ChannelConfiguration:       localization.GetString("UIText_ChannelConfiguration"),
		UIText_DiscordAdminRole:           localization.GetString("UIText_DiscordAdminRole"),
		UIText_DiscordAdminRoleInfo:       localization.GetString("UIText_DiscordAdminRoleInfo"),
		UIText_RotateServerPassword:       localization.GetString("UIText_RotateServerPassword"),
		UIText_RotateServerPasswordInfo:   localization.GetString("UIText_RotateServerPasswordInfo"),
		UIText_EventLogChannel:            localization.GetString("UIText_EventLogChannel"),
		UIText_EventLogChannelInfo:        localization.GetString("UIText_EventLogChannelInfo"),
		UIText_StatusPanelChannel:         localization.GetString("UIText_StatusPanelChannel"),
		UIText_StatusPanelChannelInfo:     localization.GetString("UIText_StatusPanelChannelInfo"),
		UIText_LogChannel:                 localization.GetString("UIText_LogChannel"),
		UIText_LogChannelInfo:             localization.GetString("UIText_LogChannelInfo"),
		UIText_BannedPlayersListPath:      localization.GetString("UIText_BannedPlayersListPath"),
		UIText_BannedPlayersListPathInfo:  localization.GetString("UIText_BannedPlayersListPathInfo"),
		UIText_DiscordIntegrationBenefits: localization.GetString("UIText_DiscordIntegrationBenefits"),
		UIText_DiscordBenefit1:            localization.GetString("UIText_DiscordBenefit1"),
		UIText_DiscordBenefit2:            localization.GetString("UIText_DiscordBenefit2"),
		UIText_DiscordBenefit3:            localization.GetString("UIText_DiscordBenefit3"),
		UIText_DiscordBenefit4:            localization.GetString("UIText_DiscordBenefit4"),
		UIText_DiscordBenefit5:            localization.GetString("UIText_DiscordBenefit5"),
		UIText_DiscordSetupInstructions:   localization.GetString("UIText_DiscordSetupInstructions"),
		UIText_DiscordVotingTitle:         localization.GetString("UIText_DiscordVotingTitle"),
		UIText_DiscordVotingInfo:          localization.GetString("UIText_DiscordVotingInfo"),
		UIText_DiscordRestartVoteEnabled:  localization.GetString("UIText_DiscordRestartVoteEnabled"),
		UIText_DiscordRestoreVoteEnabled:  localization.GetString("UIText_DiscordRestoreVoteEnabled"),
		UIText_DiscordVoteDuration:        localization.GetString("UIText_DiscordVoteDuration"),
		UIText_DiscordVoteThreshold:       localization.GetString("UIText_DiscordVoteThreshold"),
		UIText_DiscordVoteMinimum:         localization.GetString("UIText_DiscordVoteMinimum"),
		UIText_DiscordVoteCooldown:        localization.GetString("UIText_DiscordVoteCooldown"),

		UIText_CopyrightConfig1: localization.GetString("UIText_Copyright1"),
		UIText_CopyrightConfig2: localization.GetString("UIText_Copyright2"),

		// SLP Section
		UIText_SLP_Title:                  localization.GetString("UIText_SLP_Title"),
		UIText_SLP_Description:            localization.GetString("UIText_SLP_Description"),
		UIText_SLP_ReadyToInstall:         localization.GetString("UIText_SLP_ReadyToInstall"),
		UIText_SLP_InstallButton:          localization.GetString("UIText_SLP_InstallButton"),
		UIText_SLP_UploadModPackage:       localization.GetString("UIText_SLP_UploadModPackage"),
		UIText_SLP_UploadDescription:      localization.GetString("UIText_SLP_UploadDescription"),
		UIText_SLP_UploadDescriptionLink:  localization.GetString("UIText_SLP_UploadDescriptionLink"),
		UIText_SLP_InstallFirst:           localization.GetString("UIText_SLP_InstallFirst"),
		UIText_SLP_InstallFirstSubtext:    localization.GetString("UIText_SLP_InstallFirstSubtext"),
		UIText_SLP_DragDropHere:           localization.GetString("UIText_SLP_DragDropHere"),
		UIText_SLP_OrClickToSelect:        localization.GetString("UIText_SLP_OrClickToSelect"),
		UIText_SLP_UploadButton:           localization.GetString("UIText_SLP_UploadButton"),
		UIText_SLP_ManageInstallation:     localization.GetString("UIText_SLP_ManageInstallation"),
		UIText_SLP_UninstallWarning:       localization.GetString("UIText_SLP_UninstallWarning"),
		UIText_SLP_UninstallButton:        localization.GetString("UIText_SLP_UninstallButton"),
		UIText_SLP_UpdateWorkshopMods:     localization.GetString("UIText_SLP_UpdateWorkshopMods"),
		UIText_SLP_UpdateWorkshopModsDesc: localization.GetString("UIText_SLP_UpdateWorkshopModsDesc"),
		UIText_SLP_UpdateButton:           localization.GetString("UIText_SLP_UpdateButton"),
		UIText_SLP_InstalledMods:          localization.GetString("UIText_SLP_InstalledMods"),

		IsStationeersLaunchPadEnabled: isStationeersLaunchPadEnabled,

		// Expert Settings
		ShowExpertSettings:              fmt.Sprintf("%v", config.GetShowExpertSettings()),
		ShowExpertSettingsTrueSelected:  showExpertSettingsTrueSelected,
		ShowExpertSettingsFalseSelected: showExpertSettingsFalseSelected,

		// Expert Settings values
		Debug:                                    fmt.Sprintf("%v", config.GetIsDebugMode()),
		DebugTrueSelected:                        debugTrueSelected,
		DebugFalseSelected:                       debugFalseSelected,
		LogLevel:                                 fmt.Sprintf("%d", config.GetLogLevel()),
		LogClutterToConsole:                      fmt.Sprintf("%v", config.GetLogClutterToConsole()),
		LogClutterToConsoleTrueSelected:          logClutterToConsoleTrueSelected,
		LogClutterToConsoleFalseSelected:         logClutterToConsoleFalseSelected,
		IsSSCMEnabled:                            fmt.Sprintf("%v", config.GetIsSSCMEnabled()),
		IsSSCMEnabledTrueSelected:                isSSCMEnabledTrueSelected,
		IsSSCMEnabledFalseSelected:               isSSCMEnabledFalseSelected,
		IsConsoleEnabled:                         fmt.Sprintf("%v", config.GetIsConsoleEnabled()),
		IsConsoleEnabledTrueSelected:             isConsoleEnabledTrueSelected,
		IsConsoleEnabledFalseSelected:            isConsoleEnabledFalseSelected,
		SSUIWebPort:                              config.GetSSUIWebPort(),
		IsUpdateEnabled:                          fmt.Sprintf("%v", config.GetIsUpdateEnabled()),
		IsUpdateEnabledTrueSelected:              isUpdateEnabledTrueSelected,
		IsUpdateEnabledFalseSelected:             isUpdateEnabledFalseSelected,
		AllowPrereleaseUpdates:                   fmt.Sprintf("%v", config.GetAllowPrereleaseUpdates()),
		AllowPrereleaseUpdatesTrueSelected:       allowPrereleaseUpdatesTrueSelected,
		AllowPrereleaseUpdatesFalseSelected:      allowPrereleaseUpdatesFalseSelected,
		AllowMajorUpdates:                        fmt.Sprintf("%v", config.GetAllowMajorUpdates()),
		AllowMajorUpdatesTrueSelected:            allowMajorUpdatesTrueSelected,
		AllowMajorUpdatesFalseSelected:           allowMajorUpdatesFalseSelected,
		AuthEnabled:                              fmt.Sprintf("%v", config.GetAuthEnabled()),
		AuthEnabledTrueSelected:                  authEnabledTrueSelected,
		AuthEnabledFalseSelected:                 authEnabledFalseSelected,
		AuthTokenLifetime:                        fmt.Sprintf("%d", config.GetAuthTokenLifetime()),
		DiscordCharBufferSize:                    fmt.Sprintf("%d", config.GetDiscordCharBufferSize()),
		AdvertiserOverride:                       config.GetAdvertiserOverride(),
		IsStationeersLaunchPadAutoUpdatesEnabled: fmt.Sprintf("%v", config.GetIsStationeersLaunchPadAutoUpdatesEnabled()),
		IsStationeersLaunchPadAutoUpdatesEnabledTrueSelected:  isStationeersLaunchPadAutoUpdatesEnabledTrueSelected,
		IsStationeersLaunchPadAutoUpdatesEnabledFalseSelected: isStationeersLaunchPadAutoUpdatesEnabledFalseSelected,
	}
	if !data.CanViewSettings {
		data.AdvertiserOverride = ""
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
