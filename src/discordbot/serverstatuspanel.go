package discordbot

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/bwmarrin/discordgo"
)

// Button custom IDs for the server & players panel
const (
	ButtonGetPassword       = "ssui_get_password"
	ButtonGetGameVersion    = "ssui_get_game_version"
	ButtonGetNextRestart    = "ssui_get_next_restart"
	ButtonVoteMenu          = "ssui_vote_menu"
	ButtonDownloadBackupPfx = "ssui_download_backup_" // Prefix for download backup button
)

var (
	statusPanelMessageID string // tracks the message ID for editing the server status panel
	statusPanelMutex     sync.Mutex
)

// sendServerStatusPanel sends the initial now combined server info + players panel on startup
func sendServerStatusPanel() {
	if !config.GetIsDiscordEnabled() {
		return
	}

	channelID := config.GetStatusPanelChannelID()
	if channelID == "" {
		logger.Discord.Debug("Status panel channel ID not configured, skipping panel")
		return
	}

	refreshStatusPanel()
	clearMessagesAboveLastN(channelID, 1)
	logger.Discord.Info("Server status panel sent successfully")
}

// UpdateStatusPanelPlayerConnected updates the panel when a player connects
func UpdateStatusPanelPlayerConnected(username, steamID string, connectionTime time.Time, players map[string]string) {
	if !config.GetIsDiscordEnabled() || config.DiscordSession == nil {
		logger.Discord.Debug("Discord not enabled or session not initialized")
		return
	}
	channelID := config.GetStatusPanelChannelID()
	if channelID == "" {
		return
	}
	setStatusPanelPlayers(players)
	refreshStatusPanel()
}

// UpdateStatusPanelPlayerDisconnected updates the panel when a player disconnects
func UpdateStatusPanelPlayerDisconnected(steamID string, players map[string]string) {
	if !config.GetIsDiscordEnabled() || config.DiscordSession == nil {
		logger.Discord.Debug("Discord not enabled or session not initialized")
		return
	}
	channelID := config.GetStatusPanelChannelID()
	if channelID == "" {
		return
	}
	setStatusPanelPlayers(players)
	refreshStatusPanel()
}

// buildStatusPanelEmbed creates a combined server info + connected players embed
func buildStatusPanelEmbed(players map[string]string, summary *backupmgr.SaveSummary) *discordgo.MessageEmbed {
	serverName := config.GetServerName()

	embed := &discordgo.MessageEmbed{
		Title:     "🎮 Server Information",
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x5865F2, // Discord blurple
	}

	// Server info in description + inline fields (matches old style)
	//embed.Description = "Your Stationeers server is running as **" + serverName + "**"
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Server Name",
			Value:  serverName,
			Inline: true,
		},
		{
			Name:   "SSUI Version",
			Value:  config.GetVersion(),
			Inline: true,
		},
	}

	// Connected players section
	if len(players) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "👥 Connected Players",
			Value: "_No players are currently connected._",
		})
	} else {
		var lines strings.Builder
		for steamID, username := range players {
			fmt.Fprintf(&lines, "👤 [%s](https://steamcommunity.com/profiles/%s/)\n", username, steamID)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("👥 Connected Players — %d online", len(players)),
			Value: lines.String(),
		})
		embed.Color = 0x57F287 // Green when players are online
	}

	if summary != nil {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: backupStatEmoji("days", "🗓️") + " Days Played", Value: fmt.Sprintf("**%d**", summary.DaysPlayed), Inline: true},
			&discordgo.MessageEmbedField{Name: backupStatEmoji("things", "🧱") + " Things", Value: fmt.Sprintf("**%d**", summary.Things), Inline: true},
			&discordgo.MessageEmbedField{Name: backupStatEmoji("atmospheres", "🌐") + " Atmospheres", Value: fmt.Sprintf("**%d**", summary.Atmospheres), Inline: true},
			&discordgo.MessageEmbedField{Name: backupStatEmoji("rooms", "🏠") + " Rooms", Value: fmt.Sprintf("**%d**", summary.Rooms), Inline: true},
			&discordgo.MessageEmbedField{Name: backupStatEmoji("pipe_networks", "🔧") + " Pipe Networks", Value: fmt.Sprintf("**%d**", summary.PipeNetworks), Inline: true},
			&discordgo.MessageEmbedField{Name: backupStatEmoji("cable_networks", "⚡") + " Cable Networks", Value: fmt.Sprintf("**%d**", summary.CableNetworks), Inline: true},
		)
	}
	if voteField := activeVotesField(); voteField != nil {
		embed.Fields = append(embed.Fields, voteField)
	}

	return embed
}

func refreshStatusPanel() {
	if !config.GetIsDiscordEnabled() || config.DiscordSession == nil {
		return
	}
	channelID := config.GetStatusPanelChannelID()
	if channelID == "" {
		return
	}
	players, summary := statusPanelSnapshot()
	sendOrEditStatusPanel(channelID, buildStatusPanelEmbed(players, summary), buildPanelComponents())
}

// buildPanelComponents returns the action row with interactive buttons
func buildPanelComponents() []discordgo.MessageComponent {
	var buttons []discordgo.MessageComponent

	if config.GetServerPassword() != "" {
		buttons = append(buttons, discordgo.Button{
			Label:    "🔑 Get Server Password",
			Style:    discordgo.PrimaryButton,
			CustomID: ButtonGetPassword,
		})
	}

	buttons = append(buttons, discordgo.Button{
		Label:    "🎮 Get Game Version",
		Style:    discordgo.SecondaryButton,
		CustomID: ButtonGetGameVersion,
	})

	if config.GetAutoRestartServerTimer() != "0" && config.GetAutoRestartServerTimer() != "" {
		buttons = append(buttons, discordgo.Button{
			Label:    "🔄 Next Auto Restart",
			Style:    discordgo.SecondaryButton,
			CustomID: ButtonGetNextRestart,
		})
	}
	if config.GetDiscordRestartVoteEnabled() || config.GetDiscordRestoreVoteEnabled() {
		buttons = append(buttons, discordgo.Button{
			Label:    "🗳️ Vote Menu",
			Style:    discordgo.SuccessButton,
			CustomID: ButtonVoteMenu,
		})
	}

	if len(buttons) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: buttons,
		},
	}
}

// sendOrEditStatusPanel sends a new message or edits the existing one
func sendOrEditStatusPanel(channelID string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	statusPanelMutex.Lock()
	defer statusPanelMutex.Unlock()

	if statusPanelMessageID == "" {
		msg, err := config.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		})
		if err != nil {
			logger.Discord.Error("Error sending server status panel to channel " + channelID + ": " + err.Error())
			return
		}
		statusPanelMessageID = msg.ID
		logger.Discord.Debug("Sent server status panel to channel " + channelID)
	} else {
		embeds := []*discordgo.MessageEmbed{embed}
		content := ""
		_, err := config.DiscordSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    channelID,
			ID:         statusPanelMessageID,
			Content:    &content,
			Embeds:     &embeds,
			Components: &components,
		})
		if err != nil {
			logger.Discord.Error("Error editing server status panel in channel " + channelID + ": " + err.Error())
			// If editing fails (e.g., message deleted), reset and try sending a new one
			statusPanelMessageID = ""
			msg, err := config.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			})
			if err != nil {
				logger.Discord.Error("Error sending fallback server status panel to channel " + channelID + ": " + err.Error())
			} else {
				statusPanelMessageID = msg.ID
				logger.Discord.Debug("Sent new server status panel after edit failure to channel " + channelID)
			}
		}
	}
}

// handlePanelButtonInteraction handles button interactions from the combined panel
func handlePanelButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID

	switch customID {
	case ButtonGetPassword:
		handleGetPasswordButton(s, i)
	case ButtonGetGameVersion:
		handleGetGameVersionButton(s, i)
	case ButtonGetNextRestart:
		handleGetNextRestartButton(s, i)
	case ButtonVoteMenu:
		handleVoteMenuButton(s, i)
	case voteSelectCustomID:
		handleVoteSelection(s, i)
	default:
		return
	}
}

// handleGetPasswordButton sends the current server password as an ephemeral message
func handleGetPasswordButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	password := config.GetServerPassword()

	var embed *discordgo.MessageEmbed
	if password == "" {
		embed = &discordgo.MessageEmbed{
			Title:       "🔓 No Password Set",
			Description: "No password is currently configured for this server.",
			Color:       0xFFA500,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "This message will disappear in 30 seconds",
			},
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "🔑 Current Server Password",
			Description: "Use this password to connect to the server.",
			Color:       0x57F287,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Password",
					Value:  "```" + password + "```",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "🔒 Only visible to you • Disappears in 30 seconds",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to get password button: " + err.Error())
		return
	}

	go func() {
		time.Sleep(30 * time.Second)
		err := s.InteractionResponseDelete(i.Interaction)
		if err != nil {
			logger.Discord.Debug("Could not delete ephemeral password message: " + err.Error())
		}
	}()
}

// handleGetGameVersionButton sends the current game server version as an ephemeral message
func handleGetGameVersionButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	version := config.GetExtractedGameVersion()

	var embed *discordgo.MessageEmbed
	if version == "" {
		embed = &discordgo.MessageEmbed{
			Title:       "🎮 Game Version Unknown",
			Description: "The game server version has not been detected yet.",
			Color:       0xFFA500,
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title: "🎮 Game Server Version",
			Color: 0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Version",
					Value:  "```" + version + "```",
					Inline: false,
				},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to game version button: " + err.Error())
	}
}

// handleGetNextRestartButton sends the next scheduled auto-restart time as an ephemeral message
func handleGetNextRestartButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	nextRestart := config.GetNextAutoRestartTime()

	var embed *discordgo.MessageEmbed
	if nextRestart.IsZero() {
		embed = &discordgo.MessageEmbed{
			Title:       "🔄 No Restart Scheduled",
			Description: "No auto-restart is currently scheduled.",
			Color:       0xFFA500,
		}
	} else {
		unixTS := nextRestart.Unix()
		embed = &discordgo.MessageEmbed{
			Title:       "🔄 Next Auto Restart",
			Description: "Times below are shown in your local (Discord) timezone.",
			Color:       0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Scheduled Time",
					Value:  fmt.Sprintf("<t:%d>", unixTS),
					Inline: true,
				},
				{
					Name:   "Countdown",
					Value:  fmt.Sprintf("<t:%d:R>", unixTS),
					Inline: true,
				},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to next restart button: " + err.Error())
	}
}
