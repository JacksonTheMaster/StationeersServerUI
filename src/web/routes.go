package web

import (
	"io/fs"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config/configchanger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	assets, _ := fs.Sub(config.GetV1UIFS(), "SSUI/onboard_bundled/assets")
	twoBoxAssets, _ := fs.Sub(config.GetV1UIFS(), "SSUI/onboard_bundled/twoboxform")
	svelteAssets, _ := fs.Sub(config.GetV1UIFS(), "SSUI/onboard_bundled/v2/assets")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
	mux.Handle("/twoboxform/", http.StripPrefix("/twoboxform/", http.FileServer(http.FS(twoBoxAssets))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(svelteAssets))))
	mux.HandleFunc("GET /login", ServeTwoBoxFormTemplate)
	mux.HandleFunc("GET /setup", setupPage)

	mux.HandleFunc("GET /api/v3/auth/setup", api.SetupStatusHandler)
	mux.HandleFunc("POST /api/v3/auth/setup/bootstrap", api.BootstrapOwnerHandler)
	mux.HandleFunc("POST /api/v3/auth/login", api.LoginHandler)
	mux.HandleFunc("POST /api/v3/auth/logout", api.LogoutHandler)
	mux.HandleFunc("POST /api/v3/setup/settings", setupOnly(api.JSONBoundary(configchanger.SaveConfigRestful)))

	pages := http.NewServeMux()
	pages.HandleFunc("GET /", ServeIndex)
	pages.HandleFunc("GET /config", ServeConfigPage)
	pages.HandleFunc("GET /backups", ServeBackupPage)
	pages.HandleFunc("GET /detectionmanager", requirePage([]string{security.PermissionDetectionsManage}, "You don't have permission to manage custom detections.", ServeDetectionManager))
	pages.HandleFunc("GET /changeuser", requirePage([]string{security.PermissionUsersManage}, "You don't have permission to manage users.", ServeTwoBoxFormTemplate))
	pages.HandleFunc("GET /app", ServeSvelteUI)
	mux.Handle("/", api.PageIdentityMiddleware(pages))

	v3 := http.NewServeMux()
	registerIdentityRoutes(v3)
	registerServerRoutes(v3)
	registerBackupRoutes(v3)
	registerSettingsRoutes(v3)
	registerModdingRoutes(v3)
	protectedAPI := api.IdentityMiddleware(v3)
	mux.Handle("/api/v3", protectedAPI)
	mux.Handle("/api/v3/", protectedAPI)

	return mux
}

func registerIdentityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v3", api.CapabilitiesHandler)
	mux.HandleFunc("GET /api/v3/auth/session", api.SessionInfoHandler)
	mux.HandleFunc("POST /api/v3/auth/password", api.ChangeOwnPasswordHandler)
	mux.HandleFunc("GET /api/v3/auth/users", api.Require(security.PermissionUsersManage, api.UsersHandler))
	mux.HandleFunc("POST /api/v3/auth/users", api.Require(security.PermissionUsersManage, api.CreateUserHandler))
	mux.HandleFunc("PATCH /api/v3/auth/users/{id}", api.Require(security.PermissionUsersManage, api.UpdateUserHandler))
	mux.HandleFunc("DELETE /api/v3/auth/users/{id}", api.Require(security.PermissionUsersManage, api.DeleteUserHandler))
	mux.HandleFunc("GET /api/v3/auth/groups", api.RequireAny([]string{security.PermissionGroupsManage, security.PermissionUsersManage}, api.GroupsHandler))
	mux.HandleFunc("POST /api/v3/auth/groups", api.Require(security.PermissionGroupsManage, api.CreateGroupHandler))
	mux.HandleFunc("PUT /api/v3/auth/groups/{id}", api.Require(security.PermissionGroupsManage, api.UpdateGroupHandler))
	mux.HandleFunc("DELETE /api/v3/auth/groups/{id}", api.Require(security.PermissionGroupsManage, api.DeleteGroupHandler))
	mux.HandleFunc("GET /api/v3/auth/tokens", api.Require(security.PermissionTokensManage, api.TokensHandler))
	mux.HandleFunc("POST /api/v3/auth/tokens", api.Require(security.PermissionTokensManage, api.CreateTokenHandler))
	mux.HandleFunc("DELETE /api/v3/auth/tokens/{id}", api.Require(security.PermissionTokensManage, api.DeleteTokenHandler))
	mux.HandleFunc("GET /api/v3/auth/sessions", api.SessionsHandler)
	mux.HandleFunc("DELETE /api/v3/auth/sessions/{id}", api.DeleteSessionHandler)
	mux.HandleFunc("GET /api/v3/auth/audit", api.Require(security.PermissionAuditView, api.AuditHandler))
}

func registerServerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v3/server/start", api.Require(security.PermissionServerControl, api.JSONBoundary(StartServer)))
	mux.HandleFunc("POST /api/v3/server/stop", api.Require(security.PermissionServerControl, api.JSONBoundary(StopServer)))
	mux.HandleFunc("GET /api/v3/server/status", api.Require(security.PermissionServerView, api.JSONBoundary(GetGameServerRunState)))
	mux.HandleFunc("GET /api/v3/server/players", api.Require(security.PermissionServerView, api.JSONBoundary(HandleConnectedPlayersList)))
	mux.HandleFunc("GET /api/v3/monitor/status", api.Require(security.PermissionServerView, api.JSONBoundary(HandleMonitorStatus)))
	mux.HandleFunc("POST /api/v3/backend/reload", api.Require(security.PermissionBackendReload, api.JSONBoundary(HandleReloadAll)))
	mux.HandleFunc("POST /api/v3/steamcmd/run", api.Require(security.PermissionSteamCMDRun, api.JSONBoundary(HandleRunSteamCMD)))
	mux.HandleFunc("GET /api/v3/sscm/status", api.Require(security.PermissionConsoleRead, api.JSONBoundary(HandleIsSSCMEnabled)))
	mux.HandleFunc("POST /api/v3/sscm/commands", api.Require(security.PermissionConsoleWrite, api.JSONBoundary(HandleCommand)))
	mux.HandleFunc("GET /api/v3/streams/console", api.Require(security.PermissionConsoleRead, GetLogOutput))
	mux.HandleFunc("GET /api/v3/streams/events", api.Require(security.PermissionServerView, GetEventOutput))
	mux.HandleFunc("GET /api/v3/streams/logs/debug", api.Require(security.PermissionConsoleRead, GetDebugLogOutput))
	mux.HandleFunc("GET /api/v3/streams/logs/info", api.Require(security.PermissionConsoleRead, GetInfoLogOutput))
	mux.HandleFunc("GET /api/v3/streams/logs/warn", api.Require(security.PermissionConsoleRead, GetWarnLogOutput))
	mux.HandleFunc("GET /api/v3/streams/logs/error", api.Require(security.PermissionConsoleRead, GetErrorLogOutput))
	mux.HandleFunc("GET /api/v3/streams/logs/backend", api.Require(security.PermissionConsoleRead, GetBackendLogOutput))
}

func registerBackupRoutes(mux *http.ServeMux) {
	handler := backupmgr.NewHTTPHandler(backupmgr.CurrentBackupManager())
	mux.HandleFunc("GET /api/v3/backups", api.Require(security.PermissionBackupsView, api.JSONBoundary(handler.ListBackupsHandler)))
	mux.HandleFunc("GET /api/v3/backups/analysis", api.Require(security.PermissionBackupsAnalyze, api.JSONBoundary(handler.AnalyzeBackupHandler)))
	mux.HandleFunc("POST /api/v3/backups/restore", api.Require(security.PermissionBackupsRestore, api.JSONBoundary(handler.RestoreBackupHandler)))
	mux.HandleFunc("POST /api/v3/backups/download", api.Require(security.PermissionBackupsDownload, handler.DownloadBackupHandler))
}

func registerSettingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v3/settings", api.Require(security.PermissionSettingsManage, api.JSONBoundary(configchanger.SaveConfigRestful)))
	mux.HandleFunc("GET /api/v3/worldgen/catalog", api.Require(security.PermissionSettingsView, api.JSONBoundary(HandleWorldGenerationCatalog)))
	mux.HandleFunc("POST /api/v3/advertiser/override", api.Require(security.PermissionSettingsManage, api.JSONBoundary(SaveAdvertiserOverrideHandler)))
	mux.HandleFunc("POST /api/v3/tls/certificate", api.Require(security.PermissionSecurityManage, api.JSONBoundary(SaveTLSCertificateHandler)))
	mux.HandleFunc("GET /api/v3/update", api.Require(security.PermissionServerView, api.JSONBoundary(CheckUpdateHandler)))
	mux.HandleFunc("POST /api/v3/update", api.Require(security.PermissionUpdateInstall, api.JSONBoundary(TriggerUpdateHandler)))
	mux.HandleFunc("GET /api/v3/detections", api.Require(security.PermissionDetectionsManage, api.JSONBoundary(detectionmgr.HandleCustomDetection)))
	mux.HandleFunc("POST /api/v3/detections", api.Require(security.PermissionDetectionsManage, api.JSONBoundary(detectionmgr.HandleCustomDetection)))
	mux.HandleFunc("DELETE /api/v3/detections", api.Require(security.PermissionDetectionsManage, api.JSONBoundary(detectionmgr.HandleDeleteCustomDetection)))
	mux.HandleFunc("POST /api/v3/setup/finalize", api.Require(security.PermissionSettingsManage, api.JSONBoundary(SetupFinalizeHandler)))
}

func registerModdingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v3/slp/install", api.Require(security.PermissionSLPManage, api.JSONBoundary(InstallSLPHandler)))
	mux.HandleFunc("POST /api/v3/slp/uninstall", api.Require(security.PermissionSLPManage, api.JSONBoundary(UninstallSLPHandler)))
	mux.HandleFunc("POST /api/v3/slp/reinstall", api.Require(security.PermissionSLPManage, api.JSONBoundary(ReinstallSLPHandler)))
	mux.HandleFunc("POST /api/v3/slp/packages", api.Require(security.PermissionSLPManage, api.JSONBoundary(UploadModPackageHandler)))
	mux.HandleFunc("GET /api/v3/slp/mods", api.Require(security.PermissionSLPManage, api.JSONBoundary(GetInstalledModDetailsHandler)))
	mux.HandleFunc("POST /api/v3/slp/mods/update", api.Require(security.PermissionSLPManage, api.JSONBoundary(UpdateWorkshopModsHandler)))
	mux.HandleFunc("POST /api/v3/slp/mods", api.Require(security.PermissionSLPManage, api.JSONBoundary(UpdateSingleWorkshopModHandler)))
}

func setupOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !security.SetupRequired() {
			api.WriteError(w, http.StatusNotFound, "not_found", "Setup is already complete")
			return
		}
		next(w, r)
	}
}

func setupPage(w http.ResponseWriter, r *http.Request) {
	if security.SetupRequired() {
		ServeTwoBoxFormTemplate(w, r)
		return
	}
	page := requirePage(
		[]string{security.PermissionSettingsView, security.PermissionSettingsManage},
		"You don't have permission to view server settings.",
		ServeTwoBoxFormTemplate,
	)
	api.PageIdentityMiddleware(http.HandlerFunc(page)).ServeHTTP(w, r)
}
