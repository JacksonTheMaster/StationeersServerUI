package configchanger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func SaveConfigForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusInternalServerError)
		return
	}

	// Load existing configuration
	existingConfig, err := config.LoadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading existing configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// Use reflection to update fields present in the form
	v := reflect.ValueOf(existingConfig).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Tag.Get("json")
		formValues, exists := r.Form[fieldName] // Check if the field exists in the form data

		if !exists {
			continue // Skip fields not submitted in the form
		}

		// If the field exists, use the first value (even if it's empty)
		formValue := ""
		if len(formValues) > 0 {
			formValue = formValues[0]
		}

		field := v.Field(i)
		fieldType := field.Type()

		switch fieldType.Kind() {
		case reflect.String:
			field.SetString(formValue) // Set the value, even if it's empty to allow clearing the field
		case reflect.Pointer:
			if fieldType.Elem().Kind() == reflect.Bool {
				newBool := formValue == "true"       // Convert form value to bool
				field.Set(reflect.ValueOf(&newBool)) // Set the pointer to the new bool
			}
		}
	}

	// Save the updated config
	if err := SaveConfig(existingConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/config", http.StatusSeeOther)
}

func SaveConfigRestful(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Parse the JSON data into a map
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Load existing configuration
	existingConfig, err := config.LoadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading existing configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// Use reflection to update fields present in the request data
	v := reflect.ValueOf(existingConfig).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Tag.Get("json")
		value, exists := requestData[fieldName] // Check if the field exists in the request data

		if !exists {
			continue // Skip fields not submitted in the request
		}

		field := v.Field(i)
		fieldType := field.Type()

		switch fieldType.Kind() {
		case reflect.String:
			if strValue, ok := value.(string); ok {
				field.SetString(strValue)
			}
		case reflect.Pointer:
			switch fieldType.Elem().Kind() {
			case reflect.Bool:
				if boolValue, ok := value.(bool); ok {
					field.Set(reflect.ValueOf(&boolValue)) // Set the pointer to the new bool
				}
			case reflect.Int:
				intValue, ok := jsonInt(value)
				if !ok {
					http.Error(w, fmt.Sprintf("%s must be an integer", fieldName), http.StatusBadRequest)
					return
				}
				newValue := reflect.New(fieldType.Elem())
				newValue.Elem().SetInt(intValue)
				field.Set(newValue)
			}
		case reflect.Int:
			switch v := value.(type) {
			case float64: // JSON numbers are parsed as float64 by default
				field.SetInt(int64(v))
			case int:
				field.SetInt(int64(v))
			case int64:
				field.SetInt(v)
			case string:
				if intValue, err := strconv.ParseInt(v, 10, 64); err == nil {
					field.SetInt(intValue)
				}
			}
		case reflect.Map:
			// Handle map fields like Users (map[string]string)
			if mapValue, ok := value.(map[string]interface{}); ok {
				if field.Type() == reflect.TypeOf(map[string]string{}) {
					newMap := make(map[string]string)
					for k, v := range mapValue {
						if strVal, ok := v.(string); ok {
							newMap[k] = strVal
						}
					}
					field.Set(reflect.ValueOf(newMap))
				}
			}
		}
	}

	if err := validateBackupSettings(existingConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save the updated config
	if err := SaveConfig(existingConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Configuration updated successfully"})
}

func jsonInt(value interface{}) (int64, bool) {
	switch value := value.(type) {
	case float64:
		return int64(value), value == float64(int64(value))
	case int:
		return int64(value), true
	case int64:
		return value, true
	case string:
		intValue, err := strconv.ParseInt(value, 10, 64)
		return intValue, err == nil
	default:
		return 0, false
	}
}

func validateBackupSettings(cfg *config.JsonConfig) error {
	settings := []struct {
		name  string
		value *int
	}{
		{name: "backupKeepNewestCount", value: cfg.BackupKeepNewestCount},
		{name: "backupDailyRetentionDays", value: cfg.BackupDailyRetentionDays},
		{name: "backupWeeklyRetentionWeeks", value: cfg.BackupWeeklyRetentionWeeks},
		{name: "backupMonthlyRetentionMonths", value: cfg.BackupMonthlyRetentionMonths},
	}

	for _, setting := range settings {
		if setting.value != nil && *setting.value < 0 {
			return fmt.Errorf("%s cannot be negative", setting.name)
		}
	}

	if cfg.BackupCleanupIntervalHours != nil && *cfg.BackupCleanupIntervalHours <= 0 {
		return fmt.Errorf("backupCleanupIntervalHours must be greater than zero")
	}

	return nil
}
