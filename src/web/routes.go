package web

import (
	"io/fs"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config/configchanger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
)

func SetupRoutes() (*http.ServeMux, *http.ServeMux) {

	// Set up handlers with auth middleware
	mux := http.NewServeMux() // Use a mux to apply middleware globally

	// Unprotected auth routes
	twoboxformAssetsFS, _ := fs.Sub(config.GetV1UIFS(), "UIMod/onboard_bundled/twoboxform")
	mux.Handle("/twoboxform/", http.StripPrefix("/twoboxform/", http.FileServer(http.FS(twoboxformAssetsFS))))
	mux.HandleFunc("/auth/login", LoginHandler) // Token issuer
	mux.HandleFunc("/auth/logout", LogoutHandler)
	mux.HandleFunc("/login", ServeTwoBoxFormTemplate)

	// Protected routes (wrapped with middleware)
	protectedMux := http.NewServeMux()

	legacyAssetsFS, _ := fs.Sub(config.GetV1UIFS(), "UIMod/onboard_bundled/assets")
	protectedMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(legacyAssetsFS))))

	protectedMux.HandleFunc("/config", ServeConfigPage)
	protectedMux.HandleFunc("/backups", ServeBackupPage)
	protectedMux.HandleFunc("/detectionmanager", ServeDetectionManager)
	protectedMux.HandleFunc("/", ServeIndex)

	// --- SVELTE UI ---
	protectedMux.HandleFunc("/v2", ServeSvelteUI)
	svelteAssetsFS, _ := fs.Sub(config.V1UIFS, "UIMod/onboard_bundled/v2/assets")
	protectedMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(svelteAssetsFS))))
	protectedMux.HandleFunc("/api/v2/loader/reloadbackend", HandleReloadAll)

	// SSE routes
	protectedMux.HandleFunc("/console", GetLogOutput)
	protectedMux.HandleFunc("/events", GetEventOutput)
	protectedMux.HandleFunc("/logs/debug", GetDebugLogOutput)
	protectedMux.HandleFunc("/logs/info", GetInfoLogOutput)
	protectedMux.HandleFunc("/logs/warn", GetWarnLogOutput)
	protectedMux.HandleFunc("/logs/error", GetErrorLogOutput)
	protectedMux.HandleFunc("/logs/backend", GetBackendLogOutput)

	// Server Control
	protectedMux.HandleFunc("/start", StartServer)
	protectedMux.HandleFunc("/stop", StopServer)
	protectedMux.HandleFunc("/api/v2/server/start", StartServer)
	protectedMux.HandleFunc("/api/v2/server/stop", StopServer)
	protectedMux.HandleFunc("/api/v2/server/status", GetGameServerRunState)
	protectedMux.HandleFunc("/api/v2/server/status/connectedplayers", HandleConnectedPlayersList)

	backupHandler := backupmgr.NewHTTPHandler(backupmgr.GlobalBackupManager)
	protectedMux.HandleFunc("/api/v2/backups", backupHandler.ListBackupsHandler)
	protectedMux.HandleFunc("/api/v2/backups/analyze", backupHandler.AnalyzeBackupHandler)
	protectedMux.HandleFunc("/api/v2/backups/restore", backupHandler.RestoreBackupHandler)
	protectedMux.HandleFunc("/api/v2/backups/download", backupHandler.DownloadBackupHandler)

	// Configuration
	protectedMux.HandleFunc("/saveconfigasjson", configchanger.SaveConfigForm)     // legacy, used on config page
	protectedMux.HandleFunc("/api/v2/saveconfig", configchanger.SaveConfigRestful) // used on twoboxform
	protectedMux.HandleFunc("/api/v2/worldgen/catalog", HandleWorldGenerationCatalog)
	protectedMux.HandleFunc("/api/v2/advertiser/override", SaveAdvertiserOverrideHandler)
	protectedMux.HandleFunc("/api/v2/tls/certificate", SaveTLSCertificateHandler)
	protectedMux.HandleFunc("/api/v2/SSCM/run", HandleCommand)           // Command execution via SSCM (needs to be enable, config.IsSSCMEnabled)
	protectedMux.HandleFunc("/api/v2/SSCM/enabled", HandleIsSSCMEnabled) // Check if SSCM is enabled
	protectedMux.HandleFunc("/api/v2/steamcmd/run", HandleRunSteamCMD)   // Run SteamCMD
	// /api/v2/steamcmd/updatemods is defined in the SLP & Modding section below

	// Custom Detections
	protectedMux.HandleFunc("/api/v2/custom-detections", detectionmgr.HandleCustomDetection)
	protectedMux.HandleFunc("/api/v2/custom-detections/delete/", detectionmgr.HandleDeleteCustomDetection)
	// Authentication
	protectedMux.HandleFunc("/changeuser", ServeTwoBoxFormTemplate)
	protectedMux.HandleFunc("/api/v2/auth/adduser", RegisterUserHandler) // user registration and change password
	protectedMux.HandleFunc("/api/v2/auth/whoami", WhoAmIHandler)

	// Setup
	protectedMux.HandleFunc("/setup", ServeTwoBoxFormTemplate)
	protectedMux.HandleFunc("/api/v2/auth/setup/register", RegisterUserHandler) // user registration
	protectedMux.HandleFunc("/api/v2/auth/setup/apikey", RegisterAPIKeyHandler) // API Key registration
	protectedMux.HandleFunc("/api/v2/auth/setup/finalize", SetupFinalizeHandler)

	// Update
	protectedMux.HandleFunc("/api/v2/update/trigger", TriggerUpdateHandler)
	protectedMux.HandleFunc("/api/v2/update/check", CheckUpdateHandler)

	// Monitoring
	protectedMux.HandleFunc("/api/v2/monitor/gameserver/status", HandleMonitorStatus)

	// SLP & Modding
	protectedMux.HandleFunc("/api/v2/slp/install", InstallSLPHandler)
	protectedMux.HandleFunc("/api/v2/slp/uninstall", UninstallSLPHandler)
	protectedMux.HandleFunc("/api/v2/slp/reinstall", ReinstallSLPHandler)
	protectedMux.HandleFunc("/api/v2/slp/upload", UploadModPackageHandler)
	protectedMux.HandleFunc("/api/v2/slp/mods", GetInstalledModDetailsHandler)
	protectedMux.HandleFunc("/api/v2/steamcmd/updatemods", UpdateWorkshopModsHandler)
	protectedMux.HandleFunc("/api/v2/steamcmd/updatemod", UpdateSingleWorkshopModHandler)

	return mux, protectedMux
}
