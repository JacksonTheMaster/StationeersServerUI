# Quick start

For the impatient and the suspiciously confident:

1. Download the Windows or Linux amd64 build from the [SSUI download page](https://ssui.dev/).
2. Put it in a new, writable folder. Do not drop it into the Stationeers client directory.
3. On Linux, install `lib32gcc-s1`, mark the binary executable and run it as a non-root user. On Windows, run the `.exe` as your normal user.
4. Wait while SSUI installs SteamCMD and the Stationeers dedicated server.
5. Open `https://<server-LAN-IP>:8443`, accept the expected first-run certificate warning and complete the setup wizard.
6. Start Stationeers, join it, then wait for and download the first SSUI backup.

That is the entire normal path. Functioning server in a few minutes, mileage varying with SteamCMD, network speed and whether the laws of networking have selected you personally today.

### Linux commands

```bash
sudo apt-get update
sudo apt-get install -y lib32gcc-s1
mkdir -p "$HOME/ssui"
cd "$HOME/ssui"
chmod +x ./StationeersServerUI_v6.0.0_linux_amd64
./StationeersServerUI_v6.0.0_linux_amd64
```

Use the filename you actually downloaded if the version differs.

### Windows location

`%USERPROFILE%\Documents\SSUI` is a sensible installation folder. Move the downloaded executable there and run it. SSUI installs SteamCMD under `C:\SteamCMD`, so your account must also be able to write that directory.

!!! important
    Do not expose TCP `8443` to the public internet during setup. It is the administration UI, not a game port.

Need the explanations, an existing save, Docker or network help? Start at [Getting started](Getting-Started.md) or use the full [installation guide](Installation.md).

---

[Documentation home](index.md) · [Requirements](Requirements.md) · [First-time setup](First-Time-Setup.md)
