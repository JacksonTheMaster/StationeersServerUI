package discordbot

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/bwmarrin/discordgo"
)

const voteButtonPrefix = "ssui_vote_cast:"
const voteResultLifetime = 15 * time.Minute

type votePanel struct {
	// Serialize Discord writes so an older count cannot overwrite the result.
	mu                          sync.Mutex
	session                     *discordgo.Session
	channelID, messageID, token string
	retired                     bool
	vote                        *activeVote
	// Vote state is protected by discordVotes, including after the vote closes.
	state, result string
	deleteAt      time.Time
}

func newVotePanel(vote *activeVote) *votePanel {
	return &votePanel{session: config.GetDiscordSession(), channelID: config.GetStatusPanelChannelID(),
		token: rand.Text(), vote: vote, state: "active"}
}

// Caller holds discordVotes. Execution updates the same result without extending
// its lifetime: the fifteen minutes start when voting ends.
func finishVotePanelLocked(vote *activeVote, state, result string) {
	if vote == nil || vote.panel == nil {
		return
	}
	panel := vote.panel
	panel.state, panel.result = state, result
	if panel.deleteAt.IsZero() {
		panel.deleteAt = time.Now().Add(voteResultLifetime)
		if panel.session != nil {
			time.AfterFunc(voteResultLifetime, func() { deleteVotePanel(panel, 0) })
		}
	}
}

// Caller holds discordVotes. Rendering uses a stable view of the vote counts.
func buildVotePanelLocked(panel *votePanel) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	vote := panel.vote
	count := len(vote.voters)
	embed := &discordgo.MessageEmbed{
		Title:       "🗳️ " + voteKindLabel(vote.kind) + " vote",
		Description: "Use the button below to vote. Each Discord account gets one vote.",
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: "SSUI community vote"},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Votes received", Value: fmt.Sprintf("**%d**", count), Inline: true},
			{Name: "Votes required", Value: fmt.Sprintf("**%d**", vote.required), Inline: true},
			{Name: "Still needed", Value: fmt.Sprintf("**%d**", max(0, vote.required-count)), Inline: true},
		},
	}
	if vote.kind == voteRestore {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Backup", Value: shortBackupLabel(vote.target.Name, 1000)})
	}
	active := panel.state == "active"
	if active {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Voting closes", Value: fmt.Sprintf("<t:%d:R>", vote.expiresAt.Unix())})
	} else {
		label := map[string]string{"passed": "Passed", "completed": "Completed", "failed": "Action failed", "expired": "Expired", "cancelled": "Cancelled"}[panel.state]
		embed.Title += " - " + label
		embed.Description = panel.result
		embed.Color = 0xFEE75C
		if panel.state == "completed" {
			embed.Color = 0x57F287
		}
		if panel.state == "failed" || panel.state == "expired" {
			embed.Color = 0xED4245
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Result visible until", Value: fmt.Sprintf("<t:%d:R>", panel.deleteAt.Unix())})
	}
	label := "Vote for " + strings.ToLower(voteKindLabel(vote.kind))
	if !active {
		label = "Voting closed"
	}
	components := []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{CustomID: voteButtonPrefix + string(vote.kind) + ":" + panel.token, Label: label, Style: discordgo.SuccessButton, Disabled: !active},
	}}}
	return embed, components
}

func refreshVotePanel(panel *votePanel) {
	if panel == nil || panel.session == nil || panel.channelID == "" {
		return
	}
	panel.mu.Lock()
	defer panel.mu.Unlock()
	if panel.retired {
		return
	}
	discordVotes.Lock()
	embed, components := buildVotePanelLocked(panel)
	discordVotes.Unlock()
	mentions := &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}
	if panel.messageID == "" {
		message, err := panel.session.ChannelMessageSendComplex(panel.channelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embed}, Components: components, AllowedMentions: mentions,
		})
		if err != nil {
			logger.Discord.Warnf("Could not send %s vote panel: %v", panel.vote.kind, err)
			return
		}
		panel.messageID = message.ID
		return
	}
	embeds := []*discordgo.MessageEmbed{embed}
	_, err := panel.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: panel.channelID, ID: panel.messageID, Embeds: &embeds, Components: &components, AllowedMentions: mentions,
	})
	if err != nil {
		if discordMessageNotFound(err) {
			panel.retired = true
		}
		logger.Discord.Warnf("Could not update %s vote panel: %v", panel.vote.kind, err)
	}
}

func discordMessageNotFound(err error) bool {
	var apiError *discordgo.RESTError
	return errors.As(err, &apiError) && apiError.Response != nil && apiError.Response.StatusCode == http.StatusNotFound
}

func deleteVotePanel(panel *votePanel, attempt int) {
	panel.mu.Lock()
	defer panel.mu.Unlock()
	panel.retired = true
	if panel.messageID == "" {
		return
	}
	err := panel.session.ChannelMessageDelete(panel.channelID, panel.messageID)
	if err == nil || discordMessageNotFound(err) {
		return
	}
	logger.Discord.Warnf("Could not remove finished vote panel: %v", err)
	if attempt < 2 {
		time.AfterFunc(time.Minute, func() { deleteVotePanel(panel, attempt+1) })
	}
}

func handleVotePanelButton(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if !requireHubInteraction(session, interaction) || !deferHub(session, interaction) {
		return
	}
	parts := strings.Split(strings.TrimPrefix(interaction.MessageComponentData().CustomID, voteButtonPrefix), ":")
	if len(parts) != 2 {
		respondVoteInteraction(session, interaction, "Invalid vote button.")
		return
	}
	kind := voteKind(parts[0])
	if (kind == voteRestart && !config.GetDiscordRestartVoteEnabled()) || (kind == voteRestore && !config.GetDiscordRestoreVoteEnabled()) {
		respondVoteInteraction(session, interaction, "This type of vote is disabled.")
		return
	}
	discordVotes.Lock()
	vote := discordVotes.restart
	if kind == voteRestore {
		vote = discordVotes.restore
	}
	valid := (kind == voteRestart || kind == voteRestore) && vote != nil && vote.panel != nil && vote.panel.token == parts[1]
	var target restoreVoteTarget
	if valid {
		target = vote.target
	}
	discordVotes.Unlock()
	if !valid {
		respondVoteInteraction(session, interaction, "This vote has ended. This button cannot start or join a different vote.")
		return
	}
	if kind == voteRestore {
		if _, err := restoreTargetForName(target.Name); err != nil {
			respondVoteInteraction(session, interaction, "That backup is no longer available: "+err.Error())
			return
		}
	}
	result := castVote(kind, target, interactionUserID(interaction), parts[1])
	respondVoteInteraction(session, interaction, result.message)
}
