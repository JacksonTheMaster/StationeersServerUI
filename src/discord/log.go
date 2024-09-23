package discord

import (
	"StationeersServerUI/src/config"
	"fmt"
)

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

	const discordMaxMessageLength = 2000
	message := config.LogMessageBuffer

	for len(message) > 0 {
		// Determine how much of the message we can send
		chunkSize := discordMaxMessageLength
		if len(message) < discordMaxMessageLength {
			chunkSize = len(message)
		}

		// Send the chunk to Discord
		_, err := config.DiscordSession.ChannelMessageSend(config.LogChannelID, message[:chunkSize])
		if err != nil {
			fmt.Println("Error sending log to Discord:", err)
			break
		}

		// Move to the next chunk
		message = message[chunkSize:]
	}

	// Clear the buffer after sending
	config.LogMessageBuffer = ""
}
