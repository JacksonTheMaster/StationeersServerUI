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
		{Name: "/restore <index>", Value: "Restores a backup"},
		{Name: "/download [index]", Value: "Downloads a backup (most recent if no index)"},
		{Name: "/bansteamid <SteamID>", Value: "Bans a player"},
		{Name: "/unbansteamid <SteamID>", Value: "Unbans a player"},
		{Name: "/command <command>", Value: "Sends a command to the gameserver console"},
		{Name: "/announce <message>", Value: "Broadcasts an announcement to all in-game players (via announce cmd)"},
		{Name: "/help", Value: "Shows this help"},
	}
	return respond(s, i, data)
}

func handleRestore(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	index, err := strconv.Atoi(i.ApplicationCommandData().Options[0].StringValue())
	if err != nil {
		data.Title, data.Description = "Restore Failed", "Invalid index provided"
		data.Fields = []EmbedField{{Name: "Error", Value: "Please provide a valid number", Inline: true}}
		return respond(s, i, data)
	}
	data.Title, data.Description, data.Color = "Backup Restore", fmt.Sprintf("Restoring backup #%d...", index), 0xFFA500
	data.Fields = []EmbedField{{Name: "Status", Value: "🕛 Recieved", Inline: true}}
	if err := respond(s, i, data); err != nil {
		return err
	}
	gamemgr.InternalStopServer()
	if err := backupmgr.CurrentBackupManager().RestoreBackup(index); err != nil {
		SendMessageToControlChannel(fmt.Sprintf("❌Failed to restore backup %d: %v", index, err))
		SendMessageToEventLogChannel("⚠️Restore command failed")
		return nil
	}
	SendMessageToControlChannel(fmt.Sprintf("✅Backup %d restored, Starting Server...", index))
	time.Sleep(5 * time.Second)
	gamemgr.InternalStartServer()
	return nil
}

const maxDiscordFileSize = 10 * 1024 * 1024 // 10MB Discord file upload limit

func handleDownload(s *discordgo.Session, i *discordgo.InteractionCreate, data EmbedData) error {
	index := -1 // -1 means most recent

	if len(i.ApplicationCommandData().Options) > 0 {
		index = int(i.ApplicationCommandData().Options[0].IntValue())
	}

	// If no index provided, get the most recent backup index
	if index == -1 {
		backups, err := backupmgr.CurrentBackupManager().ListBackups(1)
		if err != nil || len(backups) == 0 {
			data.Title, data.Description = "Download Failed", "No backups available"
			data.Fields = []EmbedField{{Name: "Error", Value: "Could not find any backups", Inline: true}}
			return respond(s, i, data)
		}
		index = backups[0].Index
	}

	data.Title, data.Description, data.Color = "📥 Backup Download", fmt.Sprintf("Preparing backup #%d for download...", index), 0xFFA500
	data.Fields = []EmbedField{{Name: "Status", Value: "🕛 Processing", Inline: true}}
	if err := respond(s, i, data); err != nil {
		return err
	}

	sendBackupToChannel(s, i.ChannelID, index)
	return nil
}

func sendBackupToChannel(s *discordgo.Session, channelID string, index int) {
	backupData, err := backupmgr.CurrentBackupManager().GetBackupFileData(index)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ Failed to download backup #%d: %v", index, err))
		return
	}

	if backupData.Size > maxDiscordFileSize {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ Backup #%d is too large to upload (%.2f MB > 10 MB limit)", index, float64(backupData.Size)/(1024*1024)))
		return
	}

	file := &discordgo.File{
		Name:        backupData.Filename,
		ContentType: "application/octet-stream",
		Reader:      bytes.NewReader(backupData.Data),
	}

	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("📦 Backup #%d (%s)", index, backupData.SaveTime.Format("Jan 2, 2006 3:04 PM")),
		Files:   []*discordgo.File{file},
	})
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ Failed to upload backup #%d: %v", index, err))
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
			fields[j] = EmbedField{Name: fmt.Sprintf("📂 Backup #%d", b.Index), Value: b.SaveTime.Format("January 2, 2006, 3:04 PM")}
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
			buttons = append(buttons, discordgo.Button{
				Label:    fmt.Sprintf("📥 Download #%d", b.Index),
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("%s%d", ButtonDownloadBackupPfx, b.Index),
			})
		}
		components = append(components, discordgo.ActionsRow{Components: buttons})
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

	indexStr := strings.TrimPrefix(customID, ButtonDownloadBackupPfx)
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		respondToButtonError(s, i, "Invalid backup index")
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("📥 Preparing backup #%d for download...", index),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to download button: " + err.Error())
		return
	}

	go sendBackupToChannel(s, config.GetControlChannelID(), index)
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
