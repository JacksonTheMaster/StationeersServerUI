package discordbot

import (
	"maps"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
)

var statusPanelData = struct {
	sync.RWMutex
	players map[string]string
	summary *backupmgr.SaveSummary
}{players: make(map[string]string)}

func initializeDiscordBackupSummary() {
	backupmgr.SetBackupCopiedHandler(func(summary backupmgr.SaveSummary) {
		setLatestBackupSummary(summary)
		refreshStatusPanel()
	})

	manager := backupmgr.GlobalBackupManager
	if manager == nil {
		return
	}
	backups, err := manager.ListBackups(1)
	if err != nil {
		logger.Discord.Debug("Latest backup statistics are not available yet: " + err.Error())
		return
	}
	if len(backups) > 0 {
		setLatestBackupSummary(backups[0].Summary)
	}
}

func disableDiscordRuntimeState() {
	prepareDiscordRuntimeState()
	statusPanelData.Lock()
	statusPanelData.players = make(map[string]string)
	statusPanelData.Unlock()
}

func prepareDiscordRuntimeState() {
	backupmgr.SetBackupCopiedHandler(nil)
	statusPanelData.Lock()
	statusPanelData.summary = nil
	statusPanelData.Unlock()
	clearApplicationEmojiCache()
}

func setStatusPanelPlayers(players map[string]string) {
	statusPanelData.Lock()
	statusPanelData.players = maps.Clone(players)
	statusPanelData.Unlock()
}

func setLatestBackupSummary(summary backupmgr.SaveSummary) {
	statusPanelData.Lock()
	copy := summary
	statusPanelData.summary = &copy
	statusPanelData.Unlock()
}

func statusPanelSnapshot() (map[string]string, *backupmgr.SaveSummary) {
	statusPanelData.RLock()
	defer statusPanelData.RUnlock()
	players := maps.Clone(statusPanelData.players)
	if statusPanelData.summary == nil {
		return players, nil
	}
	summary := *statusPanelData.summary
	return players, &summary
}

// DisableRuntimeState stops Discord-only backup metadata work when the
// integration is disabled.
func DisableRuntimeState() {
	disableDiscordRuntimeState()
}
