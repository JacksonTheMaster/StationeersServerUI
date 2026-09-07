package web

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/loader"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

var reloadMu sync.Mutex

func ServeSvelteUI(w http.ResponseWriter, r *http.Request) {
	htmlFS, err := fs.Sub(config.V1UIFS, "SSUI/onboard_bundled/v2")
	if err != nil {
		http.Error(w, "Error accessing Svelte UI: "+err.Error(), http.StatusInternalServerError)
		return
	}

	htmlFile, err := htmlFS.Open("index.html")
	if err != nil {
		http.Error(w, "Error reading Svelte UI: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer htmlFile.Close()

	// Stream the file content to the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = io.Copy(w, htmlFile)
	if err != nil {
		http.Error(w, "Error writing Svelte UI: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleReloadAll(w http.ResponseWriter, r *http.Request) {
	logger.Web.Debug("Received reloadbackend request from API")
	reloadMu.Lock()
	defer reloadMu.Unlock()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	// Reload all loaders
	loader.ReloadBackend()

	// Set response headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "OK"}); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
