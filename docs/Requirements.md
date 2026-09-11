# Requirements and sizing

This is a sizing and platform reference, not required reading before installation. For the actual Windows/Linux dependencies and the commands to install SSUI, use [Installation](Installation.md). It is shorter and much more useful when you are currently trying to get the server running.

SSUI itself is small. Stationeers is the part simulating atmospheres, machines, scripts, networks, players and the consequences of all of the above. Size the host for the world you intend to run, not for the login page.

## The short recommendation

For a new server, start here:

- **CPU:** 4 modern x86-64 cores
- **Memory:** 8 GB RAM for a small server; 16 GB is the comfortable recommendation
- **Storage:** 20 GB of free local SSD space, plus room for your backup history
- **Operating system:** Windows 10/11, Ubuntu 24.04 LTS or newer, or Debian 13
- **Network:** a stable connection and control over the firewall; router access if internet players will connect directly

This is a practical SSUI hosting baseline, not an official promise that every world fits in it. A fresh Moon world and a multi-year Europa factory are technically both saves in the same way that a tent and an airport are both buildings.

## Supported platforms

SSUI release binaries and the published container target **x86-64 / amd64**.

| Platform | Recommendation |
|---|---|
| Windows | Windows 10 or 11, 64-bit |
| Ubuntu | Ubuntu Server 24.04 LTS or newer |
| Debian | Debian 13 "Trixie" |
| Docker | Linux host capable of running amd64 containers |

Debian 13 is the documented Debian path and is also the base of the official container image. Ubuntu 22.04 may run the SSUI binary, but Ubuntu 24.04 LTS and Debian 13 are the recommended paths for the current Stationeers runtime. Other current Linux distributions may work, but they are not the installation path this guide asks a new user to debug at 02:00. ARM hosts are not a native release target.

The dedicated server does not need a GPU or graphical desktop. You manage it from another machine through the browser.

## How much CPU and RAM?

There is no useful universal number. Base complexity matters more than the number printed next to the save file.

| World | Sensible starting point |
|---|---|
| New world, a few players | 4 cores, 8 GB RAM |
| Established base, mods or several regular players | 4-8 strong cores, 16 GB RAM |
| Very large or long-running world | 8+ strong cores and 32 GB RAM may be appropriate |

Fast individual CPU cores matter. More cores help the operating system and supporting work, but they do not make every part of one simulation parallel. Watch real memory use and simulation performance as the base grows.

Do not configure a container or VM with a hard 8 GB limit merely because the table says 8 GB. Leave headroom for SteamCMD updates, backup analysis and the host OS. If the process reaches the limit, the kernel does not award points for optimism.

<details>
<summary>Why these numbers differ from the Stationeers store page</summary>

The [Stationeers store requirements](https://store.steampowered.com/app/544550/Stationeers/) describe the playable game client and list 4 GB minimum / 8 GB recommended memory. A dedicated host has no rendering workload, but a persistent server may run a much larger world for much longer than a fresh client benchmark. The recommendations above are deliberately more conservative hosting guidance based on SSUI's operating experience.

</details>

## Storage

Use a local SSD for the first installation. Allow space for:

- SSUI and the Stationeers server installation
- the active save and the game's five rotating autosaves
- SSUI's retained backup archives
- mods, update downloads and logs

Twenty gigabytes is a reasonable empty starting point. Backup retention determines how quickly that grows. If one archive is 500 MB and you retain 100, the backups alone need roughly 50 GB. Space is rarely mysterious; it is usually multiplication wearing a fake moustache.

SSUI must be able to write its complete installation data and save tree. Read-only or oddly mapped storage will fail the startup sanity check. Network filesystems can work with the polling backup detector, but disconnects, caching and rename behavior still make them a deployment decision. See [backup storage requirements](Backup-System.md#storage-requirements).

## Network access

The host needs outbound HTTPS access to GitHub and Steam services for SSUI and game updates.

For inbound access:

- UDP `27016` is the default Stationeers game port.
- TCP `8443` is the SSUI administration interface.

Players need the game port. They do not need the administration port. Keep `8443` on a trusted LAN or VPN, or restrict it to known administrator addresses. Internet hosting usually also requires router port forwarding and a stable LAN address for the host.

## Before downloading

You are ready if you have:

- a supported x86-64 host with enough headroom;
- a new writable installation folder;
- local SSD space for the game and backups;
- administrator access to the firewall/router you actually need to configure; and
- a browser on your own computer.

Continue with [installation](Installation.md).

---

[Documentation home](index.md) · [Install SSUI](Installation.md) · [Docker](Docker-Guide.md)
