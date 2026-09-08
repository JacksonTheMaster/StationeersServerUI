package web

import (
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
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
	api.WriteData(w, http.StatusOK, loadWorldGenerationCatalog())
}
