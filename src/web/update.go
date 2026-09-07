package web

import (
	"errors"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

type updateRequest struct {
	Action            string `json:"action"`
	Version           string `json:"version"`
	ConfirmMajor      bool   `json:"confirmMajor"`
	ConfirmPrerelease bool   `json:"confirmPrerelease"`
}

func CheckUpdateHandler(w http.ResponseWriter, _ *http.Request) {
	api.WriteData(w, http.StatusOK, update.StatusSnapshot())
}

func TriggerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var request updateRequest
	if err := api.DecodeJSON(w, r, &request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if request.Action == "check" {
		status, err := update.RefreshStatus(r.Context(), true)
		if errors.Is(err, update.ErrUpdateBusy) {
			api.WriteError(w, http.StatusConflict, "update_busy", err.Error())
			return
		}
		if err != nil {
			api.WriteError(w, http.StatusBadGateway, "update_check_failed", err.Error())
			return
		}
		api.WriteData(w, http.StatusOK, status)
		return
	}
	if request.Action != "install" {
		api.WriteError(w, http.StatusBadRequest, "invalid_action", "Choose check or install")
		return
	}
	installRequest := update.InstallRequest{
		Version:           request.Version,
		ConfirmMajor:      request.ConfirmMajor,
		ConfirmPrerelease: request.ConfirmPrerelease,
	}
	if err := update.QueueInstall(installRequest); err != nil {
		if errors.Is(err, update.ErrUpdateBusy) {
			api.WriteError(w, http.StatusConflict, "update_busy", err.Error())
			return
		}
		if errors.Is(err, update.ErrContainerManaged) {
			api.WriteError(w, http.StatusConflict, "container_managed", err.Error())
			return
		}
		api.WriteError(w, http.StatusBadRequest, "update_rejected", err.Error())
		return
	}
	api.WriteData(w, http.StatusAccepted, map[string]string{"status": "installing", "version": request.Version})
}
