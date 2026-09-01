package worldgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalogDiscoversPlayableWorldData(t *testing.T) {
	dataRoot := t.TempDir()
	writeTestFile(t, filepath.Join(dataRoot, "StreamingAssets", "Worlds", "Mars", "Mars.xml"), `
<GameData><WorldSettings>
  <World Id="Mars2" Priority="1">
    <StartCondition Id="DefaultStart" IsDefault="true"/>
    <StartCondition Id="Brutal"/>
    <StartLocation Id="MarsSpawnCanyonOverlook"/>
    <RandomStartLocation Id="MarsSpawnRoundRobin"><StartLocation Id="MarsSpawnCanyonOverlook"/></RandomStartLocation>
  </World>
</WorldSettings></GameData>`)
	writeTestFile(t, filepath.Join(dataRoot, "StreamingAssets", "Worlds", "Tutorial", "Tutorial.xml"), `
<GameData><WorldSettings><World Id="Tutorial"><StartLocation Id="TutorialSpawn"/></World></WorldSettings></GameData>`)
	writeTestFile(t, filepath.Join(dataRoot, "StreamingAssets", "Worlds", "Legacy", "Legacy.xml"), `
<GameData><WorldSettings><World Id="Legacy" Priority="2" Hidden="true" Deprecated="true"/></WorldSettings></GameData>`)
	writeTestFile(t, filepath.Join(dataRoot, "StreamingAssets", "Data", "difficultySettings.xml"), `
<GameData><DifficultySettings><DifficultySetting Id="Easy"/><DifficultySetting Id="Normal" Default="true"/></DifficultySettings></GameData>`)

	catalog, err := loadCatalog(dataRoot)
	if err != nil {
		t.Fatalf("loadCatalog() error = %v", err)
	}
	if len(catalog.Worlds) != 1 || catalog.Worlds[0].ID != "Mars2" {
		t.Fatalf("worlds = %#v, want only Mars2", catalog.Worlds)
	}
	world := catalog.Worlds[0]
	if world.DefaultConditionID != "DefaultStart" {
		t.Fatalf("default condition = %q, want DefaultStart", world.DefaultConditionID)
	}
	if len(world.StartLocations) != 2 || world.StartLocations[0].Value != "MarsSpawnRoundRobin" {
		t.Fatalf("start locations = %#v, want round robin followed by concrete locations", world.StartLocations)
	}
	if len(catalog.Difficulties) != 2 || catalog.Difficulties[1].Value != "Normal" {
		t.Fatalf("difficulties = %#v", catalog.Difficulties)
	}
}

func TestLoaderFallsBackWhenGameDataIsUnavailable(t *testing.T) {
	var loader Loader
	catalog, err := loader.Load(filepath.Join(t.TempDir(), "rocketstation_DedicatedServer.x86_64"))
	if err == nil {
		t.Fatal("Load() error = nil, want unavailable game-data error")
	}
	if catalog.Source != "fallback" || len(catalog.Worlds) == 0 {
		t.Fatalf("fallback catalog = %#v", catalog)
	}
}

func TestDataRootFromExecutable(t *testing.T) {
	tests := map[string]string{
		"rocketstation_DedicatedServer.x86_64":  "rocketstation_DedicatedServer_Data",
		"bin/rocketstation_DedicatedServer.exe": filepath.Join("bin", "rocketstation_DedicatedServer_Data"),
	}
	for input, want := range tests {
		if got := DataRootFromExecutable(input); got != want {
			t.Errorf("DataRootFromExecutable(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestHumanizeIdentifier(t *testing.T) {
	if got := humanizeIdentifier("VulcanSpawnDusterPlateauNorth"); got != "Duster Plateau North" {
		t.Fatalf("humanizeIdentifier() = %q", got)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
