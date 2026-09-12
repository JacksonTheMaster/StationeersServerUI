# Stationeers Server UI documentation

!!! note "The documentation home moved"
    This older home page is kept for existing links. Open the [current documentation home](index.md) for the maintained landing page, downloads and the full navigation.

SSUI runs the Stationeers dedicated server so you do not have to babysit SteamCMD, hand-edit launch arguments or discover that your last useful autosave vanished three hours ago.

It gives you a web interface for setup and daily operation, managed backups, game updates, mods, Discord controls and enough logs to work out which part of the station caught fire first.

## New here?

The [installation guide](Installation.md) is the place to start. It contains the supported platforms, platform-specific dependencies and copyable commands. You do not need to read the entire documentation before running SSUI.

As a rough baseline, you want:

- a 64-bit Windows or Linux host (or an amd64 Docker host);
- 8 GB RAM for a small server, 16 GB if you want comfortable headroom;
- 20 GB of free SSD space, plus whatever your backup history needs; and
- control over the firewall and router if people will connect over the internet.

Then follow these three steps:

1. [Install SSUI](Installation.md). The normal path is download, put it in an empty folder and run it.
2. [Complete the first-time setup](First-Time-Setup.md).
3. Start the server and verify that SSUI can create one backup.

Not sure whether the host is large enough? The [requirements and sizing](Requirements.md) page has the longer answer. Already know your way around a server? The [quick start](Quick-Start-Guide.md) gets the ceremony out of the way.

!!! tip
    You do not need a graphical desktop on the server. SSUI runs there; the web interface opens in a browser on your own computer.

## Run your server

| I want to... | Read this |
|---|---|
| See what SSUI can do | [Features](Features.md) |
| Start, save, stop or update the server | [Server operations](Server-Operations.md) |
| Understand the dashboard and pages | [Web interface](Web-Interface.md) |
| Inspect, download or restore a save | [Backups and restores](Backup-System.md) |
| Let Discord users check or control the server | [Discord integration](Discord-Integration.md) |
| Install LaunchPad or Workshop mods | [Modding](Modding-with-StationeersLaunchPad.md) |
| Add custom log detections | [Logs and detections](Log-Detection-System.md) |
| Add users, permissions or API keys | [Access and HTTPS](Security-Considerations.md) |
| Fix something that is already on fire | [Troubleshooting](Troubleshooting.md) |
| Ask a human for help | [Create a support package](Support-Packages.md), then visit [Discord](https://discord.gg/8n3vN92MyJ) |

## Reference shelf

These pages are useful when you need one exact setting rather than a guided tour:

[Configuration](Configuration.md) · [SSCLI commands](SSUICLI-Commands.md) · [Startup flags](Command-line-flags.md) · [Game commands / SSCM](Allowed-SSCM-commands.md) · [HTTP API](API.md) · [Developer guide](Developer-Documentation.md)

## Version scope

This documentation follows the SSUI v6 development line. Version 5 has different account management, API routes, backup controls and Discord panels. Use the documentation shipped for your release while upgrading.

[Download SSUI](https://ssui.dev/) · [GitHub](https://github.com/SteamServerUI/StationeersServerUI) · [Report an issue](https://github.com/SteamServerUI/StationeersServerUI/issues) · [Discord support](https://discord.gg/8n3vN92MyJ)
