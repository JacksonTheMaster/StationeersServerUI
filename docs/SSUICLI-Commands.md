# SSCLI commands

SSCLI (previously SSUICLI) is a small, lightweight runtime terminal that runs alongside the Web UI. It remains available while SSUI and Stationeers are operating and is useful over SSH, on headless hosts, or whenever opening a browser would be excessive.
It is NOT available in the standard SSUI container.

The prompt appears as:

```text
SSCLI »
```

Type `help` or `h` for the commands available in the running build. Developer commands are deliberately hidden from ordinary help output.

## User commands

| Command | Aliases | Description |
|---|---|---|
| `help` | `h` | Show available user commands, descriptions, and aliases. |
| `startserver` | `start` | Start Stationeers. |
| `stopserver` | `stop` | Stop Stationeers; save first. |
| `update` | `u` | Check for an SSUI update. |
| `applyupdate` | `au` | Apply an available SSUI update. |
| `runsteamcmd` | `steamcmd`, `stcmd` | Run SteamCMD to install/update the game server. |
| `reloadconfig` | `rlc`, `rc` | Reload the SSUI configuration. |
| `reloadbackend` | `rlb`, `rb`, `r` | Reload backend components. |
| `restartbackend` | `rsb` | Restart the SSUI backend flow. |
| `supportmode` | `sm` | Toggle detailed support logging. |
| `supportpackage` | `sp` | Create a sanitized support ZIP; support mode must be active. |
| `exit` | `e` | Stop the game server and exit SSUI. |

`downloadworkshopupdates` (`dwu`) also exists for Workshop updates but is currently marked as a developer command in the command registry, so do not present it to ordinary users as a stable interface yet.

## Support workflow

The `sm` and `sp` commands belong in **SSCLI**, not in the Stationeers console inside the Web UI.

Enter `supportmode`, reproduce the issue, enter `supportpackage`, then enter `supportmode` again.

The final `supportmode` disables the additional debug/file logging again. See [Support Packages](Support-Packages.md).

## Developer commands

Current builds contain developer-only commands for configuration deletion, localization testing, build inspection, mod/workshop inspection, individual Workshop downloads, heap profiling, and Discord panel tests.

They are omitted from ordinary `help` output intentionally. Their behavior can change and some are destructive or create diagnostic files. Consult the current command registry in `src/cli/commands.go` when developing SSUI instead of treating this wiki page as a promise that those commands are stable.

## Command input safety

SSCLI commands control the local SSUI process. The separate Web UI/Discord server-console input sends commands to Stationeers through SSCM. They are not interchangeable, even if both happen to be black boxes containing text.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
