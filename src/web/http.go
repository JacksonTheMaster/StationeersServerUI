package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"
)

// StartServer HTTP handler
func StartServer(w http.ResponseWriter, r *http.Request) {
	logger.Web.Debug("Received start request from API")
	if err := gamemgr.InternalStartServer(); err != nil {
		logger.Web.Error("Error starting server: " + err.Error())
		if strings.Contains(err.Error(), "already running") {
			api.WriteError(w, http.StatusConflict, "server_already_running", "The game server is already running")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "server_start_failed", err.Error())
		return
	}
	logger.Web.Info("Server started.")
	api.WriteData(w, http.StatusAccepted, api.Message{Message: "Game server is starting"})
}

// StopServer HTTP handler
func StopServer(w http.ResponseWriter, r *http.Request) {
	logger.Web.Debug("Received stop request from API")
	if err := gamemgr.InternalStopServer(); err != nil {
		if err.Error() == "server not running" {
			logger.Web.Warn("Server not running or was already stopped")
			api.WriteData(w, http.StatusOK, api.Message{Message: "Game server is already stopped"})
			return
		}
		logger.Web.Error("Error stopping server: " + err.Error())
		api.WriteError(w, http.StatusInternalServerError, "server_stop_failed", err.Error())
		return
	}
	detectionmgr.ClearPlayers(detectionmgr.GetDetector())
	logger.Web.Info("Server stopped.")
	api.WriteData(w, http.StatusOK, api.Message{Message: "Game server stopped"})
}

func GetGameServerRunState(w http.ResponseWriter, r *http.Request) {
	uptime := gamemgr.GetServerUptime()
	var startedAt *time.Time
	if started := gamemgr.GetServerStartTime(); !started.IsZero() {
		startedAt = &started
	}
	api.WriteData(w, http.StatusOK, api.ServerStatus{
		Running:       config.GetIsGameServerRunning(),
		State:         string(gamemgr.GetServerState()),
		StartedAt:     startedAt,
		UptimeSeconds: int64(uptime / time.Second),
		ServerID:      gamemgr.GameServerUUID.String(),
		WorldID:       config.GetWorldID(),
	})
}

func HandleIsSSCMEnabled(w http.ResponseWriter, r *http.Request) {
	api.WriteData(w, http.StatusOK, api.SSCMStatus{Enabled: config.GetIsSSCMEnabled()})
}

// run SteamCMD from API, but only allow once every 5 minutes to "kinda" prevent concurrent executions although that woluldnt hurn.
// If the user has a 5mbit connection, I cannot help them anyways.
func HandleRunSteamCMD(w http.ResponseWriter, r *http.Request) {
	logger.Core.Info("Running SteamCMD")
	_, err := steamcmd.InstallAndRunSteamCMD()
	if err != nil {
		api.WriteError(w, http.StatusBadGateway, "steamcmd_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, api.Message{Message: "Game server files are up to date"})
}
