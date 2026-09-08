package discordbot

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

var blacklistMutex sync.Mutex

func banSteamID(steamID string) error {
	// Read the current blacklist
	blacklist, err := readBlacklist(config.GetBlackListFilePath())
	if err != nil {
		return fmt.Errorf("error reading blacklist file: %v", err)
	}

	// Check if the SteamID is already in the blacklist
	entries := strings.Split(blacklist, ",")
	for _, entry := range entries {
		if strings.TrimSpace(entry) == steamID {
			return fmt.Errorf("SteamID %s is already banned", steamID)
		}
	}

	// Add the SteamID to the blacklist
	blacklist = appendToBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	err = os.WriteFile(config.GetBlackListFilePath(), []byte(blacklist), 0644)
	if err != nil {
		return fmt.Errorf("error writing to blacklist file: %v", err)
	}

	return nil
}

func unbanSteamID(steamID string) error {
	// Read the current blacklist
	blacklist, err := readBlacklist(config.GetBlackListFilePath())
	if err != nil {
		return fmt.Errorf("error reading blacklist file: %v", err)
	}

	// Check if the SteamID is in the blacklist
	entries := strings.Split(blacklist, ",")
	exists := false
	for _, entry := range entries {
		if strings.TrimSpace(entry) == steamID {
			exists = true
			break
		}
	}

	if !exists {
		return nil // Not an error, just nothing to do
	}

	// Remove the SteamID from the blacklist
	updatedBlacklist := removeFromBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	err = os.WriteFile(config.GetBlackListFilePath(), []byte(updatedBlacklist), 0644)
	if err != nil {
		return fmt.Errorf("error writing to blacklist file: %v", err)
	}

	return nil
}

func readBlacklist(blackListFilePath string) (string, error) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

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
