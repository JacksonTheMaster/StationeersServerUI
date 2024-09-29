package api

import (
	"StationeersServerUI/src/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// LoadConfigJSON loads the configuration from the JSON file
func loadConfigJSON() (*config.Config, error) {
	configPath := "./UIMod/config.json"
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config.json: %v", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config.json: %v", err)
	}

	var config config.Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config.json: %v", err)
	}

	return &config, nil
}

func HandleConfigJSON(w http.ResponseWriter, r *http.Request) {
	config, err := loadConfigJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config.json: %v", err), http.StatusInternalServerError)
		return
	}

	htmlFile, err := os.ReadFile("./UIMod/furtherconfig.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading discord.html: %v", err), http.StatusInternalServerError)
		return
	}

	htmlContent := string(htmlFile)

	// Check the value of IsDiscordEnabled and set the appropriate option as selected
	isDiscordEnabledTrue := ""
	isDiscordEnabledFalse := ""

	if config.IsDiscordEnabled {
		isDiscordEnabledTrue = "selected"
	} else {
		isDiscordEnabledFalse = "selected"
	}

	// Replace placeholders in the HTML with actual config values, including the new errorChannelID
	replacements := map[string]string{
		"{{discordToken}}":            config.DiscordToken,
		"{{controlChannelID}}":        config.ControlChannelID,
		"{{statusChannelID}}":         config.StatusChannelID,
		"{{connectionListChannelID}}": config.ConnectionListChannelID,
		"{{logChannelID}}":            config.LogChannelID,
		"{{saveChannelID}}":           config.SaveChannelID,
		"{{controlPanelChannelID}}":   config.ControlPanelChannelID,
		"{{blackListFilePath}}":       config.BlackListFilePath,
		"{{errorChannelID}}":          config.ErrorChannelID, // New errorChannelID field
		"{{isDiscordEnabledTrue}}":    isDiscordEnabledTrue,
		"{{isDiscordEnabledFalse}}":   isDiscordEnabledFalse,
	}

	for placeholder, value := range replacements {
		htmlContent = strings.ReplaceAll(htmlContent, placeholder, value)
	}

	fmt.Fprint(w, htmlContent)
}

func SaveConfigJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		config := config.Config{
			DiscordToken:            r.FormValue("discordToken"),
			ControlChannelID:        r.FormValue("controlChannelID"),
			StatusChannelID:         r.FormValue("statusChannelID"),
			ConnectionListChannelID: r.FormValue("connectionListChannelID"),
			LogChannelID:            r.FormValue("logChannelID"),
			SaveChannelID:           r.FormValue("saveChannelID"),
			ControlPanelChannelID:   r.FormValue("controlPanelChannelID"),
			BlackListFilePath:       r.FormValue("blackListFilePath"),
			ErrorChannelID:          r.FormValue("errorChannelID"), // New errorChannelID field
			IsDiscordEnabled:        r.FormValue("isDiscordEnabled") == "true",
		}

		configPath := "./UIMod/config.json"
		file, err := os.Create(configPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating config.json: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(&config); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding config.json: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/furtherconfig", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
