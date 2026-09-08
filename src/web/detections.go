package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/detectionmgr"
	"github.com/google/uuid"
)

type createDetectionRequest struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
	Message string `json:"message"`
}

func ListDetections(w http.ResponseWriter, _ *http.Request) {
	api.WriteData(w, http.StatusOK, map[string]any{"detections": detectionmgr.GetCustomDetections()})
}

func CreateDetection(w http.ResponseWriter, r *http.Request) {
	var request createDetectionRequest
	if err := api.DecodeJSON(w, r, &request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	request.Type = strings.ToLower(strings.TrimSpace(request.Type))
	request.Pattern = strings.TrimSpace(request.Pattern)
	request.Message = strings.TrimSpace(request.Message)
	if request.Type != "keyword" && request.Type != "regex" {
		api.WriteError(w, http.StatusBadRequest, "invalid_detection_type", "Type must be keyword or regex")
		return
	}
	if request.Pattern == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_pattern", "Pattern cannot be empty")
		return
	}
	if request.Message == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_message", "Message cannot be empty")
		return
	}

	detection := detectionmgr.CustomDetection{
		ID:        uuid.NewString(),
		Type:      request.Type,
		Pattern:   request.Pattern,
		EventType: string(detectionmgr.EventCustomDetection),
		Message:   request.Message,
	}
	if err := detectionmgr.AddCustomDetection(detection); err != nil {
		if strings.Contains(err.Error(), "invalid regex pattern") {
			api.WriteError(w, http.StatusBadRequest, "invalid_pattern", err.Error())
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "detection_save_failed", "Could not save the detection")
		return
	}
	api.WriteData(w, http.StatusCreated, map[string]any{"detection": detection})
}

func DeleteDetection(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "missing_detection_id", "Detection ID is required")
		return
	}
	if err := detectionmgr.RemoveCustomDetection(id); err != nil {
		if errors.Is(err, detectionmgr.ErrDetectionNotFound) {
			api.WriteError(w, http.StatusNotFound, "detection_not_found", "Detection not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "detection_delete_failed", "Could not delete the detection")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
