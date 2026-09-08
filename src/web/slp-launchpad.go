package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/modding"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/steamcmd"
)

func InstallSLPHandler(w http.ResponseWriter, r *http.Request) {
	version, err := modding.InstallSLP()
	if err != nil {
		api.WriteError(w, http.StatusBadGateway, "slp_install_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, map[string]string{"version": version})
}

func UninstallSLPHandler(w http.ResponseWriter, r *http.Request) {

	version, err := modding.UninstallSLP()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "slp_uninstall_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, map[string]string{"version": version})
}

func ReinstallSLPHandler(w http.ResponseWriter, r *http.Request) {
	version, err := modding.ReinstallSLP()
	if err != nil {
		api.WriteError(w, http.StatusBadGateway, "slp_reinstall_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, map[string]string{"version": version})
}

func UploadModPackageHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)
	if err := modding.ProcessModPackageUpload(r.Body); err != nil {
		api.WriteError(w, http.StatusUnprocessableEntity, "mod_package_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, api.Message{Message: "Mod package imported"})
}

func GetInstalledModDetailsHandler(w http.ResponseWriter, r *http.Request) {
	mods := modding.GetModList()
	type modResponse struct {
		Name           string            `json:"name"`
		Author         string            `json:"author"`
		Version        string            `json:"version"`
		Description    string            `json:"description"`
		WorkshopHandle string            `json:"workshopId,omitempty"`
		Images         map[string]string `json:"images,omitempty"`
	}
	items := make([]modResponse, 0, len(mods))
	for _, mod := range mods {
		items = append(items, modResponse{
			Name: mod.Name, Author: mod.Author, Version: mod.Version, Description: mod.Description,
			WorkshopHandle: mod.WorkshopHandle, Images: mod.Images,
		})
	}
	api.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func UpdateWorkshopModsHandler(w http.ResponseWriter, r *http.Request) {
	logs, err := steamcmd.UpdateWorkshopItems()
	if err != nil {
		api.WriteErrorDetails(w, http.StatusBadGateway, "workshop_update_failed", err.Error(), map[string]any{"logs": logs})
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"logs": logs})
}

func UpdateSingleWorkshopModHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkshopIDs []string `json:"workshopIds"`
	}
	if err := api.DecodeJSONLimit(w, r, &req, 64<<10); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	workshopHandles, err := parseWorkshopHandles(req.WorkshopIDs)
	if err != nil {
		api.WriteError(w, http.StatusUnprocessableEntity, "invalid_workshop_id", err.Error())
		return
	}

	logs, err := steamcmd.DownloadWorkshopItems(workshopHandles)
	if err != nil {
		api.WriteErrorDetails(w, http.StatusBadGateway, "workshop_download_failed", err.Error(), map[string]any{"logs": logs})
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"workshopIds": workshopHandles, "logs": logs})
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
