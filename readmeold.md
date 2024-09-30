# Stationeers Dedicated Server Control v2.4.X

![Go](https://img.shields.io/badge/Go-1.22.1-blue)
![License](https://img.shields.io/github/license/jacksonthemaster/StationeersServerUI)
![Platform](https://img.shields.io/badge/Platform-Windows-lightgrey)

Stationeers Dedicated Server Control is a web-based tool for managing a Stationeers dedicated server. It offers an intuitive UI with a retro computer theme and a robust REST API for server operations, configuration management, and backup handling. Additionally, it features full Discord integration, allowing you to monitor and manage the server directly from your Discord server. This makes it easier for your community to manage restores and restarts. Information about this feature will be published soon™

I created this project to make it easier for me to manage my Stationeers server more efficiently, especially to restore backups. With the enhanced backup management features, such as grouping backups and improved backup deletion logic, managing server backups is now more straightforward than ever.
Also I wanted my friends to be able to start, stop and manage the Server without having to ask me to restore the lastest backup because some base exploded. So here we are.

DISCLAIMER: PUBLISHING THIS UI TO THE WEB SHOULD ONLY BE DONE BEHIND A SECURE AUTHENTICATION MECHANISM; THIS SHOULD NOT UNDER ANY CIRCUMSTANCES BE PORT FORWARDED STRAIGHT OUT!

## Features
| UI Overview | Configuration | Backup Management |
|:-----------:|:-------------:|:-----------------:|
| ![UI Overview](media/UI-1.png) | ![Configuration](media/UI-2.png) | ![Backup Management](media/UI-3.png) |
- Start and stop the server.
- View real-time server output.
- Manage server configurations.
- List and restore backups.
- Fully functional REST API for all operations.


## Requirements
- Window OS is tested, Linux is still not a priority but will be soon™.
- Administrative Privileges
- an Empty folder of your Choice



## Quick Installation Instrcutions for Administrators & Server Operators

1. Download & run the exe release file.
2. read the console output
3. Open your web browser and type `http://<IP-OF-YOUR-SERVER>:8080` in the address bar. Replace `<IP-OF-YOUR-SERVER>` with the actual IP address of your server. You can find this by opening the Command Prompt and typing `ipconfig`.
4. To allow other users to connect to your UI and the Server, open the Windows Firewall settings:
    - Go to Control Panel > System and Security > Windows Defender Firewall.
    - Click "Advanced settings" on the left.
    - In the Windows Firewall with Advanced Security window, click "Inbound Rules" on the left.
    - Click "New Rule..." on the right.
    - Select "Port" and click "Next".
    - Choose "TCP" and enter "27015, 27016, 8080" in the Specific local ports field. Click "Next".
    - Allow the connection and click "Next".
    - Select the network types to apply this rule (usually Domain, Private, and Public) and click "Next".
    - Name the rule something recognizable (e.g., "Stationeers Server Ports") and click "Finish".
    - __Note__:  Depending on your Setup, you might need to Port forward those ports on your router. For this, please consider using google or any other search engine exept bing to find a tutorial on how to do this.
5. Before starting your Server, ensure the configuration files on the /config and /furtherconfig page are set up correctly.



# First Time Setup

To Successfully run the Server, you need to do the following:

0. Follow the Installation Instructions until you get to the Magenta "First Time Setup" Message.
1. Copy your save folder to the /saves folder.
2. Edit the save file name in the /config page. (Config Button on the UI)
3. Start the Server with the Start Server Button on the UI.

# Discord Integration

## Discord Integration Features
- View server status and console output in real-time.
- Monitor server status and receive notifications.
- Manage server starts stops and restarts directly from Discord.
- Manage restores and backup control directly from Discord.
- Granular access control for server management commands, notifications, and the Server Log! (Just use Discord's role system)
- Get alerts for exceptions and errors, including Cysharp error detection.
- Get alerts for player connections and disconnections.
- Get a table of connected players and their Steam IDs
- Option to ban players by their Steam ID.

### Discord Commands

| Command | Description |
| --- | --- |
| !start | Starts the server.
| !stop | Stops the server.
| !restore:<number/backupindex> | Restores a backup at the specified index.
| !list:<number/all> | Lists the most recent backups. Defaults to 5 backups.
| !ban:SteamID | Bans a player by their SteamID.
| !unban:SteamID | Unbans a player by their SteamID.
| !update | Updates the server files if there is a game update available. (Currently Stable Branch only)
| !help | Displays a help message.

### Discord Notifications

| Notification | Description |
| --- | --- |
| Server Ready | When the server status changes to ready while startup, a notification will be sent
| Player Connection | When a player connects to the server, a notification will be sent 
| Player Disconnection | When a player disconnects from the server, a notification will be sent
| Exception | When an exception is detected, a notification will be sent 
| Error | When an error is detected, a notification will be sent 

## Discord Integration Setup
- *use the "further setup" button on the UI!*
- Create a Discord Bot and add it to your Discord server. You can find instructions on how to do this [here](https://chatgpt.com/).
- Obtain an OAuth2 Token for your Discord Bot.
- Add the OAuth2 Token to the "DiscordToken" field on the Further Setup Page.
- Enter the channel IDs for the server control, server status, and server log channels.
- These IDs can be found by right-clicking on a Discord channel and selecting "Copy ID".
- This can technically be all one channel, but I recommend using atlest two separate channels (seperate log channel by any means)
- Enable the Discord integration through the "Discord Enabled" checkbox on the Further Setup Page.
- Restart the Software by closing it and Executing it again.


## REST API Information

This server is based on Go, so it's basically a REST-API with some HTML files on top. All UI actions are API calls, so you can fully use the API to control the server.

### API Endpoints

- **Start Server**: `/start` (GET)
- **Stop Server**: `/stop` (GET)
- **Get Server Output**: `/output` (GET)
- **List Backups**: `/backups` (GET)
- **Restore Backup**: `/restore?index=<index>` (GET)
- **Edit Configuration**: `/config` (GET)
- **Save Configuration**: `/saveconfig` (POST Form Data)

### Form Data Explanation

- **SaveFileName**: The name of the save file to load. This is the name of the file without the extension. Example: `Mars`.
- **Settings**: The server settings. Use the UI to get the correct settings if you're unsure.

## UI

The web interface provides buttons to start and stop the server, edit configuration, and manage backups. The current server status and console output are displayed in real-time.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests!

## Acknowledgments

- [JacksonTheMaster](https://github.com/JacksonTheMaster) Developed with ❤️ and 💧 by J. Langisch.
- [Go](https://go.dev/) for the Go programming language.
- [RocketWerkz](https://github.com/RocketWerkz) for creating the Stationeers game.
