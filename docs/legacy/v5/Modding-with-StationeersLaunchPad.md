!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

Modding a Stationeers server running on SSUI super easy!

SSUI already comes with BepInEx by default.

## Quick instructions:

- From the SSUI dashboard, click `Edit Config`.
- Mouse over the SLP icon in the tabs to the right.
- Click the `I Understand, Continue` Beta Feature warning button.
- Click the `Install Stationeers Launchpad` button.
  - A popup will inform you SLP is downloading.
  - A second popup will inform you its installed.
- Head back into `Launchpad Mods` and drag and drop your `modpkg<date and time>.zip` file. See below on where to get your mod package file from.
- Click the `Upload Mod Package` button.
<img width="800" height="700" alt="image" src="https://github.com/user-attachments/assets/7e406d4b-5334-48f1-a2b7-39e86b0fe341" />

- SSUI will inform you its getting the file and when its done uploading.
  - Click `Close`  and refresh.
### Congratualtions, you've just installed mods on the server!


## Where to get the modpkg file:

### Generate a mod package

!!! important
    You need to have installed [BepInEx](https://docs.bepinex.dev/articles/user_guide/installation/index.html) and [SLP](https://github.com/StationeersLaunchPad/StationeersLaunchPad) on your client computer to be able to use mods and for the following steps to work.

- Using your gaming computer (your local PC where you play Stationeers from), install Mods as usual.
- Start the game.
- Click inside the black SLP loading window at the bottom of the Stationeers startup loading screen that SLP adds (gotta be fast, 3 seconds is the default) to open the StationeersLaunchPad options UI.
- Enable/disable and reorder mods to match what you want installed on the server.

!!! tip
    Some mods are client side only and have no effect on clients joining your server.

- On the `Launchpad Configuration` tab, click `Export Mod Package` to create a zip file containing the enabled mods and config file. This will be saved to the install location of Stationeers.

## Updating Mods in SSUI

This is super simple!

- Head back into the SLP config tab as before.
- Click the `Update Workshop Mods` Button.
  - SSUI will tell you its working on them with a popup and will show when it's done. 
- Head back to the SSUI dashboard and start the game server.
- Thats it, you're done.

In the SLP config page you'll see a list of installed mods with images of the workshop title cards and the installed mod version.

## SLP Issues

Sometimes SLP doesn't update correctly. Fortunately this is a one click fix!
- Head back into the SLP config tab as before.
- Click the `Reinstall SLP` button.
  - SSUI will tell you its working on them with a popup and will show when it's done. 
- Head back to the SSUI dashboard and start the game server.
- Thats it, you're done.

## Removing Mods

Removing mods is easy enough:
- Head back into the SLP config tab as before.
- Click the `Uninstall SLP` button.
  - SSUI will tell you its working on it with a popup and will show when it's done - it's pretty quick!

All mods are now removed. Just follow the steps we've already covered to put mods back on. 

## Additional information

Should you require additional help or information, visit the SSUI Discord!
