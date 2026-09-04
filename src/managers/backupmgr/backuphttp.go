package backupmgr

import (
	"encoding/json"
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

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
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
	Name     string
	SaveTime time.Time
	Summary  *SaveSummary `json:",omitempty"`
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
		http.Error(w, "backup manager is not initialized", http.StatusServiceUnavailable)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	backups, err := manager.ListBackups(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	includeSummary := strings.EqualFold(r.URL.Query().Get("include"), "summary")
	response := make([]backupListResponse, len(backups))
	for i := range backups {
		response[i] = backupListResponse{
			Name:     backups[i].Name,
			SaveTime: backups[i].SaveTime,
		}
		if includeSummary && backups[i].SummaryReady {
			response[i].Summary = &backups[i].Summary
		}
	}

	// Summary is optional while an archive is waiting for analysis.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AnalyzeBackupHandler returns lazy, cached world.xml statistics for one backup.
func (h *HTTPHandler) AnalyzeBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		http.Error(w, "backup manager is not initialized", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed, use GET", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		logger.Web.Errorf("Failed to encode backup analysis: %s", err.Error())
	}
}

// RestoreBackupHandler handles requests to restore a backup
func (h *HTTPHandler) RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		http.Error(w, "backup manager is not initialized", http.StatusServiceUnavailable)
		return
	}
	logger.Web.Debug("Received restore request")
	name, err := backupRequestName(r)
	if err == nil {
		err = CheckBackupAvailable(manager, name)
	}
	if err != nil {
		writeBackupError(w, err)
		return
	}
	gamemgr.InternalStopServer()
	if err := manager.RestoreBackup(name); err != nil {
		writeBackupError(w, err)
		return
	}

	w.Write([]byte("Server stopped & Backup restored successfully, Start the server to load the restored backup"))
}

// DownloadBackupRequest represents the JSON request for downloading a backup
type DownloadBackupRequest struct {
	Name string `json:"name"`
}

// DownloadBackupHandler handles requests to download a backup file
func (h *HTTPHandler) DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	manager := handlerManager(h)
	if manager == nil {
		http.Error(w, "backup manager is not initialized", http.StatusServiceUnavailable)
		return
	}
	logger.Web.Debug("Received backup download request")

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed, use POST"})
		return
	}

	var req DownloadBackupRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON request body"})
		return
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "expected one JSON object", http.StatusBadRequest)
		return
	}
	backupData, err := manager.GetBackupFileData(req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(backupErrorStatus(err))
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": backupData.Filename}))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", backupData.Size))
	w.Write(backupData.Data)
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
	default:
		return http.StatusInternalServerError
	}
}

func writeBackupError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), backupErrorStatus(err))
}
