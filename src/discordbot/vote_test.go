package discordbot

import (
	"strings"
	"testing"
	"time"
)

func TestRequiredVoteCountRoundsUpAndHonorsMinimum(t *testing.T) {
	tests := []struct {
		players    int
		percentage int
		minimum    int
		want       int
	}{
		{players: 0, percentage: 60, minimum: 1, want: 1},
		{players: 0, percentage: 100, minimum: 2, want: 2},
		{players: 1, percentage: 60, minimum: 1, want: 1},
		{players: 2, percentage: 60, minimum: 1, want: 2},
		{players: 4, percentage: 60, minimum: 1, want: 3},
		{players: 1, percentage: 100, minimum: 2, want: 2},
		{players: 5, percentage: 100, minimum: 2, want: 5},
	}
	for _, test := range tests {
		if got := requiredVoteCount(test.players, test.percentage, test.minimum); got != test.want {
			t.Errorf("requiredVoteCount(%d, %d, %d) = %d, want %d", test.players, test.percentage, test.minimum, got, test.want)
		}
	}
}

func TestActiveVotesFieldListsBothVotesAtPanelBottom(t *testing.T) {
	resetDiscordVotes()
	defer resetDiscordVotes()
	expires := time.Now().Add(5 * time.Minute)
	discordVotes.Lock()
	discordVotes.restart = &activeVote{required: 3, voters: map[string]struct{}{"a": {}, "b": {}}, expiresAt: expires}
	discordVotes.restore = &activeVote{target: restoreVoteTarget{Name: "040926_172441_auto.save"}, required: 2, voters: map[string]struct{}{"a": {}}, expiresAt: expires}
	discordVotes.Unlock()

	field := activeVotesField()
	if field == nil {
		t.Fatal("activeVotesField() = nil, want active vote summary")
	}
	for _, text := range []string{"VOTE FOR RESTART INITIATED", "2/3 voted", "VOTE TO RESTORE BACKUP 040926_172441_auto.save INITIATED", "1/2 voted"} {
		if !strings.Contains(field.Value, text) {
			t.Errorf("active vote field %q does not contain %q", field.Value, text)
		}
	}
}
