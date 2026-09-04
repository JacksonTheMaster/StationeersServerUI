package discordbot

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
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
	panel     *votePanel
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
	vote          *activeVote
}

func requiredVoteCount(playerCount, percentage, minimum int) int {
	required := int(math.Ceil(float64(playerCount) * float64(percentage) / 100.0))
	if required < minimum {
		return minimum
	}
	return required
}

func castDiscordVote(kind voteKind, target restoreVoteTarget, userID string) voteCastResult {
	return castVote(kind, target, userID, "")
}

func castVote(kind voteKind, target restoreVoteTarget, userID, expectedToken string) voteCastResult {
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

	if expectedToken != "" && (*current == nil || (*current).panel == nil || (*current).panel.token != expectedToken) {
		discordVotes.Unlock()
		return voteCastResult{message: "This vote has ended. Open the vote menu to start a new one."}
	}
	if *current != nil && !now.Before((*current).expiresAt) {
		vote := *current
		discordVotes.Unlock()
		expireVote(kind, vote.id)
		return voteCastResult{message: "This vote has expired."}
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
		(*current).panel = newVotePanel(*current)
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
		refreshVotePanel(vote.panel)
		refreshStatusPanel()
		return voteCastResult{message: message}
	}
	// Reserve execution before marking the vote as passed or applying cooldowns.
	// A busy admin operation must not consume an otherwise valid community vote.
	if !beginDiscordAction(string(kind) + " vote") {
		delete(vote.voters, userID)
		discordVotes.Unlock()
		refreshVotePanel(vote.panel)
		return voteCastResult{message: "Another Discord action is running. Your vote was not added; try again when it finishes."}
	}

	result := voteCastResult{
		message: fmt.Sprintf("Vote passed with **%d/%d** votes.", count, vote.required),
		kind:    kind,
		target:  vote.target,
		vote:    vote,
	}
	finishVotePanelLocked(vote, "passed", "Vote passed. The server action is in progress.")
	var cancelled *activeVote
	*current = nil
	cooldownUntil = now.Add(time.Duration(cooldownMinutes) * time.Minute)
	if kind == voteRestart {
		discordVotes.restartCooldownUntil = cooldownUntil
		if discordVotes.restore != nil {
			result.cancelledKind = voteRestore
			cancelled = discordVotes.restore
			discordVotes.restore = nil
			discordVotes.restoreCooldownUntil = now.Add(time.Duration(config.GetDiscordRestoreVoteCooldownMinutes()) * time.Minute)
		}
	} else {
		discordVotes.restoreCooldownUntil = cooldownUntil
		if discordVotes.restart != nil {
			result.cancelledKind = voteRestart
			cancelled = discordVotes.restart
			discordVotes.restart = nil
			discordVotes.restartCooldownUntil = now.Add(time.Duration(config.GetDiscordRestartVoteCooldownMinutes()) * time.Minute)
		}
	}
	finishVotePanelLocked(cancelled, "cancelled", "Another vote passed first, so this vote was cancelled.")
	discordVotes.Unlock()
	refreshVotePanel(vote.panel)
	if cancelled != nil {
		refreshVotePanel(cancelled.panel)
	}
	refreshStatusPanel()
	go executePassedVote(result)
	return result
}

func expireDiscordVote(kind voteKind, id uint64, expiresAt time.Time) {
	timer := time.NewTimer(time.Until(expiresAt))
	defer timer.Stop()
	<-timer.C
	expireVote(kind, id)
}

func expireVote(kind voteKind, id uint64) {
	now := time.Now()
	discordVotes.Lock()
	var expired *activeVote
	switch kind {
	case voteRestart:
		if discordVotes.restart != nil && discordVotes.restart.id == id {
			expired = discordVotes.restart
			discordVotes.restart = nil
			discordVotes.restartCooldownUntil = now.Add(time.Duration(config.GetDiscordRestartVoteCooldownMinutes()) * time.Minute)
		}
	case voteRestore:
		if discordVotes.restore != nil && discordVotes.restore.id == id {
			expired = discordVotes.restore
			discordVotes.restore = nil
			discordVotes.restoreCooldownUntil = now.Add(time.Duration(config.GetDiscordRestoreVoteCooldownMinutes()) * time.Minute)
		}
	}
	finishVotePanelLocked(expired, "expired", "Voting ended without enough votes. No server action was taken.")
	discordVotes.Unlock()
	if expired != nil {
		refreshVotePanel(expired.panel)
		SendMessageToEventLogChannel(fmt.Sprintf("🗳️ %s vote expired without enough votes.", voteKindLabel(kind)))
		refreshStatusPanel()
	}
}

func executePassedVote(result voteCastResult) {
	// castDiscordVote already reserved the shared execution slot.
	if result.cancelledKind != "" {
		SendMessageToEventLogChannel(fmt.Sprintf("🗳️ The active %s vote was cancelled because another vote passed.", result.cancelledKind))
	}
	SendMessageToEventLogChannel(fmt.Sprintf("🗳️ %s vote passed. Carrying out the server action.", voteKindLabel(result.kind)))
	refreshStatusPanel()
	message, err := performServerAction(string(result.kind), result.target.Name, gameActionBackend())
	finishDiscordAction(err)
	discordVotes.Lock()
	if err != nil {
		finishVotePanelLocked(result.vote, "failed", "The vote passed, but the server action failed. An administrator can check the event log for details.")
	} else {
		finishVotePanelLocked(result.vote, "completed", message)
	}
	discordVotes.Unlock()
	if result.vote != nil {
		refreshVotePanel(result.vote.panel)
	}
	if err != nil {
		reportVoteExecutionFailure(string(result.kind), err)
	} else {
		SendMessageToEventLogChannel(fmt.Sprintf("✅ %s vote action completed.", voteKindLabel(result.kind)))
	}
	refreshStatusPanel()
}

func reportVoteExecutionFailure(kind string, err error) {
	message := fmt.Sprintf("❌ **VOTED %s FAILED** — %s", strings.ToUpper(kind), err.Error())
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
	restart, restore := discordVotes.restart, discordVotes.restore
	finishVotePanelLocked(restart, "cancelled", "The Discord integration was reloaded or disabled. Start a new vote from the hub.")
	finishVotePanelLocked(restore, "cancelled", "The Discord integration was reloaded or disabled. Start a new vote from the hub.")
	discordVotes.nextID++
	discordVotes.restart = nil
	discordVotes.restore = nil
	discordVotes.restartCooldownUntil = time.Time{}
	discordVotes.restoreCooldownUntil = time.Time{}
	discordVotes.Unlock()
	if restart != nil {
		refreshVotePanel(restart.panel)
	}
	if restore != nil {
		refreshVotePanel(restore.panel)
	}
}
