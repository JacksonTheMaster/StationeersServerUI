!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

This document explains the various configuration options available for Stationeers Server UI. Configuration can be handled through different methods, applied in the following order of preference:

1. Web interface (highest priority)
2. `config.json` file (middle priority)
3. Default values (lowest priority)
4. Environment variables (deprecated)


### Configuring the Interface & Gameserver

SSUI server's user web interface (UI) provides an intuitive way to adjust settings without directly editing the configuration file. To access it:

1. Launch the server application.
2. Locate and click the **Edit Config** button on the main screen.
3. A new page will open with **Server Configuration** tab selected. You'll have access to the **Configuration Wizard** button to re-run the wizard and the **Basic Settings**, **World Generation**, **Network Settings**, and **Advacned Features** buttons, explained below.

For detailed explanations of each parameter, refer to the **Recommended Server Parameters** section below. If you're unsure about a setting, the UI includes tooltips or descriptions for most fields.

## Recommended Server Parameters
### WebUI

For a typical multiplayer server, consider the following parameters:

**Basic Settings**
- `Server Name`: Set your server's name.
- `Save Name`: Your world save name. You can call it what ever you want, but its best to call it something related to the map or scenario.
- `Max Players`: Set the maximum number of concurrent players. Depending on your server's hardware and internet connection, the maximum players your server can support will vary. If you're unsure, start with 4 and allow more if all is going smoothly.
- `Server Password`: Set a password if you don't want an open, publicly accessable server. If using the Rotating Server Password feature you can leave this empty or put in a temporary password. The Rotating Server Password system will over write it.
- `Server Auth Secret`: Put a different password here to restrict console use.
- `Auto Save`: Set to TRUE to enable automatic saving. There is only a few edge cases where you'd want this not set to TRUE.
- `Save Interval`: Set the interval between automatic saves, e.g. 300 seconds. The save files are generally fairly small, 300 seconds is a reasonable balance between save frequency and the number of files. It is common for a well played map to reach 5000+ save files if left unmanaged.
- `Auto Pause Server`: Set to TRUE to automatically pause the server when no players are connected. There is only a few edge cases where you'd want this not set to TRUE.

**World Generation**
- `World ID`: This is the world to use. The name is specific to the map, check the [Stationeers wiki](https://stationeers-wiki.com/Dedicated_Server_Guide) for world names. Mod worlds will have their custom name, check with the mod author.
- `Difficulty`: This sets the game difficulty, with the options of: Creative, Easy, Normal, Stationeer. Capitalization is important, SSUI will provide feedback.
- `Start Condition`: This sets the starting difficulty. It is world dependent, check the [Stationeers wiki](https://stationeers-wiki.com/Dedicated_Server_Guide) for the correct names. SSUI will provide feedback. Mod worlds may have their own custom start conditions, check with the mod author.
- `Start Location`: This is a world specific location, check the [Stationeers wiki](https://stationeers-wiki.com/Dedicated_Server_Guide) for vanilla worlds. For mod worlds, check with the mod author.

**Network Settings**
- `Game Port`: This is the port your clients will connect to, the default is `27016`. If you are deploying multiple servers, each one will need a different port.
- `Update Port`: This is the Steam Update port, default is `27015`. 
- `UPNP Enabled`: Setting this to TRUE will enable UPNP automatic port forwarding. Note: It is recommended to set this to FALSE and manually set up port forwarding and matching firewall rules.
- `Local IP Address`: The bind address for the server to use. Usually setting to 0.0.0.0 is what you want.
- `Server Visibile`: Setting this to TRUE will advertise your server on the in-game server list.
- `Use Steam P2P`: This setting is to allow clients to connect through the SteamP2P system, it is recommended to set this to FALSE.

**Advacned Features**
- `Scheduled Gameserver Restart`: This is a multi-format setting to enable the server to perform automatic restarts. `0` = Disabled. For an absolute time of day restart, use a time in either 12 or 24 hour format (e.g. `10:26PM` or `22:26`). For a releative restart schedule, put a value in minutes. `720` would restart the server every 12 hours. Note: the relative restart time will be from when the config was last saved or from the last server restart. This has the effect of the restart trigger time creeping.
- `Auto Start Server on Startup`: Setting this to TRUE will start the game server once SSUI has started.
- `Enable Auto Game Server Updates`: Setting this to TRUE will query for new updates and install them automatically. Note: the game server will restart to run the new version. The server will use the `say` command to inform players of the impending restart. 
!!! tip
    The default in-game chat messages are hard to see, for QoL consider the [BetterChatUI](https://steamcommunity.com/sharedfiles/filedetails/?id=3651335734) mod.
- `Create SSUI Log Files`: Default FALSE. Setting this to TRUE will create log files in `/UIMod/logs`.
- `Create Game Server Log File`: Default FALSE. Setting this to TRUE will create a new log file for each server start.
- `Game Branch`: Set the game branch you want to use. Usually this is `public`, but other options are available, e.g. `beta` or `preterrain`. 
- `Server Executable Path`: Shows the executable path. Its not editble in the UI for security reasons, but can be changed in the config.json file.
- `Admin Password`: Depreciated but available for legacy. Usually you should leave this empty.
- `Additional Parameters`: Advanced feature, only use if you know what you're doing!
- `Show Expert Settings`: Defaults to FALSE. Setting to TRUE will show more settings that are for advacned users. More info below.
!!! tip
    To the right of the **Server Configuration** tab there is access to **Discord Intergration**, **LaunchPad Mods**, **Detection Manager**, and **Themes**. See the respective pages for more information (work in progress!)

## Alternate Configuration Methods

### Configuration File (config.json)

The `config.json` file stores all settings in JSON format and is located at `./UIMod/config/config.json`. Changes to this file will be applied upon SSUI server restart.

### Default Values

If a configuration value is not specified in either environment variables or the `config.json` file, a default value will be used.

<details>
<summary>Click to read about Environment Variables example</summary>

### Environment Variables - Advanced usecase - **Their use is discouraged**

All configuration options can be set using environment variables. This is particularly useful for containerized deployments or when you want to override configuration without modifying the `config.json` file.

Environment variable names are uppercase with underscores separating words, corresponding to their JSON counterparts. For example:
- `discordToken` in JSON becomes `DISCORD_TOKEN` as an environment variable
- `ServerMaxPlayers` in JSON becomes `SERVER_MAX_PLAYERS` as an environment variable

</details>

## Configuration Reference

The following table provides a detailed breakdown of all available configuration options:
| Field Name | Environment Variable | Type | Description | Default Value |
|------------|---------------------|------|-------------|---------------|
| **Example Values** ||||
| discordToken | DISCORD_TOKEN | string | Bot token for Discord integration | "" |
| controlChannelID | CONTROL_CHANNEL_ID | string | Discord channel ID for control commands | "" |
| statusChannelID | STATUS_CHANNEL_ID | string | Channel ID for server status updates | "" |

<details>
<summary>Click to view the full table</summary>

| Field Name | Environment Variable | Type | Description | Default Value |
|------------|---------------------|------|-------------|---------------|
| **Discord Integration** ||||
| discordToken | DISCORD_TOKEN | string | Bot token for Discord integration | "" |
| controlChannelID | CONTROL_CHANNEL_ID | string | Discord channel ID for control commands | "" |
| statusChannelID | STATUS_CHANNEL_ID | string | Channel ID for server status updates | "" |
| connectionListChannelID | CONNECTION_LIST_CHANNEL_ID | string | Channel ID for player connection logs | "" |
| logChannelID | LOG_CHANNEL_ID | string | Channel ID for general server logs | "" |
| saveChannelID | SAVE_CHANNEL_ID | string | Channel ID for save notifications | "" |
| controlPanelChannelID | CONTROL_PANEL_CHANNEL_ID | string | Channel ID for interactive control panel | "" |
| DiscordCharBufferSize | DISCORD_CHAR_BUFFER_SIZE | int | Buffer size (characters) for Discord messages | 1000 |
| blackListFilePath | BLACKLIST_FILE_PATH | string | File path to the banlist | "./Blacklist.txt" |
| isDiscordEnabled | IS_DISCORD_ENABLED | bool | Enables/disables Discord integration | false |
| errorChannelID | ERROR_CHANNEL_ID | string | Channel ID for error logs | "" |
| **Backup Configuration** ||||
| backupKeepLastN | BACKUP_KEEP_LAST_N | int | Number of last backups to retain | 2000 |
| isCleanupEnabled | IS_CLEANUP_ENABLED | bool | Enables cleanup of backups | false |
| backupKeepDailyFor | BACKUP_KEEP_DAILY_FOR | int | Hours to keep daily backups | 24 |
| backupKeepWeeklyFor | BACKUP_KEEP_WEEKLY_FOR | int | Hours to keep weekly backups | 168 |
| backupKeepMonthlyFor | BACKUP_KEEP_MONTHLY_FOR | int | Hours to keep monthly backups | 730 |
| backupCleanupInterval | BACKUP_CLEANUP_INTERVAL | int | Hours between cleanup operations | 730 |
| backupWaitTime | BACKUP_WAIT_TIME | int | Seconds to wait before copying backups to Safebackups | 30 |
| **Game Settings** ||||
| gameBranch | GAME_BRANCH | string | Game branch to run (e.g., "public" or "beta") | "public" |
| ServerName | SERVER_NAME | string | Name displayed in server list | "Stationeers Server UI" |
| SaveInfo | SAVE_INFO | string | Save folder name and save file name | "Luna" |
| ServerMaxPlayers | SERVER_MAX_PLAYERS | string | Maximum number of players allowed | "8" |
| ServerPassword | SERVER_PASSWORD | string | Password for server access | "" |
| ServerAuthSecret | SERVER_AUTH_SECRET | string | Secret key for server authentication | "" |
| AdminPassword | ADMIN_PASSWORD | string | Admin password | "" |
| GamePort | GAME_PORT | string | Port for gameplay traffic | "27016" |
| UpdatePort | UPDATE_PORT | string | Port for server updates | "27015" |
| UPNPEnabled | UPNP_ENABLED | bool | Enables UPNP for automatic port forwarding | false |
| AutoSave | AUTO_SAVE | bool | Enables automatic saving | true |
| SaveInterval | SAVE_INTERVAL | string | Interval between autosaves in seconds | "300" |
| AutoPauseServer | AUTO_PAUSE_SERVER | bool | Pauses server when no players are connected | true |
| LocalIpAddress | LOCAL_IP_ADDRESS | string | IP address to bind server to | "" |
| StartLocalHost | START_LOCAL_HOST | bool | Restricts server to local network | true |
| ServerVisible | SERVER_VISIBLE | bool | Makes server visible in public listings | true |
| UseSteamP2P | USE_STEAM_P2P | bool | Enables Steam Peer-to-Peer networking | true |
| **Server UI Settings** ||||
| ExePath | EXE_PATH | string | Path to server executable | (platform-dependent) |
| AdditionalParams | ADDITIONAL_PARAMS | string | Extra command-line parameters | "" |
| users | SSUI_USERS | map[string]string | Map of usernames to hashed passwords | {} |
| authEnabled | SSUI_AUTH_ENABLED | bool | Enables authentication for UI access | false |
| JwtKey | SSUI_JWT_KEY | string | JWT key for authentication | (auto-generated) |
| AuthTokenLifetime | SSUI_AUTH_TOKEN_LIFETIME | int | Lifetime of auth token in minutes | 1440 |
| Debug | DEBUG | bool | Enables debug mode | false |
| CreateSSUILogFile | CREATE_SSUI_LOGFILE | bool | Enables log file writing | false |
| LogLevel | LOG_LEVEL | int | Sets the log level | 20 |
| subsystemFilters | SUBSYSTEM_FILTERS | []string | Filters for specific subsystem logs | [] |
| **Update Settings** ||||
| isUpdateEnabled | IS_UPDATE_ENABLED | bool | Enables Auto Updater | true |
| allowPrereleaseUpdates | ALLOW_PRERELEASE_UPDATES | bool | Allows installation of prerelease updates | false |
| allowMajorUpdates | ALLOW_MAJOR_UPDATES | bool | Allows installation of major version updates | false |
</details>

This list is slightly outdated, sorry about that

### Field Details

#### Logging Configuration

- **LogLevel:** Sets the logging severity level:
  - 10: Debug
  - 20: Info (Default)
  - 30: Warn
  - 40: Error

- **CreateSSUILogFile:** When enabled, logs operations to `./UIMod/logs/`. Default: false.

- **SubsystemFilters:** Array of subsystem names to filter logs, allowing you to focus on specific parts of the application.

**Note:** Gameserver logs are only saved to a file on Linux systems, where `debug.log` is written by the gameserver and are not included in the SSUI log files.

#### Authentication Configuration

- **Users:** Map of usernames to hashed passwords for UI authentication.
- **AuthEnabled:** When true, requires authentication to access the UI.
- **JwtKey:** Secret key used for JWT token generation.
- **AuthTokenLifetime:** Duration in minutes that authentication tokens remain valid


#### Update Configuration

- **IsUpdateEnabled:** When true, automatically checks for and downloads new executable updates from GitHub. Default: true.
- **AllowPrereleaseUpdates:** When true, includes prerelease versions in update checks. Default: false.
- **AllowMajorUpdates:** When true, allows installation of major version updates that might include breaking changes. Default: false.

### Save Info Format

The `SaveInfo` parameter takes a single phrase (e.g. `Moon` or `my-awesome-world`). For compatability, refrain from using spaces in the name.
When this parameter is processed, the system automatically sets:
- The folder name in `saves/<SaveInfo>/` 
- The name of the game save file itself within the `saves` folder `<saveInfo>.save`. e.g. Using `my-awesome-world` will create `saves/my-awesome-world/my-awesome-world.save` along with `SafeBackups`, `autosave`, `manualsave`, and `quicksave` folders.

## Example Configuration

### Example JSON Configuration

<details>
<summary>Click to view the full example</summary>

```json
{
  "discordToken": "",
  "controlChannelID": "",
  "statusChannelID": "",
  "connectionListChannelID": "",
  "logChannelID": "",
  "saveChannelID": "",
  "controlPanelChannelID": "",
  "DiscordCharBufferSize": 1000,
  "blackListFilePath": "./Blacklist.txt",
  "isDiscordEnabled": false,
  "errorChannelID": "",
  "backupKeepLastN": 2000,
  "isCleanupEnabled": true,
  "backupKeepDailyFor": 24,
  "backupKeepWeeklyFor": 168,
  "backupKeepMonthlyFor": 730,
  "backupCleanupInterval": 730,
  "backupWaitTime": 30,
  "gameBranch": "public",
  "ServerName": "StationeersServerWithUI",
  "SaveInfo": "Moon Moon",
  "ServerMaxPlayers": "8",
  "ServerPassword": "secret",
  "ServerAuthSecret": "",
  "AdminPassword": "",
  "GamePort": "27016",
  "UpdatePort": "27015",
  "UPNPEnabled": false,
  "AutoSave": true,
  "SaveInterval": "300",
  "AutoPauseServer": true,
  "LocalIpAddress": "0.0.0.0",
  "StartLocalHost": false,
  "ServerVisible": true,
  "UseSteamP2P": true,
  "ExePath": "",
  "AdditionalParams": "",
  "users": null,
  "authEnabled": false,
  "JwtKey": "",
  "AuthTokenLifetime": 1440,
  "Debug": false,
  "CreateSSUILogFile": true,
  "LogLevel": 20,
  "subsystemFilters": [],
  "isUpdateEnabled": true,
  "allowPrereleaseUpdates": false,
  "allowMajorUpdates": false
}
```
Slightly outdated too, sorry about that
</details>

### Example Environment Variables Setup

!!! important
    The Environment Variables can be shown for completeness, but their use is discouraged. Use `config.json` modification instead.

<details>
<summary>Click to view the Environment Variables example</summary>

Using Environment Variables is disencouraged in never versions.

```bash
# Discord Configuration
export DISCORD_TOKEN="your_discord_token"
export IS_DISCORD_ENABLED=true
export CONTROL_CHANNEL_ID="123456789012345678"

# Game Settings
export SERVER_NAME="My Stationeers Server"
export SAVE_INFO="Mars Mars"
export SERVER_MAX_PLAYERS="10"
export SERVER_PASSWORD="mysecretpassword"

# Backup Configuration
export BACKUP_KEEP_LAST_N=500
export IS_CLEANUP_ENABLED=true

# UI Authentication
export SSUI_AUTH_ENABLED=true
export SSUI_USERS="admin:hashedpassword123,moderator:hashedpassword456"
export SSUI_JWT_KEY="your-secure-jwt-key"

# Update Settings
export IS_UPDATE_ENABLED=true
export ALLOW_PRERELEASE_UPDATES=false

# Logging
export LOG_LEVEL=10
export CREATE_SSUI_LOGFILE=true
export SUBSYSTEM_FILTERS="gameserver,backup,discord"
```

</details>


## Additional Parameters (Gameserver)

You can also customize your gameserver even further by adding parameters in the **Additional Parameters** field in the UI, or in the `config.json` file. These parameters allow you to control even more aspects of the game, from weird graphics stuff to experimental unity server configuration.

<details>
<summary>Click to view the full list of available parameters</summary>

- `get_Path`
- `set_Path`
- `Save`
- `Load`
- `Equals`
- `GetHashCode`
- `GetType`
- `ToString`
- `.ctor`
- `Path`
- `SettingsVersion`
- `ShowFps`
- `ShowLatency`
- `AutoSave`
- `SaveInterval`
- `SavePath`
- `HUDScale`
- `TooltipOpacity`
- `IngamePortrait`
- `ExtendedTooltips`
- `ChatFadeTimer`
- `DayLength`
- `LegacyInventory`
- `ShowSlotToolTips`
- `DeleteSkeletonOnDecay`
- `Monitor`
- `ScreenWidth`
- `ScreenHeight`
- `RefreshRate`
- `GraphicQuality`
- `TextureQuality`
- `FullScreen`
- `Vsync`
- `Shadows`
- `ShadowResolution`
- `ShadowDistance`
- `LightShadowDistance`
- `RoomControlTickSpeed`
- `ShadowNearPlaneOffset`
- `ShadowCascades`
- `ThingShadowMode`
- `ThingShadowDistanceMultiplier`
- `RenderDistance`
- `WorldOrigin`
- `Brightness`
- `FieldOfView`
- `ColorBlind`
- `ParticleQuality`
- `SoftParticles`
- `EnvironmentElements`
- `ExtendedTerrain`
- `VolumeLight`
- `PixelLightCount`
- `MaxThingLights`
- `Antialiasing`
- `FrameLock`
- `AtmosphericScattering`
- `AmbientOcclusion`
- `LensFlares`
- `ChunkRenderDistance`
- `MineableRenderDistance`
- `DisableWaterVisualizer`
- `Clouds`
- `HelmetOverlay`
- `WeatherEventQuality`
- `MasterVolume`
- `SoundVolume`
- `VoiceNotificationVolume`
- `MusicVolume`
- `InterfaceVolume`
- `VirtualVoices`
- `RealVoices`
- `UserSpeakerMode`
- `ServerName`
- `StartLocalHost`
- `ServerVisible`
- `ServerPassword`
- `AdminPassword`
- `ServerAuthSecret`
- `ServerMaxPlayers`
- `UpdatePort`
- `GamePort`
- `UPNPEnabled`
- `UseSteamP2P`
- `DisconnectTimeout`
- `NetworkDebugFrequency`
- `LocalIpAddress`
- `AutoPauseServer`
- `LanguageCode`
- `VoiceLanguageCode`
- `Voice`
- `PopupChat`
- `CameraSensitivity`
- `KeyList`
- `InvertMouse`
- `InvertMouseWheelInventory`
- `MenuLite`
- `MouseWheelZoom`
- `FirstRun`
- `VoiceNotifications`
- `CompletedTutorials`
- `CompletedScenarios`
- `DisplayHelperHints`
- `AutoExpandHelperHints`
- `VerticalMovementAxis`
- `HorizontalMovementAxis`
- `ForwardMovementAxis`
- `VerticalLookAxis`
- `HorizontalLookAxis`
- `UseCustomWorkThreadsCount`
- `MinWorkerThreads`
- `MinCompletionPortThreads`
- `MaxWorkerThreads`
- `MaxCompletionPortThreads`
- `CoroutineTimeBudget`
- `SmoothTerrain`
- `SmoothTerrainAngle`
- `ConsoleBufferSize`
- `LegacyCpu`
</details>
