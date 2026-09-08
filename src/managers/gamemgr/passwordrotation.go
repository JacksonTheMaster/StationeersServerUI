// passwordrotation.go
package gamemgr

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// rotatePasswordIfEnabled checks if password rotation is enabled and sets a new random password
func rotatePasswordIfEnabled() {

	if !config.GetIsDiscordEnabled() || config.GetStatusPanelChannelID() == "" || !config.GetRotateServerPassword() {
		return
	}

	// Seed the random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random 6-digit password
	newPassword := strconv.Itoa(rng.Intn(900000) + 100000)

	// Set the new password in config
	err := config.SetServerPassword(newPassword)
	if err != nil {
		logger.Core.Error("Failed to set rotated server password: " + err.Error())
		return
	}

	logger.Core.Info("Password rotation enabled - new server password set: " + newPassword)
}
