package web

import (
	"net/http"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/worldgen"
)

func loadWorldGenerationCatalog() worldgen.Catalog {
	catalog, err := worldgen.DefaultLoader.Load(config.GetExePath())
	if err != nil {
		logger.Web.Warn("Could not load world-generation options from the installed game files; using bundled fallback: " + err.Error())
	}
	return catalog
}

func HandleWorldGenerationCatalog(w http.ResponseWriter, r *http.Request) {
	api.WriteData(w, http.StatusOK, loadWorldGenerationCatalog())
}
