package backupmgr

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	Index    int
	SaveFile string
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

	// Check if classic mode is requested
	mode := r.URL.Query().Get("mode")
	if mode == "classic" {
		// Format the response in the classic format
		classicResponses := make([]string, 0, len(backups))
		for _, backup := range backups {
			// Format according to classic view: "BackupIndex: X, Created: DD.MM.YYYY HH:MM:SS"
			classicLine := fmt.Sprintf("BackupIndex: %d, Created: %s",
				backup.Index,
				backup.SaveTime.Format("02.01.2006 15:04:05"))
			classicResponses = append(classicResponses, classicLine)
		}

		// Return plain text response for classic mode
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Join(classicResponses, "\n")))
		return
	}

	includeSummary := strings.EqualFold(r.URL.Query().Get("include"), "summary")
	response := make([]backupListResponse, len(backups))
	for i := range backups {
		response[i] = backupListResponse{
			Index:    backups[i].Index,
			SaveFile: backups[i].SaveFile,
			SaveTime: backups[i].SaveTime,
		}
		if includeSummary && backups[i].SummaryReady {
			response[i].Summary = &backups[i].Summary
		}
	}

	// Default JSON response. Summary remains opt-in for legacy API consumers.
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

	index, err := strconv.Atoi(r.URL.Query().Get("index"))
	if err != nil || index < 0 {
		http.Error(w, "valid index parameter is required", http.StatusBadRequest)
		return
	}
	analysis, err := getBackupAnalysis(r.Context(), manager, index, r.URL.Query().Get("file"))
	if err != nil {
		if strings.Contains(err.Error(), "out of range") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "index parameter is required", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "invalid index parameter", http.StatusBadRequest)
		return
	}

	gamemgr.InternalStopServer()

	if file := r.URL.Query().Get("file"); file != "" {
		err = manager.RestoreBackupFile(file)
	} else {
		err = manager.RestoreBackup(index)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Server stopped & Backup restored successfully, Start the server to load the restored backup"))
}

// DownloadBackupRequest represents the JSON request for downloading a backup
type DownloadBackupRequest struct {
	Index    int    `json:"index"`
	SaveFile string `json:"saveFile,omitempty"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON request body"})
		return
	}

	backupData, err := getBackupFileData(manager, req.Index, req.SaveFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "out of range") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", backupData.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", backupData.Size))
	w.Write(backupData.Data)
}
