package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// LoadConfig loads the configuration from an XML file
func loadConfig() (*Config, error) {
	configPath := "./UIMod/config.xml"
	xmlFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = xml.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %v", err)
	}

	return &config, nil
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	config, err := loadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	htmlFile, err := os.ReadFile("./UIMod/config.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading config.html: %v", err), http.StatusInternalServerError)
		return
	}

	htmlContent := string(htmlFile)

	// Split the settings string into a map for easier access
	settingsMap := make(map[string]string)
	settings := strings.Split(config.Server.Settings, " ")
	for i := 0; i < len(settings)-1; i += 2 {
		settingsMap[settings[i]] = settings[i+1]
	}

	// Replace placeholders with actual values
	replacements := map[string]string{
		"{{ExePath}}":          config.Server.ExePath,
		"{{StartLocalHost}}":   settingsMap["StartLocalHost"],
		"{{ServerVisible}}":    settingsMap["ServerVisible"],
		"{{GamePort}}":         settingsMap["GamePort"],
		"{{UpdatePort}}":       settingsMap["UpdatePort"],
		"{{AutoSave}}":         settingsMap["AutoSave"],
		"{{SaveInterval}}":     settingsMap["SaveInterval"],
		"{{LocalIpAddress}}":   settingsMap["LocalIpAddress"],
		"{{ServerPassword}}":   settingsMap["ServerPassword"],
		"{{AdminPassword}}":    settingsMap["AdminPassword"],
		"{{ServerMaxPlayers}}": settingsMap["ServerMaxPlayers"],
		"{{ServerName}}":       settingsMap["ServerName"],
		"{{AdditionalParams}}": getAdditionalParams(settings),
		"{{SaveFileName}}":     config.SaveFileName,
	}

	for placeholder, value := range replacements {
		htmlContent = strings.ReplaceAll(htmlContent, placeholder, value)
	}

	fmt.Fprint(w, htmlContent)
}

func getAdditionalParams(settings []string) string {
	// List of known parameters
	knownParams := map[string]bool{
		"StartLocalHost":   true,
		"ServerVisible":    true,
		"GamePort":         true,
		"UpdatePort":       true,
		"AutoSave":         true,
		"SaveInterval":     true,
		"LocalIpAddress":   true,
		"ServerPassword":   true,
		"AdminPassword":    true,
		"ServerMaxPlayers": true,
		"ServerName":       true,
	}

	var additionalParams []string
	for i := 0; i < len(settings)-1; i += 2 {
		if !knownParams[settings[i]] {
			additionalParams = append(additionalParams, settings[i]+" "+settings[i+1])
		}
	}

	return strings.Join(additionalParams, " ")
}

// SaveConfig saves the updated configuration to the XML file
func SaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		settings := []string{
			"StartLocalHost", r.FormValue("StartLocalHost"),
			"ServerVisible", r.FormValue("ServerVisible"),
			"GamePort", r.FormValue("GamePort"),
			"UpdatePort", r.FormValue("UpdatePort"),
			"AutoSave", r.FormValue("AutoSave"),
			"SaveInterval", r.FormValue("SaveInterval"),
			"LocalIpAddress", r.FormValue("LocalIpAddress"),
			"ServerPassword", r.FormValue("ServerPassword"),
			"AdminPassword", r.FormValue("AdminPassword"),
			"ServerMaxPlayers", r.FormValue("ServerMaxPlayers"),
			"ServerName", r.FormValue("ServerName"),
		}

		// Append additional parameters if any
		additionalParams := r.FormValue("AdditionalParams")
		if additionalParams != "" {
			settings = append(settings, strings.Split(additionalParams, " ")...)
		}

		settingsStr := strings.Join(settings, " ")

		config := Config{
			Server: struct {
				ExePath  string `xml:"exePath"`
				Settings string `xml:"settings"`
			}{
				ExePath: "./rocketstation_DedicatedServer.exe",
				// hardcoded ../rocketstation_DedicatedServer.exe for now, otherwise this is a security risk because an attacker could run a malicious exe and or command on the server
				// explaination: if the exepath is set to powershell.exe again, and matching parameters are set, the server would be able to run arbitrary code on the server.
				// this is very much an RCE vulnerability, and thus should be avoided at all costs.
				// with the hardcoded exepath, the server will not be able to run arbitrary code, but it will still be able to run the server with the given parameters.
				//more info: https://github.com/JacksonTheMaster/StationeersServerUI/issues/
				Settings: settingsStr,
			},
			SaveFileName: r.FormValue("saveFileName"),
		}

		configPath := "./UIMod/config.xml"
		file, err := os.Create(configPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating config file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		encoder := xml.NewEncoder(file)
		encoder.Indent("", "  ")
		if err := encoder.Encode(config); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding config: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
