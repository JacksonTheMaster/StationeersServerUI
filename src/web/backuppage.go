package web

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/localization"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

type backupPageTemplateData struct {
	Version                        string
	SSUIIdentifier                 string
	UITextBackToDashboard          string
	UITextBackupManager            string
	UITextBackupManagerDescription string
	UITextBackupHistory            string
	UITextBackupCreated            string
	UITextBackupDaysPlayed         string
	UITextBackupThings             string
	UITextBackupAtmospheres        string
	UITextBackupRooms              string
	UITextBackupPipeNetworks       string
	UITextBackupCableNetworks      string
	UITextBackupPlayers            string
	UITextBackupPlayersAlive       string
	UITextBackupPlayersUnconscious string
	UITextBackupFurnaces           string
	UITextBackupDestroyedFurnaces  string
	UITextBackupExpand             string
	UITextBackupCollapse           string
	UITextBackupAnalysisLoading    string
	UITextBackupAnalysisFailed     string
	UITextBackupRetry              string
	UITextBackupGameVersion        string
	UITextBackupArchiveSize        string
	UITextBackupDownload           string
	UITextBackupRestore            string
	UITextBackupTechnicalDetails   string
	UITextBackupInfrastructure     string
	UITextBackupLast               string
	UITextBackupAll                string
	UITextBackupLoading            string
	UITextBackupNone               string
	UITextBackupDownloadFailed     string
	UITextRefresh                  string
}

func ServeBackupPage(w http.ResponseWriter, r *http.Request) {
	htmlFS, err := fs.Sub(config.V1UIFS, "UIMod/onboard_bundled/ui")
	if err != nil {
		http.Error(w, "Error accessing Virt FS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(htmlFS, "backups.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Core.Error("failed to serve backups.html")
		return
	}

	identifier := " (" + config.GetBranch() + ")"
	if config.GetSSUIIdentifier() != "" {
		identifier = ": " + config.GetSSUIIdentifier()
	}

	data := backupPageTemplateData{
		Version:                        config.GetVersion(),
		SSUIIdentifier:                 identifier,
		UITextBackToDashboard:          localization.GetString("UIText_BackToDashboard"),
		UITextBackupManager:            localization.GetString("UIText_Backup_Manager"),
		UITextBackupManagerDescription: localization.GetString("UIText_BackupManagerDescription"),
		UITextBackupHistory:            localization.GetString("UIText_BackupHistory"),
		UITextBackupCreated:            localization.GetString("UIText_BackupCreated"),
		UITextBackupDaysPlayed:         localization.GetString("UIText_BackupDaysPlayed"),
		UITextBackupThings:             localization.GetString("UIText_BackupThings"),
		UITextBackupAtmospheres:        localization.GetString("UIText_BackupAtmospheres"),
		UITextBackupRooms:              localization.GetString("UIText_BackupRooms"),
		UITextBackupPipeNetworks:       localization.GetString("UIText_BackupPipeNetworks"),
		UITextBackupCableNetworks:      localization.GetString("UIText_BackupCableNetworks"),
		UITextBackupPlayers:            localization.GetString("UIText_BackupPlayers"),
		UITextBackupPlayersAlive:       localization.GetString("UIText_BackupPlayersAlive"),
		UITextBackupPlayersUnconscious: localization.GetString("UIText_BackupPlayersUnconscious"),
		UITextBackupFurnaces:           localization.GetString("UIText_BackupFurnaces"),
		UITextBackupDestroyedFurnaces:  localization.GetString("UIText_BackupDestroyedFurnaces"),
		UITextBackupExpand:             localization.GetString("UIText_BackupExpand"),
		UITextBackupCollapse:           localization.GetString("UIText_BackupCollapse"),
		UITextBackupAnalysisLoading:    localization.GetString("UIText_BackupAnalysisLoading"),
		UITextBackupAnalysisFailed:     localization.GetString("UIText_BackupAnalysisFailed"),
		UITextBackupRetry:              localization.GetString("UIText_BackupRetry"),
		UITextBackupGameVersion:        localization.GetString("UIText_BackupGameVersion"),
		UITextBackupArchiveSize:        localization.GetString("UIText_BackupArchiveSize"),
		UITextBackupDownload:           localization.GetString("UIText_BackupDownload"),
		UITextBackupRestore:            localization.GetString("UIText_BackupRestore"),
		UITextBackupTechnicalDetails:   localization.GetString("UIText_BackupTechnicalDetails"),
		UITextBackupInfrastructure:     localization.GetString("UIText_BackupInfrastructure"),
		UITextBackupLast:               localization.GetString("UIText_BackupLast"),
		UITextBackupAll:                localization.GetString("UIText_BackupAll"),
		UITextBackupLoading:            localization.GetString("UIText_BackupLoading"),
		UITextBackupNone:               localization.GetString("UIText_BackupNone"),
		UITextBackupDownloadFailed:     localization.GetString("UIText_BackupDownloadFailed"),
		UITextRefresh:                  localization.GetString("UIText_Refresh"),
	}

	if data.Version == "" {
		data.Version = "unknown"
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
