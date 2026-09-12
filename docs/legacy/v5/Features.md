!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

## Features

StationeersServerUI v5.x delivers a modernized, secure, and scalable toolset for effortless server management:

- **Zero-Config Setup**: Drop the executable in an empty folder of an OS of your choice and run it. Stationeers server will to spin up with sensible defaults that can then be customized from the UI or a config file.
- **Automatic Updates**: Game server auto-updates on software startup, leveraging streamlined SteamCMD integration.
- **SteamCMD Management**: Fully automated SteamCMD setup with improved permission handling for Windows and Linux.
- **Server Control**: Start, stop, and manage the Stationeers server via an intuitive UI, Disocord intergration or REST API.
- **Mod Support**: One click SLP install! Drag and drop mod package zip, then one click mod updates.
- **Real-time Monitoring**: Stream server console output live with HTTP/2-powered SSE, supporting tons of streams!
- **Configuration Management**: Edit server settings through a sleek UI or JSON config files.
- **Backup System**: Customisable backup interval, list, restore and download backups directly from the UI. Your backups are in a separated backup directory, safe from the game's quirks.
- **Backup System** From v5.x, Improved backup architecture with thread safety.
- **API Support**: Robust REST API over HTTPS (port 443) for advanced control, secured with TLS and HTTP/2.
- **Security**: TLS encryption and JWT-based authentication with a polished login flow
- **Modern Discord Integration**: Server monitoring and management via Discord, Slash commands, embedded responses, and server status displays.
- **Content Creator Discord Features**: Use Discord roles to restrict server access with a rolling server passcode, retrievable by a bot in channel.
- **Cross-Platform**: Full Linux and Windows support!


## Next Steps

- [Installation](Installation.md) - Install the latest server control package.
- [First-Time Setup](First-Time-Setup.md) - Configure JSON settings and secure your server.
