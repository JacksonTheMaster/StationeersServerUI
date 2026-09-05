package discordbot

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/bwmarrin/discordgo"
)

const hubBackupPageSize = 10
const maxDiscordFileSize = 10 * 1024 * 1024

var hubDownloads = make(chan struct{}, 2)

func showBackupBrowser(s *discordgo.Session, i *discordgo.InteractionCreate, page, limit int) {
	if !deferHub(s, i) {
		return
	}
	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		editHubResult(s, i, "Backups unavailable", "The backup manager is not ready.", false)
		return
	}
	backups, err := manager.ListBackups(limit)
	if err != nil {
		editHubResult(s, i, "Backups unavailable", err.Error(), false)
		return
	}
	if len(backups) == 0 {
		editHubResult(s, i, "Backup archives", "No archived backups yet.", true)
		return
	}
	embed, components, err := buildBackupPage(i, backups, page, limit)
	if err != nil {
		editHubResult(s, i, "Backups unavailable", err.Error(), false)
		return
	}
	editHub(s, i, embed, components)
}

func buildBackupPage(i *discordgo.InteractionCreate, backups []backupmgr.BackupSaveFile, page, limit int) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	if len(backups) == 0 {
		return nil, nil, fmt.Errorf("no backups available")
	}
	page = min(max(0, page), (len(backups)-1)/hubBackupPageSize)
	start := page * hubBackupPageSize
	end := min(start+hubBackupPageSize, len(backups))
	dialog := hubDialog{screen: "backups", page: page, limit: limit}
	var options []discordgo.SelectMenuOption
	var description strings.Builder
	for n, backup := range backups[start:end] {
		dialog.names = append(dialog.names, backup.Name)
		meta := backup.SaveTime.Format("02 Jan 2006 · 15:04 MST")
		if backup.SummaryReady {
			meta += fmt.Sprintf(" · Day %d", backup.Summary.DaysPlayed)
		}
		options = append(options, discordgo.SelectMenuOption{Label: shortBackupLabel(backup.Name, 100), Value: strconv.Itoa(n), Description: shortBackupLabel(meta, 100), Emoji: &discordgo.ComponentEmoji{Name: "📦"}})
		fmt.Fprintf(&description, "**%s**\n<t:%d:f>", shortBackupLabel(backup.Name, 130), backup.SaveTime.Unix())
		if backup.SummaryReady {
			fmt.Fprintf(&description, " · Day %d · %d things", backup.Summary.DaysPlayed, backup.Summary.Things)
		}
		description.WriteString("\n\n")
	}
	id, err := storeHubDialog(i, dialog)
	if err != nil {
		return nil, nil, err
	}
	embed := hubEmbed("🗃️ Backup archives", description.String(), 0x5865F2)
	embed.Footer.Text = fmt.Sprintf("Page %d/%d · %d backups · Only visible to you", page+1, (len(backups)+hubBackupPageSize-1)/hubBackupPageSize, len(backups))
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{CustomID: hubDialogPrefix + id + ":select", Placeholder: "Choose a backup", Options: options}}},
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{CustomID: hubDialogPrefix + id + ":prev", Label: "Previous", Style: discordgo.SecondaryButton, Disabled: page == 0},
			discordgo.Button{CustomID: hubDialogPrefix + id + ":next", Label: "Next", Style: discordgo.SecondaryButton, Disabled: end == len(backups)},
			discordgo.Button{CustomID: hubDialogPrefix + id + ":cancel", Label: "Admin actions", Style: discordgo.SecondaryButton},
		}},
	}
	return embed, components, nil
}

func showBackupDetails(s *discordgo.Session, i *discordgo.InteractionCreate, name string, page, limit int) {
	if !deferHub(s, i) {
		return
	}
	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		editHubResult(s, i, "Backups unavailable", "The backup manager is not ready.", false)
		return
	}
	backups, err := manager.ListBackups(0)
	if err != nil {
		editHubResult(s, i, "Backups unavailable", err.Error(), false)
		return
	}
	var selected *backupmgr.BackupSaveFile
	for n := range backups {
		if backups[n].Name == name {
			selected = &backups[n]
			break
		}
	}
	if selected == nil {
		editHubResult(s, i, "Backup unavailable", "This backup was removed. Open the backup browser again.", false)
		return
	}
	embed := hubEmbed("📦 Backup details", name+fmt.Sprintf("\n\nSaved <t:%d:f> · <t:%d:R>", selected.SaveTime.Unix(), selected.SaveTime.Unix()), 0x5865F2)
	if selected.SummaryReady {
		appendSaveStats(embed, &selected.Summary)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: backupStatEmoji("archive_size", "📦") + " Archive size", Value: fmt.Sprintf("%.2f MiB", float64(selected.Summary.ArchiveSize)/(1024*1024)), Inline: true})
	} else {
		embed.Description += "\n\nAnalysis is still pending. The backup manager will fill in its metadata in the background."
	}
	id, err := storeHubDialog(i, hubDialog{screen: "backup", argument: name, page: page, limit: limit})
	if err != nil {
		editHubResult(s, i, "Backups unavailable", err.Error(), false)
		return
	}
	editHub(s, i, embed, []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{CustomID: hubDialogPrefix + id + ":download", Label: "Download", Style: discordgo.PrimaryButton, Emoji: &discordgo.ComponentEmoji{Name: "📥"}},
		discordgo.Button{CustomID: hubDialogPrefix + id + ":restore", Label: "Restore", Style: discordgo.DangerButton, Emoji: &discordgo.ComponentEmoji{Name: "⏪"}},
		discordgo.Button{CustomID: hubDialogPrefix + id + ":back", Label: "Back to backups", Style: discordgo.SecondaryButton},
	}}})
}

// The reply was deferred before any disk I/O. Never fall back to uploading a
// private backup into a public channel when an interaction expires or fails.
func downloadBackupReply(s *discordgo.Session, i *discordgo.InteractionCreate, manager *backupmgr.BackupManager, name string) {
	select {
	case hubDownloads <- struct{}{}:
		defer func() { <-hubDownloads }()
	default:
		editHubResult(s, i, "Download busy", "Two downloads are already being prepared. Please try again shortly.", false)
		return
	}
	if manager == nil {
		editHubResult(s, i, "Download unavailable", "The backup manager is not ready.", false)
		return
	}
	data, err := backupmgr.ReadBackupFileData(manager, name, maxDiscordFileSize)
	if errors.Is(err, backupmgr.ErrBackupTooLarge) {
		editHubResult(s, i, "Backup too large", "This backup exceeds SSUI's 10 MiB Discord upload limit. Download it through the web UI instead.", false)
		return
	}
	if err != nil {
		editHubResult(s, i, "Download failed", err.Error(), false)
		return
	}
	embeds := []*discordgo.MessageEmbed{hubEmbed("📥 Backup download", name, 0x57F287)}
	components := adminHomeButton()
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds, Components: &components,
		Files: []*discordgo.File{{Name: data.Filename, ContentType: "application/octet-stream", Reader: bytes.NewReader(data.Data)}}})
	if err != nil {
		editHubResult(s, i, "Upload failed", "Discord could not accept the file. Try again or download it through the web UI.", false)
	}
}
