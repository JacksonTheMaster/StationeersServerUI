package discordbot

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/commandmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"
	"github.com/bwmarrin/discordgo"
)

var discordAction = struct {
	sync.Mutex
	busy         bool
	name, result string
	changed      time.Time
}{}

func beginDiscordAction(name string) bool {
	discordAction.Lock()
	defer discordAction.Unlock()
	if discordAction.busy {
		return false
	}
	discordAction.busy, discordAction.name, discordAction.result = true, name, "In progress"
	discordAction.changed = time.Now()
	return true
}

func finishDiscordAction(err error) {
	discordAction.Lock()
	discordAction.busy = false
	discordAction.result = "Completed"
	if err != nil {
		discordAction.result = "Failed — see the event log"
	}
	discordAction.changed = time.Now()
	discordAction.Unlock()
}

func discordActionStatus() string {
	discordAction.Lock()
	defer discordAction.Unlock()
	if discordAction.name == "" {
		return ""
	}
	return fmt.Sprintf("**%s** · %s · <t:%d:R>", discordAction.name, discordAction.result, discordAction.changed.Unix())
}

func chooseAdminAction(s *discordgo.Session, i *discordgo.InteractionCreate, action, argument string) {
	switch action {
	case "backups":
		showBackupBrowser(s, i, 0, 0)
	case "download":
		if !deferHub(s, i) {
			return
		}
		downloadBackupReply(s, i, backupmgr.CurrentBackupManager(), argument)
	case "stop", "restart", "update", "restore":
		showActionConfirmation(s, i, action, argument)
	case "start":
		startAdminAction(s, i, action, argument)
	case "command", "announce", "bansteamid", "unbansteamid":
		if argument == "" {
			showAdminInput(s, i, action)
			return
		}
		startAdminAction(s, i, action, argument)
	default:
		hubError(s, i, "Unknown admin action.")
	}
}

func showActionConfirmation(s *discordgo.Session, i *discordgo.InteractionCreate, action, argument string) {
	description := map[string]string{
		"stop":    "Stop the game server? Connected players will be disconnected.",
		"restart": "Stop and restart the game server? Connected players will be disconnected.",
		"update":  "Stop the game server and update it through SteamCMD? The server will remain stopped afterwards so you can choose when to start it.",
		"restore": "Restore this backup? The game server will be stopped, its current save replaced, and the server started again.\n\n**Backup**\n" + argument,
	}[action]
	if description == "" || (action == "restore" && argument == "") {
		hubError(s, i, "Choose a valid action and backup first.")
		return
	}
	id, err := storeHubDialog(i, hubDialog{screen: "confirm", action: action, argument: argument})
	if err != nil {
		hubError(s, i, err.Error())
		return
	}
	respondHub(s, i, hubEmbed("Confirm "+action, description, 0xFEE75C), []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{CustomID: hubDialogPrefix + id + ":confirm", Label: "Confirm " + action, Style: discordgo.DangerButton},
		discordgo.Button{CustomID: hubDialogPrefix + id + ":cancel", Label: "Cancel", Style: discordgo.SecondaryButton},
	}}})
}

func validateAdminInput(action, argument string) error {
	switch action {
	case "start", "stop", "restart", "update":
		if argument != "" {
			return fmt.Errorf("this action does not accept an argument")
		}
	case "restore":
		if argument == "" {
			return fmt.Errorf("choose a backup first")
		}
	case "command", "announce":
		if strings.TrimSpace(argument) == "" || utf8.RuneCountInString(argument) > 1000 || strings.ContainsFunc(argument, unicode.IsControl) {
			return fmt.Errorf("enter one line of text, at most 1000 characters, without control characters")
		}
	case "bansteamid", "unbansteamid":
		if len(argument) != 17 || strings.ContainsFunc(argument, func(r rune) bool { return r < '0' || r > '9' }) {
			return fmt.Errorf("enter a 17-digit SteamID64")
		}
	default:
		return fmt.Errorf("unknown admin action")
	}
	return nil
}

func startAdminAction(s *discordgo.Session, i *discordgo.InteractionCreate, action, argument string) {
	// Check again at execution, including modal submissions and confirmations.
	if !requireDiscordAdmin(s, i) {
		return
	}
	if err := validateAdminInput(action, argument); err != nil {
		hubError(s, i, err.Error())
		return
	}
	if !beginDiscordAction(action) {
		hubError(s, i, "Another Discord action is still running. Check the hub and try again once it finishes.")
		return
	}
	if !deferHub(s, i) {
		finishDiscordAction(fmt.Errorf("interaction acknowledgement failed"))
		return
	}
	editHub(s, i, hubEmbed("Working on it", "Running **"+action+"**. You can check the hub if this takes a while.", 0xFEE75C), nil)
	go func() {
		// Do not put command arguments (which can contain passwords) in logs.
		userID := interactionUserID(i)
		logger.Discord.Infof("Admin %s requested %s", userID, action)
		SendMessageToEventLogChannel(fmt.Sprintf("🛠️ Admin %s requested %s.", userID, action))
		refreshStatusPanel()
		message, err := performDiscordAction(action, argument)
		finishDiscordAction(err)
		if err != nil {
			logger.Discord.Warnf("Discord action %s failed: %v", action, err)
			SendMessageToEventLogChannel(fmt.Sprintf("❌ Admin action %s failed. Details were sent privately to the requester.", action))
			editHubResult(s, i, "Action failed", err.Error(), false)
		} else {
			SendMessageToEventLogChannel(fmt.Sprintf("✅ Admin action %s completed.", action))
			editHubResult(s, i, "Action completed", message, true)
		}
		refreshStatusPanel()
	}()
}

// Small function dependencies let us test stop/update/restore ordering without
// ever launching a real server. The same sequence is used by passed votes.
type serverActionBackend struct {
	running              func() bool
	start, stop, update  func() error
	checkBackup, restore func(string) error
}

func gameActionBackend() serverActionBackend {
	manager := backupmgr.CurrentBackupManager()
	return serverActionBackend{
		running: gamemgr.InternalIsServerRunning, start: gamemgr.InternalStartServer, stop: gamemgr.InternalStopServer,
		update: func() error { _, err := steamcmd.InstallAndRunSteamCMD(); return err },
		checkBackup: func(name string) error {
			if manager == nil {
				return fmt.Errorf("backup manager is not initialized")
			}
			return backupmgr.CheckBackupAvailable(manager, name)
		},
		restore: func(name string) error { return manager.RestoreBackup(name) },
	}
}

func performServerAction(action, name string, backend serverActionBackend) (string, error) {
	if action == "restore" {
		if err := backend.checkBackup(name); err != nil {
			return "", err
		}
	}
	if action == "start" {
		if backend.running() {
			return "The game server is already running.", nil
		}
		return "The server process has started. Follow startup progress in the hub.", backend.start()
	}
	if action != "stop" && action != "restart" && action != "update" && action != "restore" {
		return "", fmt.Errorf("unknown server action")
	}
	if backend.running() {
		if err := backend.stop(); err != nil {
			return "", fmt.Errorf("could not stop the game server: %w", err)
		}
	}
	switch action {
	case "stop":
		return "The game server is stopped.", nil
	case "update":
		return "Update completed. The game server is stopped; start it when you are ready.", backend.update()
	case "restore":
		if err := backend.restore(name); err != nil {
			return "", fmt.Errorf("restore failed; the server remains stopped: %w", err)
		}
	}
	if err := backend.start(); err != nil {
		return "", fmt.Errorf("could not start the game server: %w", err)
	}
	return "The server process has started. Follow startup progress in the hub.", nil
}

func performDiscordAction(action, argument string) (string, error) {
	switch action {
	case "start", "stop", "restart", "update", "restore":
		return performServerAction(action, argument, gameActionBackend())
	case "command", "announce":
		if !config.GetIsSSCMEnabled() || !gamemgr.InternalIsServerRunning() {
			return "", fmt.Errorf("the game server must be running with SSCM enabled")
		}
		if action == "announce" {
			argument = "announce " + argument
		}
		return "Command submitted through SSCM. This is not an acknowledgement from the game server.", commandmgr.WriteCommand(argument)
	case "bansteamid":
		return "Blacklist updated. Restart the game server for the change to take effect.", banSteamID(argument)
	case "unbansteamid":
		return "Blacklist updated. Restart the game server for the change to take effect.", unbanSteamID(argument)
	}
	return "", fmt.Errorf("unknown admin action")
}
