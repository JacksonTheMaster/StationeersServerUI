package backupmgr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
)

// HTTPHandler provides HTTP endpoints for backup operations
type HTTPHandler struct {
	mu      sync.RWMutex
	manager *BackupManager
}

func handlerManager(h *HTTPHandler) *BackupManager {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.manager
}

type backupListResponse struct {
	Name     string       `json:"name"`
	SaveTime time.Time    `json:"saveTime"`
	Summary  *SaveSummary `json:"summary,omitempty"`
}

type backupCollection struct {
	Items      []backupListResponse `json:"items"`
	NextCursor string               `json:"nextCursor,omitempty"`
}

// NewHTTPHandler creates a new HTTP handler for backups
func NewHTTPHandler(manager *BackupManager) *HTTPHandler {
	handler := &HTTPHandler{manager: manager}
	RegisterHTTPHandler(handler) // Register this handler for automatic updates
	return handler
}

// ListBackupsHandler handles requests to list available backups
func (h *HTTPHandler) ListBackupsHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "backup_unavailable", "Backup manager is not initialized")
		return
	}
	query := r.URL.Query()
	for key := range query {
		if key != "limit" && key != "include" && key != "cursor" {
			api.WriteError(w, http.StatusBadRequest, "invalid_query", "Unknown query parameter: "+key)
			return
		}
	}
	limitStr := query.Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 500 {
			api.WriteError(w, http.StatusBadRequest, "invalid_limit", "Limit must be between 1 and 500")
			return
		}
	}

	include := query.Get("include")
	if include != "" && !strings.EqualFold(include, "summary") {
		api.WriteError(w, http.StatusBadRequest, "invalid_include", "Include must be summary")
		return
	}

	backups, err := manager.ListBackups(0)
	if err != nil {
		writeBackupError(w, err)
		return
	}
	start := 0
	if cursor := query.Get("cursor"); cursor != "" {
		start = -1
		for i := range backups {
			if backups[i].Name == cursor {
				start = i + 1
				break
			}
		}
		if start < 0 {
			api.WriteError(w, http.StatusBadRequest, "invalid_cursor", "Backup cursor is no longer available")
			return
		}
	}
	end := len(backups)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	includeSummary := strings.EqualFold(include, "summary")
	items := make([]backupListResponse, 0, end-start)
	for i := start; i < end; i++ {
		item := backupListResponse{
			Name:     backups[i].Name,
			SaveTime: backups[i].SaveTime,
		}
		if includeSummary && backups[i].SummaryReady {
			item.Summary = &backups[i].Summary
		}
		items = append(items, item)
	}

	response := backupCollection{Items: items}
	if end < len(backups) && len(items) > 0 {
		response.NextCursor = items[len(items)-1].Name
	}
	api.WriteData(w, http.StatusOK, response)
}

// AnalyzeBackupHandler returns lazy, cached world.xml statistics for one backup.
func (h *HTTPHandler) AnalyzeBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "backup_unavailable", "Backup manager is not initialized")
		return
	}

	name, err := backupRequestName(r)
	if err != nil {
		writeBackupError(w, err)
		return
	}
	analysis, err := manager.AnalyzeBackup(r.Context(), name)
	if err != nil {
		writeBackupError(w, err)
		return
	}

	api.WriteData(w, http.StatusOK, analysis)
}

// RestoreBackupHandler handles requests to restore a backup
func (h *HTTPHandler) RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "backup_unavailable", "Backup manager is not initialized")
		return
	}
	logger.Web.Debug("Received restore request")
	var request struct {
		Name string `json:"name"`
	}
	if err := api.DecodeJSONLimit(w, r, &request, 4096); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validateBackupName(request.Name); err != nil {
		writeBackupError(w, err)
		return
	}
	err := CheckBackupAvailable(manager, request.Name)
	if err != nil {
		writeBackupError(w, err)
		return
	}
	if config.GetIsGameServerRunning() {
		if err := gamemgr.InternalStopServer(); err != nil {
			api.WriteError(w, http.StatusConflict, "server_stop_failed", "The game server could not be stopped: "+err.Error())
			return
		}
	}
	if err := manager.RestoreBackup(request.Name); err != nil {
		writeBackupError(w, err)
		return
	}

	api.WriteData(w, http.StatusOK, api.Message{Message: "Backup restored. Start the game server to load it"})
}

// DownloadBackupHandler handles requests to download a backup file
func (h *HTTPHandler) DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "backup_unavailable", "Backup manager is not initialized")
		return
	}
	logger.Web.Debug("Received backup download request")

	name, err := backupRequestName(r)
	if err != nil {
		writeBackupError(w, err)
		return
	}
	file, backup, err := OpenBackupFile(manager, name)
	if err != nil {
		writeBackupError(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": backup.Filename}))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", backup.Size))
	if _, err := io.Copy(w, file); err != nil {
		logger.Web.Errorf("Failed to stream backup %s: %v", name, err)
	}
}

func backupRequestName(r *http.Request) (string, error) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", ErrInvalidBackupName
	}
	values := query["name"]
	if len(query) != 1 || len(values) != 1 {
		return "", fmt.Errorf("%w: provide exactly one name parameter", ErrInvalidBackupName)
	}
	if err := validateBackupName(values[0]); err != nil {
		return "", err
	}
	return values[0], nil
}

func backupErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidBackupName):
		return http.StatusBadRequest
	case errors.Is(err, os.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, context.Canceled):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func writeBackupError(w http.ResponseWriter, err error) {
	code := "backup_failed"
	switch {
	case errors.Is(err, ErrInvalidBackupName):
		code = "invalid_backup_name"
	case errors.Is(err, os.ErrNotExist):
		code = "backup_not_found"
	case errors.Is(err, context.Canceled):
		code = "backup_unavailable"
	}
	api.WriteError(w, backupErrorStatus(err), code, err.Error())
}
