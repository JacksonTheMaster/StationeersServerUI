package localization

import (
	"encoding/json"
	"io/fs"
	"strings"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// translations stores language code to key-value pairs
var translations = make(map[string]map[string]string)
var mu sync.RWMutex
var currentLanguage string

const fallbackLanguage = "en-us"

// reloads all translations and resets to the current language
func ReloadLocalizer() {
	logger.Localization.Info("Reloading localization data")
	currentLanguage = strings.ToLower(config.GetLanguageSetting())
	loadTranslations()
}

// loadTranslations reads all JSON files from virtFS
func loadTranslations() {
	mu.Lock()
	defer mu.Unlock()

	// Clear existing translations
	translations = make(map[string]map[string]string)

	virtFS, err := fs.Sub(config.V1UIFS, "SSUI/onboard_bundled/localization")
	if err != nil {
		logger.Localization.Error("Failed to access virtual filesystem: " + err.Error())
		return
	}

	entries, err := fs.ReadDir(virtFS, ".")
	if err != nil {
		logger.Localization.Error("Failed to read virtual filesystem directory: " + err.Error())
		return
	}

	// Process each JSON file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue // Skip directories and non-JSON files
		}

		langCode := strings.ToLower(strings.TrimSuffix(entry.Name(), ".json"))
		file, err := virtFS.Open(entry.Name())
		if err != nil {
			logger.Localization.Error("Failed to open file " + entry.Name() + ": " + err.Error())
			continue
		}

		data, err := fs.ReadFile(virtFS, entry.Name())
		if err != nil {
			logger.Localization.Error("Failed to read file " + entry.Name() + ": " + err.Error())
			file.Close()
			continue
		}
		file.Close()

		var rawData map[string]any
		if err := json.Unmarshal(data, &rawData); err != nil {
			logger.Localization.Error("Failed to parse JSON in " + entry.Name() + ": " + err.Error())
			continue
		}

		langMap := make(map[string]string)
		flattenJSON(rawData, langMap)

		translations[langCode] = langMap
		logger.Localization.Debug("Loaded translations for language: " + langCode)
	}

	if _, exists := translations[fallbackLanguage]; !exists {
		logger.Localization.Warn("Fallback language en-us not found in localization files")
	}
}

func flattenJSON(data map[string]interface{}, output map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			flattenJSON(v, output)
		case string:
			// stor str vals directly
			output[key] = v
		default:
			logger.Localization.Warn("Ignoring non-string value for key: " + key)
		}
	}
}

// GetString returns the localized string for the given key
func GetString(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	//logger.Localization.Warn("Looking up key: " + key + " in language: " + currentLanguage)

	// Try current language
	if langMap, exists := translations[currentLanguage]; exists {
		if translation, exists := langMap[key]; exists {
			return translation
		}
	}

	// Fall back to en-us
	if langMap, exists := translations[fallbackLanguage]; exists {
		if translation, exists := langMap[key]; exists {
			logger.Localization.Debug("Falling back to en-us for key: " + key)
			return translation
		}
	}

	// Return key as final fallback
	logger.Localization.Debug("Translation not found for key: " + key)
	return key
}
