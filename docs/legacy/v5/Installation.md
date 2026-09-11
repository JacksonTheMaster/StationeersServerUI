!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# Installation

The install is supposed to be as easy as possible. 

**No matter which platform (Linux/Windows), you can essentially just download the latest executable and run it.**

SSUI will create all necessary files and folders for you.

[![Download latest Version](https://img.shields.io/badge/Open-DownloadPage%20-orange?style=for-the-badge)](https://ssui.dev/)
[![Download latest Version](https://img.shields.io/badge/Direct-Linux%20Download-%23FCC624?style=for-the-badge&logo=linux&logoColor=%23FCC624)](https://ssui.dev/?download=linux)
[![Download latest Version](https://img.shields.io/badge/Direct-Windows%20Download-blue?style=for-the-badge)](https://ssui.dev/?download=windows)

## Windows Installation
<details>

<summary>Open Section</summary>

  - Download the latest release executable file from the [Download Page](https://ssui.dev) or using the Direct Download Button below.

[![Download latest Version](https://img.shields.io/badge/Open-DownloadPage%20-orange?style=for-the-badge)](https://ssui.dev/)
[![Download latest Version](https://img.shields.io/badge/Direct-Windows%20Download-blue?style=for-the-badge)](https://ssui.dev/?download=windows)

   - Place the executable in an **empty** folder of your choice 
     - Make sure you have full permission to this folder. It is recommended to not create this folder under the root of your `C:` drive. Your Documents folder, e.g: `%USERPROFILE%\Documents\SSUI`, would be a good starting point.)
   - Run the executable. 
     - SSUI will download Steamcmd, Steamcmd will download Stationeers. 
   - A console window will open, displaying output. **Read it** and Navigate to `https://<IP-OF-YOUR-SERVER>:8443` in a Browser of your choice.
     - Follow the setup track

</details>


## Linux Installation

<details>

<summary>Open Section</summary>

_(use Ubuntu or Debian 13. They have been verified to work with the Stationeers Dedicated Server. If using Debian, make sure you use Debain 13 "Trixie"). **Fedora is NOT recommended.**_

1: Update your system

   ```sh
   sudo apt-get update 
   ```


2: Visit the [Download Page](https://ssui.dev) or use the Direct Download Button below.

[![Download latest Version](https://img.shields.io/badge/Open-DownloadPage%20-orange?style=for-the-badge)](https://ssui.dev/)

[![Download latest Version](https://img.shields.io/badge/Direct-Linux%20Download-%23FCC624?style=for-the-badge&logo=linux&logoColor=%23FCC624)](https://ssui.dev/?download=linux)

!!! note
    Command examples for `wget`, `curl` and `Powershell` can be found on the [Download Page](https://ssui.dev)



2: Execute the Downloaded file

```sh
./StationeersServerControlX.X.X.x86_64
```
**XXX should be replaced with the version you Downloaded, check the filename for that.**

3: **Access the WebUI** at https://[your-server-ip]:8443

## Automatically starting SSUI

To run SSUI on Startup, you could use a cronjob:
Example:
```sh
echo "@reboot $(whoami) $(pwd)/StationeersServerControl* >/dev/null 2>&1" | crontab -
```
This runs SSUI as the user you are currently logged in as at every server startup and suppresses the logs. Edit to your liking.


5. **VERY OPTIONAL If you run into weird issues, try this:**
_**Check and Install all sorts of possible dependencies I found while experimenting**_
   ```sh
   sudo dpkg --add-architecture i386; sudo apt update; sudo apt install bsdmainutils bzip2 jq lib32gcc-s1 lib32stdc++6 libsdl2-2.0-0:i386 libxml2-utils netcat pigz steamcmd unzip
   ```

</details>

## Docker Installation

<details>

<summary>Open Section</summary>

Visit the [Docker Guide](Docker-Guide.md)

</details>

## Storage
!!! warning
    For reliable backup operations, use local storage or **block-level** network storage (iSCSI, Fibre Channel). Network shares like SMB, NFS1/2, CIFS are NOT supported and cause issues. 
    See [Backup-System-v2/Important Notes on Filesystem Support](https://github.com/SteamServerUI/StationeersServerUI/wiki/Backup-System-v2#important-notes-on-filesystem-support) for further details if you want to use Network Storage.


## Updating StationeersServerUI

To update to the latest version, you should use the built-in update system. Can be disabled in the configuration (see [Configuration](https://github.com/JacksonTheMaster/StationeersServerUI/wiki/Configuration) `IsUpdateEnabled`).

## 🔌 Network Setup for External Players

1. Forward these ports on your router:
   - UDP `27016`: Game port
   - UDP `27015`: Update port - unused in current game versions, no longer needed
   - TCP `8443`: the SSUI WebUI (optional, only if remote admin needed - set a password!)

2. Configure Windows Firewall:
   - Add inbound rules for ports `27016`, `27015`, and `8443`

### 🛠️ Connection Troubleshooting

- **Can't Connect**: Verify port forwarding and firewall settings

## Next Steps

- [First-Time Setup](First-Time-Setup.md) - Configure your server for first use
- [Docker Guide](Docker-Guide.md) - For containerized deployment options
- [Web Interface](Web-Interface.md) - Learn how to use the web interface
