package web

import (
	"encoding/json"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/worldgen"
)

func loadWorldGenerationCatalog() worldgen.Catalog {
	catalog, err := worldgen.DefaultLoader.Load(config.GetExePath())
	if err != nil {
		logger.Web.Warn("Could not load world-generation options from the installed game files; using bundled fallback: " + err.Error())
	}
	return catalog
}

func HandleWorldGenerationCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(loadWorldGenerationCatalog()); err != nil {
		logger.Web.Error("Failed to encode world-generation catalog: " + err.Error())
	}
}
