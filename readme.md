# Stationeers Dedicated Server Control v2.4.1

![Go](https://img.shields.io/badge/Go-1.22.1-blue)
![License](https://img.shields.io/github/license/jacksonthemaster/StationeersServerUI)
![Platform](https://img.shields.io/badge/Platform-Windows-lightgrey)

| UI Overview | Configuration | Backup Management |
|:-----------:|:-------------:|:-----------------:|
| ![UI Overview](media/UI-1.png) | ![Configuration](media/UI-2.png) | ![Backup Management](media/UI-3.png) |

## Introduction

Stationeers Dedicated Server Control is a user-friendly, web-based tool for managing a Stationeers dedicated server. It features an intuitive retro computer-themed interface, allowing you to easily start and stop the server, view real-time server output, manage configurations, and handle backups—all from your web browser.

Additionally, it offers full Discord integration, enabling you and your community to monitor and manage the server directly from a Discord server. Features include real-time server status updates, console output, and the ability to start, stop, and restore backups via Discord commands.

**Important:** For security reasons, do not expose this UI directly to the internet without a secure authentication mechanism. Do not port forward the UI directly.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [First-Time Setup](#first-time-setup)
- [Discord Integration](#discord-integration)
  - [Features](#discord-integration-features)
  - [Setup Instructions](#discord-integration-setup)
- [Usage](#usage)
- [License](#license)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)

## Features

- Start and stop the Stationeers server with ease.
- View real-time server console output.
- Manage server configurations through a user-friendly interface.
- List and restore backups, with enhanced backup management features.
- Fully functional REST API for advanced operations (optional).
- Full Discord integration for server monitoring and management.

## Requirements

- Windows OS (tested on Windows; Linux support coming soon).
- Administrative privileges on the server machine.
- An empty folder of your choice to install the server control software.

## Installation

### Quick Installation Instructions

1. **Download and Run the Application**

   - Download the latest release executable file (`.exe`) from the [releases page](https://github.com/JacksonTheMaster/StationeersServerUI/releases).
   - Place it in an empty folder of your choice.
   - Run the executable. A console window will open, displaying output.

2. **Access the Web Interface**

   - Open your web browser.
   - Navigate to `http://<IP-OF-YOUR-SERVER>:8080`.
     - Replace `<IP-OF-YOUR-SERVER>` with the local IP address of your server. You can find this by opening Command Prompt and typing `ipconfig`.

3. **Allow External Connections (Optional)**

   - If you want others on your network to access the server UI or the gameserver, you'll need to adjust your Windows Firewall settings:
     - Go to **Control Panel > System and Security > Windows Defender Firewall**.
     - Click on **Advanced settings**.
     - Select **Inbound Rules** and click on **New Rule...**.
     - Choose **Port** and click **Next**.
     - for the gameserver, select **TCP** and enter `27015, 27016` in the **Specific local ports** field.
     - for the WebUI(This Software), select **TCP** and enter `8080` in the **Specific local ports** field.
     - Click **Next**.
     - Choose **Allow the connection** and click **Next**.
     - Select the network profiles of your choise (Domain, Private, Public) and click **Next**.
     - Name the rule (e.g., "Stationeers Server Ports") and click **Finish**.
   - **Note:** Depending on your network setup, you may need to configure port forwarding on your router to allow external connections. Please refer to your router's documentation for instructions.


## First-Time Setup

To successfully run the server for the first time, follow these steps:
Follow the Installation Instructions above.
Only turn to this section when the magenta Text in the Console tells you to do so.

1. **Prepare Your Save File**

   - Copy an existing Stationeers save folder into the `/saves` directory created during the installation.

2. **Configure the Save File Name**

   - In the web interface, click on the **Config** button.
   - Enter the name of your save folder in the **Save File Name** field.
   - You might restart the Software at this point to be sure, but it's technically not necessary.

3. **Start the Server**

   - Return to the main page of the web interface.
   - Click on the **Start Server** button.
   - The server will begin to start up, and you can monitor the console output in real-time.

## Discord Integration

### Discord Integration Features

- **Real-Time Monitoring:**
  - View server status and console output directly in Discord.
  - Receive notifications for server events such as player connections/disconnections, exceptions, and errors.
- **Server Management Commands:**
  - Start, stop, and restart the server.
  - Restore backups.
  - Ban and unban players by their Steam ID.
  - Update server files (currently supports the stable branch only).
- **Access Control:**
  - Utilize Discord's role system for granular access control over server management commands and notifications.

### Discord Notifications

The bot can send notifications for the following events:

- **Server Ready:** Notifies when the server status changes to ready.
- **Player Connection/Disconnection:** Alerts when a player connects or disconnects.
- **Exceptions and Errors:** Sends notifications when exceptions or errors are detected, including Cysharp error detection.
- **Player List:** Provides a table of connected players and their Steam IDs.

## Discord Integration Setup

1. **Create a Discord Bot**

   - Follow the instructions on [Discord's Developer Portal](https://discord.com/developers/applications) to create a new bot and add it to your Discord server.

2. **Obtain the Bot Token**

   - In the bot settings, under the **Bot** tab, copy the **Token**. Keep this token secure.

3. **Configure the Bot in the Server Control UI**

   - In the web interface, click on the **Further Setup** button.
   - Enter the bot's token in the **Discord Token** field.
   - Create a Discord Server if not already done.
   - Create a Discord Channel for the Server Control (commands), Server Status, and Server Log, and the Control Panel. Additionally, create a Discord Channel for the Error Channel.

   - Input the **Channel IDs** on the further Setup Page.
     - **Server Control Channel ID**: For sending commands to the bot.
     - **Server Status Channel ID**: For receiving server status notifications.
     - **Server Log Channel ID**: For viewing real-time console output.
     - **Control Panel Channel ID**: For the Control Panel.
     - **Error Channel ID**: For the Error Channel.
   - **Note:** To get a channel's ID, right-click on the channel in Discord and select **Copy ID**.

4. **Enable Discord Integration**

   - In the **Further Setup** page, check the **Discord Enabled** checkbox.

5. **Restart the Application**

   - Close the application and run the executable again to apply the changes.

## Usage

### Web Interface

- **Start/Stop Server:** Use the **Start Server** and **Stop Server** buttons on the main page.
- **View Server Output:** Monitor real-time console output directly in the web interface.
- **Manage Configurations:**
  - Click on the **Config** button to edit server settings.
  - Ensure all settings are correct before starting the server.
- **Backup Management:**
  - Access the **Backups** page to list and restore backups.
  - Backups are grouped and have improved deletion logic for easier management.

#### Discord Commands

| Command                       | Description                                                         |
|-------------------------------|---------------------------------------------------------------------|
| `!start`                      | Starts the server.                                                  |
| `!stop`                       | Stops the server.                                                   |
| `!restore:<backup_index>`     | Restores a backup at the specified index.                           |
| `!list:<number/all>`          | Lists recent backups (defaults to 5 if number not specified).       |
| `!ban:<SteamID>`              | Bans a player by their SteamID.                                     |
| `!unban:<SteamID>`            | Unbans a player by their SteamID.                                   |
| `!update`                     | Updates the server files if a game update is available.             |
| `!help`                       | Displays help information for the bot commands.                     |

### Important Notes

- **Do Not Expose the UI Publicly:** For security reasons, do not expose the UI directly to the internet without proper authentication mechanisms.
- **Server Updates:** Currently, only the stable branch is supported for updates via Discord commands.

## License

This project is licensed under the MIT License. See the [LICENSE](link-to-license-file) file for details.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## Acknowledgments

- **[JacksonTheMaster](https://github.com/JacksonTheMaster):** Developed with ❤️ and 💧 by J. Langisch.
- **[Go](https://go.dev/):** For the Go programming language.
- **[RocketWerkz](https://rocketwerkz.com/):** For creating the Stationeers game.
