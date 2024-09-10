package discord

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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

func handleRestoreByIndex(s *discordgo.Session, channelID string, index int) {
	// Stop the server before restoring
	SendCommandToAPI("/stop")

	url := fmt.Sprintf("http://localhost:8080/restore?index=%d", index)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌Failed to restore backup at index %d.", index))
		sendMessageToStatusChannel(fmt.Sprintf("⚠️Restore command received, but failed to restore backup at index %d.", index))
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("✅Backup %d restored successfully. Restarting server...", index))
	sendMessageToStatusChannel(fmt.Sprintf("✅Backup %d restored and server is restarting.", index))

	// Start the server after restoring
	SendCommandToAPI("/start")
}
