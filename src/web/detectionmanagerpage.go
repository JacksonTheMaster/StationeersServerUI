package web

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func ServeDetectionManager(w http.ResponseWriter, r *http.Request) {
	detectionmanagerFS, err := fs.Sub(config.V1UIFS, "SSUI/onboard_bundled/detectionmanager")
	if err != nil {
		http.Error(w, "Error accessing Virt FS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	htmlFile, err := detectionmanagerFS.Open("detectionmanager.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading detectionmanager.html: %v", err), http.StatusInternalServerError)
		return
	}
	defer htmlFile.Close()

	htmlContent, err := io.ReadAll(htmlFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading detectionmanager.html content: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}
