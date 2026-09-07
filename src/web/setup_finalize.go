package web

import (
	"encoding/json"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config/configchanger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/loader"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func SetupFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	if security.SetupRequired() {
		http.Error(w, "Create the owner account before finalizing setup", http.StatusBadRequest)
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		return
	}
	authEnabled := true
	cfg.AuthEnabled = &authEnabled
	if err := configchanger.SaveConfig(cfg, false); err != nil {
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}
	_ = config.SetIsFirstTimeSetup(false)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Setup finalized successfully",
	})
	logger.Web.Info("User setup finalized successfully")
	go loader.ReloadBackend()
}
