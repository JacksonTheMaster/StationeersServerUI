package configchanger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func TestApplySettingsPatchOnlyChangesPresentFields(t *testing.T) {
	serverName := "Europa"
	autoSave := false
	keepNewest := 25
	cfg := config.JsonConfig{
		ServerName:            "Mars",
		ServerPassword:        "keep-me",
		AutoSave:              boolPointer(true),
		BackupKeepNewestCount: intPointer(5),
	}
	patch := api.SettingsPatch{
		ServerName:            &serverName,
		AutoSave:              &autoSave,
		BackupKeepNewestCount: &keepNewest,
	}

	applySettingsPatch(&cfg, &patch)

	if cfg.ServerName != "Europa" || cfg.ServerPassword != "keep-me" {
		t.Fatalf("unexpected string settings after patch: %+v", cfg)
	}
	if cfg.AutoSave == nil || *cfg.AutoSave || cfg.BackupKeepNewestCount == nil || *cfg.BackupKeepNewestCount != 25 {
		t.Fatalf("pointer settings were not applied: %+v", cfg)
	}
}

func TestSettingsFromConfigCopiesValues(t *testing.T) {
	cfg := config.JsonConfig{ServerName: "Europa", AutoSave: boolPointer(true)}
	settings := settingsFromConfig(&cfg)
	*cfg.AutoSave = false

	if settings.ServerName == nil || *settings.ServerName != "Europa" {
		t.Fatal("server name missing from settings")
	}
	if settings.AutoSave == nil || !*settings.AutoSave {
		t.Fatal("settings shared a bool pointer with config")
	}
}

func TestPatchSettingsRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/api/v3/settings", bytes.NewBufferString(`{"JwtKey":"nope"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	PatchSettings(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestGetSettingsKeepsInternalSecretsWriteOnly(t *testing.T) {
	config.ConfigMu.Lock()
	oldDiscordToken := config.DiscordToken
	oldServerPassword := config.ServerPassword
	config.DiscordToken = "discord-secret"
	config.ServerPassword = "game-secret"
	config.ConfigMu.Unlock()
	t.Cleanup(func() {
		config.ConfigMu.Lock()
		config.DiscordToken = oldDiscordToken
		config.ServerPassword = oldServerPassword
		config.ConfigMu.Unlock()
	})

	response := httptest.NewRecorder()
	GetSettings(response, httptest.NewRequest(http.MethodGet, "/api/v3/settings", nil))
	body := response.Body.String()

	if strings.Contains(body, "discord-secret") {
		t.Fatal("settings response exposed the Discord token")
	}
	if !strings.Contains(body, `"discordTokenConfigured":true`) || !strings.Contains(body, "game-secret") {
		t.Fatalf("settings response did not describe configured secrets: %s", body)
	}
}

func boolPointer(value bool) *bool { return &value }
func intPointer(value int) *int    { return &value }
