!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

StationeersServerUI offers comprehensive Discord integration, allowing you to monitor and manage your server directly from Discord.

## Features

### Real-Time Monitoring

- View server status and console output directly in Discord
- Receive notifications for server events such as:
  - Player connections/disconnections
  - Current join password
  - Server status changes
  - Version info

### Server Management Commands

- Start, stop, and restart the server
- Restore backups
- Ban and unban players by their Steam ID
- Update server files
- And more - see [Discord Commands](https://github.com/SteamServerUI/StationeersServerUI/wiki/Discord-Integration/#discord-commands)

### Access Control

- Utilize Discord's role system for granular access control
- Restrict server management commands to specific roles

## Discord Notifications

The Bot can send notifications for the following events:

- **Server Ready**: Notifies when the server status changes to ready.
- **Player Connection/Disconnection**: Alerts when a player connects or disconnects.
- **Player List**: Provides a table of connected players and their Steam IDs.
- **Request for Current Server Join Pass Code**: The Bot will retrieve and display the current server password.
- **Game Server Version**: The Bot will retrieve the current game version installed.
- **Time Until Restart**: The Bot will retrieve the time remaining until the next auto restart, if enabled.

## Setup Instructions

1. **Create a Discord Bot**

   - Head to the [Discord Developer Portal](https://discord.com/developers/applications) and log in with your Discord account.
   - Click **New Application**, give it a name (e.g., "MyStationeersBot"), and hit **Create**.
   - Go to the **Bot** section in the sidebar
   - Customize your Bot’s name or icon if you want (optional).

2. **Obtain the Bot Token**

   - In the **Bot** tab, find the **Token** section and click **Reset Token** to obtain a new one.
   - Paste it somewhere secure for later (like a private text file) - never share this, as it’s your bot’s key to Discord.

3. **Invite the Bot to Your Server**
   
   3.1 (Easy):
   - Go to the **OAuth2** sidebar section
    - Copy the **Client ID** from _Client information_
    - Edit the following URL to include the just copied ClientID where the URL says <CLIENTIDHERE> (remove the <> brackets)
      `https://discord.com/oauth2/authorize?client_id=<CLIENTIDHERE>&permissions=2147609680&integration_type=0&scope=bot`
   - Copy that URL, paste it in a browser, pick your server (you need “Manage Server” perms), and click **Authorize**.
   - Your Bot should now appear in your server’s member list (offline until you run StationeersServerUI with the token and Discord Enabled).

!!! important
    If you used the Easy way (above) there is no need to do this step the Advanced way again. Skip this step and proceed with Step 4.

   3.2 (Advanced): If you want to setup the URL yourself in the Discord Developer portal, follow these steps:
   - Proceed in the section **OAuth2 URL Generator**.
   - Under **Scopes**, check **Bot**.
   - Under **Bot Permissions**, select at least:
<img width="750" height="450" alt="image" src="https://github.com/user-attachments/assets/d6c77e5b-480c-48b9-8cc7-daa6af672319" />

   - Copy the "Permissions Integer" at the bottom of the panel
   - Go to the **OAuth2** sidebar section
    - Copy the **Client ID** from _Client information_
    - Edit the following URL to include the just copied ClientID where the URL says <CLIENTIDHERE> and the Permisson Integer where it says <PERMISSIONSINTEGER> (remove the <> brackets)
      `https://discord.com/oauth2/authorize?client_id=<CLIENTIDHERE>&permissions=<PERMISSIONSINTEGER>&integration_type=0&scope=bot`
   - Copy that URL, paste it in a browser, pick your server (you need “Manage Server” perms), and click **Authorize**.
   - Copy the generated URL, paste it in a browser, pick your server (you need “Manage Server” perms), and click **Authorize**.
   - Your Bot should now appear in your server’s member list (offline until you run StationeersServerUI with the token and Discord Enabled).


4. **Discord channels**
   
   Make new channels for each of the required Bot functions. While you *can* name the channels whatever you like, descriptive names are always a good idea.
   For Disord's help on making channels, look [here](https://support.discord.com/hc/en-us/articles/6208479917079-Forum-Channels-FAQ)

   For each channel you have the Bot active in, you need to adjust the channel's permissions to allow it to function correctly. This can be synced from the channel's Category or from the Bot's permissions at setup. If you are wanting to have unsynced permissions in your channels, you should set the permissions per channel.

   For best security, you should have different levels of permissions for general users, administration teams and Bots. The use of Discord roles is highly recommended. When referring to administration teams, it is inferred that they are not the channel owner. Listed in this section are the bare minimum permissions for the channels to function as intended. 
   For Discord's help on managing permissions, look [here](https://support.discord.com/hc/en-us/articles/6208479917079-Forum-Channels-FAQ)

!!! important
    When making these channels private, be sure to add the Bot to them.

   The Bot requires the same permissions for all the channels it's in. Use the following:
      - View Channel
      - Send Messages
      - Embed Links
      - Attach Files
      - Add Reactions
      - Manage Messages
      - Read Message History
      - Use Application Commands

!!! caution
    SSUI Discord intergration fields: For the bot to send information to your corresponding Discord channels, SSUI is expecting a Discord **ChannelID** to be entered. A *Discord ChannelID* is a string of numbers, 18 to 19 digits long. DO NOT ENTER A CHANNEL NAME - THAT WON'T WORK!

!!! tip
    To get a channel's ID, enable developer mode in discord, right-click on the channel and select **Copy ID**. Alternatively, right click on the channel then select 'Copy Link' and the last string of numbers in the URL is the ChannelID

   - **Server Status Channel**: This will display currently connected players, a button to retrieve server password, server version info and time to next auto restart - if enabled. Your players will need to interact with this channel.
   
      **User permissions:**
      - View Channel
  
      **Admin permissions:**
      - View Channel

   - **Admin Channel**: This allows for using `/` commands to interact with SSUI (see below for commands). Your admin team role will need to interact with the Bot, but your players should be restricted from accessing it.
  
      **User permissions:**
      - None
      
      **Admin permissions:**
      - View Channel
      - Use Application Commands
    
   - **Control Panel Channel**: The Bot will show a panel of quick commands activated by clicking on the emoji reactions below the panel. Your admin team role will need to interact with the Bot, but your players should be restricted from accessing it.
   
      **User permissions:**
      - None
      
      **Admin permissions:**
      - View Channel

   !!! tip
       You can combine the Admin Commands and Control Panel Channels.

   - **Game Event Log Channel**: For viewing real-time SSUI Events output.
    
      **User permissions:**
      - None
      
      **Admin permissions:**
      - View Channel
      - Read Message History (optional)

   - **Log Channel**: The backend SSUI logs will be sent here. This channel is optional and can be quite verbose. If you enable this channel, turn off notifications for it. 
   
      **User permissions:**
      - None
      
      **Admin permissions:**
      - View Channel
      - Read Message History (optional) 
   

!!! warning
    **Same Channel**: While you *could* use the same channel for *all* of SSUI's streams, it is **highly discouraged** to do so.
    The Bot will **remove previous messages**, including any user messages. This is why it is important to have dedicated channels for the Bot.

!!! tip
    For an explanation of Discord channel permissions, there a help article [here](https://support.discord.com/hc/en-us/articles/10543994968087-Channel-Permissions-Settings-101)

5. **Configure the Bot in the Server Control UI**

   Follow these simple steps to integrate your Discord Bot into SSUI:

!!! caution
    SSUI Discord intergration fields: For the bot to send information to your corresponding Discord channels, SSUI is expecting a Discord **ChannelID** to be entered. A *Discord ChannelID* is a string of numbers, 18 to 19 digits long. DO NOT ENTER A CHANNEL NAME - THAT WON'T WORK!

!!! tip
    To get a channel's ID, enable developer mode in discord, right-click on the channel and select **Copy ID**. Alternatively, right click on the channel then select 'Copy Link' and the last string of numbers in the URL is the ChannelID

   - From the SSUI dashboard, click `Edit Config`, then click on the tab with the **Discord icon** to open the **Discord Integration** page.
   - Change the `Discord integration` from `Disabled` to  `Enabled`.
   - Using the token obtained in **Section 2** above, enter the Bot's token in the **Discord Token** field.
   - In the `Channel Configuration` section, add the Discord **Channel ID** for the sections required. 
     - **Admin Command Channel** The Bot will accept `/` commands, listed below in **Discord Bot Commands** section. Restrict this to admin roles. 
     - **Control Panel Channel** A neat little control panel to start/stop/restart the game server from within Discord.
     - **Rotate Server Password** SSUI will make a new six digit join pass code every time the game server is started or restarted.
     - **Game Even Log Channel** An optional log of player joins, saves, etc. It is the output of the `Events` tab on the SSUI dashboard.
     - **Server Status Panel Channel** Status messages, current connected players and buttons for retrieving the current server join pass code, game server version and time until next auto restart.
     - **Server Log Channel** ALL Server Logs will be output here. It can be very verbose.
   
6. **Restart SSUI** or run `restartbackend ` (also has alias `rsb`) from SSUICLI

## Discord Bot Commands

| Command                       | Description                                                         |
|-------------------------------|---------------------------------------------------------------------|
| `/start`                      | Starts the server                                                  |
| `/stop`                       | Stops the server                                                   |
| `/restore <index>`            | Restores a backup at the specified index                           |
| `/list [limit]`               | Lists recent backups (defaults to 5 if number not specified)       |
| `/ban <SteamID>`              | Bans a player by their SteamID                                     |
| `/unban <SteamID>`            | Unbans a player by their SteamID                                   |
| `/help`                       | Displays help information for the Bot commands                     |
| `/command <command>`          | Allows you to run server console commands from discord             |
| `/update`                     | Runs Steamcmd to update the game server                             |

## Next Steps

- [Web Interface](Web-Interface.md) - Learn how to use the web interface
- [Security Considerations](Security-Considerations.md) - Important security notes
