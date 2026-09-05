package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

type UpdateTriggerRequest struct {
	AllowUpdate      bool   `json:"allowUpdate"`
	AllowMajorUpdate bool   `json:"allowMajorUpdate"`
	Version          string `json:"version"`
}

func CheckUpdateHandler(w http.ResponseWriter, _ *http.Request) {
	status := update.GetStatus()
	response := map[string]any{
		"status":          "success",
		"operationState":  status.State,
		"updateAvailable": boolString(status.Available),
		"version":         status.Version,
		"majorUpdate":     boolString(status.Major),
	}
	if status.Error != "" {
		response["message"] = status.Error
	}
	writeUpdateJSON(w, http.StatusOK, response)
}

func TriggerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeUpdateJSON(w, http.StatusMethodNotAllowed, map[string]any{"status": "error", "message": "Method not allowed"})
		return
	}

	var request UpdateTriggerRequest
	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeUpdateJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "message": "Invalid update request"})
			return
		}
	}

	if !update.TryStartOperation() {
		writeUpdateJSON(w, http.StatusConflict, map[string]any{"status": "busy", "message": "An update operation is already in progress"})
		return
	}

	if !request.AllowUpdate {
		err, version := update.CheckForUpdates()
		update.SetCheckResult(err, version)
		update.FinishOperation()
		if err != nil {
			writeUpdateJSON(w, http.StatusBadGateway, map[string]any{"status": "failed", "message": err.Error()})
			return
		}
		writeUpdateJSON(w, http.StatusOK, updateResponse(version))
		return
	}

	status := update.GetStatus()
	if !status.Available || status.Version == "" || request.Version != status.Version {
		update.FinishOperation()
		writeUpdateJSON(w, http.StatusConflict, map[string]any{
			"status":  "refresh-required",
			"message": "The available update changed. Refresh the page and review it again",
		})
		return
	}
	if status.Major && !request.AllowMajorUpdate {
		update.FinishOperation()
		writeUpdateJSON(w, http.StatusConflict, map[string]any{
			"status":      "approval-required",
			"message":     "This major update needs explicit approval",
			"version":     status.Version,
			"majorUpdate": true,
		})
		return
	}

	update.SetApplying(status.Version)
	go applyUpdate(status.Version, request.AllowMajorUpdate)
	writeUpdateJSON(w, http.StatusAccepted, map[string]any{
		"status":  "installing",
		"version": status.Version,
		"message": "Update started",
	})
}

func applyUpdate(version string, majorApproved bool) {
	defer update.FinishOperation()

	err, availableVersion := update.ApplyVersion(version, majorApproved)
	if err != nil {
		logger.Install.Error("Manual update failed: " + err.Error())
		update.SetUpdateFailed(version, err)
		return
	}
	if availableVersion != "" {
		err = errors.New("update " + availableVersion + " was not applied")
		logger.Install.Warn(err.Error())
		update.SetUpdateFailed(availableVersion, err)
	}
}

func updateResponse(version string) map[string]any {
	return map[string]any{
		"status":          "success",
		"operationState":  "idle",
		"updateAvailable": boolString(version != ""),
		"version":         version,
		"majorUpdate":     boolString(update.IsMajorUpdate(version)),
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func writeUpdateJSON(w http.ResponseWriter, statusCode int, response map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}
