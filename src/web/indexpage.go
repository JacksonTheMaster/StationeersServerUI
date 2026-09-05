package web

import (
	"io/fs"
	"net/http"
	"text/template"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/localization"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	htmlFS, err := fs.Sub(config.V1UIFS, "UIMod/onboard_bundled/ui")
	if err != nil {
		http.Error(w, "Error accessing Virt FS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(htmlFS, "index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Core.Error("failed to serve v1 Index.html")
		return
	}

	var Identifier string

	if config.SSUIIdentifier == "" {
		Identifier = " (" + config.GetBranch() + ")"
	} else {
		Identifier = ": " + config.GetSSUIIdentifier()
	}

	data := IndexTemplateData{
		UIText_UpdateAvailable:         localization.GetString("UIText_UpdateAvailable"),
		UIText_UpdateLater:             localization.GetString("UIText_UpdateLater"),
		UIText_UpdateNow:               localization.GetString("UIText_UpdateNow"),
		UIText_UpdateInstalling:        localization.GetString("UIText_UpdateInstalling"),
		UIText_UpdateFailed:            localization.GetString("UIText_UpdateFailed"),
		UIText_UpdateDescription:       localization.GetString("UIText_UpdateDescription"),
		UIText_MajorUpdateTitle:        localization.GetString("UIText_MajorUpdateTitle"),
		UIText_MajorUpdateWarning:      localization.GetString("UIText_MajorUpdateWarning"),
		UIText_MajorUpdateConfirm:      localization.GetString("UIText_MajorUpdateConfirm"),
		UIText_UpdateReleaseNotes:      localization.GetString("UIText_UpdateReleaseNotes"),
		UIText_UpdateRetry:             localization.GetString("UIText_UpdateRetry"),
		Version:                        config.GetVersion(),
		Branch:                         config.GetBranch(),
		SSUIIdentifier:                 Identifier,
		UIText_StartButton:             localization.GetString("UIText_StartButton"),
		UIText_StopButton:              localization.GetString("UIText_StopButton"),
		UIText_Settings:                localization.GetString("UIText_Settings"),
		UIText_Update_SteamCMD:         localization.GetString("UIText_Update_SteamCMD"),
		UIText_Console:                 localization.GetString("UIText_Console"),
		UIText_Detection_Events:        localization.GetString("UIText_Detection_Events"),
		UIText_Backend_Log:             localization.GetString("UIText_Backend_Log"),
		UIText_Backup_Manager:          localization.GetString("UIText_Backup_Manager"),
		UIText_Connected_PlayersHeader: localization.GetString("UIText_Connected_PlayersHeader"),
		UIText_GameServer:              localization.GetString("UIText_GameServer"),
		UIText_Uptime:                  localization.GetString("UIText_Uptime"),
		UIText_PlayersOnline:           localization.GetString("UIText_PlayersOnline"),
		UIText_LatestBackup:            localization.GetString("UIText_LatestBackup"),
		UIText_ServerOutput:            localization.GetString("UIText_ServerOutput"),
		UIText_ServerActivity:          localization.GetString("UIText_ServerActivity"),
		UIText_StreamConnected:         localization.GetString("UIText_StreamConnected"),
		UIText_StreamConnecting:        localization.GetString("UIText_StreamConnecting"),
		UIText_StreamReconnecting:      localization.GetString("UIText_StreamReconnecting"),
		UIText_StreamPaused:            localization.GetString("UIText_StreamPaused"),
		UIText_Pause:                   localization.GetString("UIText_Pause"),
		UIText_Resume:                  localization.GetString("UIText_Resume"),
		UIText_Clear:                   localization.GetString("UIText_Clear"),
		UIText_NoPlayers:               localization.GetString("UIText_NoPlayers"),
		UIText_PlayersUnavailable:      localization.GetString("UIText_PlayersUnavailable"),
		UIText_RecentBackups:           localization.GetString("UIText_RecentBackups"),
		UIText_BackupHistory:           localization.GetString("UIText_BackupHistory"),
		UIText_ViewAllBackups:          localization.GetString("UIText_ViewAllBackups"),
		UIText_ShowRecentBackups:       localization.GetString("UIText_ShowRecentBackups"),
		UIText_Refresh:                 localization.GetString("UIText_Refresh"),
		UIText_Close:                   localization.GetString("UIText_Close"),
		UIText_StateUncertain:          localization.GetString("UIText_StateUncertain"),
		UIText_StateStopped:            localization.GetString("UIText_StateStopped"),
		UIText_StateStarting:           localization.GetString("UIText_StateStarting"),
		UIText_StateLoadingMap:         localization.GetString("UIText_StateLoadingMap"),
		UIText_StateHostingSession:     localization.GetString("UIText_StateHostingSession"),
		UIText_StateRunning:            localization.GetString("UIText_StateRunning"),
		UIText_StateStopping:           localization.GetString("UIText_StateStopping"),
		UIText_Discord_Info:            localization.GetString("UIText_Discord_Info"),
		UIText_API_Info:                localization.GetString("UIText_API_Info"),
		UIText_Copyright1:              localization.GetString("UIText_Copyright1"),
		UIText_Copyright2:              localization.GetString("UIText_Copyright2"),
	}
	if data.Version == "" {
		data.Version = "unknown"
	}
	if data.Branch == "" {
		data.Branch = "unknown"
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
