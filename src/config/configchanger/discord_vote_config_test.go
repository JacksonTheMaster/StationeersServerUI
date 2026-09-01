package configchanger

import (
	"testing"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func validDiscordVoteConfig() *config.JsonConfig {
	return &config.JsonConfig{
		DiscordVoteDurationMinutes:        configInt(5),
		DiscordRestartVoteThreshold:       configInt(60),
		DiscordRestartVoteMinimum:         configInt(1),
		DiscordRestartVoteCooldownMinutes: configInt(30),
		DiscordRestoreVoteThreshold:       configInt(100),
		DiscordRestoreVoteMinimum:         configInt(2),
		DiscordRestoreVoteCooldownMinutes: configInt(60),
	}
}

func TestValidateDiscordVoteSettingsAcceptsDefaults(t *testing.T) {
	if err := validateDiscordVoteSettings(validDiscordVoteConfig()); err != nil {
		t.Fatalf("validateDiscordVoteSettings() returned unexpected error: %v", err)
	}
}

func TestValidateDiscordVoteSettingsRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*config.JsonConfig)
	}{
		{name: "zero duration", mutate: func(cfg *config.JsonConfig) { cfg.DiscordVoteDurationMinutes = configInt(0) }},
		{name: "zero threshold", mutate: func(cfg *config.JsonConfig) { cfg.DiscordRestartVoteThreshold = configInt(0) }},
		{name: "threshold over 100", mutate: func(cfg *config.JsonConfig) { cfg.DiscordRestoreVoteThreshold = configInt(101) }},
		{name: "zero minimum", mutate: func(cfg *config.JsonConfig) { cfg.DiscordRestoreVoteMinimum = configInt(0) }},
		{name: "negative cooldown", mutate: func(cfg *config.JsonConfig) { cfg.DiscordRestartVoteCooldownMinutes = configInt(-1) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validDiscordVoteConfig()
			test.mutate(cfg)
			if err := validateDiscordVoteSettings(cfg); err == nil {
				t.Fatal("validateDiscordVoteSettings() returned nil, want validation error")
			}
		})
	}
}
