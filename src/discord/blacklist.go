package discord

import (
	"io"
	"os"
	"strings"
)

func readBlacklist(blackListFilePath string) (string, error) {
	file, err := os.Open(blackListFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close() // Ensure the file is closed after reading

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func appendToBlacklist(blacklist, steamID string) string {
	if len(blacklist) > 0 && !strings.HasSuffix(blacklist, ",") {
		blacklist += ","
	}
	return blacklist + steamID
}

func removeFromBlacklist(blacklist, steamID string) string {
	entries := strings.Split(blacklist, ",")
	var updatedEntries []string
	for _, entry := range entries {
		if strings.TrimSpace(entry) != steamID {
			updatedEntries = append(updatedEntries, entry)
		}
	}
	return strings.Join(updatedEntries, ",")
}
