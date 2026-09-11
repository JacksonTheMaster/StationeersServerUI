# Command-line flags

Startup flags set selected values during SSUI startup. Many use persistent configuration setters, so removing a flag does not necessarily undo its effect. They are useful for recovery, development, containers, and unusual network layouts. Persistent ordinary settings belong in the Web UI or `config.json`.

Use the spelling and capitalization shown below.

| Flag | Alias | Value | Description |
|---|---|---|---|
| `-BackendEndpointPort` | `-p` | port | Override `SSUIWebPort`. Despite the historical name, this is the SSUI HTTPS endpoint port. |
| `-GamePort` | - | port | Override the Stationeers game port. |
| `-GameBranch` | `-b` | branch | Override the Steam game branch, such as `beta`. |
| `-RecoveryPassword` | `-r` | password | Add/replace a `recovery` user with the supplied password. |
| `-LogLevel` | `-ll` | integer | Override log level; normally 10 debug, 20 info, 30 warning, 40 error. |
| `-IsDebugMode` | `-debug` | boolean flag | Enable debug mode and debug-level logging. |
| `-LogToFiles` | `-lf` | boolean flag | Enable SSUI log files. |
| `-NoSteamCMD` | - | boolean flag | Skip SteamCMD installation/initialization. |
| `-NoSanityCheck` | - | boolean flag | Skip startup sanity checks. Not recommended. |
| `-AdvertiserOverride` | - | address | Override advertised address with `auto`, IPv4, or DNS hostname. `ServerVisible` must be false for this mode. |
| `-dev` | - | boolean flag | Development-only mode with known development credentials. Never use on a real exposed server. |

## Examples

Windows PowerShell:

```powershell
.\StationeersServerUI_v6.0.0_windows_amd64.exe -GameBranch beta -LogLevel 10 -LogToFiles
```

Linux:

```bash
./StationeersServerUI_v6.0.0_linux_amd64 -GameBranch beta -LogLevel 10 -LogToFiles
```

## Recovery user

If normal accounts are inaccessible, `-RecoveryPassword` creates or replaces a local owner named `recovery`, revokes that account's existing sessions and API tokens, and stores its password securely in the identity file.

```text
-RecoveryPassword "use-a-strong-temporary-password"
```

!!! warning
    Command-line arguments can be visible in shell history and process listings. Run recovery locally, use a temporary password, change it after signing in and remove the command from any retained history.

## Dangerous flags

`-NoSanityCheck` bypasses checks intended to catch an invalid installation or configuration. SSUI deliberately pauses and warns when it is used. If you bypass the sanity check, the dragons mentioned in the console are your problem now.

`-dev` enables a known development login and is explicitly unsuitable for production or any network where another person can reach SSUI.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
