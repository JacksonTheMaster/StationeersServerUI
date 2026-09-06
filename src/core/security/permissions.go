package security

const (
	PermissionServerView       = "server.view"
	PermissionServerControl    = "server.control"
	PermissionConsoleRead      = "console.read"
	PermissionConsoleWrite     = "console.write"
	PermissionBackupsView      = "backups.view"
	PermissionBackupsAnalyze   = "backups.analyze"
	PermissionBackupsDownload  = "backups.download"
	PermissionBackupsRestore   = "backups.restore"
	PermissionSettingsView     = "settings.view"
	PermissionSettingsManage   = "settings.manage"
	PermissionSteamCMDRun      = "steamcmd.run"
	PermissionSLPManage        = "slp.manage"
	PermissionDetectionsManage = "detections.manage"
	PermissionUpdateInstall    = "update.install"
	PermissionUsersManage      = "users.manage"
	PermissionGroupsManage     = "groups.manage"
	PermissionTokensManage     = "tokens.manage"
	PermissionAuditView        = "audit.view"
	PermissionBackendReload    = "backend.reload"
	PermissionSecurityManage   = "security.manage"
)

var AllPermissions = []string{
	PermissionServerView,
	PermissionServerControl,
	PermissionConsoleRead,
	PermissionConsoleWrite,
	PermissionBackupsView,
	PermissionBackupsAnalyze,
	PermissionBackupsDownload,
	PermissionBackupsRestore,
	PermissionSettingsView,
	PermissionSettingsManage,
	PermissionSteamCMDRun,
	PermissionSLPManage,
	PermissionDetectionsManage,
	PermissionUpdateInstall,
	PermissionUsersManage,
	PermissionGroupsManage,
	PermissionTokensManage,
	PermissionAuditView,
	PermissionBackendReload,
	PermissionSecurityManage,
}

func IsPermission(value string) bool {
	for _, permission := range AllPermissions {
		if value == permission {
			return true
		}
	}
	return false
}
