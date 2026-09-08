package web

import (
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config/configchanger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/loader"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func SetupFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	if security.SetupRequired() {
		api.WriteError(w, http.StatusConflict, "setup_incomplete", "Create the owner account before finalizing setup")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "config_load_failed", "Failed to load configuration")
		return
	}
	authEnabled := true
	cfg.AuthEnabled = &authEnabled
	if err := configchanger.SaveConfig(cfg, false); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "config_save_failed", "Failed to save configuration")
		return
	}
	_ = config.SetIsFirstTimeSetup(false)
	api.WriteData(w, http.StatusOK, api.Message{Message: "Setup finalized successfully"})
	logger.Web.Info("User setup finalized successfully")
	go loader.ReloadBackend()
}
