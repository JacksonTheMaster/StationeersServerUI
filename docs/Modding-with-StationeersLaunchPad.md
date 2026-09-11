# LaunchPad and Workshop mods

SSUI can install [StationeersLaunchPad](https://github.com/StationeersLaunchPad/StationeersLaunchPad), import an exported mod package and download Workshop items. Start with a working unmodded server so a mod failure is distinguishable from an installation failure.

## Install a mod set

1. Save and stop Stationeers. Copy the world and any existing mod configuration before changing them.
2. On your game client, prepare and test the desired mods using the current [LaunchPad instructions](https://github.com/StationeersLaunchPad/StationeersLaunchPad). Export the mod package from LaunchPad's configuration interface.
3. In SSUI, open **Edit Config → LaunchPad Mods** and select **Install StationeersLaunchPad**. Wait for completion.
4. Upload the exported ZIP through **Upload Mod Package**.
5. Check the installed mod list, then start the game and watch BepInEx/SLP output. Join with the compatible client mod set.

!!! caution
    Package import replaces the existing `mods` contents and removes `modconfig.xml` before extraction. It is not an additive merge or an atomic rollback operation. Keep the previous package and files so a failed import does not leave you improvising.

Server and client compatibility depends on the individual mods. Some mods are client-only. Uploading a ZIP does not turn those into server mods through force of optimism.

## Add or update Workshop items

Stop the game. In **Add Workshop Mods**, paste numeric Workshop IDs or full Steam Workshop URLs, separated by spaces, commas or new lines. Submit once and inspect the result.

SSUI validates and deduplicates entries, downloads them through SteamCMD and copies successful items into the mod directory. A batch can partly succeed and still return an error. Read the failed-item output before retrying; an aggregate error does not undo successful installs.

Use **Update Workshop Mods** for the installed Workshop set. Start the game afterwards and verify loading and client compatibility.

## Automatic LaunchPad updates

`IsStationeersLaunchPadAutoUpdatesEnabled` defaults to true. SSUI applies it to `CheckForUpdate` and `AutoUpdateOnStart` in `BepInEx/config/stationeers.launchpad.cfg` when that file exists. It controls LaunchPad's own startup updater; it is not SSUI's Workshop batch-download switch.

Keep a known working package and world backup when permitting automatic updates. Mods execute code inside the game; use sources you trust.

## Repair or remove

| Action | Files affected |
|---|---|
| Reinstall SLP | Replaces `BepInEx/plugins/StationeersLaunchPad`; preserves `mods` and `modconfig.xml` |
| Uninstall SLP | Removes the SLP plugin, `mods` and `modconfig.xml` |
| Import package | Clears existing mods/configuration and imports the package |

Stop the game and back up first. Removing a mod can make an existing world unloadable or alter its objects. Consult the mod's instructions before loading that world without it.

## Diagnose a failure

For an import error, preserve the ZIP and inspect SSUI's first extraction/import error. For Workshop failures, check each ID, Steam access, disk space and write permissions. For game startup failures, read the console from the first BepInEx message and compare game/SLP/mod versions and load order.

Restore the previous mod set before restoring a world that depends on it. [Support packages](Support-Packages.md) help with SSUI errors; failures inside a specific mod may need its author.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
