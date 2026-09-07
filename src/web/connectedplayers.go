package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
)

// PrintConnectedPlayersHandler handles HTTP requests to list connected players.
func HandleConnectedPlayersList(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	detector := detectionmgr.GetDetector()
	players := detectionmgr.GetPlayers(detector)

	type connectedPlayer struct {
		Username string `json:"username"`
		SteamID  string `json:"steamId"`
	}
	playerList := make([]connectedPlayer, 0, len(players))
	for steamID, username := range players {
		playerList = append(playerList, connectedPlayer{Username: username, SteamID: steamID})
	}
	sort.Slice(playerList, func(i, j int) bool { return playerList[i].Username < playerList[j].Username })

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(playerList); err != nil {
		http.Error(w, "Failed to encode player list", http.StatusInternalServerError)
	}
}
