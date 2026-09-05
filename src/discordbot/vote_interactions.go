package discordbot

import (
	"fmt"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/bwmarrin/discordgo"
)

const voteSelectCustomID = "ssui_vote_select"

func handleVoteMenuButton(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if !deferHub(session, interaction) {
		return
	}
	options := buildVoteMenuOptions()
	if len(options) == 0 {
		respondVoteInteraction(session, interaction, "No voting options are currently available.")
		return
	}

	editHub(session, interaction, hubEmbed("🗳️ Community votes", "Choose an action to start or join a vote.", 0x5865F2),
		[]discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{CustomID: voteSelectCustomID, Placeholder: "Choose a vote", Options: options},
		}}})
}

func buildVoteMenuOptions() []discordgo.SelectMenuOption {
	var options []discordgo.SelectMenuOption
	if config.GetDiscordRestartVoteEnabled() {
		label := "Start Restart Vote"
		description := "Request a vote to restart the game server"
		discordVotes.Lock()
		if discordVotes.restart != nil {
			label = "Vote for Restart"
			description = fmt.Sprintf("Current result: %d/%d", len(discordVotes.restart.voters), discordVotes.restart.required)
		}
		discordVotes.Unlock()
		options = append(options, discordgo.SelectMenuOption{Label: label, Value: "restart", Description: description, Emoji: &discordgo.ComponentEmoji{Name: "🔄"}})
	}

	if !config.GetDiscordRestoreVoteEnabled() {
		return options
	}
	discordVotes.Lock()
	if activeRestore := discordVotes.restore; activeRestore != nil {
		option := discordgo.SelectMenuOption{
			Label:       shortBackupLabel("Vote to Restore "+activeRestore.target.Name, 100),
			Value:       "restore:" + activeRestore.target.Name,
			Description: fmt.Sprintf("Current result: %d/%d", len(activeRestore.voters), activeRestore.required),
			Emoji:       &discordgo.ComponentEmoji{Name: "⏪"},
		}
		discordVotes.Unlock()
		return append(options, discordgo.SelectMenuOption{
			Label: option.Label, Value: option.Value, Description: option.Description, Emoji: option.Emoji,
		})
	}
	discordVotes.Unlock()

	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		return options
	}
	backups, err := manager.ListBackups(3)
	if err != nil {
		logger.Discord.Debug("Could not populate restore vote menu: " + err.Error())
		return options
	}
	for _, backup := range backups {
		if len("restore:"+backup.Name) > 100 {
			continue
		}
		description := fmt.Sprintf("Day %d • %s", backup.Summary.DaysPlayed, backup.SaveTime.Format("Jan 02 15:04"))
		options = append(options, discordgo.SelectMenuOption{
			Label:       shortBackupLabel("Restore "+backup.Name, 100),
			Value:       "restore:" + backup.Name,
			Description: description,
			Emoji:       &discordgo.ComponentEmoji{Name: "⏪"},
		})
	}
	return options
}

func handleVoteSelection(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if !deferHub(session, interaction) {
		return
	}
	values := interaction.MessageComponentData().Values
	if len(values) != 1 {
		respondVoteInteraction(session, interaction, "Invalid vote selection.")
		return
	}
	userID := interactionUserID(interaction)
	if userID == "" {
		respondVoteInteraction(session, interaction, "Could not identify your Discord account.")
		return
	}

	selection := values[0]
	var result voteCastResult
	switch {
	case selection == "restart":
		if !config.GetDiscordRestartVoteEnabled() {
			respondVoteInteraction(session, interaction, "Restart voting is disabled.")
			return
		}
		result = castDiscordVote(voteRestart, restoreVoteTarget{}, userID)
	case strings.HasPrefix(selection, "restore:"):
		if !config.GetDiscordRestoreVoteEnabled() {
			respondVoteInteraction(session, interaction, "Restore voting is disabled.")
			return
		}
		target, err := restoreTargetForName(strings.TrimPrefix(selection, "restore:"))
		if err != nil {
			respondVoteInteraction(session, interaction, "The selected backup is no longer available: "+err.Error())
			return
		}
		result = castDiscordVote(voteRestore, target, userID)
	default:
		respondVoteInteraction(session, interaction, "Unknown vote selection.")
		return
	}
	respondVoteInteraction(session, interaction, result.message)
}

func restoreTargetForName(name string) (restoreVoteTarget, error) {
	manager := backupmgr.CurrentBackupManager()
	if manager == nil {
		return restoreVoteTarget{}, fmt.Errorf("backup manager is not initialized")
	}
	if err := backupmgr.CheckBackupAvailable(manager, name); err != nil {
		return restoreVoteTarget{}, err
	}
	return restoreVoteTarget{Name: name}, nil
}

func interactionUserID(interaction *discordgo.InteractionCreate) string {
	var user *discordgo.User
	if interaction.Member != nil {
		user = interaction.Member.User
	}
	if user == nil {
		user = interaction.User
	}
	if user == nil {
		return ""
	}
	return user.ID
}

func respondVoteInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate, message string) {
	editHub(session, interaction, hubEmbed("🗳️ Community votes", message, 0x5865F2), nil)
}
