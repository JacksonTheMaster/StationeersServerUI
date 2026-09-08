package web

import (
	"net/http"
	"sort"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
)

// PrintConnectedPlayersHandler handles HTTP requests to list connected players.
func HandleConnectedPlayersList(w http.ResponseWriter, r *http.Request) {
	detector := detectionmgr.GetDetector()
	players := detectionmgr.GetPlayers(detector)

	playerList := make([]api.Player, 0, len(players))
	for steamID, username := range players {
		playerList = append(playerList, api.Player{Username: username, SteamID: steamID})
	}
	sort.Slice(playerList, func(i, j int) bool { return playerList[i].Username < playerList[j].Username })
	api.WriteData(w, http.StatusOK, map[string]any{"players": playerList})
}
