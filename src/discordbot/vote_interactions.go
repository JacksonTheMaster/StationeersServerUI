package discordbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/backupmgr"
	"github.com/bwmarrin/discordgo"
)

const voteSelectCustomID = "ssui_vote_select"

func handleVoteMenuButton(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := buildVoteMenuOptions()
	if len(options) == 0 {
		respondVoteInteraction(session, interaction, "No voting options are currently available.")
		return
	}

	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Choose an action to start or join a vote:",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{CustomID: voteSelectCustomID, Placeholder: "Choose a vote", Options: options},
			}}},
		},
	})
	if err != nil {
		logger.Discord.Error("Error opening vote menu: " + err.Error())
	}
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
			Label:       fmt.Sprintf("Vote to Restore Backup #%d", activeRestore.target.Index),
			Value:       "restore:active",
			Description: fmt.Sprintf("Current result: %d/%d", len(activeRestore.voters), activeRestore.required),
			Emoji:       &discordgo.ComponentEmoji{Name: "⏪"},
		}
		discordVotes.Unlock()
		return append(options, discordgo.SelectMenuOption{
			Label: option.Label, Value: option.Value, Description: option.Description, Emoji: option.Emoji,
		})
	}
	discordVotes.Unlock()

	if backupmgr.CurrentBackupManager() == nil {
		return options
	}
	backups, err := backupmgr.CurrentBackupManager().ListBackups(3)
	if err != nil {
		logger.Discord.Debug("Could not populate restore vote menu: " + err.Error())
		return options
	}
	for _, backup := range backups {
		description := fmt.Sprintf("Day %d • %s", backup.Summary.DaysPlayed, backup.SaveTime.Format("Jan 02 15:04"))
		options = append(options, discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("Restore Backup #%d", backup.Index),
			Value:       fmt.Sprintf("restore:%d", backup.Index),
			Description: description,
			Emoji:       &discordgo.ComponentEmoji{Name: "⏪"},
		})
	}
	return options
}

func handleVoteSelection(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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
	case selection == "restore:active":
		if !config.GetDiscordRestoreVoteEnabled() {
			respondVoteInteraction(session, interaction, "Restore voting is disabled.")
			return
		}
		discordVotes.Lock()
		target := restoreVoteTarget{}
		if discordVotes.restore != nil {
			target = discordVotes.restore.target
		}
		discordVotes.Unlock()
		if target.SaveFile == "" {
			respondVoteInteraction(session, interaction, "That restore vote is no longer active.")
			return
		}
		result = castDiscordVote(voteRestore, target, userID)
	case strings.HasPrefix(selection, "restore:"):
		if !config.GetDiscordRestoreVoteEnabled() {
			respondVoteInteraction(session, interaction, "Restore voting is disabled.")
			return
		}
		index, err := strconv.Atoi(strings.TrimPrefix(selection, "restore:"))
		if err != nil {
			respondVoteInteraction(session, interaction, "Invalid backup selection.")
			return
		}
		target, err := restoreTargetForIndex(index)
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

func restoreTargetForIndex(index int) (restoreVoteTarget, error) {
	if backupmgr.CurrentBackupManager() == nil {
		return restoreVoteTarget{}, fmt.Errorf("backup manager is not initialized")
	}
	backups, err := backupmgr.CurrentBackupManager().ListBackups(0)
	if err != nil {
		return restoreVoteTarget{}, err
	}
	for _, backup := range backups {
		if backup.Index == index {
			return restoreVoteTarget{Index: backup.Index, SaveFile: backup.SaveFile}, nil
		}
	}
	return restoreVoteTarget{}, fmt.Errorf("backup #%d was removed", index)
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
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Discord.Error("Error responding to vote interaction: " + err.Error())
	}
}
