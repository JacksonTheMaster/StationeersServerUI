---
description: Troubleshoot Stationeers dedicated server setup, ports, permissions, updates, backups and SSUI runtime problems.
---

# Troubleshooting

Stuck? Start with [Get help](Support.md) and come ask in the [SSUI Discord](https://discord.gg/8n3vN92MyJ) before you spend three hours fighting the wrong layer. We genuinely do not bite, usually respond pretty quickly, and are happy to help you work through a setup problem or tell you when something is actually broken. We even have a self-service bot that can actually help you unstuck yourself.

Start with the symptom, change one thing at a time, and keep the first error rather than only the final cascade. SSUI's terminal output and subsystem logs usually say which layer failed: browser, backend, SteamCMD, Stationeers, backups, Discord, or mods.

| Symptom | Jump to |
|---|---|
| Browser cannot connect | [Web UI](#web-ui-does-not-open) |
| Cannot sign in | [Login](#login-or-session-fails) |
| Game stalls or exits | [Startup](#stationeers-does-not-start-or-become-ready) |
| Game runs but nobody can join | [Network](#players-cannot-connect) |
| Spacecat is red | [Network](#players-cannot-connect) |
| Missing or failing backups | [Backups](#backups-do-not-appear) · [Restore](#restore-or-backup-analysis-fails) |
| Bot or mod errors | [Discord](#discord-bot-or-panels-fail) · [Mods](#slp-or-workshop-mods-fail) |

## Before changing anything

1. Record the SSUI version/branch, operating system, and whether this is native or Docker.
2. Reproduce the problem once while watching the relevant live log.
3. Note the first error and the action that triggered it.
4. Create a [Support Package](Support-Packages.md) if the cause is not obvious.

Do not publish raw `config.json`, Discord tokens, JWT keys, passwords, cookies, TLS private keys, real backup files, or unreviewed logs. A support package sanitizes known configuration secrets; still review it before sharing.

## Web UI does not open

- Use `https://`, not `http://`, and the configured `SSUIWebPort` (default `8443`).
- Test `https://localhost:8443` on the SSUI host. If Windows/Hyper-V resolves `localhost` unexpectedly, also test `https://127.0.0.1:8443`, then the host's LAN address from another device.
- Confirm SSUI is still running and no other process owns the port.
- Allow the port through the host firewall only for the networks that should administer SSUI.
- A first-run self-signed certificate causes a browser warning; verify the host before accepting it.
- In Docker/dev containers, confirm the container publishes the port to the host; merely exposing/documenting a port does not publish it.
- For persistent-volume permission errors, verify that the mounted `/app` exists and is writable inside the container. The official image currently runs as root specifically to avoid host UID/GID mismatches; it does not need `--privileged`, host networking or the Docker socket.
- With a reverse proxy, test SSUI directly first. Disable response caching/buffering for live SSE routes, use HTTP/1.1 upstream where required, and allow long-lived read timeouts.

Players need the Stationeers game port, not necessarily the Web UI port. Do not expose administration publicly as a side effect of fixing game connectivity.

## Login or session fails

- Confirm cookies are enabled and the browser is using the same scheme/host consistently.
- Check system time on the server and client; large clock errors can invalidate signed sessions.
- Complete first-time user registration before exposing the instance.
- If credentials are lost, use the local [recovery flag](Command-line-flags.md#recovery-user). The temporary password can remain in shell history or process listings, so change it after signing in.

## Stationeers does not start or become ready

- Read the first Stationeers/process error in the console.
- Run SSUI from its installation directory and confirm the service/container working directory points there.
- Confirm SteamCMD finished and `ExePath` targets the installed dedicated-server executable.
- Confirm the SSUI account can read/write the installation, saves, logs, and SteamCMD directories.
- On Windows, install the current Microsoft Visual C++ runtime required by Stationeers if the executable reports missing runtime DLLs.
- On Linux, use a supported distribution/runtime and inspect missing-library errors rather than repeatedly reinstalling the game.
- Remove newly added launch parameters temporarily and verify the selected game branch is supported.
- Stop duplicate Stationeers/SteamCMD processes before retrying.

## Players cannot connect

- Confirm the server is running and ready, not merely that SSUI itself is online.
- If Spacecat reports the gameserver as unreachable, start with the configured UDP port, host firewall, router forwarding, Docker publishing and advertised address. A red Spacecat result is a network signal, not an SSUI web-interface failure.
- Verify the configured UDP game port (default `27016`) in the host firewall, router/NAT rule, and container publishing.
- Leave `LocalIpAddress` at `0.0.0.0` unless the host specifically requires another bind address.
- Test LAN connectivity before public connectivity.
- Check double NAT or carrier-grade NAT if port forwarding appears correct but internet clients cannot connect.
- Confirm client/server game versions, branch, password, and required mod set match.
- If listing is wrong after a public-IP change, review `AdvertiserOverride` and advertiser logs.

## Backups do not appear

- Confirm `SaveName` exactly matches the active save directory.
- Confirm Stationeers writes files under `saves/<SaveName>/autosave`.
- Confirm SSUI can write `SSUI/savebackups/<SaveName>`.
- Allow two unchanged observations at least 45 seconds apart, plus archive validation, copy and analysis time.
- BMv4 does not need filesystem events. On network/ZFS storage, check actual file visibility, reconnects, permissions and write/rename behavior; compare with a local test directory if needed.
- Open `/backups` and refresh after the safe copy completes.

If cleanup removes files sooner than expected, disable retention, review all newest/daily/weekly/monthly values together, and check logs. SSUI applies the safe-backup policy before purging source autosaves; a safe-directory cleanup failure leaves the source files alone. See [Backup System](Backup-System.md).

## Restore or backup analysis fails

- Backup actions use filenames in both the Web UI and Discord. Refresh pre-update numeric Discord buttons/menus and migrate API clients to `name`. An unavailable selected file fails rather than selecting another backup.
- Stop active gameplay before restoring and confirm there is enough disk space.
- Check that the selected archive is readable and still exists.
- Deep statistics read `world.xml` and can take longer on large saves; summary metadata comes from the much smaller `world_meta.xml`.
- A malformed or incomplete archive may still be downloadable even when its metadata cannot be parsed. Preserve it before attempting repairs.

## Discord bot or panels fail

- Confirm Discord integration is enabled and the token is current.
- Use numeric channel IDs, not channel names.
- Check hub permissions: View Channel, Send Messages, Embed Links and Read Message History; Attach Files for downloads. Check member application-command access separately.
- Use the configured hub and give administrators the exact `discordAdminRoleID` role. Discord Administrator alone does not bypass it.
- Refresh/reload the integration after channel or token changes and inspect Discord subsystem errors.
- Voting controls appear only when their feature is enabled and a usable status-panel channel exists.

## SLP or Workshop mods fail

- Stop Stationeers before changing the mod stack.
- Confirm outbound SteamCMD access, free disk space, and directory permissions.
- Validate every Workshop ID/URL; a multi-item request can install successful items and still report an aggregate failure for the others.
- Compare the game, BepInEx, SLP, and mod versions plus client/server mod order.
- Restore the last exported SLP package and a safe world backup if a mod change damaged the save.

## What to include in a support request

- SSUI version/branch and installation method.
- Operating system or container image.
- The exact action and reproducible steps.
- The first relevant error and timestamp.
- Whether the failure also occurs with direct/local access or an unmodded test.
- A reviewed [Support Package](Support-Packages.md).

Keep the original installation intact until diagnosis is complete. Deleting logs, configs, saves, containers or all Docker data may remove the only evidence - and occasionally the only backup.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
