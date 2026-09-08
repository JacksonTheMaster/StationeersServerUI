package web

import (
	"net/http"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/commandmgr"
)

type CommandRequest struct {
	Command string `json:"command"`
}

// CommandResponse represents the JSON response structure.
// CommandHandler handles POST requests to execute commands via commandmgr.
func HandleCommand(w http.ResponseWriter, r *http.Request) {
	if !config.GetIsSSCMEnabled() {
		api.WriteError(w, http.StatusConflict, "sscm_disabled", "SSCM is not enabled")
		return
	}

	var req CommandRequest
	if err := api.DecodeJSONLimit(w, r, &req, 16<<10); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		api.WriteError(w, http.StatusUnprocessableEntity, "command_required", "Command cannot be empty")
		return
	}

	if err := commandmgr.WriteCommand(req.Command); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "command_failed", err.Error())
		return
	}

	api.WriteData(w, http.StatusAccepted, api.CommandReceipt{Command: req.Command, State: "queued"})
}
