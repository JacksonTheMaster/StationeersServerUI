package discordbot

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/commandmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"

	"github.com/bwmarrin/discordgo"
)

type commandHandler func(*discordgo.Session, *discordgo.InteractionCreate, EmbedData) error

// Command handlers map
var handlers = map[string]commandHandler{
	"start":        handleStart,
	"stop":         handleStop,
	"status":       handleStatus,
	"help":         handleHelp,
	"restore":      handleRestore,
	"list":         handleList,
	"download":     handleDownload,
	"bansteamid":   handleBan,
	"unbansteamid": handleUnban,
	"update":       handleUpdate,
	"command":      handleCommand,
	"announce":     handleAnnounce,
}

// Check channel and handle initial validation
func listenToSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if i.ChannelID != config.GetControlChannelID() {
		respond(s, i, EmbedData{
			Title: "Wrong Channel", Description: "Commands must be sent to the configured control channel",
			Color: 0xFF0000, Fields: []EmbedField{{Name: "Accepted Channel", Value: fmt.Sprintf("<#%s>", config.GetControlChannelID()), Inline: true}},
		})
		return
	}

	cmd := i.ApplicationCommandData().Name
	if handler, ok := handlers[cmd]; ok {
		data := EmbedData{Title: "Command Error", Color: 0xFF0000}
		if err := handler(s, i, data); err != nil {
			logger.Discord.Error("Error handling " + cmd + ": " + err.Error())
		}
	}
}

// Generic response function
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, embed EmbedData) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{generateEmbed(embed)},
		},
	})
}

func handleStart(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	data.Title, data.Description, data.Color = "Server Control", "Starting the server...", 0x00FF00
	data.Fields = []EmbedField{{Name: "Status", Value: "🕛 Recieved", Inline: true}}
	if err := respond(s, i, data); err != nil {
		return err
	}
	gamemgr.InternalStartServer()
	SendMessageToEventLogChannel("🕛Start command received, Server is Starting...")
	return nil
}

func handleStop(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	data.Title, data.Description, data.Color = "Server Control", "Stopping the server...", 0xFF0000
	data.Fields = []EmbedField{{Name: "Status", Value: "🕛 Recieved", Inline: true}}
	if err := respond(s, i, data); err != nil {
		return err
	}
	gamemgr.InternalStopServer()
	SendMessageToEventLogChannel("🕛Stop command received, flatlining Server in 5 Seconds...")
	return nil
}

func handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	isRunning := gamemgr.InternalIsServerRunning()
	data.Title = "🎮 Server Status"
	data.Description = "Current process state for the Stationeers game server.\n*Note: 'Started' indicates a running process was found, but not necessarily fully operational.*"
	data.Color = map[bool]int{true: 0x00FF00, false: 0xFF0000}[isRunning]
	data.Fields = []EmbedField{
		{Name: "Status:", Value: map[bool]string{true: "🟢 Started", false: "🔴 Stopped"}[isRunning], Inline: true},
		{Name: "Checked:", Value: time.Now().Format("15:04:05 MST"), Inline: true},
	}
	return respond(s, i, data)
}

func handleUpdate(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	thinkingData := EmbedData{
		Title:       "🎮 Gameserver Update",
		Description: "The Backend is processing the gameserver update via SteamCMD. Please wait, this may take a while...",
		Color:       0xFFA500, // Orange color for in-progress
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{generateEmbed(thinkingData)},
		},
	})
	if err != nil {
		return err
	}
	data.Title = "🎮 Gameserver Update"
	data.Description = "Gameserver update completed."
	data.Color = 0x00FF00 // Green for completion (will adjust if error)

	_, err = steamcmd.InstallAndRunSteamCMD()

	data.Fields = []EmbedField{
		{Name: "Update Status:", Value: map[bool]string{true: "🟢 Success", false: "🔴 Failed"}[err == nil], Inline: true},
	}
	if err != nil {
		data.Color = 0xFF0000 // Red for error
		data.Fields = append(data.Fields, EmbedField{Name: "Error:", Value: err.Error(), Inline: true})
	}

	// Edit the original message with "update completed" embed
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{generateEmbed(data)},
	})
	return err
}

func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	data.Title, data.Description, data.Color = "Command Help", "Available Commands:", 0x1E90FF
	data.Fields = []EmbedField{
		{Name: "/start", Value: "Starts the server"},
		{Name: "/stop", Value: "Stops the server"},
		{Name: "/status", Value: "Gets the running status of the gameserver process"},
		{Name: "/update", Value: "Updates the gameserver via SteamCMD"},
		{Name: "/list [limit]", Value: "Lists recent backups (default: 5)"},
		{Name: "/restore <name>", Value: "Restores a backup"},
		{Name: "/download [name]", Value: "Downloads a backup (most recent if no name)"},
		{Name: "/bansteamid <SteamID>", Value: "Bans a player"},
		{Name: "/unbansteamid <SteamID>", Value: "Unbans a player"},
		{Name: "/command <command>", Value: "Sends a command to the gameserver console"},
		{Name: "/announce <message>", Value: "Broadcasts an announcement to all in-game players (via announce cmd)"},
		{Name: "/help", Value: "Shows this help"},
	}
	return respond(s, i, data)
}

func handleRestore(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	name, err := backupCommandName(i.ApplicationCommandData().Options)
	manager := backupmgr.CurrentBackupManager()
	if err == nil && manager == nil {
		err = fmt.Errorf("backup manager is not initialized")
	}
	if err == nil {
		err = backupmgr.CheckBackupAvailable(manager, name)
	}
	if err != nil {
		data.Title, data.Description = "Restore Failed", err.Error()
		return respond(s, i, data)
	}
	data.Title, data.Description, data.Color = "Backup Restore", fmt.Sprintf("Restoring %s...", name), 0xFFA500
	if err := respond(s, i, data); err != nil {
		return err
	}
	gamemgr.InternalStopServer()
	if err := manager.RestoreBackup(name); err != nil {
		SendMessageToControlChannel(fmt.Sprintf("❌ Failed to restore %s: %v", name, err))
		SendMessageToEventLogChannel("⚠️ Restore command failed")
		return nil
	}
	SendMessageToControlChannel(fmt.Sprintf("✅ Restored %s. Starting server...", name))
	time.Sleep(5 * time.Second)
	gamemgr.InternalStartServer()
	return nil
}

const maxDiscordFileSize = 10 * 1024 * 1024 // 10MB Discord file upload limit

func handleDownload(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		data.Title, data.Description = "Download Failed", "Backup manager is not initialized"
		return respond(s, i, data)
	}
	name, err := backupCommandName(i.ApplicationCommandData().Options)
	if len(i.ApplicationCommandData().Options) == 0 {
		var backups []backupmgr.BackupSaveFile
		backups, err = manager.ListBackups(1)
		if err == nil && len(backups) == 0 {
			err = fmt.Errorf("no backups available")
		}
		if err == nil {
			name = backups[0].Name
		}
	}
	if err == nil {
		err = backupmgr.CheckBackupAvailable(manager, name)
	}
	if err != nil {
		data.Title, data.Description = "Download Failed", err.Error()
		return respond(s, i, data)
	}
	data.Title, data.Description, data.Color = "📥 Backup Download", fmt.Sprintf("Preparing %s...", name), 0xFFA500
	if err := respond(s, i, data); err != nil {
		return err
	}
	sendBackupToChannel(s, i.ChannelID, manager, name)
	return nil
}

// Reject stale registered index options instead of interpreting them as filenames.
func backupCommandName(options []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	if len(options) != 1 || options[0] == nil || options[0].Name != "name" || options[0].Type != discordgo.ApplicationCommandOptionString {
		return "", fmt.Errorf("provide a backup name from /list")
	}
	name, ok := options[0].Value.(string)
	if !ok || name == "" {
		return "", fmt.Errorf("provide a backup name from /list")
	}
	return name, nil
}

func sendBackupToChannel(s *discordgo.Session, channelID string, manager *backupmgr.BackupManager, name string) {
	backupData, err := manager.GetBackupFileData(name)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ Failed to download %s: %v", name, err))
		return
	}
	if backupData.Size > maxDiscordFileSize {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ %s is too large to upload (%.2f MB > 10 MB limit)", name, float64(backupData.Size)/(1024*1024)))
		return
	}
	file := &discordgo.File{Name: backupData.Filename, ContentType: "application/octet-stream", Reader: bytes.NewReader(backupData.Data)}
	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("📦 %s (%s)", name, backupData.SaveTime.Format("Jan 2, 2006 3:04 PM")),
		Files:   []*discordgo.File{file},
	})
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ Failed to upload %s: %v", name, err))
	}
}

func handleList(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	limit := 5
	if len(i.ApplicationCommandData().Options) > 0 {
		limitStr := i.ApplicationCommandData().Options[0].StringValue()
		if strings.ToLower(limitStr) == "all" {
			limit = 0
		} else if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			data.Title, data.Description = "List Failed", "Invalid limit provided"
			data.Fields = []EmbedField{{Name: "Error", Value: "Use a number or 'all'", Inline: true}}
			return respond(s, i, data)
		}
	}

	backups, err := backupmgr.CurrentBackupManager().ListBackups(limit)
	if err != nil {
		data.Title, data.Description = "List Failed", "Error fetching backups"
		data.Fields = []EmbedField{{Name: "Error", Value: "Failed to fetch backup list", Inline: true}}
		return respond(s, i, data)
	}
	if len(backups) == 0 {
		data.Title, data.Description, data.Color = "Backup List", "No backups found", 0xFFD700
		return respond(s, i, data)
	}

	sort.Slice(backups, func(i, j int) bool { return backups[i].SaveTime.After(backups[j].SaveTime) })
	batchSize := 20
	embeds := []*discordgo.MessageEmbed{}
	for start := 0; start < len(backups); start += batchSize {
		end := start + batchSize
		if end > len(backups) {
			end = len(backups)
		}
		fields := make([]EmbedField, end-start)
		for j, b := range backups[start:end] {
			fields[j] = EmbedField{Name: "📂 " + b.Name, Value: b.SaveTime.Format("January 2, 2006, 3:04 PM")}
		}
		embeds = append(embeds, generateEmbed(EmbedData{
			Title: "📜 Backup Archives", Description: fmt.Sprintf("Showing %d-%d of %d backups", start+1, end, len(backups)),
			Color: 0xFFD700, Fields: fields,
		}))
	}

	// Add download buttons if showing 5 or fewer backups
	var components []discordgo.MessageComponent
	if len(backups) <= 5 {
		var buttons []discordgo.MessageComponent
		for _, b := range backups {
			if len(ButtonDownloadBackupPfx+b.Name) > 100 {
				continue // Long nested names remain available through /download name.
			}
			buttons = append(buttons, discordgo.Button{
				Label:    shortBackupLabel("📥 "+b.Name, 80),
				Style:    discordgo.SecondaryButton,
				CustomID: ButtonDownloadBackupPfx + b.Name,
			})
		}
		if len(buttons) > 0 {
			components = append(components, discordgo.ActionsRow{Components: buttons})
		}
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embeds[0]},
			Components: components,
		},
	}); err != nil {
		return err
	}
	for _, embed := range embeds[1:] {
		time.Sleep(500 * time.Millisecond)
		s.ChannelMessageSendEmbed(i.ChannelID, embed)
	}
	return nil
}

func handleBan(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	return handleBanUnban(s, i, data, banSteamID, "Banned", "Ban Failed", 0xFF0000)
}

func handleUnban(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	return handleBanUnban(s, i, data, unbanSteamID, "Unbanned", "Unban Failed", 0x00FF00)
}

func handleBanUnban(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData, fn func(string) error, successTitle, failTitle string, color int) error {
	if len(i.ApplicationCommandData().Options) == 0 {
		data.Title, data.Description = failTitle, "No SteamID provided"
		data.Fields = []EmbedField{{Name: "Error", Value: "Please provide a SteamID", Inline: true}}
		return respond(s, i, data)
	}
	steamID := i.ApplicationCommandData().Options[0].StringValue()
	if err := fn(steamID); err != nil {
		data.Title, data.Description = failTitle, fmt.Sprintf("Could not %s SteamID %s", strings.ToLower(failTitle[:len(failTitle)-6]), steamID)
		data.Fields = []EmbedField{{Name: "Error", Value: err.Error(), Inline: true}}
		return respond(s, i, data)
	}
	data.Title, data.Description, data.Color = successTitle, fmt.Sprintf("SteamID %s has been %s", steamID, strings.ToLower(successTitle)), color
	data.Fields = []EmbedField{{Name: "Status", Value: "✅ Completed", Inline: true}}
	return respond(s, i, data)
}

func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	cmd := ""
	if len(i.ApplicationCommandData().Options) > 0 {
		cmd = i.ApplicationCommandData().Options[0].StringValue()
	}
	data.Title, data.Description, data.Color = "Server Control", "Sending a command to the gameserver console...", 0x00FF00
	data.Fields = []EmbedField{
		{Name: "Command", Value: "`" + cmd + "`", Inline: false},
		{Name: "Status", Value: "❌ Failed, is the server running and SSCM enabled?", Inline: true},
	}
	data.Color = 0xFF0000
	if gamemgr.InternalIsServerRunning() {
		data.Color = 0x00FF00
		err := commandmgr.WriteCommand(cmd)
		if err != nil {
			data.Fields = []EmbedField{
				{Name: "Command", Value: "`" + cmd + "`", Inline: false},
				{Name: "Error", Value: err.Error(), Inline: true},
			}
			return respond(s, i, data)
		}
		data.Fields = []EmbedField{
			{Name: "Command", Value: "`" + cmd + "`", Inline: false},
			{Name: "Status", Value: "✅ Gameserver received command", Inline: true},
		}
	}

	if err := respond(s, i, data); err != nil {
		return err
	}
	return nil
}

func handleAnnounce(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	cmdData := i.ApplicationCommandData()
	msg := ""
	if len(cmdData.Options) > 0 {
		msg = cmdData.Options[0].StringValue()
	}
	fullCmd := "announce " + msg

	data.Title = "📢 Server Announcement"
	data.Description = "Broadcasting message to connected players via the game announce command."
	data.Color = 0x1E90FF
	data.Fields = []EmbedField{
		{Name: "Command Sent", Value: "`" + fullCmd + "`", Inline: false},
		{Name: "Status", Value: "🕛 Received", Inline: true},
	}

	if !gamemgr.InternalIsServerRunning() {
		data.Color = 0xFF0000
		data.Fields = append(data.Fields, EmbedField{Name: "Warning", Value: "Server not running - command may have no effect", Inline: true})
	} else if !config.GetIsSSCMEnabled() {
		data.Color = 0xFFA500
		data.Fields = append(data.Fields, EmbedField{Name: "Warning", Value: "SSCM is not enabled - announcement not delivered", Inline: true})
	} else {
		if err := commandmgr.WriteCommand(fullCmd); err != nil {
			data.Color = 0xFF0000
			data.Fields = []EmbedField{
				{Name: "Command Sent", Value: "`" + fullCmd + "`", Inline: false},
				{Name: "Error", Value: err.Error(), Inline: true},
			}
			return respond(s, i, data)
		}
		data.Fields = []EmbedField{
			{Name: "Command Sent", Value: "`" + fullCmd + "`", Inline: false},
			{Name: "Status", Value: "✅ Announcement sent to players", Inline: true},
		}
	}

	return respond(s, i, data)
}

// handleDownloadButtonInteraction handles button interactions for downloading backups
func handleDownloadButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	if !strings.HasPrefix(customID, ButtonDownloadBackupPfx) {
		return
	}

	name := strings.TrimPrefix(customID, ButtonDownloadBackupPfx)
	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		respondToButtonError(s, i, "Backup manager is not initialized")
		return
	}
	if err := backupmgr.CheckBackupAvailable(manager, name); err != nil {
		respondToButtonError(s, i, "That backup is no longer available. Run /list again.")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("📥 Preparing %s for download...", name),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to download button: " + err.Error())
		return
	}

	go sendBackupToChannel(s, config.GetControlChannelID(), manager, name)
}

func respondToButtonError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func shortBackupLabel(label string, limit int) string {
	runes := []rune(label)
	if len(runes) > limit {
		return string(runes[:limit-1]) + "…"
	}
	return label
}
