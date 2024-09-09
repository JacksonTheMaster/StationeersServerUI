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
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	config.DiscordSession.AddHandler(messageCreate)

	err = config.DiscordSession.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is now running.")
	// Start the buffer flush ticker to send the remaining buffer every 5 seconds
	config.BufferFlushTicker = time.NewTicker(5 * time.Second)
	sendMessageToStatusChannel("Bot reconnected to Discord. Good Morning, Stationeers!")
	go func() {
		for range config.BufferFlushTicker.C {
			flushLogBufferToDiscord()
		}
	}()

	select {} // Keep the program running
}

func AddToLogBuffer(logMessage string) {
	config.LogMessageBuffer += logMessage + "\n" // Add the log message to the buffer with a newline
	checkForKeywords(logMessage)
	// If the buffer exceeds the max size, send it to Discord
	if len(config.LogMessageBuffer) >= config.MaxBufferSize {
		flushLogBufferToDiscord()
	}
}

func flushLogBufferToDiscord() {
	if len(config.LogMessageBuffer) == 0 {
		return // No messages to send
	}

	if config.DiscordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := config.DiscordSession.ChannelMessageSend(config.LogChannelID, config.LogMessageBuffer)
	if err != nil {
		fmt.Println("Error sending log to Discord:", err)
	} else {
		//fmt.Println("Flushed log buffer to Discord.")
		config.LogMessageBuffer = "" // Clear the buffer after sending
	}
}

func checkForKeywords(logMessage string) {
	// List of keywords to detect and their corresponding messages
	keywordActions := map[string]string{
		"Ready": "Attention! Server is ready to connect!",
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
				message := fmt.Sprintf("Client %s (Steam ID: %s) is ready!", username, steamID)
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
				message := fmt.Sprintf("Client %s disconnected.", username)
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
				message := fmt.Sprintf("World Saved: BackupIndex: %s UTCTime: %s", backupIndex, currentTime)
				sendMessageToSavesChannel(message)
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

func SendMessageToControlChannel(message string) {
	if config.DiscordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := config.DiscordSession.ChannelMessageSend(config.ControlChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to control channel:", err)
	} else {
		fmt.Println("Sent message to control channel:", message)
	}
}

func sendMessageToStatusChannel(message string) {
	if config.DiscordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := config.DiscordSession.ChannelMessageSend(config.StatusChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to status channel:", err)
	} else {
		fmt.Println("Sent message to status channel:", message)
	}
}

func sendAndEditMessageInConnectedPlayersChannel(channelID, message string) {
	if config.DiscordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	if config.ConnectedPlayersMessageID == "" {
		// Send a new message if there's no existing message to edit
		msg, err := config.DiscordSession.ChannelMessageSend(channelID, message)
		if err != nil {
			fmt.Printf("Error sending message to channel %s: %v\n", channelID, err)
		} else {
			config.ConnectedPlayersMessageID = msg.ID
			fmt.Printf("Sent message to channel %s: %s\n", channelID, message)
		}
	} else {
		// Edit the existing message
		_, err := config.DiscordSession.ChannelMessageEdit(channelID, config.ConnectedPlayersMessageID, message)
		if err != nil {
			fmt.Printf("Error editing message in channel %s: %v\n", channelID, err)
		} else {
			fmt.Printf("Updated message in channel %s: %s\n", channelID, message)
		}
	}
}

func updateBotStatus(s *discordgo.Session) {
	playerCount := len(config.ConnectedPlayers)
	statusMessage := fmt.Sprintf("%d Employees connected", playerCount)
	err := s.UpdateGameStatus(0, statusMessage)
	if err != nil {
		fmt.Println("Error updating bot status:", err)
	}
}

func sendMessageToSavesChannel(message string) {
	if config.DiscordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := config.DiscordSession.ChannelMessageSend(config.SaveChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to saves channel:", err)
	} else {
		fmt.Println("Sent message to saves channel:", message)
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
		s.ChannelMessageSend(m.ChannelID, "Server is starting...")
		sendMessageToStatusChannel("Start command received from @Server Controller, Server is trying to start...")

	case strings.HasPrefix(content, "!stop"):
		SendCommandToAPI("/stop")
		s.ChannelMessageSend(m.ChannelID, "Server is stopping...")
		sendMessageToStatusChannel("Stop command received from @Server Controller, Server is stopping...")

	case strings.HasPrefix(content, "!restore"):
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

	default:
		// Optionally handle unrecognized commands or ignore them
	}
}

func parseBackupList(rawData string) string {
	lines := strings.Split(rawData, "\n")
	var formattedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) == 2 {
			formattedLines = append(formattedLines, fmt.Sprintf("**%s** - %s", parts[0], parts[1]))
		}
	}
	return strings.Join(formattedLines, "\n")
}
