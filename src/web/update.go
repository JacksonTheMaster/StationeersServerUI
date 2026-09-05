package web

import (
	"encoding/json"
	"net/http"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

type UpdateTriggerRequest struct {
	AllowUpdate bool `json:"allowUpdate"`
}

func CheckUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	update.UpdateInfo.RLock()
	if update.UpdateInfo.Available {
		version := update.UpdateInfo.Version
		update.UpdateInfo.RUnlock()

		json.NewEncoder(w).Encode(map[string]string{
			"status":          "success",
			"updateAvailable": "true",
			"version":         version,
			"message":         "Update to " + version + " available",
		})
		return
	}
	update.UpdateInfo.RUnlock()
	// no update available
	json.NewEncoder(w).Encode(map[string]string{
		"status":          "success",
		"updateAvailable": "false",
		"message":         "No update available",
	})
}

func TriggerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"status":"error","message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	applyUpdate := false
	if r.Body != http.NoBody {
		var req UpdateTriggerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			applyUpdate = req.AllowUpdate
		}
	}

	// Prevent concurrent runs
	if !update.UpdateInfo.TryLock() {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "busy",
			"message": "An update operation is already in progress - please try again later",
		})
		return
	}

	// If we're applying the update, do it in background (may restart process)
	if applyUpdate {
		go func() {
			defer update.UpdateInfo.Unlock()

			err, newVersion := update.Update(true)
			if err != nil {
				logger.Install.Error("Manual update (apply) failed: " + err.Error())
			} else {
				logger.Install.Info("Manual update (apply) completed successfully")
				// Note: if it replaced the binary and restarts, we won't get here reliably
			}

			// Still try to update state if we're still running
			if newVersion != "" {
				update.UpdateInfo.Available = true
				update.UpdateInfo.Version = newVersion
			} else {
				update.UpdateInfo.Available = false
				update.UpdateInfo.Version = ""
			}
		}()

		json.NewEncoder(w).Encode(map[string]string{
			"status":          "running",
			"updateAvailable": "true",
			"message":         "Update started - applying in background. Check logs for progress.",
		})
		return
	}

	// --- Check-only mode: do synchronously ---
	err, newVersion := update.Update(false)

	if err == nil && newVersion != "" {
		update.UpdateInfo.Available = true
		update.UpdateInfo.Version = newVersion
	} else {
		update.UpdateInfo.Available = false
		update.UpdateInfo.Version = ""
	}

	update.UpdateInfo.Unlock()

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Update check failed: " + err.Error(),
		})
		return
	}

	if newVersion != "" {
		json.NewEncoder(w).Encode(map[string]string{
			"status":          "success",
			"updateAvailable": "true",
			"version":         newVersion,
			"message":         "Update to " + newVersion + " available",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"status":          "success",
			"updateAvailable": "false",
			"message":         "No update available - you are up to date",
		})
	}
}
