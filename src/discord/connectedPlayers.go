package discord

import (
	"StationeersServerUI/src/config"
	"fmt"
	"strings"
)

func formatConnectedPlayers() string {
	if len(config.ConnectedPlayers) == 0 {
		return "No players are currently connected."
	}

	var sb strings.Builder
	sb.WriteString("**Connected Players:**\n")
	sb.WriteString("```\n")
	sb.WriteString("Username              | Steam ID\n")
	sb.WriteString("----------------------|------------------------\n")

	for steamID, username := range config.ConnectedPlayers {
		sb.WriteString(fmt.Sprintf("%-20s | %s\n", username, steamID))
	}

	sb.WriteString("```")
	return sb.String()
}
