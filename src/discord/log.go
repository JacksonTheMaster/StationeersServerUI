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

	_, err := config.DiscordSession.ChannelMessageSend(config.LogChannelID, config.LogMessageBuffer)
	if err != nil {
		fmt.Println("Error sending log to Discord:", err)
	} else {
		//fmt.Println("Flushed log buffer to Discord.")
		config.LogMessageBuffer = "" // Clear the buffer after sending
	}
}
