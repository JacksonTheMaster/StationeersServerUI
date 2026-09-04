package discordbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/bwmarrin/discordgo"
)

// Slash commands are shortcuts into the same private flows as the hub.
func listenToSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if !requireHubInteraction(s, i) {
		return
	}
	data := i.ApplicationCommandData()
	if data.Name == "status" {
		players, summary := statusPanelSnapshot()
		respondHub(s, i, buildStatusPanelEmbed(players, summary), nil)
		return
	}
	if !requireDiscordAdmin(s, i) {
		return
	}
	switch data.Name {
	case "help":
		showAdminMenu(s, i)
	case "start", "stop", "restart", "update":
		if len(data.Options) != 0 {
			hubError(s, i, "This command does not accept options.")
			return
		}
		chooseAdminAction(s, i, data.Name, "")
	case "list":
		limit := 0
		if len(data.Options) > 0 {
			value, err := stringCommandOption(data.Options, "limit")
			if err != nil {
				hubError(s, i, err.Error())
				return
			}
			if strings.ToLower(value) != "all" {
				limit, err = strconv.Atoi(value)
				if err != nil || limit < 1 {
					hubError(s, i, "Use a positive number or 'all' for the limit.")
					return
				}
			}
		}
		showBackupBrowser(s, i, 0, limit)
	case "restore", "download":
		name, err := backupCommandName(data.Options)
		if data.Name == "download" && len(data.Options) == 0 {
			if !deferHub(s, i) {
				return
			}
			manager := backupmgr.CurrentBackupManager()
			if manager == nil {
				editHubResult(s, i, "Download unavailable", "The backup manager is not ready.", false)
				return
			}
			backups, listErr := manager.ListBackups(1)
			if listErr != nil || len(backups) == 0 {
				editHubResult(s, i, "Download unavailable", "No backup is available yet.", false)
				return
			}
			// Pin 'latest' once; the downloader never resolves the name again.
			downloadBackupReply(s, i, manager, backups[0].Name)
			return
		}
		if err != nil {
			hubError(s, i, err.Error())
			return
		}
		chooseAdminAction(s, i, data.Name, name)
	case "command", "announce", "bansteamid", "unbansteamid":
		key := map[string]string{"command": "command", "announce": "message", "bansteamid": "steamid", "unbansteamid": "steamid"}[data.Name]
		value, err := stringCommandOption(data.Options, key)
		if err != nil {
			hubError(s, i, err.Error())
			return
		}
		chooseAdminAction(s, i, data.Name, value)
	default:
		hubError(s, i, "This command has been replaced. Open Server Admin Actions in the hub.")
	}
}

func stringCommandOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (string, error) {
	if len(options) != 1 || options[0] == nil || options[0].Name != key || options[0].Type != discordgo.ApplicationCommandOptionString {
		return "", fmt.Errorf("provide a valid %s option", key)
	}
	value, ok := options[0].Value.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("provide a valid %s option", key)
	}
	return value, nil
}

func backupCommandName(options []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	return stringCommandOption(options, "name")
}

func shortBackupLabel(label string, limit int) string {
	runes := []rune(label)
	if len(runes) > limit {
		return string(runes[:limit-1]) + "…"
	}
	return label
}
