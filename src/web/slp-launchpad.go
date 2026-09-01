package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/modding"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"
)

func InstallSLPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := modding.InstallSLP(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func UninstallSLPHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	if _, err := modding.UninstallSLP(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func ReinstallSLPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	version, err := modding.ReinstallSLP()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"version": version,
		"message": "SLP reinstalled successfully",
	})
}

func UploadModPackageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := modding.ProcessModPackageUpload(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mod package uploaded and extracted successfully",
	})
}

func GetInstalledModDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mods := modding.GetModList()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"mods":    mods,
	})
}

func UpdateWorkshopModsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logs, err := steamcmd.UpdateWorkshopItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"logs":    logs,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Workshop mods updated successfully",
		"logs":    logs,
	})
}

func UpdateSingleWorkshopModHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var req struct {
		WorkshopHandle  string   `json:"workshopHandle"`
		WorkshopHandles []string `json:"workshopHandles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	inputs := append([]string{}, req.WorkshopHandles...)
	if req.WorkshopHandle != "" {
		inputs = append(inputs, req.WorkshopHandle)
	}

	workshopHandles, err := parseWorkshopHandles(inputs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	logs, err := steamcmd.DownloadWorkshopItems(workshopHandles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"logs":    logs,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Workshop mods downloaded successfully",
		"logs":    logs,
	})
}

func parseWorkshopHandles(inputs []string) ([]string, error) {
	seen := make(map[string]bool)
	var handles []string

	for _, input := range inputs {
		for _, value := range strings.FieldsFunc(input, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
		}) {
			candidate := strings.TrimSpace(value)
			if parsed, err := url.Parse(candidate); err == nil && parsed.Scheme != "" {
				candidate = parsed.Query().Get("id")
			}

			workshopID, err := strconv.ParseUint(candidate, 10, 64)
			if err != nil || workshopID == 0 {
				return nil, fmt.Errorf("invalid Steam Workshop ID or URL %q", value)
			}

			handle := strconv.FormatUint(workshopID, 10)
			if !seen[handle] {
				seen[handle] = true
				handles = append(handles, handle)
			}
		}
	}

	if len(handles) == 0 {
		return nil, fmt.Errorf("at least one Steam Workshop ID or URL is required")
	}

	return handles, nil
}
