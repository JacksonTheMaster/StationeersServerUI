package discord

import (
	"fmt"
	"strings"
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
