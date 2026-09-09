package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

type UpdateTriggerRequest struct {
	Action            string `json:"action"`
	Version           string `json:"version"`
	ConfirmMajor      bool   `json:"confirmMajor"`
	ConfirmPrerelease bool   `json:"confirmPrerelease"`
}

func CheckUpdateHandler(w http.ResponseWriter, _ *http.Request) {
	writeUpdateJSON(w, http.StatusOK, update.StatusSnapshot())
}

func TriggerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeUpdateJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var request UpdateTriggerRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeUpdateJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid update request"})
		return
	}
	if request.Action == "check" {
		status, err := update.RefreshStatus(r.Context(), true)
		if errors.Is(err, update.ErrUpdateBusy) {
			writeUpdateJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if err != nil {
			writeUpdateJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeUpdateJSON(w, http.StatusOK, status)
		return
	}
	if request.Action != "install" {
		writeUpdateJSON(w, http.StatusBadRequest, map[string]string{"error": "Choose check or install"})
		return
	}
	err := update.QueueInstall(update.InstallRequest{
		Version:           request.Version,
		ConfirmMajor:      request.ConfirmMajor,
		ConfirmPrerelease: request.ConfirmPrerelease,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, update.ErrUpdateBusy) {
			status = http.StatusConflict
		}
		writeUpdateJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeUpdateJSON(w, http.StatusAccepted, map[string]string{"status": "installing", "version": request.Version})
}

func writeUpdateJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
