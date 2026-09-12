!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# The Holy SSUI Debug Flowchart

> *Based on the Data of ALL support tickets so far*

**Navigation:** Start at [Chart 1](#chart-1-main-triage). Each chart tells you where to go next.

---

## Chart 1 - Main Triage

*Start here. Always.*

```mermaid
flowchart TD
    START(["🆘 HELP!\nSomething is wrong!"]):::alert --> Q1{"What's the problem?"}

    Q1 -->|"Can't access\nWeb UI"| C2["➡️ Go to Chart 2\nWeb UI & Login Issues"]:::nav
    Q1 -->|"Can't log in"| C2
    Q1 -->|"Server\nwon't start"| C3["➡️ Go to Chart 3\nServer Startup Issues"]:::nav
    Q1 -->|"Players\ncan't connect"| C4["➡️ Go to Chart 4\nConnection & Networking"]:::nav
    Q1 -->|"Server disappears\nfrom server list"| C5["➡️ Go to Chart 5\nServer List & Advertiser"]:::nav
    Q1 -->|"Docker\nbeing Docker"| C6["➡️ Go to Chart 6\nDocker Issues"]:::nav
    Q1 -->|"Discord bot /\nSaves / Crashes /\nOther"| C7["➡️ Go to Chart 7\nEverything Else™"]:::nav

    Q1 -->|"It's actually fine\nI'm just anxious"| RESOLVED

    RESOLVED(["✅ Crisis Averted!\nServer goes brrr.\nGo play Stationeers."]):::success

    classDef alert fill:#e74c3c,color:#fff,stroke:#c0392b,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 2 - Web UI & Login Issues

*Can't reach the UI or can't log in? Start here.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 2\nWeb UI & Login Issues"]):::header --> MODE{"What's the issue?"}

    MODE -->|"Can't reach\nthe Web UI"| WEB
    MODE -->|"Can't log in"| LOGIN

    %% --- WEB UI ACCESS ---
    WEB{"Are you using\nhttps:// in the URL?"}
    WEB -->|"No / what?"| HTTPS_FIX["Use https://IP:8443\nSSUI does NOT serve HTTP.\nNever has. Never will."]
    HTTPS_FIX --> HTTPS_OK{"Fixed?"}
    HTTPS_OK -->|"Yes!"| RESOLVED
    HTTPS_OK -->|"Nope"| WEB_LOCAL
    WEB -->|"Yes"| WEB_LOCAL

    WEB_LOCAL{"Using 'localhost'\non Windows?"}
    WEB_LOCAL -->|"Yes"| HYPER_FIX["Try 127.0.0.1:8443 instead.\nHyper-V likes to eat\nlocalhost DNS for breakfast."]
    HYPER_FIX --> HYPER_OK{"Fixed?"}
    HYPER_OK -->|"Yes!"| RESOLVED
    HYPER_OK -->|"No"| WEB_PROXY
    WEB_LOCAL -->|"No"| WEB_PROXY

    WEB_PROXY{"Behind a\nreverse proxy?"}
    WEB_PROXY -->|"nginx"| NGINX_FIX["Add to your nginx config:\n• proxy_buffering off\n• proxy_cache off\n• proxy_http_version 1.1\n• proxy_read_timeout 86400s\nClear browser cache after!"]
    WEB_PROXY -->|"Traefik"| TRAEFIK_FIX["Add insecureSkipVerify\nfor SSUI's self-signed cert.\nCheck serversTransport config."]
    WEB_PROXY -->|"Apache"| APACHE_FIX["SSUI needs RSA certs,\nnot ECDSA!\nUse: certbot --key-type rsa"]
    WEB_PROXY -->|"No proxy"| WEB_FW{"Firewall blocking\nport 8443 TCP?"}
    WEB_FW -->|"Yes / maybe"| FW_FIX["Open port 8443 TCP\nin your firewall."]
    WEB_FW -->|"No, it's open"| JACKSON

    NGINX_FIX --> PROXY_OK{"Fixed?"}
    TRAEFIK_FIX --> PROXY_OK
    APACHE_FIX --> PROXY_OK
    FW_FIX --> PROXY_OK
    PROXY_OK -->|"Yes!"| RESOLVED
    PROXY_OK -->|"No"| JACKSON

    %% --- LOGIN ---
    LOGIN{"What's happening?"}
    LOGIN -->|"Forgot password"| PW_RESET["1. Edit UiMod/config/config.json\n2. Set authEnabled: false\n3. Restart SSUI\n4. Go to /setup\n\nOr use the -r flag on startup."]
    LOGIN -->|"'Invalid token'\nspam in logs"| JWT_FIX["Old browser tab from a\nprevious install is haunting you. 👻\nClose ALL SSUI tabs."]
    LOGIN -->|"Just doesn't work"| PW_RESET

    PW_RESET --> PW_OK{"Fixed?"}
    JWT_FIX --> PW_OK
    PW_OK -->|"Yes!"| RESOLVED
    PW_OK -->|"No"| JACKSON

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 3 - Server Startup Issues

*Server won't start? Find your platform.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 3\nServer Startup Issues"]):::header --> PLAT{"What platform?"}

    PLAT -->|"Windows"| WIN_START
    PLAT -->|"Linux\n(bare metal)"| LIN_START
    PLAT -->|"Docker"| DOCK_START

    %% --- WINDOWS ---
    WIN_START{"Installed VC++\nRedistributable?"}
    WIN_START -->|"No / What?"| VCPP["Install the latest VC++ Redist\nfrom Microsoft. Stationeers\nforgets to include it.\nEspecially on Win11."]
    VCPP --> VCPP_OK{"Fixed?"}
    VCPP_OK -->|"Yes!"| RESOLVED
    VCPP_OK -->|"No"| CHECK_CONSOLE
    WIN_START -->|"Yes"| CHECK_CONSOLE

    %% --- LINUX ---
    LIN_START{"What distro / version?"}
    LIN_START -->|"Ubuntu 22.04\nor older"| OLD_OS["Your OS is too old.\nStationeers needs newer glibc.\nUpgrade to Debian 13+\nor Ubuntu 24.04+.\n\nOr just use the Docker image."]
    OLD_OS --> RESOLVED
    LIN_START -->|"Debian 13+ /\nUbuntu 24.04+"| STEAM_ERR

    STEAM_ERR{"SteamCMD errors?"}
    STEAM_ERR -->|"Error 0x2 /\nupdate fails"| IPV6_FIX["Disable IPv6:\nsysctl -w\nnet.ipv6.conf.all.disable_ipv6=1\n\nSteamCMD and IPv6\nhave beef. 🥩"]
    IPV6_FIX --> IPV6_OK{"Fixed?"}
    IPV6_OK -->|"Yes!"| RESOLVED
    IPV6_OK -->|"No"| CHECK_CONSOLE
    STEAM_ERR -->|"'No such file'\nerrors"| PATIENCE["Is SteamCMD still downloading?\nWait for it to finish.\nPatience, young grasshopper. 🧘"]
    PATIENCE --> PAT_OK{"Fixed?"}
    PAT_OK -->|"Yes!"| RESOLVED
    PAT_OK -->|"Still broken"| STEAM_PATH["Ran SteamCMD manually?\nIt puts files in ~/Steam\ninstead of ~/SSUI/Steam.\nDelete stale folders,\nlet SSUI manage it."]
    STEAM_PATH --> SP_OK{"Fixed?"}
    SP_OK -->|"Yes!"| RESOLVED
    SP_OK -->|"No"| CHECK_CONSOLE
    STEAM_ERR -->|"No errors"| CHECK_CONSOLE

    %% --- DOCKER ---
    DOCK_START{"Sanity check\nfailing?"}
    DOCK_START -->|"Yes"| SANITY["containerd / K8s?\nSSUI doesn't detect all runtimes.\nUse -NoSanityCheck flag\nor create /.dockerenv"]
    SANITY --> SAN_OK{"Fixed?"}
    SAN_OK -->|"Yes!"| RESOLVED
    SAN_OK -->|"No"| CHECK_CONSOLE
    DOCK_START -->|"No"| CHECK_CONSOLE

    %% --- SHARED CONSOLE CHECK ---
    CHECK_CONSOLE{"Check SSUI console.\nIndexOutOfRangeException\nin TerrainSystem?"}
    CHECK_CONSOLE -->|"Yes!"| ONE_CORE["Running a 1-core VM?\nGame does ProcessorCount - 1\n= 0 workers on 1 core.\n\nFix: Add to AdditionalParams:\nMaxConcurrentWorkers 1"]
    ONE_CORE --> RESOLVED
    CHECK_CONSOLE -->|"No / other errors"| SUPPORT_PKG["Generate a Support Package!\nWiki → Support-Packages"] --> JACKSON

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 4 - Connection & Networking

*Server runs but players can't connect? The big one. Deep breaths.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 4\nConnection & Networking"]):::header --> PF

    PF{"Did you port forward\nUDP 27016 on router?"}
    PF -->|"No / What?"| PF_FIX["Forward your game port\n(default: UDP 27016)\non your router.\nSee portforward.com"]
    PF_FIX --> PF_OK{"Fixed?"}
    PF_OK -->|"Yes!"| RESOLVED
    PF_OK -->|"No"| TYPO
    PF -->|"Yes"| TYPO

    TYPO{"Double-check:\nIs it really 27016\nor did you fat-finger it?\n(Looking at you, 27106)"}
    TYPO -->|"Oops, typo!"| PF_FIX
    TYPO -->|"It's correct"| LIP

    LIP{"Is LocalIpAddress\nset to 0.0.0.0?"}
    LIP -->|"No, it's my\nmachine's IP"| LIP_FIX["Set LocalIpAddress to 0.0.0.0\nYes, the name is misleading.\nNo, don't put your actual IP.\nBlame Rocketwerkz. 🤷"]
    LIP_FIX --> LIP_OK{"Fixed?"}
    LIP_OK -->|"Yes!"| RESOLVED
    LIP_OK -->|"No"| CGNAT
    LIP -->|"Yes"| CGNAT

    CGNAT{"Behind CGNAT?\n(ISP shares your\npublic IP with neighbors)"}
    CGNAT -->|"Yes / Probably"| CGNAT_FIX["Port forwarding is impossible.\n1. Rent a VPS + FRP tunnel\n2. VPN tunnel solution\n3. AdvertiserOverride = VPS IP\n\nOr switch ISP. Seriously."]
    CGNAT_FIX --> RESOLVED
    CGNAT -->|"No"| DOCK

    DOCK{"Running in Docker?"}
    DOCK -->|"Yes"| UPNP["UPnP does NOT work in Docker.\nManual port forwarding only.\nCheck host networking mode."]
    UPNP --> UP_OK{"Fixed?"}
    UP_OK -->|"Yes!"| RESOLVED
    UP_OK -->|"No"| RAKNET
    DOCK -->|"No"| RAKNET

    RAKNET{"RakNet errors\nin server logs?"}
    RAKNET -->|"Yes"| RK_Q{"On Kubernetes?"}
    RK_Q -->|"Yes"| K8S["Use hostNetwork: true\nin your pod spec.\nRakNet can't handle\nkube-proxy SNAT. At all."]
    K8S --> RESOLVED
    RK_Q -->|"No"| RKPORT["Try moving game port\nto 50000+ range.\nRakNet is picky about ports.\n\n(Stationeers bug, not SSUI.\nWe're innocent. 🤷)"]
    RKPORT --> RK_OK{"Fixed?"}
    RK_OK -->|"Yes!"| RESOLVED
    RK_OK -->|"No"| JACKSON
    RAKNET -->|"No"| STUCK

    STUCK{"Connection stuck\nat 0%?"}
    STUCK -->|"Yes"| CLIENT["Is the CLIENT PC also\nport-forwarding those ports?\nRemove them on client!\nOnly the SERVER forwards."]
    STUCK -->|"'Incompatible\nversion'"| VER["Server + client versions\nmust match. Re-run\nSteamCMD update."]
    STUCK -->|"Neither"| DNAT{"Double NAT?\n(Modem + Router\nboth doing NAT)"}
    DNAT -->|"Yes"| DFIX["Put modem in bridge mode\nor port forward on BOTH\ndevices. Double NAT is pain."]
    DNAT -->|"No / Don't know"| JACKSON

    CLIENT --> C_OK{"Fixed?"}
    VER --> C_OK
    DFIX --> C_OK
    C_OK -->|"Yes!"| RESOLVED
    C_OK -->|"No"| JACKSON

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 5 - Server List & Advertiser

*Server runs, people connected before, but now it vanishes from the list?*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 5\nServer List & Advertiser"]):::header --> WHEN

    WHEN{"When does it\ndisappear?"}
    WHEN -->|"After some\nhours / days"| DYN{"Dynamic IP\nfrom your ISP?"}
    WHEN -->|"Never appears\nat all"| NEVER["Server probably isn't\nreachable at all.\n➡️ Go to Chart 4\nConnection & Networking"]:::nav
    WHEN -->|"Shows as\n'incompatible'"| VER["Server + client versions\nmust match.\nRe-run SteamCMD update.\nCheck you're not on\nan old beta branch."]

    VER --> V_OK{"Fixed?"}
    V_OK -->|"Yes!"| RESOLVED
    V_OK -->|"No"| JACKSON

    DYN -->|"Yes"| ADV["Enable the SSUI Advertiser\nwith 'auto' mode.\nUses ipify to detect your\npublic IP and re-advertises\nautomatically."]
    ADV --> A_OK{"Fixed?"}
    A_OK -->|"Yes!"| RESOLVED
    A_OK -->|"No"| ERR
    DYN -->|"No / Not sure"| ERR

    ERR{"Advertiser stopped?\nCheck logs for error count."}
    ERR -->|"Yes, error\nthreshold hit"| UPDATE["Update SSUI to latest!\nOld versions killed the\nadvertiser after 5 transient\nerrors. Forever.\nFixed with incremental backoff."]
    UPDATE --> RESOLVED
    ERR -->|"No errors visible"| JACKSON

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 6 - Docker Issues

*Docker being Docker. A tale as old as containerization itself.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 6\nDocker Issues"]):::header --> WHAT

    WHAT{"What Docker problem?"}
    WHAT -->|"Sanity check\nfails"| SANITY["containerd / K8s?\nSSUI doesn't detect all runtimes.\nUse -NoSanityCheck flag\nor create /.dockerenv\nin the container."]
    WHAT -->|"Mods not\nloading"| MODS
    WHAT -->|"Files owned\nby root"| ROOT["Add user/uid/gid mapping\nin docker-compose.yml\nto match your host user."]
    WHAT -->|"CLI commands\ndon't work"| CLI["SSUI manages the server\nthrough the Web UI.\nYou don't need docker exec\nfor game server commands.\nThat's the whole point. 😉"]
    WHAT -->|"HTTP 400 on\nfirst access"| HTTPS["Are you using https:// ?\nSSUI does NOT serve HTTP.\n➡️ See Chart 2 if stuck"]:::nav
    WHAT -->|"Server works then\nstops responding"| TIMEOUT["Enable SSUI log file:\nCreateSSUILogFile: true\nin config.json.\nWait for next occurrence."]
    WHAT -->|"Players can't\nconnect"| CONN["UPnP doesn't work in Docker.\nManual port forwarding only.\n➡️ See Chart 4 for more"]:::nav

    SANITY --> S_OK{"Fixed?"}
    S_OK -->|"Yes!"| RESOLVED
    S_OK -->|"No"| JACKSON

    ROOT --> RESOLVED
    CLI --> RESOLVED
    TIMEOUT --> JACKSON

    MODS{"Is BepInEx/core\npresent in the container?"}
    MODS -->|"No / Missing"| NUKE["Nuclear option:\n1. Remove all app files\n2. docker volume prune\n3. Re-create with bind mount:\n   ./app:/app:rw\n4. Re-install SLP + mods\n5. Verify BepInEx/core exists"]
    MODS -->|"Yes"| VER["Check SLP server DLL\nmatches your game version.\nRe-copy the mod package.\nWatch for XML\ndeserialization errors."]
    NUKE --> M_OK{"Fixed?"}
    VER --> M_OK
    M_OK -->|"Yes!"| RESOLVED
    M_OK -->|"No"| JACKSON

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 7 - Discord Bot, Saves, Crashes & Everything Else™

*The grab-bag chart. If you made it here, things are getting interesting.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 7\nEverything Else™"]):::header --> CAT{"Pick your poison:"}

    CAT -->|"Discord bot\nacting weird"| DISCORD
    CAT -->|"Save / backup\nproblems"| SAVES
    CAT -->|"Server crashes"| CRASH
    CAT -->|"Something else"| ELSE

    %% --- DISCORD ---
    DISCORD{"What's the bot doing?"}
    DISCORD -->|"403 errors"| OAUTH["Regenerate the bot invite\nwith the updated OAuth URL.\nCheck channel permissions."]
    DISCORD -->|"Spamming\nmessages"| SPAM["Restart SSUI.\nUsually a Stationeers memory\nleak causing repeated events.\nClassic Rocketwerkz moment. 🙄"]
    DISCORD -->|"Retrieve password\nbroken"| BOTPW["Update to SSUI v5.13+\nRace condition was fixed."]
    DISCORD -->|"Save notifications\nmissing"| BOTSAVE["Check Discord OAuth URL\nis up to date."]
    DISCORD -->|"Other"| JACKSON

    OAUTH --> B_OK{"Fixed?"}
    BOTSAVE --> B_OK
    B_OK -->|"Yes!"| RESOLVED
    B_OK -->|"No"| JACKSON
    SPAM --> SP_OK{"Still spamming\nafter restart?"}
    SP_OK -->|"No!"| RESOLVED
    SP_OK -->|"YES"| JACKSON
    BOTPW --> RESOLVED

    %% --- SAVES ---
    SAVES{"What save issue?"}
    SAVES -->|"Server ignores\nmy save"| SNAME["Does SaveName in config\nEXACTLY match folder name?\n\n'Luna' ≠ 'Lunar'\nCase. Matters. Every. Letter."]
    SAVES -->|"Backup manager\nis empty"| SBACKUP["It only watches the\nautosaves directory.\nUse 'file save' to trigger\nautosaves instead."]
    SAVES -->|"Autosave\ncopy broken"| SBUG["Update to SSUI v5.9.1+\nBug after ~2400 autosaves.\nYes, really."]
    SAVES -->|"Commands like\n'say' don't work"| SCMD["Save name mismatch!\nSSCM needs exact world name\nfor command privileges."]
    SAVES -->|"Copied save\nfrom another server"| SCOPY{"SaveName matches\nfolder name?\nPermissions OK?"}

    SNAME --> S_OK{"Fixed?"}
    SCMD --> S_OK
    S_OK -->|"Yes!"| RESOLVED
    S_OK -->|"No"| JACKSON
    SBACKUP --> RESOLVED
    SBUG --> RESOLVED
    SCOPY -->|"Fixed it!"| RESOLVED
    SCOPY -->|"Still broken"| JACKSON

    %% --- CRASHES ---
    CRASH{"When does it crash?"}
    CRASH -->|"During save\noperations"| CSAVE["Update SSUI!\nSSCM had a threading bug:\nsaves ran off the main\nUnity thread. Fixed in PR #119."]
    CRASH -->|"On loading\na save"| CLOAD
    CRASH -->|"Random crashes\nwith mods"| CMODS
    CRASH -->|"Other / random"| COTHER["Generate a Support Package!\nWiki → Support-Packages\nEnable CreateSSUILogFile: true"]

    CSAVE --> RESOLVED
    COTHER --> JACKSON

    CLOAD{"IndexOutOfRange\nor save corruption?"}
    CLOAD -->|"IndexOutOfRange"| CTABLET["Known Stationeers bug:\ntablet cartridge switched\nduring a save operation.\nJackson has a patched DLL."]
    CLOAD -->|"Other error"| COTHER
    CTABLET --> JACKSON

    CMODS{"Crashes without\nmods too?"}
    CMODS -->|"No, mods only"| MODISSUE["Mod issue, not SSUI.\nRunning beta branch?\nGood luck, brave soul. ⚔️"]
    MODISSUE --> SKILL_ISSUE
    CMODS -->|"Yes"| COTHER

    %% --- OTHER ---
    ELSE{"Be more specific:"}
    ELSE -->|"AutoPause\ndoesn't work"| APAUSE["Stationeers game bug.\nSSUI sends the command.\nGame just... ignores it. 🫠"]
    ELSE -->|"Detection manager\nnot firing"| DETECT["Known bug.\nCheck GitHub issues."]
    ELSE -->|"Config changes\ndon't apply"| RESTART{"Restarted SSUI\nafter config change?"}
    ELSE -->|"Run 2 servers\non 1 machine"| DUAL["Change SSUIWebPort in\nconfig.json or use -p=PORT.\nChange game port in\nNetwork settings too."]
    ELSE -->|"Orphan process\nafter update"| ORPHAN["GH issue #159.\nStop server BEFORE updating.\nKill orphan manually via PID."]

    APAUSE --> SKILL_ISSUE
    DETECT --> JACKSON
    RESTART -->|"...no"| RFIX["Restart SSUI. Always\nrestart after config changes.\nALWAYS."] --> RESOLVED
    RESTART -->|"Yes!"| JACKSON
    DUAL --> RESOLVED
    ORPHAN --> RESOLVED

    %% --- TERMINALS ---
    JACKSON["➡️ Go to Chart 8\n📣 Ping Jackson"]:::nav
    RESOLVED(["✅ Crisis Averted!"]):::success
    SKILL_ISSUE(["🏆 SKILL ISSUE™\nCertified by SSUI Support Dept."]):::skillissue

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef skillissue fill:#e67e22,color:#fff,stroke:#d35400,stroke-width:4px
    classDef nav fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:2px
```

---

## Chart 8 - The Jackson Endpoint™

*You've tried everything. There's only one hope left.*

```mermaid
flowchart TD
    ENTRY(["📋 Chart 8\nThe Jackson Endpoint™"]):::header --> JACKSON

    JACKSON["📣 Ping @JacksonTheMaster\nin the SSUI Discord #support\n\nBring:\n• SSUI version number\n• A support package\n• Your config.json\n• Patience\n• Snacks (recommended)"]:::jackson

    JACKSON --> J_OK{"Did Jackson fix it?"}
    J_OK -->|"Yes! 🎉"| RESOLVED
    J_OK -->|"No response yet"| J_WAIT["He sleeps sometimes.\nApparently he's 'human'\nand has a 'life'.\nWait 24-48h."]:::wait
    J_WAIT --> J_OK2{"How about now?"}
    J_OK2 -->|"Fixed! 🎉"| RESOLVED
    J_OK2 -->|"Still nothing"| SKILL_ISSUE
    J_OK -->|"Even Jackson\ncan't fix this"| SKILL_ISSUE

    RESOLVED(["✅ Crisis Averted!\nServer goes brrr.\nGo play Stationeers."]):::success
    SKILL_ISSUE(["🏆 SKILL ISSUE™\n\nCertified by the\nSSUI Support Department.\n\nWe tried. We really did.\nMaybe try Minecraft?"]):::skillissue

    classDef header fill:#2c3e50,color:#fff,stroke:#1a252f,stroke-width:3px
    classDef jackson fill:#3498db,color:#fff,stroke:#2980b9,stroke-width:3px
    classDef wait fill:#9b59b6,color:#fff,stroke:#8e44ad,stroke-width:2px
    classDef success fill:#27ae60,color:#fff,stroke:#1e8449,stroke-width:3px
    classDef skillissue fill:#e67e22,color:#fff,stroke:#d35400,stroke-width:4px
```

---

## Quick Reference: The Flowchart Cheat Sheet

For those who don't like pictures, here's the TL;DR version:

### Tier 1: Check These First (Solves 60% of tickets)
| Symptom | Fix |
|---|---|
| Can't reach Web UI | Use `https://` not `http://`, port 8443 |
| `localhost` doesn't work on Windows | Use `127.0.0.1` instead (Hyper-V) |
| Forgot password | `authEnabled: false` in config → restart → `/setup` |
| Players can't connect | Port forward UDP 27016 on router |
| "Incompatible version" | Update game server via SteamCMD |
| Server won't start on Windows | Install VC++ Redistributable |

### Tier 2: Platform-Specific Gotchas (Solves another 25%)
| Platform | Gotcha | Fix |
|---|---|---|
| Linux | SteamCMD error 0x2 | Disable IPv6 |
| Linux | Ubuntu 22.04 too old | Upgrade to Debian 13+ / Ubuntu 24.04+ |
| Linux | 1-core VM crash | `MaxConcurrentWorkers 1` in AdditionalParams |
| Docker | UPnP doesn't work | Manual port forwarding only |
| Docker | Containerd not detected | `-NoSanityCheck` flag |
| Docker | Mods broken after volume change | Full cleanup + rebuild with bind mount |
| Docker | Files owned by root | Add `uid`/`gid` to docker-compose |
| Windows | Hyper-V eats localhost | Use `127.0.0.1` |
| Kubernetes | Player disconnects | `hostNetwork: true` in pod spec |

### Tier 3: Networking Nightmares (The hard ones)
| Issue | Fix |
|---|---|
| Behind CGNAT | VPS + FRP tunnel, or VPN |
| Double NAT | Bridge mode on modem, or forward on both devices |
| RakNet errors | Try port 50000+ range; it's a game bug |
| Reverse proxy (nginx) | `proxy_buffering off`, `proxy_cache off`, timeouts |
| Reverse proxy (Apache) | `certbot --key-type rsa` (ECDSA not supported) |
| Reverse proxy (Traefik) | `insecureSkipVerify` for self-signed cert |
| Server disappears from list | Enable SSUI Advertiser with `auto` mode |
| `LocalIpAddress` confusion | Set to `0.0.0.0`, NOT your actual IP |

### Tier 4: Ping @JacksonTheMaster
When all else fails. Bring logs, a support package, and snacks.

### Tier 5: 🏆 Skill Issue™
You've reached the end. We tried.

---

*No Jacksons were harmed in the making of this chart. 3000 GPUs worked hard on this analysis (probably)*
