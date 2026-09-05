package discordbot

import "testing"

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
