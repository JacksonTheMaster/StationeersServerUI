# Rework in Progress: Stationeers Dedicated Server Control v1.1

![Go](https://img.shields.io/badge/Go-1.22.1-blue)
![License](https://img.shields.io/github/license/jacksonthemaster/StationeersServerUI)
![Platform](https://img.shields.io/badge/Platform-Windows-lightgrey)

Stationeers Dedicated Server Control is a web-based tool for managing a Stationeers dedicated server. It offers an intuitive UI and a robust REST API for server operations, configuration management, and backup handling.
I created this project to make it easier for me to manage my Stationeers server more efficiently, especially to restore backups.
I found that the Stationeers server was not very user-friendly, and I wanted to create a tool that would make it easier to manage the server.
Also I wanted my friends to be able to start, stop and manage the Server without having to ask me to restore the lastest backup because some base exploded. So here we are.

DISCLAIMER: PUBLISHING THIS UI TO THE WEB SHOULD ONLY BE DONE BEHIND A SECURE AUTHENTICATION MECHANISM; THIS SHOULD NOT UNDER ANY CIRCUMSTANCES BE PORT FORWARDED STRAIGHT OUT!

## Coming Features & Current Bugs

So, in the land of "things I didn't want to happen," I've discovered a lovely little bug in the backup system that's causing—wait for it—*backups to disappear*. Yes, exactly what you want in a backup system, right? Now, whether this catastrophe is entirely my fault or a byproduct of the game being... itself... remains to be seen. But, while I channel my inner ostrich (head firmly in the sand), I've slapped together a *temporary* production fix on our server, SpaceInc: backups will be more granular going forward. Yay, progress!

Now, here's the fun part. The software is going to copy any backup to a secure location on your disk, all neatly segregated by date like some sort of highly organized digital filing cabinet. Restoring them? Oh yeah, that'll be more granular too. You'll be able to sort them by date and feel really smug about it.

Oh, and I made that Discord integration: It’s coming along. Expect a control channel for starting, stopping, and restoring the server, along with other channels for logs, saves, and player data, all nicely formatted and ready for your Discord flex. Plus, live updates for connected players because we all need to know how many people are (or aren't) using the server, right?

## When is "The Future"?

You want to know when all of this magic will be available? Well, let’s just say it’ll be ready "soon™." As someone with a house, a wife, and approximately zero free time, this whole development thing gets... tricky. The changes are mostly done in a super-secret private repo, and I reckon I need about 5 to 15 hours to wrap this thing up. But hey, who's counting?

## Linux Support

For all my Linux people out there: yes, it’s coming. But, as of right now, I’m relying on Powershell to execute the game server, so... yeah. That’s pretty much it. Stay tuned for when the Linux fairies finally decide to bless this project.
I'd say you can expect a Docker release then aswell. Because who wouldn't make a Docker release then.

## Features
| UI Overview | Configuration | Backup Management |
|:-----------:|:-------------:|:-----------------:|
| ![UI Overview](media/UI-1.png) | ![Configuration](media/UI-2.png) | ![Backup Management](media/UI-3.png) |

- Start and stop the server (because sometimes, the server just needs a break too).
- View real-time server output (so you can stare at the logs and pretend you know what's going on).
- Manage server configurations
- List and restore backups
- Fully functional REST API for all operations (because who doesn’t love APIs, right..?).

### Coming *Soon™*:
- Discord integration with a shiny control channel, log channel, save channel, and player log channel (all the channels you never knew you needed).
- Granular backups with date-sorted directories, because losing unsaved progress is *so 2023*.
- Linux support (yes, it’s real, and yes, it’s happening—just not right now, *soon™*)

## Requirements
- Windows OS
- Downloaded and installed the Stationeers Dedicated Server.
- Administrative Privileges (Hostesly i havnt tested it without running as admin, but  it..should'nt run without..?!

## Quick Installation Instrcutions for Administrators & Server Operators

1. Download & Extract release ZIP from GitHub.
2. Move "startStatoneersServerUI.exe" and the "UIMod" folder to the server's executable directory.
3. Run "startStatoneersServerUI.exe". (If you start "UIMod/Stationeers-ServerUI.exe", the Server wont auto restart)
4. Access UI at `http://<server-ip>:8080`.
5. Open firewall ports 27015, 27016, 8080.
6. Check /config before starting the server.


## Detailed Installation Instrcutions for "Normal" Windows Users

1. Go to the link: https://github.com/jacksonthemaster/StationeersServerUI/releases.
2. Find the latest release and click to download the ZIP file.
3. Once downloaded, locate the ZIP file, right-click on it, and select "Extract All...".
4. Choose a folder where you want to save the extracted files and click "Extract".
5. Open the folder with the extracted files and locate "startStatoneersServerUI.exe".
6. Cut (Ctrl+X) or copy (Ctrl+C) "startStatoneersServerUI.exe".
7. Navigate to the folder where you have installed your Stationeers Dedicated Server.
8. Paste (Ctrl+V) "startStatoneersServerUI.exe" into this folder.
9. Go back to the extracted files folder and find the "UIMod" folder.
10. Cut (Ctrl+X) or copy (Ctrl+C) the "UIMod" folder.
11. Paste (Ctrl+V) the "UIMod" folder into the same folder where your Stationeers Dedicated Server executable is located.
13. Double-click "startStatoneersServerUI.exe" to run it. Do not run "UIMod/Stationeers-ServerUI.exe" unless you DONT want the server to auto restart.
14. Open your web browser and type `http://<IP-OF-YOUR-SERVER>:8080` in the address bar. Replace `<IP-OF-YOUR-SERVER>` with the actual IP address of your server. You can find this by opening the Command Prompt and typing `ipconfig`.
15. To allow other users to connect to your UI and the Server, open the Windows Firewall settings:
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
16. Before starting your server, ensure the configuration files on the /config page are set up correctly.


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
