package discord

import (
	"fmt"
	"time"

	"StationeersServerUI/src/config"

	"github.com/bwmarrin/discordgo"
)

func SendControlMessage() {
	messageContent := "Control Panel:\n\nReact with the following to perform actions:\n" +
		"▶️ Start the server\n\n" +
		"⏹️ Stop the server\n\n" +
		"♻️ Restart the server\n\n"

	msg, err := config.DiscordSession.ChannelMessageSend(config.ControlPanelChannelID, messageContent)
	if err != nil {
		fmt.Println("Error sending control message:", err)
		return
	}

	// Add reactions (acting as buttons) to the control message
	config.DiscordSession.MessageReactionAdd(config.ControlPanelChannelID, msg.ID, "▶️") // Start
	config.DiscordSession.MessageReactionAdd(config.ControlPanelChannelID, msg.ID, "⏹️") // Stop
	config.DiscordSession.MessageReactionAdd(config.ControlPanelChannelID, msg.ID, "♻️") // Restart
	config.ControlMessageID = msg.ID
	clearMessagesAboveLastN(config.ControlPanelChannelID, 1) // Store the message ID for later reference
}

// reactionAddHandler - Handles reactions added to messages
func reactionAddHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore bot's own reactions
	if r.UserID == s.State.User.ID {
		return
	}

	// Check if the reaction was added to the control message for server control
	if r.MessageID == config.ControlMessageID {
		handleControlReactions(s, r)
	}

	// Optionally, add more message-specific handlers here for other features
}

// handleControlReactions - Handles reactions for server control actions
func handleControlReactions(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	var actionMessage string

	switch r.Emoji.Name {
	case "▶️": // Start action
		SendCommandToAPI("/start")
		actionMessage = "🕛Server is Starting..."
	case "⏹️": // Stop action
		SendCommandToAPI("/stop")
		actionMessage = "🛑Server is Stopping..."
	case "♻️": // Restart action
		actionMessage = "♻️Server is restarting..."
		SendCommandToAPI("/stop")
		//sleep 5 sec
		time.Sleep(5 * time.Second)
		SendCommandToAPI("/start")

	default:
		fmt.Println("Unknown reaction:", r.Emoji.Name)
		return
	}

	// Get the user who triggered the action
	user, err := s.User(r.UserID)
	if err != nil {
		fmt.Printf("Error fetching user details: %v\n", err)
		return
	}
	username := user.Username

	// Send the action message to the control channel
	sendMessageToStatusChannel(fmt.Sprintf("%s triggered by %s.", actionMessage, username))

	// Remove the reaction after processing
	err = s.MessageReactionRemove(config.ControlPanelChannelID, r.MessageID, r.Emoji.APIName(), r.UserID)
	if err != nil {
		fmt.Printf("Error removing reaction: %v\n", err)
	}
}
