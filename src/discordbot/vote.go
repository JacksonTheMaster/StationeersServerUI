package discordbot

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/bwmarrin/discordgo"
)

type voteKind string

const (
	voteRestart voteKind = "restart"
	voteRestore voteKind = "restore"
)

type restoreVoteTarget struct {
	Name string
}

type activeVote struct {
	id        uint64
	kind      voteKind
	target    restoreVoteTarget
	required  int
	voters    map[string]struct{}
	expiresAt time.Time
}

var discordVotes = struct {
	sync.Mutex
	nextID               uint64
	restart              *activeVote
	restore              *activeVote
	restartCooldownUntil time.Time
	restoreCooldownUntil time.Time
}{}

type voteCastResult struct {
	message       string
	kind          voteKind
	target        restoreVoteTarget
	cancelledKind voteKind
}

func requiredVoteCount(playerCount, percentage, minimum int) int {
	required := int(math.Ceil(float64(playerCount) * float64(percentage) / 100.0))
	if required < minimum {
		return minimum
	}
	return required
}

func castDiscordVote(kind voteKind, target restoreVoteTarget, userID string) voteCastResult {
	now := time.Now()
	players, _ := statusPanelSnapshot()
	playerCount := len(players)

	discordVotes.Lock()
	var current **activeVote
	var cooldownUntil time.Time
	var threshold, minimum, cooldownMinutes int
	switch kind {
	case voteRestart:
		current = &discordVotes.restart
		cooldownUntil = discordVotes.restartCooldownUntil
		threshold = config.GetDiscordRestartVoteThreshold()
		minimum = config.GetDiscordRestartVoteMinimum()
		cooldownMinutes = config.GetDiscordRestartVoteCooldownMinutes()
	case voteRestore:
		current = &discordVotes.restore
		cooldownUntil = discordVotes.restoreCooldownUntil
		threshold = config.GetDiscordRestoreVoteThreshold()
		minimum = config.GetDiscordRestoreVoteMinimum()
		cooldownMinutes = config.GetDiscordRestoreVoteCooldownMinutes()
	default:
		discordVotes.Unlock()
		return voteCastResult{message: "Unknown vote type."}
	}

	if *current == nil {
		if now.Before(cooldownUntil) {
			discordVotes.Unlock()
			return voteCastResult{message: fmt.Sprintf("That vote is on cooldown until <t:%d:R>.", cooldownUntil.Unix())}
		}
		discordVotes.nextID++
		*current = &activeVote{
			id:        discordVotes.nextID,
			kind:      kind,
			target:    target,
			required:  requiredVoteCount(playerCount, threshold, minimum),
			voters:    make(map[string]struct{}),
			expiresAt: now.Add(time.Duration(config.GetDiscordVoteDurationMinutes()) * time.Minute),
		}
		go expireDiscordVote(kind, (*current).id, (*current).expiresAt)
	} else if kind == voteRestore && (*current).target.Name != target.Name && target.Name != "" {
		discordVotes.Unlock()
		return voteCastResult{message: "A restore vote for a different backup is already active."}
	}

	vote := *current
	if _, exists := vote.voters[userID]; exists {
		message := fmt.Sprintf("You have already voted. Current result: **%d/%d**.", len(vote.voters), vote.required)
		discordVotes.Unlock()
		return voteCastResult{message: message}
	}
	vote.voters[userID] = struct{}{}
	count := len(vote.voters)
	if count < vote.required {
		message := fmt.Sprintf("Vote recorded: **%d/%d**. The vote ends <t:%d:R>.", count, vote.required, vote.expiresAt.Unix())
		discordVotes.Unlock()
		refreshStatusPanel()
		return voteCastResult{message: message}
	}
	// Reserve execution before marking the vote as passed or applying cooldowns.
	// A busy admin operation must not consume an otherwise valid community vote.
	if !beginDiscordAction(string(kind) + " vote") {
		delete(vote.voters, userID)
		discordVotes.Unlock()
		return voteCastResult{message: "Another Discord action is running. Your vote was not added; try again when it finishes."}
	}

	result := voteCastResult{
		message: fmt.Sprintf("Vote passed with **%d/%d** votes.", count, vote.required),
		kind:    kind,
		target:  vote.target,
	}
	*current = nil
	cooldownUntil = now.Add(time.Duration(cooldownMinutes) * time.Minute)
	if kind == voteRestart {
		discordVotes.restartCooldownUntil = cooldownUntil
		if discordVotes.restore != nil {
			result.cancelledKind = voteRestore
			discordVotes.restore = nil
			discordVotes.restoreCooldownUntil = now.Add(time.Duration(config.GetDiscordRestoreVoteCooldownMinutes()) * time.Minute)
		}
	} else {
		discordVotes.restoreCooldownUntil = cooldownUntil
		if discordVotes.restart != nil {
			result.cancelledKind = voteRestart
			discordVotes.restart = nil
			discordVotes.restartCooldownUntil = now.Add(time.Duration(config.GetDiscordRestartVoteCooldownMinutes()) * time.Minute)
		}
	}
	discordVotes.Unlock()
	refreshStatusPanel()
	go executePassedVote(result)
	return result
}

func expireDiscordVote(kind voteKind, id uint64, expiresAt time.Time) {
	timer := time.NewTimer(time.Until(expiresAt))
	defer timer.Stop()
	<-timer.C

	now := time.Now()
	discordVotes.Lock()
	expired := false
	switch kind {
	case voteRestart:
		if discordVotes.restart != nil && discordVotes.restart.id == id {
			discordVotes.restart = nil
			discordVotes.restartCooldownUntil = now.Add(time.Duration(config.GetDiscordRestartVoteCooldownMinutes()) * time.Minute)
			expired = true
		}
	case voteRestore:
		if discordVotes.restore != nil && discordVotes.restore.id == id {
			discordVotes.restore = nil
			discordVotes.restoreCooldownUntil = now.Add(time.Duration(config.GetDiscordRestoreVoteCooldownMinutes()) * time.Minute)
			expired = true
		}
	}
	discordVotes.Unlock()
	if expired {
		SendMessageToEventLogChannel(fmt.Sprintf("🗳️ %s vote expired without enough votes.", voteKindLabel(kind)))
		refreshStatusPanel()
	}
}

func executePassedVote(result voteCastResult) {
	// castDiscordVote already reserved the shared execution slot.
	if result.cancelledKind != "" {
		SendMessageToEventLogChannel(fmt.Sprintf("🗳️ The active %s vote was cancelled because another vote passed.", result.cancelledKind))
	}
	sendVotePanelMessage(fmt.Sprintf("🗳️ **%s vote passed** — carrying out the server action.", voteKindLabel(result.kind)))
	refreshStatusPanel()
	_, err := performServerAction(string(result.kind), result.target.Name, gameActionBackend())
	finishDiscordAction(err)
	if err != nil {
		reportVoteExecutionFailure(string(result.kind), err)
	} else {
		SendMessageToEventLogChannel(fmt.Sprintf("✅ %s vote action completed.", voteKindLabel(result.kind)))
	}
	refreshStatusPanel()
}

func sendVotePanelMessage(message string) {
	session := config.GetDiscordSession()
	if session == nil {
		return
	}

	if !config.GetIsDiscordEnabled() {
		return
	}
	channelID := config.GetStatusPanelChannelID()
	if channelID == "" {
		return
	}
	if _, err := session.ChannelMessageSend(channelID, message); err != nil {
		logger.Discord.Error("Error sending vote result to status panel channel: " + err.Error())
	}
}

func reportVoteExecutionFailure(kind string, err error) {
	message := fmt.Sprintf("❌ **VOTED %s FAILED** — %s", strings.ToUpper(kind), err.Error())
	sendVotePanelMessage(message)
	SendMessageToEventLogChannel(message)
}

func voteKindLabel(kind voteKind) string {
	if kind == voteRestore {
		return "Restore"
	}
	return "Restart"
}

func activeVotesField() *discordgo.MessageEmbedField {
	discordVotes.Lock()
	defer discordVotes.Unlock()
	var lines []string
	if vote := discordVotes.restart; vote != nil {
		lines = append(lines, fmt.Sprintf("🔄 **VOTE FOR RESTART INITIATED** — %d/%d voted • ends <t:%d:R>", len(vote.voters), vote.required, vote.expiresAt.Unix()))
	}
	if vote := discordVotes.restore; vote != nil {
		lines = append(lines, fmt.Sprintf("⏪ **VOTE TO RESTORE BACKUP %s INITIATED** — %d/%d voted • ends <t:%d:R>", vote.target.Name, len(vote.voters), vote.required, vote.expiresAt.Unix()))
	}
	if len(lines) == 0 {
		return nil
	}
	return &discordgo.MessageEmbedField{Name: "🗳️ Active Votes", Value: strings.Join(lines, "\n"), Inline: false}
}

func resetDiscordVotes() {
	discordVotes.Lock()
	discordVotes.nextID++
	discordVotes.restart = nil
	discordVotes.restore = nil
	discordVotes.restartCooldownUntil = time.Time{}
	discordVotes.restoreCooldownUntil = time.Time{}
	discordVotes.Unlock()
}
