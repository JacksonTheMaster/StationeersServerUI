package web

import (
	"errors"
	"net/http"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/connectivity"
)

func GetConnectivityStatus(w http.ResponseWriter, _ *http.Request) {
	api.WriteData(w, http.StatusOK, connectivity.StatusSnapshot())
}

func RunConnectivityCheck(w http.ResponseWriter, _ *http.Request) {
	status, err := connectivity.Check()
	if errors.Is(err, connectivity.ErrCheckInProgress) {
		api.WriteData(w, http.StatusConflict, status)
		return
	}
	api.WriteData(w, http.StatusOK, status)
}
