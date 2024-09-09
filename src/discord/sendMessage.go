package discord

import (
	"StationeersServerUI/src/config"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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

// CONNECTION LIST
func updateConnectedPlayersMessage(channelID string) {
	content := formatConnectedPlayers()
	sendAndEditMessageInConnectedPlayersChannel(channelID, content)
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

// BOT STATUS
func updateBotStatus(s *discordgo.Session) {
	playerCount := len(config.ConnectedPlayers)
	statusMessage := fmt.Sprintf("%d Employees connected", playerCount)
	err := s.UpdateGameStatus(0, statusMessage)
	if err != nil {
		fmt.Println("Error updating bot status:", err)
	}
}
