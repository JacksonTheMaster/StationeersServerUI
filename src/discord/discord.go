package discord

import (
	"StationeersServerUI/src/config"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func StartDiscordBot() {
	var err error
	config.DiscordSession, err = discordgo.New("Bot " + config.DiscordToken)
	fmt.Println("Discord token:", config.DiscordToken)
	fmt.Println("ControlChannelID:", config.ControlChannelID)
	fmt.Println("StatusChannelID:", config.StatusChannelID)
	fmt.Println("ConnectionListChannelID:", config.ConnectionListChannelID)
	fmt.Println("LogChannelID:", config.LogChannelID)
	fmt.Println("SaveChannelID:", config.SaveChannelID)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}
	fmt.Println("Bot is now running and connected")

	config.DiscordSession.AddHandler(messageCreate)
	config.DiscordSession.AddHandler(reactionAddHandler)

	err = config.DiscordSession.Open()
	if err != nil {
		fmt.Println("Error opening Discord connection:", err)
		return
	}

	fmt.Println("Bot is now running.")
	// Start the buffer flush ticker to send the remaining buffer every 5 seconds
	config.BufferFlushTicker = time.NewTicker(5 * time.Second)
	sendMessageToStatusChannel("🤖Bot v1.33 SpaceInc. Prod-Release connected to Discord.")
	go func() {
		for range config.BufferFlushTicker.C {
			flushLogBufferToDiscord()
		}
	}()
	SendControlMessage()
	select {} // Keep the program running
}

func checkForKeywords(logMessage string) {
	// List of keywords to detect and their corresponding messages
	keywordActions := map[string]string{
		"Ready": "🔔Server is ready to connect!",
		// "No clients connected.": "No clients connected. Server is going into idle mode!",
	}

	// Iterate through the keywordActions map
	for keyword, actionMessage := range keywordActions {
		if strings.Contains(logMessage, keyword) {
			sendMessageToStatusChannel(actionMessage)
		}
	}

	// Detect more complex patterns using regex
	complexPatterns := []struct {
		pattern *regexp.Regexp
		handler func(matches []string)
	}{
		{
			// Example: "Client Jacksonthemaster (76561198334231312) is ready!"
			pattern: regexp.MustCompile(`Client\s+(.+)\s+\((\d+)\)\s+is\s+ready!`),
			handler: func(matches []string) {
				username := matches[1]
				steamID := matches[2]
				message := fmt.Sprintf("🔗Client %s (Steam ID: %s) is ready!", username, steamID)
				sendMessageToStatusChannel(message)

				config.ConnectedPlayers[steamID] = username
				updateConnectedPlayersMessage(config.ConnectionListChannelID)
			},
		},
		{
			// Example: "Client disconnected: 135108291984612402 | Jacksonthemaster"
			pattern: regexp.MustCompile(`Client\s+disconnected:\s+\d+\s+\|\s+(.+)\s+connectTime:\s+\d+,\d+s,\s+ClientId:\s+(\d+)`),
			handler: func(matches []string) {
				username := matches[1]
				steamID := matches[2]
				message := fmt.Sprintf("\u200dClient %s disconnected.", username)
				sendMessageToStatusChannel(message)

				delete(config.ConnectedPlayers, steamID)
				updateConnectedPlayersMessage(config.ConnectionListChannelID)
				updateBotStatus(config.DiscordSession) // Update bot status
			},
		},
		{
			// Enhanced "World Saved" pattern: "World Saved: C:/SteamCMD/Stationeers/saves/EuropaProd, BackupIndex: 1057"
			pattern: regexp.MustCompile(`World Saved:\s.*,\sBackupIndex:\s(\d+)`),
			handler: func(matches []string) {
				backupIndex := matches[1]
				currentTime := time.Now().UTC().Format(time.RFC3339)
				message := fmt.Sprintf("💾World Saved: BackupIndex: %s UTCTime: %s", backupIndex, currentTime)
				SendMessageToSavesChannel(message)
				updateBotStatus(config.DiscordSession) // Update bot status
			},
		},
		// Add more complex patterns and handlers here
	}

	for _, cp := range complexPatterns {
		if matches := cp.pattern.FindStringSubmatch(logMessage); matches != nil {
			cp.handler(matches)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.ChannelID != config.ControlChannelID {
		return
	}

	content := strings.TrimSpace(m.Content)

	switch {
	case strings.HasPrefix(content, "!start"):
		SendCommandToAPI("/start")
		s.ChannelMessageSend(m.ChannelID, "🕛Server is starting...")
		sendMessageToStatusChannel("🕛Start command received from Server Controller, Server is Starting...")

	case strings.HasPrefix(content, "!stop"):
		SendCommandToAPI("/stop")
		s.ChannelMessageSend(m.ChannelID, "🕛Server is stopping...")
		sendMessageToStatusChannel("🕛Stop command received from Server Controller, flatlining Server in 5 Seconds...")

	case strings.HasPrefix(content, "!restore"):
		sendMessageToStatusChannel("⚠️Restore command received, flatlining and restoring Server in 5 Seconds. Server will come back online in about 60 Seconds.")
		handleRestoreCommand(s, m, content)

	case strings.HasPrefix(content, "!list"):
		handleListCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!update"):
		handleUpdateCommand(s, m.ChannelID)

	case strings.HasPrefix(content, "!help"):
		handleHelpCommand(s, m.ChannelID)

	case strings.HasPrefix(content, "!ban"):
		handleBanCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!unban"):
		handleUnbanCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!validate"):
		handleValidateCommand(s, m.ChannelID)
	default:
		// Optionally handle unrecognized commands or ignore them
	}
}
