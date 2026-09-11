# Configuration

SSUI can be configured through the Web UI, `SSUI/config/config.json`, and environment variables. For normal installations, use the Web UI: it provides descriptions, validation, grouped settings, and fewer opportunities to invent a JSON syntax error at midnight.

## Configuration sources

SSUI resolves most values in this order:

1. Value loaded from `config.json`.
2. Matching environment variable when the JSON value is absent/empty.
3. Built-in default.

Startup flags override selected settings through configuration setters and can persist those values. See [Command-line Flags](Command-line-flags.md).

!!! important
    Existing JSON takes precedence over environment variables, and SSUI saves resolved values back to disk. Changing a container environment variable may therefore do nothing until the persisted setting is changed.

Booleans preserve explicit `false`; optional retention/vote integers preserve `0`. Ordinary integer fields treat `0` as absent. Empty strings fall through to environment/default; a present empty user map does not. `SSUI_USERS` expects comma-separated `username:hash` pairs.

Environment variables remain useful for containers and automated deployments, but ordinary installations should prefer the Web UI or JSON file.

## Applying changes

The Web UI saves configuration to disk. Depending on the component, a change may require one of:

- A configuration reload.
- An SSUI backend reload/restart.
- A Stationeers server restart.

Useful SSCLI commands:

```text
reloadconfig
reloadbackend
restartbackend
```

Do not edit `config.json` while SSUI is simultaneously writing settings. Stop SSUI or use the supported interface, then keep a copy before making large manual changes.

## Game and world settings

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `gameBranch` | `GAME_BRANCH` | `public` | Steam branch. Legacy terrain branches are rejected by current SSUI. |
| `ServerName` | `SERVER_NAME` | `Stationeers Server UI` | Name shown to players/server listings. |
| `SaveName` | `SAVE_NAME` | `MyMapName` | Save directory and active save name. |
| `WorldID` | `WORLD_ID` | `Lunar` | World/scenario identifier. |
| `ServerMaxPlayers` | `SERVER_MAX_PLAYERS` | `6` | Maximum simultaneous players. |
| `ServerPassword` | `SERVER_PASSWORD` | empty | Join password. Can be managed by Discord rotation. |
| `ServerAuthSecret` | `SERVER_AUTH_SECRET` | empty | Restricts access to the Stationeers server console. |
| `AdminPassword` | `ADMIN_PASSWORD` | empty | Legacy game administrator field; normally leave empty. |
| `Difficulty` | `DIFFICULTY` | empty | Stationeers difficulty identifier. |
| `StartCondition` | `START_CONDITION` | empty | World-specific start condition. |
| `StartLocation` | `START_LOCATION` | empty | World-specific starting location. |
| `AutoSave` | `AUTO_SAVE` | `true` | Enable Stationeers automatic saves. |
| `SaveInterval` | `SAVE_INTERVAL` | `300` | Seconds between automatic saves. |
| `AutoPauseServer` | `AUTO_PAUSE_SERVER` | `true` | Pause simulation while no players are connected. |
| `AdditionalParams` | `ADDITIONAL_PARAMS` | empty | Extra Stationeers launch arguments for advanced use. |

`SaveInfo` (`SAVE_INFO`) is a deprecated compatibility field. Older values are migrated into `SaveName` and `WorldID`; new configuration should use the separate fields.

The Web UI discovers available worlds, difficulties, start conditions, and start locations from the installed Stationeers XML files. The controls remain editable, so exact IDs from mods or a newer game build can still be entered manually. World-generation arguments are positional: `StartCondition` requires `Difficulty`, and `StartLocation` requires both preceding values. SSUI validates that sequence before saving instead of letting Stationeers misread a later value as an earlier one.

`IsNewTerrainAndSaveSystem` (`ENABLE_DOT_SAVES`) also remains in the JSON schema for compatibility, but current SSUI enforces it as `true` for supported game branches. It is not a useful user toggle anymore.

Current SSUI supports the modern save/terrain system. Legacy branches including `preterrain`, `prerocket`, `prephasechange`, and `multiplayersafe` require an older SSUI release that still supports their save layout.

## Network settings

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `GamePort` | `GAME_PORT` | `27016` | UDP port used by Stationeers clients. |
| `UpdatePort` | `UPDATE_PORT` | `27015` | Legacy/update port retained for game compatibility. |
| `LocalIpAddress` | `LOCAL_IP_ADDRESS` | `0.0.0.0` | Address Stationeers binds to. |
| `StartLocalHost` | `START_LOCAL_HOST` | `true` | Start with localhost/local-host behavior enabled. |
| `ServerVisible` | `SERVER_VISIBLE` | `true` | Advertise the server in supported listings. |
| `UseSteamP2P` | `USE_STEAM_P2P` | `false` | Enable Steam peer-to-peer networking. |
| `UPNPEnabled` | `UPNP_ENABLED` | `false` | Ask the network router to create port mappings automatically. |
| `AdvertiserOverride` | `ADVERTISER_OVERRIDE` | empty | Override the address advertised by SSUI; supports `auto`, IPv4, or hostname modes. |
| `ConnectivityCheckEnabled` | `CONNECTIVITY_CHECK_ENABLED` | `true` | Run the Spacecat gameserver reachability check while the Stationeers server is stopped. |

Manual router/firewall configuration is more predictable than UPNP. Each SSUI/game-server instance on the same host needs unique ports.

## SSUI operation

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `SSUIWebPort` | `SSUI_WEB_PORT` | `8443` | HTTPS port for Web UI and API. |
| `ExePath` | `EXE_PATH` | platform-specific | Stationeers dedicated-server executable path. |
| `IsSSCMEnabled` | `IS_SSCM_ENABLED` | `true` | Enable Stationeers Server Command Manager integration. |
| `AutoRestartServerTimer` | `AUTO_RESTART_SERVER_TIMER` | `0` | `0`, interval in minutes, or daily `HH:MM`/`HH:MMPM` time. |
| `AutoRestartCountdown` | `AUTO_RESTART_COUNTDOWN` | `65` | Seconds of in-game warning before scheduled restart. |
| `IsConsoleEnabled` | `IS_CONSOLE_ENABLED` | `true` | Enable interactive SSCLI console. |
| `LanguageSetting` | `LANGUAGE_SETTING` | `en-US` | Interface language/locale. |
| `AutoStartServerOnStartup` | `AUTO_START_SERVER_ON_STARTUP` | `false` | Start Stationeers after SSUI starts. |
| `ShowExpertSettings` | `SHOW_EXPERT_SETTINGS` | `false` | Show the separate **Expert Settings** tab in the Web UI. Turning it off hides the tab; it does not erase its values. |
| `LogClutterToConsole` | `LOG_CLUTTER_TO_CONSOLE` | `false` | Include normally suppressed noisy game/Mono output. |

`SSUIIdentifier` (`SSUI_IDENTIFIER`) and runtime `BackendUUID` are identifiers managed by SSUI. Do not copy them between unrelated installations unless you know why you need to.

## Updates

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `IsUpdateEnabled` | `IS_UPDATE_ENABLED` | `true` | Enable SSUI update checks (including manual check/apply calls) against GitHub. |
| `AllowPrereleaseUpdates` | `ALLOW_PRERELEASE_UPDATES` | `false` | Allow prerelease SSUI builds. |
| `AllowMajorUpdates` | `ALLOW_MAJOR_UPDATES` | `false` | Allow automatic movement to a new major SSUI version. |
| `AllowAutoGameServerUpdates` | `ALLOW_AUTO_GAME_SERVER_UPDATES` | `false` | Enable additional automatic game updates; startup still runs SteamCMD by default. |

Disabling `IsUpdateEnabled` stops automatic SSUI update checks. You must then track and apply SSUI releases manually.

The Web UI can display newer major and prerelease versions for deliberate manual installation even when their automatic flags are false. It labels those releases and requires separate confirmation for each risk. The flags decide what SSUI may choose without that interaction; they are not a filter that pretends other releases do not exist.

## Authentication and HTTPS

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `users` | `SSUI_USERS` | empty map | Legacy v5 users imported into the v6 identity store on first start. |
| `authEnabled` | `SSUI_AUTH_ENABLED` | legacy | Retained for config compatibility; v6 authentication is always enabled. |
| `JwtKey` | `SSUI_JWT_KEY` | legacy | Retained for config compatibility; v6 does not accept old JWTs. |
| `AuthTokenLifetime` | `SSUI_AUTH_TOKEN_LIFETIME` | legacy | Retained for config compatibility; v6 session lifetimes are managed by the identity system. |

TLS files normally live below `SSUI/tls/`. Use the custom-certificate interface instead of pasting private-key material into unrelated settings.

See [Security Considerations](Security-Considerations.md).

## Logging and diagnostics

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `Debug` | `DEBUG` | `false` | Enable debug/profiling behavior. Not for ordinary production use. |
| `CreateSSUILogFile` | `CREATE_SSUI_LOGFILE` | `false` | Write SSUI logs below `SSUI/logs`. |
| `CreateGameServerLogFile` | `CREATE_GAMESERVER_LOGFILE` | `false` | Keep a separate Stationeers log for each server run. |
| `LogLevel` | `LOG_LEVEL` | `20` | Minimum log severity: 10 debug, 20 info, 30 warning, 40 error. |
| `subsystemFilters` | `SUBSYSTEM_FILTERS` | empty | Limit output to selected SSUI subsystems. |

Support mode temporarily enables detailed logging and file output. See [Support Packages](Support-Packages.md).

## Backup settings

| JSON field | Environment variable | Default | Unit | Description |
|---|---|---:|---|---|
| `backupRetentionEnabled` | `BACKUP_RETENTION_ENABLED` | `false` | - | Enable scheduled cleanup of autosaves and safe backups. |
| `backupKeepNewestCount` | `BACKUP_KEEP_NEWEST_COUNT` | `2` | backups | Always retain this many newest safe backups. `0` disables this rule. |
| `backupDailyRetentionDays` | `BACKUP_DAILY_RETENTION_DAYS` | `7` | calendar days | Keep one representative backup for each covered local calendar day. `0` disables this rule. |
| `backupWeeklyRetentionWeeks` | `BACKUP_WEEKLY_RETENTION_WEEKS` | `4` | ISO weeks | Keep one representative backup for each covered ISO calendar week. `0` disables this rule. |
| `backupMonthlyRetentionMonths` | `BACKUP_MONTHLY_RETENTION_MONTHS` | `3` | calendar months | Keep one representative backup for each covered calendar month. `0` disables this rule. |
| `backupCleanupIntervalHours` | `BACKUP_CLEANUP_INTERVAL_HOURS` | `24` | hours | Time between cleanup runs; must produce a positive valid duration. |

Backup and safe-backup paths are derived from the active `SaveName` in current SSUI. See [Backup System](Backup-System.md) before enabling cleanup.

The polling/stability interval is fixed at 45 seconds in normal application configuration. Handling requires unchanged observations at least one interval apart and a complete valid archive; it is not a blind 45-second delay after creation. The former `isCleanupEnabled`, `backupKeepLastN`, `backupKeepDailyFor`, `backupKeepWeeklyFor`, `backupKeepMonthlyFor`, `backupCleanupInterval`, and `backupWaitTime` keys are intentionally not migrated. An upgraded installation therefore returns to cleanup disabled with the new conservative defaults. Review and explicitly enable the new policy.

Negative retention values and invalid/overflowing cleanup intervals loaded directly from JSON or environment variables are replaced with safe defaults and automatic cleanup is disabled. The Web UI/API rejects those values before saving.

## Discord settings

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `isDiscordEnabled` | `IS_DISCORD_ENABLED` | `false` | Enable the Discord bot. |
| `discordToken` | `DISCORD_TOKEN` | empty | Secret bot token. Never share it. |
| `discordAdminRoleID` | `DISCORD_ADMIN_ROLE_ID` | empty | Role required for all hub admin actions and slash commands except `/status`. Empty disables admin actions. |
| `eventLogChannelID` | `GAME_EVENT_LOG_CHANNEL_ID` | empty | Parsed game events and detections. |
| `statusPanelChannelID` | `STATUS_PANEL_CHANNEL_ID` | empty | Shared server hub, private admin flows and slash commands. |
| `logChannelID` | `LOG_CHANNEL_ID` | empty | Verbose log output. |
| `DiscordCharBufferSize` | `DISCORD_CHAR_BUFFER_SIZE` | `1000` | Maximum buffered text size used for Discord output. |
| `blackListFilePath` | `BLACKLIST_FILE_PATH` | `./Blacklist.txt` | Steam-ID ban-list file. |
| `rotateServerPassword` | `ROTATE_SERVER_PASSWORD` | `false` | Generate a new six-digit join code when Stationeers starts. |
| `discordRestartVoteEnabled` | `DISCORD_RESTART_VOTE_ENABLED` | `false` | Show restart voting in the member-facing status panel. |
| `discordRestoreVoteEnabled` | `DISCORD_RESTORE_VOTE_ENABLED` | `false` | Allow votes to restore one of the three most recent safe backups. |
| `discordVoteDurationMinutes` | `DISCORD_VOTE_DURATION_MINUTES` | `5` | Minutes before an unfinished vote expires. |
| `discordRestartVoteThreshold` | `DISCORD_RESTART_VOTE_THRESHOLD` | `60` | Percentage of currently connected players used for the restart-vote target. |
| `discordRestartVoteMinimum` | `DISCORD_RESTART_VOTE_MINIMUM` | `1` | Minimum Discord votes required for restart. |
| `discordRestartVoteCooldownMinutes` | `DISCORD_RESTART_VOTE_COOLDOWN_MINUTES` | `30` | Restart-vote cooldown after pass or expiry. |
| `discordRestoreVoteThreshold` | `DISCORD_RESTORE_VOTE_THRESHOLD` | `100` | Percentage of currently connected players used for the restore-vote target. |
| `discordRestoreVoteMinimum` | `DISCORD_RESTORE_VOTE_MINIMUM` | `2` | Minimum Discord votes required for restore. |
| `discordRestoreVoteCooldownMinutes` | `DISCORD_RESTORE_VOTE_COOLDOWN_MINUTES` | `60` | Restore-vote cooldown after pass or expiry. |

Deprecated Discord fields are migrated where possible:

- `connectionListChannelID` → `statusPanelChannelID`
- `statusChannelID` → `eventLogChannelID`
- `saveChannelID` → removed; save events now use `eventLogChannelID`

See [Discord Integration](Discord-Integration.md).

## StationeersLaunchPad settings

| JSON field | Environment variable | Default | Description |
|---|---|---:|---|
| `IsStationeersLaunchPadEnabled` | `IS_SLP_MODDING_ENABLED` | `false` | Enable the SLP server-mod workflow. |
| `IsStationeersLaunchPadAutoUpdatesEnabled` | `IS_SLP_MODDING_AUTO_UPDATES_ENABLED` | `true` | Set SLP's `CheckForUpdate` and `AutoUpdateOnStart` flags; not the Workshop batch-update switch. |

See [Modding with StationeersLaunchPad](Modding-with-StationeersLaunchPad.md).

## Example configuration

<details>
<summary>Partial JSON example for an already configured installation</summary>

This example intentionally omits generated identifiers and secrets. SSUI writes a complete normalized file after loading.

```json
{
  "gameBranch": "public",
  "GamePort": "27016",
  "ServerName": "My Extremely Sensible Server Name",
  "SaveName": "EuropaBase",
  "WorldID": "Europa",
  "ServerMaxPlayers": "6",
  "ServerPassword": "",
  "AutoSave": true,
  "SaveInterval": "300",
  "AutoPauseServer": true,
  "LocalIpAddress": "0.0.0.0",
  "ServerVisible": true,
  "UseSteamP2P": false,
  "SSUIWebPort": "8443",
  "AutoRestartServerTimer": "04:00",
  "AutoRestartCountdown": "65",
  "AutoStartServerOnStartup": true,
  "IsUpdateEnabled": true,
  "AllowPrereleaseUpdates": false,
  "AllowMajorUpdates": false,
  "AllowAutoGameServerUpdates": true,
  "backupRetentionEnabled": true,
  "backupKeepNewestCount": 2,
  "backupDailyRetentionDays": 7,
  "backupWeeklyRetentionWeeks": 4,
  "backupMonthlyRetentionMonths": 3,
  "backupCleanupIntervalHours": 24,
  "isDiscordEnabled": false,
  "IsStationeersLaunchPadEnabled": false
}
```

Retention values above are examples, not universal recommendations. Choose them based on save frequency, world size, available storage, and how far back you genuinely need to recover.

</details>

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
