package configchanger

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func GetSettings(w http.ResponseWriter, _ *http.Request) {
	cfg := config.Snapshot()
	settings := api.Settings{
		SettingsPatch:          settingsFromConfig(&cfg),
		DiscordTokenConfigured: cfg.DiscordToken != "",
	}
	// Bot credentials are write-only. Game server secrets intentionally remain
	// visible to users with settings.view.
	settings.DiscordToken = nil
	api.WriteData(w, http.StatusOK, settings)
}

func PatchSettings(w http.ResponseWriter, r *http.Request) {
	var patch api.SettingsPatch
	if err := api.DecodeJSON(w, r, &patch); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	restartRequired := false
	err := UpdateConfig(func(cfg *config.JsonConfig) error {
		if patch.SSUIWebPort != nil && *patch.SSUIWebPort != cfg.SSUIWebPort {
			restartRequired = true
		}
		applySettingsPatch(cfg, &patch)
		if patch.WorldID != nil || patch.Difficulty != nil || patch.StartCondition != nil || patch.StartLocation != nil {
			if err := validateWorldGenerationSettings(cfg); err != nil {
				return &ValidationError{Err: err}
			}
		}
		return nil
	})
	if err != nil {
		var validationError *ValidationError
		if errors.As(err, &validationError) {
			api.WriteError(w, http.StatusBadRequest, "invalid_settings", err.Error())
		} else {
			api.WriteError(w, http.StatusInternalServerError, "settings_update_failed", "Failed to update settings")
		}
		return
	}

	result := api.SettingsUpdate{Message: "Configuration updated"}
	if restartRequired {
		result.RestartRequired = []string{"ssuiWebPort"}
	}
	api.WriteData(w, http.StatusOK, result)
}

func PatchSetupSettings(w http.ResponseWriter, r *http.Request) {
	var patch api.SetupSettingsPatch
	if err := api.DecodeJSON(w, r, &patch); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := UpdateConfig(func(cfg *config.JsonConfig) error {
		applyPatch(cfg, &patch)
		return nil
	}); err != nil {
		var validationError *ValidationError
		if errors.As(err, &validationError) {
			api.WriteError(w, http.StatusBadRequest, "invalid_settings", err.Error())
		} else {
			api.WriteError(w, http.StatusInternalServerError, "settings_update_failed", "Failed to update settings")
		}
		return
	}
	api.WriteData(w, http.StatusOK, api.Message{Message: "Configuration updated"})
}

func settingsFromConfig(cfg *config.JsonConfig) api.SettingsPatch {
	var settings api.SettingsPatch
	destination := reflect.ValueOf(&settings).Elem()
	source := reflect.ValueOf(cfg).Elem()
	for i := 0; i < destination.NumField(); i++ {
		output := destination.Field(i)
		input := source.FieldByName(destination.Type().Field(i).Name)
		if !input.IsValid() {
			continue
		}
		if input.Kind() == reflect.Pointer {
			if !input.IsNil() {
				copy := reflect.New(input.Type().Elem())
				copy.Elem().Set(input.Elem())
				output.Set(copy)
			}
			continue
		}
		copy := reflect.New(input.Type())
		copy.Elem().Set(input)
		output.Set(copy)
	}
	return settings
}

func applySettingsPatch(cfg *config.JsonConfig, patch *api.SettingsPatch) {
	applyPatch(cfg, patch)
}

func applyPatch(cfg *config.JsonConfig, patch any) {
	destination := reflect.ValueOf(cfg).Elem()
	source := reflect.ValueOf(patch).Elem()
	for i := 0; i < source.NumField(); i++ {
		input := source.Field(i)
		if input.IsNil() {
			continue
		}
		output := destination.FieldByName(source.Type().Field(i).Name)
		if output.Kind() == reflect.Pointer {
			copy := reflect.New(input.Elem().Type())
			copy.Elem().Set(input.Elem())
			output.Set(copy)
		} else {
			output.Set(input.Elem())
		}
	}
}
