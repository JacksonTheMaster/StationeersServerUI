package discordbot

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"path"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/bwmarrin/discordgo"
)

const discordBackupEmojiAssetDir = "SSUI/onboard_bundled/assets/backupstats/discord"

var backupStatEmojiAssets = map[string]string{
	"days":           "days.webp",
	"things":         "things.webp",
	"atmospheres":    "atmospheres.webp",
	"rooms":          "rooms.webp",
	"pipe_networks":  "pipe-networks.webp",
	"cable_networks": "cable-networks.webp",
	"players":        "../players.png",
	"archive_size":   "../archive-size.png",
}

var applicationEmojis = struct {
	sync.RWMutex
	byName map[string]*discordgo.Emoji
}{byName: make(map[string]*discordgo.Emoji)}

// syncApplicationEmojis ensures the backup-stat artwork is owned by the user's
// Discord application. Application emojis work in every guild where the bot is
// installed and do not consume a guild's emoji slots.
func syncApplicationEmojis(session *discordgo.Session) {
	clearApplicationEmojiCache()
	if session == nil || session.State == nil || session.State.User == nil {
		return
	}

	appID := session.State.User.ID
	existing, err := session.ApplicationEmojis(appID)
	if err != nil {
		logger.Discord.Warn("Could not list Discord application emojis; using text fallbacks: " + err.Error())
		return
	}

	byName := make(map[string]*discordgo.Emoji, len(existing)+len(backupStatEmojiAssets))
	for _, emoji := range existing {
		if emoji != nil {
			byName[emoji.Name] = emoji
		}
	}

	assets := config.GetV1UIFS()
	for shortName, filename := range backupStatEmojiAssets {
		name := "ssui_" + shortName
		if _, ok := byName[name]; ok {
			continue
		}

		data, readErr := fs.ReadFile(assets, path.Join(discordBackupEmojiAssetDir, filename))
		if readErr != nil {
			logger.Discord.Warn(fmt.Sprintf("Could not read bundled emoji %s: %v", filename, readErr))
			continue
		}

		mediaType := "image/webp"
		if path.Ext(filename) == ".png" {
			mediaType = "image/png"
		}
		created, createErr := session.ApplicationEmojiCreate(appID, &discordgo.EmojiParams{
			Name:  name,
			Image: "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(data),
		})
		if createErr != nil {
			logger.Discord.Warn(fmt.Sprintf("Could not create Discord application emoji %s: %v", name, createErr))
			continue
		}
		byName[name] = created
	}

	applicationEmojis.Lock()
	applicationEmojis.byName = byName
	applicationEmojis.Unlock()
	logger.Discord.Debugf("Loaded %d SSUI Discord application emojis", len(byName))
}

func backupStatEmoji(name, fallback string) string {
	applicationEmojis.RLock()
	emoji := applicationEmojis.byName["ssui_"+name]
	applicationEmojis.RUnlock()
	if emoji == nil {
		return fallback
	}
	return emoji.MessageFormat()
}

func clearApplicationEmojiCache() {
	applicationEmojis.Lock()
	applicationEmojis.byName = make(map[string]*discordgo.Emoji)
	applicationEmojis.Unlock()
}
