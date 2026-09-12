# Getting started

There are only a few clicks between you and a working Stationeers server. None require a degree in orbital mechanics.

The normal SSUI path is intentionally simple: download one executable, place it in a writable folder, run it and follow the browser wizard. This guide helps you choose the host and understand the few decisions that matter. You do not need to learn the advanced pages before the first server starts.

## 1. Choose where SSUI will run

| Host | Good choice when... | Start here |
|---|---|---|
| Windows 10 or 11 | You want the simplest familiar setup or already have a Windows server | [Windows installation](Installation.md#windows) |
| Ubuntu LTS or Debian 13 | You want a lightweight native server and are comfortable with a terminal | [Linux installation](Installation.md#linux) |
| Docker on Linux amd64 | You already operate containers and want the complete installation in one volume | [Docker guide](Docker-Guide.md) |

The [installation guide](Installation.md) covers the actual dependencies and commands for each platform. Not sure whether the machine is large enough? Read [requirements and sizing](Requirements.md). The world simulation consumes most of the resources, not the web interface.

## 2. Install it

The native installation is deliberately boring:

1. Download the correct SSUI executable.
2. Put it in a new writable folder.
3. Run it as your normal user.
4. Let SSUI install SteamCMD and Stationeers.
5. Open `https://<server-address>:8443`.

The [installation guide](Installation.md) includes copyable commands, the Windows runtime dependency and Linux permissions. If you have done this sort of thing before, use the [quick start](Quick-Start-Guide.md).

## 3. Build the first configuration

The browser setup asks for a server name, a save name, a world, network settings and the first owner account. The [first-time setup guide](First-Time-Setup.md) explains what each important choice means and how to bring an existing world.

You are finished when all four of these are true:

- You can sign in to SSUI again.
- Stationeers reaches a joinable running state.
- A player can connect through the network you intend to use.
- SSUI creates and can download a backup after an autosave.

That last item is not optional busywork. A backup system first tested after the disaster is just a confidence-themed folder.

## After the first launch

- Learn the safe save, stop and update sequence in [server operations](Server-Operations.md).
- Understand archive retention and restores in [backups](Backup-System.md).
- Add other administrators and restrict their permissions in [access and HTTPS](Security-Considerations.md).
- Set up the [Discord hub](Discord-Integration.md) or [mods](Modding-with-StationeersLaunchPad.md) when the basic server is stable.

If something fails, start with [troubleshooting](Troubleshooting.md). If the logs are mostly hieroglyphics, create a [support package](Support-Packages.md) and ask in the [SSUI Discord](https://discord.gg/8n3vN92MyJ).

---

[Documentation home](index.md) · [Requirements](Requirements.md) · [Install SSUI](Installation.md)
