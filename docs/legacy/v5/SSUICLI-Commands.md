!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

### What is SSUICLI?

**SSUICLI** (Stationeers Server UI Command-Line Interface) is the built-in runtime console.

It provides a command-line interface that runs alongside the web UI, allowing server administrators to perform common management tasks directly from the terminal without needing to access the web interface. The console is optional (can be disabled in config) and appears with a green prompt: `SSUICLI »`.


<img width="525" height="150" alt="image" src="https://github.com/user-attachments/assets/9d43a957-14ee-4d76-b21b-406eaad91b60" />

_Example: Reloading the backend configuration from CLI_

### Key Features and Capabilities

- **Interactive Control**: Type commands in real-time while the server and web UI continue running.
- **Server Management**: Start/stop the dedicated game server, reload configuration, or restart the web backend.
- **Updates & Maintenance**: Check for StationeersServerUI updates, apply them, manually run SteamCMD, or delete the config for a fresh setup.
- **Debugging & Support**: Enable a temporary "support mode" for verbose logging, test localization, generate a support package (ZIP with logs, sanitized config, and system info), or print current config/build details.
- **Convenience**: Many commands have short aliases (e.g., `start` for `startserver`, `u` for `update`) for quick typing.
- **Help System**: Type `help` (or `h`) to list all available commands and their aliases.

For a full list of available commands and their functions, see the table below:

#### General
| Command | Aliases | Description                                      |
|---------|---------|--------------------------------------------------|
| help    | h       | Lists all available runtime CLI commands with their aliases. |
| exit    | e       | Exits the StationeersServerUI application immediately. |

#### Server Control
| Command         | Aliases     | Description                                      |
|-----------------|-------------|--------------------------------------------------|
| startserver     | start       | Starts the dedicated Stationeers game server.    |
| stopserver      | stop        | Stops the running dedicated Stationeers game server. |

#### Configuration & Backend
| Command         | Aliases         | Description                                                                 |
|-----------------|-----------------|-----------------------------------------------------------------------------|
| reloadconfig    | rlc, rc         | Reloads the configuration file without restarting the backend.              |
| reloadbackend   | rlb, rb, r      | Reloads the backend (restarts the web server without stopping the game server). |
| restartbackend  | rsb             | DESTRUCTIVE Restarts the backend/web server immediately (stops the server and SSUI then restarts SSUI). |
| printconfig     | pc              | Prints the current configuration details to the log.                        |
| deleteconfig    | delc, dc        | Deletes the current config file (`config.json`), forcing a fresh setup on next start. |

#### Updates & Steam
| Command         | Aliases         | Description                                                                 |
|-----------------|-----------------|-----------------------------------------------------------------------------|
| update          | u               | Checks for StationeersServerUI updates and reports if a new version is available. |
| applyupdate     | au              | Applies a pending update to StationeersServerUI (downloads and installs if available). |
| runsteamcmd     | steamcmd, stcmd | Manually runs SteamCMD installation/update process.                         |
| getbuildid      | gbid            | Prints the current detected build ID for the active game branch.            |

#### Debugging & Support
| Command          | Aliases | Description                                                                 |
|------------------|---------|-----------------------------------------------------------------------------|
| supportmode      | sm      | Toggles support/debug mode: enables debug logging, max verbosity, and log files. When enabled, allows generating a support package. |
| supportpackage   | sp      | Creates a ZIP support package containing logs, sanitized config, and system info (only works when support mode is active). |
| testlocalization | tl      | Tests the current localization by printing the "Start Server" button text in the active language. |
| setdummybuildid | sdbid           | Sets a dummy build ID (used for testing or bypassing update checks).        |
