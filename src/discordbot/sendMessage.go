package discordbot

import (
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func SendMessageToEventLogChannel(message string) {
	session := config.GetDiscordSession()

	if !config.GetIsDiscordEnabled() {
		return
	}

	if config.GetEventLogChannelID() == "" {
		return
	}

	if session == nil {
		logger.Discord.Error("Discord Error: Discord is enabled but session is not initialized")
		return
	}

	if len(message) <= 2000 {
		_, err := session.ChannelMessageSend(config.GetEventLogChannelID(), message)
		if err != nil {
			logger.Discord.Error("Error sending message to EventLog channel: " + err.Error())
		}
		return
	}

	maxMessageLength := 2000 // Discord's message character limit

	// Function to split the message into chunks and send each one in case the message (primilariy exception stack traces) exceeds the Discord message limit
	for len(message) > 0 {
		if len(message) > maxMessageLength {
			// Find a safe split point, for example, the last newline before the limit
			splitIndex := strings.LastIndex(message[:maxMessageLength], "\n")
			if splitIndex == -1 {
				splitIndex = maxMessageLength // No newline found, force split at max length
			}

			// Send the chunk
			_, err := session.ChannelMessageSend(config.GetEventLogChannelID(), message[:splitIndex])
			if err != nil {
				logger.Discord.Error("Error sending message to EventLog channel: " + err.Error())
				return
			}

			// Remove the sent chunk from the message
			message = message[splitIndex:]
		} else {
			// Send the remaining part of the message
			_, err := session.ChannelMessageSend(config.GetEventLogChannelID(), message)
			if err != nil {
				logger.Discord.Error("Error sending message to EventLog channel: " + err.Error())
				return
			}
			break
		}
	}
}
