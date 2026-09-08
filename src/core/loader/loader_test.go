package loader

import (
	"testing"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func TestPlanConfigReloadIgnoresGameServerSettings(t *testing.T) {
	previous := &config.JsonConfig{ServerName: "Europa"}
	next := &config.JsonConfig{ServerName: "Mars"}

	if plan := planConfigReload(previous, next); plan != (configReloadPlan{}) {
		t.Fatalf("game server setting unexpectedly reloads a subsystem: %+v", plan)
	}
}

func TestPlanConfigReloadTargetsChangedSubsystems(t *testing.T) {
	oldFalse, newTrue := false, true
	oldKeep, newKeep := 5, 10
	previous := &config.JsonConfig{
		BackupKeepNewestCount:      &oldKeep,
		IsDiscordEnabled:           &oldFalse,
		AllowAutoGameServerUpdates: &oldFalse,
	}
	next := &config.JsonConfig{
		BackupKeepNewestCount:      &newKeep,
		IsDiscordEnabled:           &newTrue,
		AllowAutoGameServerUpdates: &oldFalse,
	}

	plan := planConfigReload(previous, next)
	if !plan.backupManager || !plan.discord {
		t.Fatalf("expected backup and Discord reloads: %+v", plan)
	}
	if plan.localizer || plan.appInfoPoller || plan.sscm || plan.slpAutoUpdates {
		t.Fatalf("unrelated subsystem was selected: %+v", plan)
	}
}
