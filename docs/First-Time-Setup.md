# First-time setup

You have run SSUI and reached `https://<server-address>:8443`. This guide turns the setup wizard into a server you can actually join and recover.

Keep the SSUI startup output visible for now. A native installation prints it in the console; Docker users see it in `docker compose logs ssui`. You will use the browser wizard to create the first owner before normal login is enabled.

## Before clicking through

Decide whether you are creating a new world or bringing an existing save.

### Create a new world

Choose a short save name such as `EuropaBase`. This becomes the folder identity of the world. Then choose the actual world or celestial body separately in the wizard.

### Bring an existing world

1. Stop the old game server and make a copy of the save before moving it.
2. Find the folder containing `world.xml` and `world_meta.xml`.
3. Copy that whole folder into SSUI's `saves` directory.
4. Enter the folder name exactly as **Save Name** in the wizard.
5. Select the matching **World**. This is used if Stationeers needs to create the world and for its launch arguments.

For example, if the files are under `saves/EuropaBase/world.xml`, use `EuropaBase` as the save name. Do not copy only `world.xml`; a save is a directory, not a collectible card.

!!! caution
    Do not place an irreplaceable save into a new server without retaining the original copy. SSUI backups begin after the game produces autosaves; they cannot travel backwards through time to protect the import.

## Walk through the wizard

### 1. Game branch

Choose **public** for a normal server. Beta branches are for people who deliberately want a particular Stationeers branch and accept that saves, clients and mods must agree with it.

### 2. Server name and save name

- **Server Name** is what players recognize in the server browser.
- **Save Name** identifies the save directory on disk.

They may be the same, but they do different jobs. Changing the display name later does not rename the save folder.

### 3. World

Choose Moon, Mars, Europa or another entry from the installed game's world catalog. Difficulty, start condition and start location can be adjusted later in the terrain/configuration page.

<details>
<summary>Difficulty, start conditions and modded world IDs</summary>

The selectable values come from the Stationeers installation. Exact modded IDs can be entered in the full configuration. Difficulty, start condition and start location form an ordered set: a start condition needs a difficulty, and a start location needs both. Leave optional values empty if you want the game's default behavior.

</details>

### 4. Players and game password

Set the maximum number of simultaneous players. A game password is optional and controls who can join Stationeers.

This is not your SSUI password. One opens the airlock; the other controls the building containing the airlock.

### 5. Network

For a normal first run, keep these defaults:

| Setting | First-run value | Meaning |
|---|---|---|
| Game port | `27016` | UDP port used by players |
| Local IP address | `0.0.0.0` | Listen on the host's available interfaces |
| UPnP | Disabled | Configure the firewall/router deliberately |

The old update-port setting may still appear, but `27016/UDP` is the important player port. If you change the game port, change the firewall, router forwarding and Docker mapping to match.

Leave the local address at `0.0.0.0` unless you know exactly which host interface should be used. Do not paste your router's public IP into a process that does not own that address.

<details>
<summary>Making the server reachable over the internet</summary>

Give the host a stable LAN address, allow the chosen UDP game port through its firewall, then forward that UDP port from your router to the host. Test from outside your own network; testing the public address from inside may fail on routers without NAT loopback even when the forwarding is correct.

Keep the SSUI web port (`8443/TCP`) on a trusted LAN or VPN where possible. If remote administrators truly need direct internet access, restrict source addresses and read [Access and HTTPS](Security-Considerations.md).

</details>

### 6. Create the owner

Create the first owner with a unique username and a strong password. During this initial setup window, the wizard is the trusted path for creating the first account. Do this before exposing the administration interface to the public internet.

The owner can manage users, permissions, API keys, configuration, backups and the server itself. Create narrower accounts for other people later instead of sharing this password.

If you upgraded from v5 and already had valid users, SSUI imports those users as owners and does not require a new bootstrap account. Legacy API-key users are not imported; v6 uses scoped personal access tokens.

### 7. Finish and sign in

Complete the wizard and sign in. Open **Edit Config** once before starting the game and check:

- **Auto Save** is enabled.
- **Save Interval** is `300` seconds unless you intentionally want another interval.
- **Auto Pause** matches whether an empty server should keep simulating.
- **Auto Start Server on Startup** remains off until this first launch works.

## Start the first server

1. Return to the dashboard and select **Start Server**.
2. Watch the game console through loading and session registration.
3. Wait until the server is actually joinable, not merely until a process exists.
4. Join from a Stationeers client on the same network.
5. If internet players are intended, test from outside the host network too.

If the server does not appear in the browser, try Stationeers direct connect using the host address and game port. Then check [connection troubleshooting](Troubleshooting.md) before randomly changing five network settings at once. Randomness is excellent for world seeds and terrible for diagnosis.

## Verify the first backup

Stationeers normally creates an autosave every five minutes and retains five of them. SSUI watches that autosave directory by polling, waits for a candidate to remain stable across scans and verifies its save files before archiving it.

After the world has been running long enough:

1. Open **Backups**.
2. Wait for a new archive to appear. Detection needs two stable observations at least 45 seconds apart, followed by copy and analysis time.
3. Open its details and confirm the world metadata looks plausible.
4. Download the archive and verify that it is not empty.

For a valuable world, perform one test restore while the server is new and the stakes are low. Read [restore a backup](Backup-System.md#restore-a-backup) first.

## Learn the safe stop sequence

Before shutting down your first session:

1. Send `save` through the game-command input.
2. Wait for the world-saved event.
3. Select **Stop Server**.
4. Confirm that the state becomes **Stopped**.

The Stop button does not itself guarantee a final save. Full behavior is explained in [server operations](Server-Operations.md#save-and-stop).

You now have a working SSUI installation, a joinable server and a tested path to your save. That is a considerably better milestone than "the process did not crash immediately."

---

[Documentation home](index.md) · [Server operations](Server-Operations.md) · [Backups and restores](Backup-System.md)
