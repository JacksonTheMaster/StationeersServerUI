package discordbot

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
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
	statusPanelChannelID string
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
	logger.Discord.Debug("Initial hub refresh requested")
}

// UpdateStatusPanelPlayerConnected updates the panel when a player connects
func UpdateStatusPanelPlayerConnected(username, steamID string, connectionTime time.Time, players map[string]string) {
	session := config.GetDiscordSession()
	if session == nil {
		return
	}

	if !config.GetIsDiscordEnabled() {
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
	session := config.GetDiscordSession()
	if session == nil {
		return
	}

	if !config.GetIsDiscordEnabled() {
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

// buildStatusPanelEmbed uses cached game and backup state; rendering never scans a save.
func buildStatusPanelEmbed(players map[string]string, summary *backupmgr.SaveSummary) *discordgo.MessageEmbed {
	state, color := hubServerState(gamemgr.GetServerState())
	name := strings.TrimSpace(config.GetSSUIIdentifier())
	if name == "" {
		name = config.GetServerName()
	}
	embed := &discordgo.MessageEmbed{
		Title:       "🛰️ " + shortBackupLabel(name, 200),
		Description: state,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer:      &discordgo.MessageEmbedFooter{Text: "SSUI Hub • v" + config.GetVersion() + " • Updated"},
	}
	version := config.GetExtractedGameVersion()
	if version == "" {
		version = "Not detected yet"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "🎮 Game version", Value: shortBackupLabel(version, 100), Inline: true})
	if started := gamemgr.GetServerStartTime(); !started.IsZero() {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "⏱️ Process started", Value: fmt.Sprintf("<t:%d:R>", started.Unix()), Inline: true})
	}
	if next := config.GetNextAutoRestartTime(); !next.IsZero() {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "🔄 Next restart", Value: fmt.Sprintf("<t:%d:R>", next.Unix()), Inline: true})
	}
	var names []string
	for id := range players {
		names = append(names, id)
	}
	sort.Strings(names)
	var lines strings.Builder
	for _, id := range names {
		label := strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_").Replace(shortBackupLabel(players[id], 50))
		line := fmt.Sprintf("👤 %s\n", label)
		if len(id) == 17 && !strings.ContainsFunc(id, func(r rune) bool { return r < '0' || r > '9' }) {
			line = fmt.Sprintf("👤 [%s](https://steamcommunity.com/profiles/%s/)\n", label, id)
		}
		if lines.Len()+len(line) > 850 {
			lines.WriteString("…more players connected\n")
			break
		}
		lines.WriteString(line)
	}
	value := lines.String()
	if value == "" {
		value = "_The airlock is quiet. Nobody is connected._"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("%s Crew online · %d", backupStatEmoji("players", "👥"), len(players)), Value: value,
	})
	if summary != nil {
		appendSaveStats(embed, summary)
		if !summary.SavedAt.IsZero() {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "📦 Latest archived save", Value: fmt.Sprintf("<t:%d:f> · <t:%d:R>\nStatistics reflect this save, not the live world.", summary.SavedAt.Unix(), summary.SavedAt.Unix())})
		}
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "📦 Latest archived save", Value: "Waiting for backup metadata."})
	}
	if voteField := activeVotesField(); voteField != nil {
		embed.Fields = append(embed.Fields, voteField)
	}
	if action := discordActionStatus(); action != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "🛠️ Server actions", Value: action})
	}
	return embed
}

func hubServerState(state gamemgr.ServerState) (string, int) {
	switch state {
	case gamemgr.ServerStateRunning:
		return "🟢 **Online** · Ready for your next expedition.", 0x57F287
	case gamemgr.ServerStateStopped:
		return "🔴 **Offline** · The game server is stopped.", 0xED4245
	case gamemgr.ServerStateStarting:
		return "🟡 **Starting** · Booting the game server.", 0xFEE75C
	case gamemgr.ServerStateLoadingMap:
		return "🟡 **Loading world** · Preparing your station.", 0xFEE75C
	case gamemgr.ServerStateHostingSession:
		return "🟡 **Opening session** · Almost there.", 0xFEE75C
	case gamemgr.ServerStateStopping:
		return "🟠 **Stopping** · Shutting down the session.", 0xFEE75C
	default:
		return "⚪ **Status uncertain** · No confirmed ready state yet.", 0x95A5A6
	}
}

func appendSaveStats(embed *discordgo.MessageEmbed, summary *backupmgr.SaveSummary) {
	for _, stat := range []struct {
		name, emoji, fallback string
		value                 int64
	}{
		{"Days played", "days", "🗓️", summary.DaysPlayed},
		{"Things", "things", "🧱", summary.Things},
		{"Atmospheres", "atmospheres", "🌐", summary.Atmospheres},
		{"Rooms", "rooms", "🏠", summary.Rooms},
		{"Pipe networks", "pipe_networks", "🔧", summary.PipeNetworks},
		{"Cable networks", "cable_networks", "⚡", summary.CableNetworks},
	} {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: backupStatEmoji(stat.emoji, stat.fallback) + " " + stat.name, Value: fmt.Sprintf("**%d**", stat.value), Inline: true})
	}
}

// Reuse only our own hub (or the previous SSUI status panel). Never delete
// unrelated messages when adopting a channel as the hub.
func findExistingHub(s *discordgo.Session, channelID string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return "", err
	}
	for _, message := range messages {
		if message.Author == nil || s.State == nil || s.State.User == nil || message.Author.ID != s.State.User.ID {
			continue
		}
		for _, embed := range message.Embeds {
			if embed.Footer != nil && strings.HasPrefix(embed.Footer.Text, "SSUI Hub •") {
				return message.ID, nil
			}
			if embed.Title == "🎮 Server Information" && len(message.Components) > 0 {
				return message.ID, nil
			}
		}
	}
	return "", nil
}

func refreshStatusPanel() {
	session := config.GetDiscordSession()
	if session == nil {
		return
	}

	if !config.GetIsDiscordEnabled() {
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

	if config.GetDiscordRestartVoteEnabled() || config.GetDiscordRestoreVoteEnabled() {
		buttons = append(buttons, discordgo.Button{
			Label:    "🗳️ Vote Menu",
			Style:    discordgo.SuccessButton,
			CustomID: ButtonVoteMenu,
		})
	}

	var rows []discordgo.MessageComponent
	if len(buttons) > 0 {
		rows = append(rows, discordgo.ActionsRow{Components: buttons})
	}
	return append(rows, discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{CustomID: ButtonAdminActions, Label: "Server Admin Actions", Style: discordgo.PrimaryButton, Emoji: &discordgo.ComponentEmoji{Name: "🛠️"}},
	}})
}

// sendOrEditStatusPanel sends a new message or edits the existing one
func sendOrEditStatusPanel(channelID string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	session := config.GetDiscordSession()
	if session == nil {
		return
	}

	statusPanelMutex.Lock()
	defer statusPanelMutex.Unlock()
	if channelID != statusPanelChannelID {
		statusPanelMessageID = ""
		statusPanelChannelID = channelID
	}

	if statusPanelMessageID == "" {
		var err error
		statusPanelMessageID, err = findExistingHub(session, channelID)
		if err != nil {
			logger.Discord.Warnf("Could not find the existing hub; check Read Message History permission: %v", err)
			return
		}
	}
	if statusPanelMessageID == "" {
		msg, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
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
		_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    channelID,
			ID:         statusPanelMessageID,
			Content:    &content,
			Embeds:     &embeds,
			Components: &components,
		})
		if err != nil {
			logger.Discord.Error("Error editing server status panel in channel " + channelID + ": " + err.Error())
			var apiError *discordgo.RESTError
			if !errors.As(err, &apiError) || apiError.Response == nil || apiError.Response.StatusCode != http.StatusNotFound {
				return // A rate limit or permissions failure must not create duplicate panels.
			}
			// If editing fails (e.g., message deleted), reset and try sending a new one
			statusPanelMessageID = ""
			msg, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
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
	case ButtonGetPassword, ButtonGetGameVersion, ButtonGetNextRestart, ButtonVoteMenu, voteSelectCustomID:
		if !requireHubInteraction(s, i) {
			return
		}
	default:
		return
	}

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
