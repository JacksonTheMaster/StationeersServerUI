!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# First-Time Setup Guide

This guide walks you through configuring and launching your Stationeers Server UI for the first time after you've completed the [Installation](Installation.md) process.

## Prerequisites

- You've downloaded and installed the Stationeers Dedicated Server Control software
- You can access the WebUI at `https://[server-ip]:8443`

## 🧙‍♂️ First-Time Setup Wizard

The Stationeers Server UI now features a sleek guided setup wizard that helps you configure your server quickly and easily. When you first access the WebUI, you'll be automatically directed to the setup process.

### Setup Wizard Features

- **User-friendly interface** with helpful placeholders at each step
- **Guided configuration** with clear instructions
- **Optional steps** - skip anything you want to configure later
- **Automatic configuration saving** as you progress through each step
- **Discord integration setup** with channel configuration
- **Network configuration** optimized for both Windows and Linux environments
- **Admin account creation** for secure server management

## Detailed Setup Process

### 1. Prepare Your Game World

You have two options for setting up a game world:

#### Option A: Create a New World

You can create a fresh world directly from the server:

- Think of a name for your new world like `MyBase` and keep it in your head for now
- The server will automatically generate a new world with this name when you start it for the first time.

#### Option B: Use an Existing Save File

If you already have a Stationeers world you want to use:

- Locate your existing Stationeers save folder (typically found in `Documents\My Games\Stationeers\saves` on Windows)
- Copy the entire save folder into the `/saves` directory created during installation
- Make sure the first letter of your Save foler name is capitalized and **has no spaces** (`MyBase` instead of `mybase` or `stationeers euroa base`)
- Make note of the exact folder name as you'll need it in the configuration step

### 1. Accessing the Setup Wizard

When you first access the WebUI at `https://[server-ip]:8443`, you'll automatically be directed to the setup wizard. If you need to re-open the wizard, go to `/setup`, click the Wizard button on the config page or use the logout button If authentication is not yet configured, and you'll go straight to the setup process.

### 2. Welcome Screen

The wizard begins with a welcome screen. Click **Start Setup** to begin the configuration process.

### 3. Basic Server Configuration
!!! tip
    You do not have to configure all fields. You can skip any that you don't need, or skip the wizard entirely and configure later through the Config page.

1. **Server Name**
   - Enter a descriptive name for your server as it will appear in the server browser
   - Spaces are allowed (e.g., "Space Station 13")

2. - **Save Identifier**: 
      - Enter the exact name of your save folder (e.g., `MyBase`)
     - If using an existing save, this must match your folder name exactly
     - The first letter of your Save foler name MUST be capitalized and **have no spaces** (`MyBase` instead of `mybase` or `stationeers euroa base`) for Stationeers to recognize it.
     - If creating a new world, enter your desired world type after the save name (e.g., `MyBase Moon`)
     - Choose from Moon, Mars, Europa, etc. [About World Types (Stationeers Wiki)](https://stationeers-wiki.com/Dedicated_Server_Guide)

!!! important
    The first letter of your Save foler name MUST be capitalized and **have no spaces** (`MyBase` instead of >`mybase` or `stationeers euroa base`) for Stationeers to recognize it.

3. **Player Limit**
   - Set the maximum number of concurrent players allowed
   - Default suggestion is 8 players

4. **Server Password**
   - Set an optional password to restrict access to your server
   - Leave blank for a public server

5. **Game Branch**
   - Enter a beta branch name or skip to use the release version
   - Remember to restart SSUI after completing the wizard if changing branches

### 4. Discord Integration (Optional)

See [Discord Integration](Discord-Integration.md) for more information.

1. **Enable Discord**
   - Type 'yes' to enable Discord integration or skip to disable
   
2. **Discord Bot Token**
   - Enter your Discord bot token for server integration
   
3. **Discord Channel Setup**
   - Configure up to 6 different channel IDs for various functions:
     - Control Panel Channel
     - Save Channel
     - Log Channel
     - Connection List Channel
     - Status Channel
     - Control Channel

### 5. Network Configuration

!!! caution
    Local IP Address is required for Linux servers.

1. **Network Configuration Choice**
   - Choose whether to configure network settings or use defaults
   - Network configuration is especially important for Linux servers
   
2. **Game Port**
   - Set the port for game connections (default: 27016)
   
3. **Update Port**
   - Set the port for update connections (default: 27015)
   
4. **UPnP Configuration**
   - Enable or disable UPnP for automatic port forwarding
   
5. **Server Visibility**
   - Choose whether your server appears in the public server list
   
6. **Steam P2P Configuration**
   - Enable or disable Steam P2P networking
   
7. **Local IP Address**
   - Set the server's local IP address (format: 0.0.0.0, no CIDR notation)
   - **REQUIRED** for Linux, recommended for Windows

### 6. Admin Account Setup

1. **Create Admin Credentials**
   - Set up a username and password for administrative access
   - This secures your server from unauthorized access
   - You can add additional user accounts later through `/adduser`
      - There is no real button besides the link on the APIinfo page for this, but it exists. :)

### 7. Finalize Setup

Once you've completed all the steps (or skipped those you don't need), you'll reach the finalization screen. Your configuration has been saved throughout the process.

- Click **Return to Start** if you want to revisit any settings
- All options can be changed later through the Config tab in the WebUI

## Starting Your Server

After completing the setup wizard:

1. **Access the Dashboard**
   - You'll be directed to the main dashboard after setup completion
   - Get familiar with the layout and available options

2. **Launch the Server**
   - Click the **Start Server** button on the dashboard
   - The server will initialize and load your world
   - Monitor startup progress in the real-time console output display

3. **Verify Server Status**
   - Watch for the "Server started successfully" message in the console
   - The server should now be visible in the Stationeers server browser in-game

## Connect to Your Server

1. **From the Same Network**:
   - Launch Stationeers
   - Go to Join Game
   - Your server should appear in the list (if not, use Direct Connect)

2. **From Outside Your Network**:
   - Make sure you've set up port forwarding on your router for selected ports (TCP)
     - If you want others to access the server UI or the gameserver, you'll need to adjust your Windows Firewall settings as well:
     - Go to **Control Panel > System and Security > Windows Defender Firewall**.
     - Click on **Advanced settings**.
     - Select **Inbound Rules** and click on **New Rule...**.
     - Choose **Port** and click **Next**.
     - For the gameserver, select **TCP** and enter `27015, 27016` or the ports you configured in the **Specific local ports** field.
     - For the WebUI (This Software), select **TCP** and enter `8443` in the **Specific local ports** field.
     - Click **Next**.
     - Choose **Allow the connection** and click **Next**.
     - Select the network profiles of your choice (Domain, Private, Public) and click **Next**.
     - Name the rule (e.g., "Stationeers Server Ports") and click **Finish**.
   - **Note:** Depending on your network setup, you may need to configure port forwarding on your router to allow external connections. Please refer to your router's documentation for instructions.

## 🔐 Authentication & Multi-User Support

The Stationeers Server UI now supports multiple user accounts for better security and team management:

- **Admin Account Creation** during first-time setup
- **Add Additional Users** through the WebUI on the `/changeuser` page
- **User Management** for updating passwords or removing accounts is also done on the `/changeuser` page
- **Secure Access Control** to protect your server from unauthorized access

To manage users after initial setup:
1. Navigate to the user management section in the WebUI
2. Add new users or update existing credentials
3. Assign appropriate permissions as needed

## Troubleshooting Common Issues

### Server Won't Start

- Check the console output on the UI and the CLI for error messages
- Verify the save file name matches exactly (including capitalization)
- Ensure you have sufficient disk space, CPU power, and RAM available
- Try deleting the UIMod folder and restarting the control software, enabling a redownload of the UIMod files.
**The Stationeers Server on Windows** requires [The Visual C++ Redistributable](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist?view=msvc-170) on Windows wich is included in some windows versions - make sure to install it manually to be sure.

### Can't Connect to Server

- Set the Local IP Address to the correct IP address for your server.
- Verify the server is registered with [RocketNet's backend](http://40.82.200.175:8081/List)
- For local network: Make sure your firewall allows connections to selected ports
- For external connections: Verify your port forwarding is set up correctly
- Check if the server has a password that players need to enter

### World Generation Failed

- Try using "Moon Moon" as the save name.
- Ensure you have the latest version of the control software
- Read the logs for more information.
- Open an issue on the [GitHub Issues](https://github.com/jacksonthemaster/StationeersServerUI/issues) page.

## Next Steps

After completing your initial setup, consider exploring these additional features:

- [Security Considerations](Security-Considerations.md)
- [Docker Guide](Docker-Guide.md) - For containerized deployment options
