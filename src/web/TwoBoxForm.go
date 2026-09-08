package web

import (
	"io/fs"
	"net/http"
	"text/template"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/security"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/localization"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

func ServeTwoBoxFormTemplate(w http.ResponseWriter, r *http.Request) {
	type Step struct {
		ID                 string
		Title              string
		HeaderTitle        string
		StepMessage        string
		PrimaryLabel       string
		SecondaryLabel     string
		SecondaryLabelType string
		SecondaryOptions   []struct {
			Display string // Text shown in dropdown
			Value   string // Value sent on submission
		}
		SubmitButtonText         string
		SkipButtonText           string
		PrimaryPlaceholderText   string
		SecondaryPlaceholderText string
		ConfigField              string // Field name to save in config
		NextStep                 string // ID of the next step
	}

	type TemplateData struct {
		IsFirstTimeSetup   bool
		Path               string
		Title              string
		HeaderTitle        string
		StepMessage        string
		PrimaryLabel       string
		SecondaryLabel     string
		SecondaryLabelType string
		SecondaryOptions   []struct {
			Display string
			Value   string
		}
		SubmitButtonText         string
		SkipButtonText           string
		Mode                     string
		ShowExtraButtons         bool
		FooterText               string
		FooterTextInfo           string
		FinalizeSubmitButtonText string
		Step                     string
		ConfigField              string
		NextStep                 string
		PrimaryPlaceholderText   string
		SecondaryPlaceholderText string
		SetupSecretRequired      bool
		Steps                    []Step
	}

	twoboxformAssetsFS, err := fs.Sub(config.GetV1UIFS(), "SSUI/onboard_bundled/twoboxform")
	if err != nil {
		logger.Web.Error("Failed to get bundled FS")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(twoboxformAssetsFS, "twoboxform.html")
	if err != nil {
		logger.Web.Error("Failed to parse 2BoxForm template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	path := r.URL.Path
	stepID := r.URL.Query().Get("step")
	if stepID == "" && path == "/setup" {
		stepID = "welcome" // Start with welcome page for first-time setup
	}

	var gameBranchOptions = []struct{ Display, Value string }{
		{Display: "Stable branch (default)", Value: "public"},
		{Display: "Beta branch", Value: "beta"},
		{Display: "Previous", Value: "previous"},
		// Legacy terrain branches (preterrain, prerocket, prephasechange, multiplayersafe) removed per #172.
		// Using them now triggers hard error at startup.
	}

	var worldOptions = []struct{ Display, Value string }{
		{Display: "Moon", Value: "Moon"},
		{Display: "Vulcan", Value: "Vulcan"},
		{Display: "Venus", Value: "Venus"},
		{Display: "Mars", Value: "Mars"},
		{Display: "Europa", Value: "Europa"},
		{Display: "Mimas", Value: "Mimas"},
	}

	if config.GetIsNewTerrainAndSaveSystem() {
		catalog := loadWorldGenerationCatalog()
		worldOptions = make([]struct{ Display, Value string }, 0, len(catalog.Worlds))
		for _, world := range catalog.Worlds {
			worldOptions = append(worldOptions, struct{ Display, Value string }{
				Display: world.Label,
				Value:   world.ID,
			})
		}
	}

	// Define all steps in a map for easy access and modification
	steps := map[string]Step{
		"welcome": {
			ID:                 "welcome",
			Title:              localization.GetString("UIText_Welcome_Title"),
			HeaderTitle:        localization.GetString("UIText_Welcome_HeaderTitle"),
			StepMessage:        "",
			PrimaryLabel:       "",
			SecondaryLabel:     "",
			SecondaryLabelType: "hidden",
			SubmitButtonText:   localization.GetString("UIText_Welcome_SubmitButton"),
			SkipButtonText:     localization.GetString("UIText_Welcome_SkipButton"),
			NextStep:           "pls_read",
		},
		"pls_read": {
			ID:                 "pls_read",
			Title:              localization.GetString("UIText_PlsRead_Title"),
			HeaderTitle:        localization.GetString("UIText_PlsRead_HeaderTitle"),
			StepMessage:        localization.GetString("UIText_PlsRead_StepMessage"),
			PrimaryLabel:       "",
			SecondaryLabel:     "",
			SecondaryLabelType: "hidden",
			SubmitButtonText:   localization.GetString("UIText_PlsRead_SubmitButton"),
			SkipButtonText:     localization.GetString("UIText_PlsRead_SkipButton"),
			NextStep:           "game_branch",
		},
		"game_branch": {
			ID:                       "game_branch",
			Title:                    localization.GetString("UIText_GameBranch_Title"),
			HeaderTitle:              localization.GetString("UIText_GameBranch_HeaderTitle"),
			StepMessage:              localization.GetString("UIText_GameBranch_StepMessage"),
			SecondaryPlaceholderText: localization.GetString("UIText_GameBranch_SecondaryPlaceholder"),
			SecondaryLabel:           localization.GetString("UIText_GameBranch_SecondaryLabel"),
			SecondaryLabelType:       "dropdown",
			SecondaryOptions:         gameBranchOptions,
			SubmitButtonText:         localization.GetString("UIText_GameBranch_SubmitButton"),
			SkipButtonText:           localization.GetString("UIText_GameBranch_SkipButton"),
			ConfigField:              "gameBranch",
			NextStep:                 "server_name",
		},
		"server_name": {
			ID:                     "server_name",
			Title:                  localization.GetString("UIText_ServerName_Title"),
			HeaderTitle:            localization.GetString("UIText_ServerName_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_ServerName_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_ServerName_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_ServerName_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_ServerName_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_ServerName_SkipButton"),
			ConfigField:            "ServerName",
			NextStep:               "save_name",
		},
		"save_name": {
			ID:                     "save_name",
			Title:                  localization.GetString("UIText_SaveName_Title"),
			HeaderTitle:            localization.GetString("UIText_SaveName_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_SaveName_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_SaveName_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_SaveName_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_SaveName_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_SaveName_SkipButton"),
			ConfigField:            "SaveName",
			NextStep:               "world_id",
		},
		"world_id": {
			ID:                       "world_id",
			Title:                    localization.GetString("UIText_WorldID_Title"),
			HeaderTitle:              localization.GetString("UIText_WorldID_HeaderTitle"),
			StepMessage:              localization.GetString("UIText_WorldID_StepMessage"),
			SecondaryLabel:           localization.GetString("UIText_WorldID_SecondaryLabel"),
			SecondaryLabelType:       "dropdown",
			SecondaryPlaceholderText: localization.GetString("UIText_WorldID_SecondaryPlaceholder"),
			SecondaryOptions:         worldOptions,
			SubmitButtonText:         localization.GetString("UIText_WorldID_SubmitButton"),
			SkipButtonText:           localization.GetString("UIText_WorldID_SkipButton"),
			ConfigField:              "WorldID",
			NextStep:                 "max_players",
		},
		"max_players": {
			ID:                     "max_players",
			Title:                  localization.GetString("UIText_MaxPlayers_Title"),
			HeaderTitle:            localization.GetString("UIText_MaxPlayers_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_MaxPlayers_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_MaxPlayers_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_MaxPlayers_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_MaxPlayers_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_MaxPlayers_SkipButton"),
			ConfigField:            "ServerMaxPlayers",
			NextStep:               "server_password",
		},
		"server_password": {
			ID:                     "server_password",
			Title:                  localization.GetString("UIText_ServerPassword_Title"),
			HeaderTitle:            localization.GetString("UIText_ServerPassword_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_ServerPassword_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_ServerPassword_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_ServerPassword_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_ServerPassword_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_ServerPassword_SkipButton"),
			ConfigField:            "ServerPassword",
			NextStep:               "network_config_choice",
		},
		"network_config_choice": {
			ID:                     "network_config_choice",
			Title:                  localization.GetString("UIText_NetworkConfigChoice_Title"),
			HeaderTitle:            localization.GetString("UIText_NetworkConfigChoice_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_NetworkConfigChoice_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_NetworkConfigChoice_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_NetworkConfigChoice_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_NetworkConfigChoice_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_NetworkConfigChoice_SkipButton"),
			ConfigField:            "",          // No config field, just for branching
			NextStep:               "game_port", // Default next step if they choose to configure
			// The actual next step will be determined by JS based on the answer
		},

		"game_port": {
			ID:                     "game_port",
			Title:                  localization.GetString("UIText_GamePort_Title"),
			HeaderTitle:            localization.GetString("UIText_GamePort_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_GamePort_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_GamePort_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_GamePort_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_GamePort_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_GamePort_SkipButton"),
			ConfigField:            "GamePort",
			NextStep:               "update_port",
		},
		"update_port": {
			ID:                     "update_port",
			Title:                  localization.GetString("UIText_UpdatePort_Title"),
			HeaderTitle:            localization.GetString("UIText_UpdatePort_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_UpdatePort_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_UpdatePort_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_UpdatePort_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_UpdatePort_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_UpdatePort_SkipButton"),
			ConfigField:            "UpdatePort",
			NextStep:               "upnp_enabled",
		},
		"upnp_enabled": {
			ID:                     "upnp_enabled",
			Title:                  localization.GetString("UIText_UPnPEnabled_Title"),
			HeaderTitle:            localization.GetString("UIText_UPnPEnabled_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_UPnPEnabled_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_UPnPEnabled_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_UPnPEnabled_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_UPnPEnabled_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_UPnPEnabled_SkipButton"),
			ConfigField:            "UPNPEnabled",
			NextStep:               "local_ip_address",
		},
		"local_ip_address": {
			ID:                     "local_ip_address",
			Title:                  localization.GetString("UIText_LocalIPAddress_Title"),
			HeaderTitle:            localization.GetString("UIText_LocalIPAddress_HeaderTitle"),
			StepMessage:            localization.GetString("UIText_LocalIPAddress_StepMessage"),
			PrimaryPlaceholderText: localization.GetString("UIText_LocalIPAddress_PrimaryPlaceholder"),
			PrimaryLabel:           localization.GetString("UIText_LocalIPAddress_PrimaryLabel"),
			SecondaryLabel:         "",
			SecondaryLabelType:     "hidden",
			SubmitButtonText:       localization.GetString("UIText_LocalIPAddress_SubmitButton"),
			SkipButtonText:         localization.GetString("UIText_LocalIPAddress_SkipButton"),
			ConfigField:            "LocalIpAddress",
			NextStep:               "admin_account",
		},
		"admin_account": {
			ID:                       "admin_account",
			Title:                    localization.GetString("UIText_AdminAccount_Title"),
			HeaderTitle:              localization.GetString("UIText_AdminAccount_HeaderTitle"),
			StepMessage:              localization.GetString("UIText_AdminAccount_StepMessage"),
			PrimaryPlaceholderText:   localization.GetString("UIText_AdminAccount_PrimaryPlaceholder"),
			PrimaryLabel:             localization.GetString("UIText_AdminAccount_PrimaryLabel"),
			SecondaryLabel:           localization.GetString("UIText_AdminAccount_SecondaryLabel"),
			SecondaryPlaceholderText: localization.GetString("UIText_AdminAccount_SecondaryPlaceholder"),
			SecondaryLabelType:       "password",
			SubmitButtonText:         localization.GetString("UIText_AdminAccount_SubmitButton"),
			SkipButtonText:           localization.GetString("UIText_AdminAccount_SkipButton"),
			ConfigField:              "",
			NextStep:                 "finalize",
		},
		"finalize": {
			ID:               "finalize",
			Title:            localization.GetString("UIText_Finalize_Title"),
			HeaderTitle:      localization.GetString("UIText_Finalize_HeaderTitle"),
			StepMessage:      localization.GetString("UIText_Finalize_StepMessage"),
			PrimaryLabel:     "",
			SecondaryLabel:   "",
			SubmitButtonText: localization.GetString("UIText_Finalize_SubmitButton"),
			SkipButtonText:   localization.GetString("UIText_Finalize_SkipButton"),
			NextStep:         "welcome", // Return to first step if "Return to Setup" is clicked
		},
	}

	data := TemplateData{
		IsFirstTimeSetup:         config.GetIsFirstTimeSetup(),
		Path:                     path,
		Step:                     stepID,
		FooterText:               localization.GetString("UIText_FooterText"),
		FooterTextInfo:           localization.GetString("UIText_FooterTextInfo"),
		FinalizeSubmitButtonText: localization.GetString("UIText_FinalizeSubmitButtonText"),
	}

	switch {

	case path == "/login" && security.SetupRequired():
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return

	case path == "/setup":
		data.Mode = "setup"
		data.ShowExtraButtons = stepID != "welcome" && stepID != "admin_account" && stepID != "finalize"

		// Find the current step in our map
		if step, exists := steps[stepID]; exists {
			data.Title = step.Title
			data.HeaderTitle = step.HeaderTitle
			data.StepMessage = step.StepMessage
			data.PrimaryLabel = step.PrimaryLabel
			data.SecondaryLabel = step.SecondaryLabel
			data.SecondaryLabelType = step.SecondaryLabelType
			data.SubmitButtonText = step.SubmitButtonText
			data.SkipButtonText = step.SkipButtonText
			data.ConfigField = step.ConfigField
			data.NextStep = step.NextStep
			data.PrimaryPlaceholderText = step.PrimaryPlaceholderText
			data.SecondaryPlaceholderText = step.SecondaryPlaceholderText
			data.SecondaryOptions = step.SecondaryOptions
			data.SetupSecretRequired = stepID == "admin_account"
		} else {
			// Default to welcome page if step is invalid
			welcomeStep := steps["welcome"]
			data.Title = welcomeStep.Title
			data.HeaderTitle = welcomeStep.HeaderTitle
			data.StepMessage = welcomeStep.StepMessage
			data.PrimaryLabel = welcomeStep.PrimaryLabel
			data.SecondaryLabel = welcomeStep.SecondaryLabel
			data.SecondaryLabelType = welcomeStep.SecondaryLabelType
			data.SubmitButtonText = welcomeStep.SubmitButtonText
			data.SkipButtonText = welcomeStep.SkipButtonText
			data.ConfigField = welcomeStep.ConfigField
			data.NextStep = welcomeStep.NextStep
			data.Step = "welcome"
		}
		stepOrder := []string{
			"welcome", "pls_read", "game_branch", "server_name", "save_name", "world_id", "max_players",
			"server_password", "discord_enabled", "discord_token", "control_panel_channel", "save_channel",
			"log_channel", "connection_list_channel", "status_channel", "control_channel",
			"network_config_choice", "game_port", "update_port", "upnp_enabled",
			"local_ip_address", "admin_account", "finalize",
		}
		var stepSlice []Step
		for _, id := range stepOrder {
			if step, exists := steps[id]; exists {
				stepSlice = append(stepSlice, step)
			}
		}
		data.Steps = stepSlice

	case path == "/changeuser":
		data.Title = localization.GetString("UIText_ChangeUser_Title")
		data.HeaderTitle = localization.GetString("UIText_ChangeUser_HeaderTitle")
		data.PrimaryLabel = localization.GetString("UIText_ChangeUser_PrimaryLabel")
		data.SecondaryLabel = localization.GetString("UIText_ChangeUser_SecondaryLabel")
		data.SecondaryPlaceholderText = localization.GetString("UIText_ChangeUser_SecondaryPlaceholder")
		data.SecondaryLabelType = "password"
		data.SubmitButtonText = localization.GetString("UIText_ChangeUser_SubmitButton")
		data.Mode = "changeuser"
		data.ShowExtraButtons = false

	default:
		data.Title = localization.GetString("UIText_Login_Title")
		data.HeaderTitle = localization.GetString("UIText_Login_HeaderTitle") + config.GetSSUIIdentifier()
		data.PrimaryLabel = localization.GetString("UIText_Login_PrimaryLabel")
		data.SecondaryLabel = localization.GetString("UIText_Login_SecondaryLabel")
		data.PrimaryPlaceholderText = localization.GetString("UIText_Login_PrimaryPlaceholder")
		data.SecondaryPlaceholderText = localization.GetString("UIText_Login_SecondaryPlaceholder")
		data.SecondaryLabelType = "password"
		data.SubmitButtonText = localization.GetString("UIText_Login_SubmitButton")
		data.Mode = "login"
		data.Step = ""
		data.ShowExtraButtons = false
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		logger.Web.Error("Failed to execute 2BoxForm template: %v" + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
