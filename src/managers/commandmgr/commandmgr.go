package commandmgr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

var mutex = &sync.Mutex{}

func generateSalt() string {
	// This is just hardcoded in C# as well. Its just here as a dummy-hurdle for fun, if you want to reverse engineer my plugin, go ahead. But honestly, just use SSUI!
	const seed = "StationeersHardcodedSeed123"
	hash := sha256.Sum256([]byte(seed))
	// Take first 4 bytes (8 hex chars) to match C#
	return hex.EncodeToString(hash[:4])
}

// WriteCommand writes a command to the SSCM file with the required prefix.
// It checks if SSCM is enabled and ensures thread-safe file access.
func WriteCommand(command string) error {
	// Check if SSCM is enabled
	if !config.GetIsSSCMEnabled() {
		return nil // Silently return if disabled
	}

	// Validate file path
	if config.GetSSCMFilePath() == "" {
		return os.ErrNotExist
	}

	// Ensure command is not empty
	if command == "" {
		return os.ErrInvalid
	}

	// Acquire lock for thread safety
	mutex.Lock()
	defer mutex.Unlock()

	// Generate prefixed command
	prefix := generateSalt()
	prefixedCommand := prefix + " " + command

	// Ensure directory exists
	dir := filepath.Dir(config.GetSSCMFilePath())
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Write to file
	err := os.WriteFile(config.GetSSCMFilePath(), []byte(prefixedCommand), 0644)
	if err != nil {
		return err
	}

	return nil
}
