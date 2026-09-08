package discordbot

import (
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"

	"github.com/bwmarrin/discordgo"
)

// registerSlashCommands defines and registers slash commands if they have not been registered already.
func registerSlashCommands(s *discordgo.Session, applicationID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "start",
			Description: "Start the server",
		},
		{
			Name:        "stop",
			Description: "Stop the server",
		},
		{
			Name:        "restart",
			Description: "Restart the game server (confirmation required)",
		},
		{
			Name:        "status",
			Description: "Gets the running status of the gameserver process",
		},
		{
			Name:        "help",
			Description: "Show command help",
		},
		{
			Name:        "update",
			Description: "Update the gameserver via SteamCMD. Feedback will take a while, please be patient.",
		},
		{
			Name:        "command",
			Description: "send a console command to the gameserver via SSCM",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Console Command to send to the gameserver. Has NO command syntax validation!",
					Required:    true,
				},
			},
		},
		{
			Name:        "announce",
			Description: "Send an announcement message to all in-game players)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "The announcement text to broadcast to players",
					Required:    true,
				},
			},
		},
		{
			Name:        "restore",
			Description: "Restore a backup at the specified filename",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Backup filename from /list",
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "List the most recent backups",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "limit",
					Description: "Number of backups to browse or 'all' (default: all, paginated)",
					Required:    false,
				},
			},
		},
		{
			Name:        "download",
			Description: "Download a backup file (most recent if no name given)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Backup filename from /list (default: most recent)",
					Required:    false,
				},
			},
		},
		{
			Name:        "bansteamid",
			Description: "Bans a player by their SteamID. Needs a Server restart to take effect.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "steamid",
					Description: "SteamID to ban",
					Required:    true,
				},
			},
		},
		{
			Name:        "unbansteamid",
			Description: "Unbans a player by their SteamID. Needs a Server restart to take effect.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "steamid",
					Description: "SteamID to unban",
					Required:    true,
				},
			},
		},
	}

	logger.Discord.Info("Checking and registering slash commands with Discord...")
	for _, command := range commands {
		dmPermission := false
		command.DMPermission = &dmPermission
	}

	// Fetch existing commands from Discord
	existingCmds, err := s.ApplicationCommands(applicationID, "")
	if err != nil {
		logger.Discord.Error("Failed to fetch existing commands: " + err.Error())
		return
	}

	// Map existing commands by name for quick lookup
	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCmds {
		existingMap[cmd.Name] = cmd
	}

	// Compare and register only what’s necessary
	for _, desiredCmd := range commands {
		existing, exists := existingMap[desiredCmd.Name]
		needsUpdate := !exists || !commandsAreEqual(desiredCmd, existing)

		if !needsUpdate {
			logger.Discord.Debug("Command " + desiredCmd.Name + " already up-to-date, skipping")
			continue
		}

		startTime := time.Now()
		_, err := s.ApplicationCommandCreate(applicationID, "", desiredCmd)
		duration := time.Since(startTime)
		if err != nil {
			logger.Discord.Error("Error registering command " + desiredCmd.Name + ": " + err.Error())
		} else {
			logger.Discord.Debug("Successfully registered command " + desiredCmd.Name + " took:" + duration.String())
		}
	}

	logger.Discord.Info("Finished processing slash commands.")
}

// This is used to determine if a slash command needs to be registered with the discord server we are connected to or if it already exists.
// commandsAreEqual (helper) checks if two discrd commands are functionally identical
func commandsAreEqual(desired, existing *discordgo.ApplicationCommand) bool {
	if (desired.DMPermission == nil) != (existing.DMPermission == nil) {
		return false
	}
	if desired.DMPermission != nil && *desired.DMPermission != *existing.DMPermission {
		return false
	}
	if desired.Name != existing.Name || desired.Description != existing.Description {
		return false
	}

	// Compare options (nil vs empty slice handling)
	if len(desired.Options) != len(existing.Options) {
		return false
	}

	for i, desiredOpt := range desired.Options {
		existingOpt := existing.Options[i]
		if desiredOpt.Type != existingOpt.Type ||
			desiredOpt.Name != existingOpt.Name ||
			desiredOpt.Description != existingOpt.Description ||
			desiredOpt.Required != existingOpt.Required {
			return false
		}
	}

	return true
}
