package discordbot

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/bwmarrin/discordgo"
)

const (
	ButtonAdminActions = "ssui_admin_actions"
	adminSelectID      = "ssui_admin_select"
	hubDialogPrefix    = "ssui_admin:"
	hubDialogLifetime  = 15 * time.Minute
	maxHubDialogs      = 512
)

// A screen gets an opaque, single-use ID. Names need not fit into Discord's
// custom IDs, and old buttons cannot accidentally select a different save.
type hubDialog struct {
	userID, guildID, channelID string
	screen, action, argument   string
	names                      []string
	page, limit                int
	expires                    time.Time
}

var hubDialogs = struct {
	sync.Mutex
	items map[string]hubDialog
}{items: make(map[string]hubDialog)}

func storeHubDialog(i *discordgo.InteractionCreate, dialog hubDialog) (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	id := hex.EncodeToString(random[:])
	dialog.userID, dialog.guildID, dialog.channelID = interactionUserID(i), i.GuildID, i.ChannelID
	dialog.expires = time.Now().Add(hubDialogLifetime)
	dialog.names = slices.Clone(dialog.names)
	hubDialogs.Lock()
	defer hubDialogs.Unlock()
	for key, old := range hubDialogs.items {
		if time.Now().After(old.expires) || (old.userID == dialog.userID && old.guildID == dialog.guildID) {
			delete(hubDialogs.items, key)
		}
	}
	if len(hubDialogs.items) >= maxHubDialogs {
		return "", fmt.Errorf("too many open menus; please try again later")
	}
	hubDialogs.items[id] = dialog
	return id, nil
}

func takeHubDialog(i *discordgo.InteractionCreate, id string) (hubDialog, bool) {
	hubDialogs.Lock()
	defer hubDialogs.Unlock()
	dialog, ok := hubDialogs.items[id]
	if !ok || dialog.userID != interactionUserID(i) || dialog.guildID != i.GuildID || dialog.channelID != i.ChannelID {
		return hubDialog{}, false
	}
	delete(hubDialogs.items, id)
	return dialog, time.Now().Before(dialog.expires)
}

func resetHubDialogs() {
	hubDialogs.Lock()
	hubDialogs.items = make(map[string]hubDialog)
	hubDialogs.Unlock()
}

func isHubInteraction(i *discordgo.InteractionCreate) bool {
	return i != nil && i.Interaction != nil && config.GetIsDiscordEnabled() && i.GuildID != "" && i.Member != nil &&
		interactionUserID(i) != "" && config.GetStatusPanelChannelID() != "" && i.ChannelID == config.GetStatusPanelChannelID()
}

func requireHubInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if isHubInteraction(i) {
		return true
	}
	hubError(s, i, "Please use the configured SSUI hub channel in your Discord server.")
	return false
}

func hasDiscordAdminRole(i *discordgo.InteractionCreate) bool {
	role := config.GetDiscordAdminRoleID()
	return isHubInteraction(i) && role != "" && slices.Contains(i.Member.Roles, role)
}

func requireDiscordAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if hasDiscordAdminRole(i) {
		return true
	}
	title := "😔 Sorry!"
	message := "You don't have permission to use this."
	if config.GetDiscordAdminRoleID() == "" {
		title = "🛠️ A little setup first"
		message = "Server admin actions haven't been set up yet. Ask your server admin to configure the Discord Admin Role ID in SSUI.\n\n[Open the SSUI docs](https://github.com/steamserverui/stationeersserverui/wiki/)"
	}
	privateHubNotice(s, i, title, message)
	return false
}

func hubEmbed(title, description string, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{Title: title, Description: shortBackupLabel(description, 3500), Color: color,
		Footer: &discordgo.MessageEmbedFooter{Text: "SSUI • Only visible to you"}}
}

func privateComponent(i *discordgo.InteractionCreate) bool {
	return i.Type == discordgo.InteractionMessageComponent && i.Message != nil && i.Message.Flags&discordgo.MessageFlagsEphemeral != 0
}

func respondHub(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	typ, flags := discordgo.InteractionResponseChannelMessageWithSource, discordgo.MessageFlagsEphemeral
	if privateComponent(i) {
		typ, flags = discordgo.InteractionResponseUpdateMessage, 0
	}
	if components == nil {
		components = []discordgo.MessageComponent{}
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: typ, Data: &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed}, Components: components, Flags: flags,
		AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}},
	}})
	if err != nil {
		logger.Discord.Warnf("Could not respond to hub interaction: %v", err)
	}
}

func hubError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	privateHubNotice(s, i, "Action unavailable", message)
}

func privateHubNotice(s *discordgo.Session, i *discordgo.InteractionCreate, title, message string) {
	// Never update a public message or somebody else's private menu on denial.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Embeds: []*discordgo.MessageEmbed{hubEmbed(title, message, 0xED4245)}}})
	if err != nil {
		logger.Discord.Warnf("Could not send private interaction error: %v", err)
	}
}

func deferHub(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	typ, flags := discordgo.InteractionResponseDeferredChannelMessageWithSource, discordgo.MessageFlagsEphemeral
	if privateComponent(i) {
		typ, flags = discordgo.InteractionResponseDeferredMessageUpdate, 0
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: typ, Data: &discordgo.InteractionResponseData{Flags: flags}})
	if err != nil {
		logger.Discord.Warnf("Could not acknowledge hub action: %v", err)
	}
	return err == nil
}

func editHub(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	if components == nil {
		components = []discordgo.MessageComponent{}
	}
	embeds := []*discordgo.MessageEmbed{embed}
	content := ""
	attachments := []*discordgo.MessageAttachment{}
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content, Embeds: &embeds, Components: &components,
		Attachments:     &attachments,
		AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}})
	if err != nil {
		logger.Discord.Warnf("Could not update private hub response: %v", err)
	}
}

func adminHomeButton() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{CustomID: ButtonAdminActions, Label: "Admin actions", Style: discordgo.SecondaryButton},
	}}}
}

func editHubResult(s *discordgo.Session, i *discordgo.InteractionCreate, title, text string, success bool) {
	color := 0xED4245
	if success {
		color = 0x57F287
	}
	editHub(s, i, hubEmbed(title, text, color), adminHomeButton())
}

func showAdminMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var options []discordgo.SelectMenuOption
	for _, item := range []struct{ value, label, emoji, description string }{
		{"start", "Start server", "🟢", "Launch the game server"},
		{"stop", "Stop server", "🛑", "Stop the current session"},
		{"restart", "Restart server", "🔄", "Stop and start the game server"},
		{"update", "Update server", "⬆️", "Stop, update via SteamCMD, and leave stopped"},
		{"backups", "Manage backups", "🗃️", "Browse, inspect, download or restore a save"},
		{"announce", "Send announcement", "📣", "Broadcast a message to in-game players"},
		{"command", "Send console command", "⌨️", "Send a single command through SSCM"},
		{"bansteamid", "Ban player", "🔨", "Add a SteamID to the blacklist"},
		{"unbansteamid", "Unban player", "🔓", "Remove a SteamID from the blacklist"},
	} {
		options = append(options, discordgo.SelectMenuOption{Value: item.value, Label: item.label, Description: item.description, Emoji: &discordgo.ComponentEmoji{Name: item.emoji}})
	}
	respondHub(s, i, hubEmbed("Server Admin Actions", "Choose an action below. Changes that interrupt the server require confirmation.", 0x5865F2),
		[]discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{CustomID: adminSelectID, Placeholder: "Choose an action", Options: options}}}})
}

func handleAdminInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var id string
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		id = i.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		id = i.ModalSubmitData().CustomID
	default:
		return
	}
	if id != ButtonAdminActions && id != adminSelectID && !strings.HasPrefix(id, hubDialogPrefix) && !strings.HasPrefix(id, ButtonDownloadBackupPfx) {
		return
	}
	if !requireDiscordAdmin(s, i) {
		return
	}
	if id == ButtonAdminActions {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}
		showAdminMenu(s, i)
		return
	}
	if id == adminSelectID {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}
		values := i.MessageComponentData().Values
		if len(values) != 1 {
			hubError(s, i, "Invalid action selection.")
			return
		}
		chooseAdminAction(s, i, values[0], "")
		return
	}
	if strings.HasPrefix(id, ButtonDownloadBackupPfx) {
		hubError(s, i, "This backup menu has been replaced. Open Server Admin Actions in the hub.")
		return
	}
	parts := strings.Split(strings.TrimPrefix(id, hubDialogPrefix), ":")
	if len(parts) != 2 {
		hubError(s, i, "Invalid menu.")
		return
	}
	dialog, ok := takeHubDialog(i, parts[0])
	if !ok {
		hubError(s, i, "This menu has expired or was already used. Open Server Admin Actions again.")
		return
	}
	verb := parts[1]
	if verb == "cancel" {
		showAdminMenu(s, i)
		return
	}
	switch dialog.screen {
	case "confirm":
		if verb == "confirm" && i.Type == discordgo.InteractionMessageComponent {
			startAdminAction(s, i, dialog.action, dialog.argument)
			return
		}
	case "input":
		if verb == "submit" && i.Type == discordgo.InteractionModalSubmit {
			value, err := hubModalValue(i.ModalSubmitData())
			if err != nil {
				hubError(s, i, err.Error())
				return
			}
			chooseAdminAction(s, i, dialog.action, value)
			return
		}
	case "backups":
		if i.Type != discordgo.InteractionMessageComponent {
			break
		}
		switch verb {
		case "prev":
			showBackupBrowser(s, i, max(0, dialog.page-1), dialog.limit)
			return
		case "next":
			showBackupBrowser(s, i, dialog.page+1, dialog.limit)
			return
		case "select":
			values := i.MessageComponentData().Values
			if len(values) != 1 {
				break
			}
			n, err := strconv.Atoi(values[0])
			if err != nil || n < 0 || n >= len(dialog.names) {
				break
			}
			showBackupDetails(s, i, dialog.names[n], dialog.page, dialog.limit)
			return
		}
	case "backup":
		switch verb {
		case "back":
			showBackupBrowser(s, i, dialog.page, dialog.limit)
			return
		case "restore", "download":
			chooseAdminAction(s, i, verb, dialog.argument)
			return
		}
	}
	hubError(s, i, "This action does not belong to the current menu.")
}

func showAdminInput(s *discordgo.Session, i *discordgo.InteractionCreate, action string) {
	label := map[string]string{"announce": "Announcement", "command": "Console command", "bansteamid": "SteamID to ban", "unbansteamid": "SteamID to unban"}[action]
	id, err := storeHubDialog(i, hubDialog{screen: "input", action: action})
	if err != nil {
		hubError(s, i, err.Error())
		return
	}
	maxLength := 1000
	if action == "bansteamid" || action == "unbansteamid" {
		maxLength = 17
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{CustomID: hubDialogPrefix + id + ":submit", Title: label,
			Components: []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.TextInput{CustomID: "value", Label: label, Style: discordgo.TextInputShort, Required: true, MinLength: 1, MaxLength: maxLength},
			}}}}})
	if err != nil {
		logger.Discord.Warnf("Could not open admin input: %v", err)
	}
}

func hubModalValue(data discordgo.ModalSubmitInteractionData) (string, error) {
	if len(data.Components) == 1 {
		row, ok := data.Components[0].(*discordgo.ActionsRow)
		if ok && len(row.Components) == 1 {
			input, ok := row.Components[0].(*discordgo.TextInput)
			if ok && input.CustomID == "value" && strings.TrimSpace(input.Value) != "" {
				return input.Value, nil
			}
		}
	}
	return "", fmt.Errorf("please enter a value")
}
