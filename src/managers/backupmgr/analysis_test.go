package backupmgr

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeAnalysisSave(t testing.TB, metaXML, worldXML string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "analysis.save")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, contents := range map[string]string{
		worldMetaFilename: metaXML,
		worldFilename:     worldXML,
	} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func analysisFixture(t testing.TB) string {
	t.Helper()

	meta := fmt.Sprintf(`<?xml version="1.0"?>
<WorldMetaData Id="world-id">
  <GameVersion>0.2.test</GameVersion>
  <DateTime>%d</DateTime>
  <DaysPast>67</DaysPast>
  <WorldName>Europa</WorldName>
  <WorldFileName>TestWorld</WorldFileName>
  <NumberOfRooms>11</NumberOfRooms>
  <NumberOfPipeNetworks>22</NumberOfPipeNetworks>
  <NumberOfCableNetworks>33</NumberOfCableNetworks>
  <NumberOfThings>6</NumberOfThings>
  <NumberOfAtmospheres>2330</NumberOfAtmospheres>
</WorldMetaData>`, filetimeEpochOffset+123456789)

	world := `<?xml version="1.0"?>
<WorldData xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <Things>
    <ThingSaveData xsi:type="HumanSaveData">
      <PrefabName>Character</PrefabName>
      <States><State><StateName>Mode</StateName><State>99</State></State></States>
      <State>Alive</State>
    </ThingSaveData>
    <ThingSaveData xsi:type="HumanSaveData">
      <PrefabName>Character</PrefabName>
      <State>Unconscious</State>
    </ThingSaveData>
    <ThingSaveData xsi:type="HumanSaveData">
      <PrefabName>Character</PrefabName>
      <State>Disconnected</State>
    </ThingSaveData>
    <ThingSaveData xsi:type="AdvancedFurnaceSaveData">
      <PrefabName>StructureAdvancedFurnace</PrefabName>
      <HasSpawnedWreckage>false</HasSpawnedWreckage>
    </ThingSaveData>
    <ThingSaveData xsi:type="DeviceImportExportSaveData">
      <PrefabName>StructureArcFurnace</PrefabName>
      <HasSpawnedWreckage>true</HasSpawnedWreckage>
    </ThingSaveData>
    <ThingSaveData xsi:type="StackableSaveData">
      <PrefabName>ItemKitAdvancedFurnace</PrefabName>
    </ThingSaveData>
  </Things>
</WorldData>`

	return writeAnalysisSave(t, meta, world)
}

func TestReadSaveSummary(t *testing.T) {
	path := analysisFixture(t)

	summary, err := ReadSaveSummary(path)
	if err != nil {
		t.Fatal(err)
	}

	if summary.ID != "world-id" || summary.WorldName != "Europa" || summary.WorldFileName != "TestWorld" {
		t.Fatalf("unexpected identity metadata: %+v", summary)
	}
	if summary.DaysPlayed != 67 || summary.Things != 6 || summary.Atmospheres != 2330 {
		t.Fatalf("unexpected world counters: %+v", summary)
	}
	if summary.Rooms != 11 || summary.PipeNetworks != 22 || summary.CableNetworks != 33 {
		t.Fatalf("unexpected network counters: %+v", summary)
	}
	if summary.SavedAt.Unix() != 12 || summary.SavedAt.Nanosecond() != 345678900 {
		t.Fatalf("unexpected FILETIME conversion: %s", summary.SavedAt)
	}
	if summary.ArchiveSize <= 0 || summary.WorldXMLSize == 0 || summary.WorldXMLCompressed == 0 {
		t.Fatalf("missing archive size information: %+v", summary)
	}
}

func TestAnalyzeSaveStreamsThingStatistics(t *testing.T) {
	analysis, err := AnalyzeSave(context.Background(), analysisFixture(t))
	if err != nil {
		t.Fatal(err)
	}

	if analysis.Players != 3 || analysis.PlayersAlive != 1 || analysis.PlayersUnconscious != 1 {
		t.Fatalf("unexpected player counts: %+v", analysis)
	}
	if analysis.Furnaces != 2 || analysis.DestroyedFurnaces != 1 {
		t.Fatalf("unexpected furnace counts: %+v", analysis)
	}
}

func TestAnalyzeSaveHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := AnalyzeSave(ctx, analysisFixture(t))
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestReadSaveSummaryRejectsMissingWorld(t *testing.T) {
	incomplete := filepath.Join(t.TempDir(), "incomplete.save")
	file, err := os.Create(incomplete)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create(worldMetaFilename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(`<WorldMetaData />`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadSaveSummary(incomplete); err == nil {
		t.Fatal("expected missing world.xml error")
	}
}

func TestAnalyzeConfiguredSave(t *testing.T) {
	path := os.Getenv("SSUI_ANALYSIS_TEST_SAVE")
	if path == "" {
		t.Skip("set SSUI_ANALYSIS_TEST_SAVE to inspect a real save")
	}

	analysis, err := AnalyzeSave(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("analysis: %+v", analysis)
}

func BenchmarkAnalyzeSave(b *testing.B) {
	path := os.Getenv("SSUI_ANALYSIS_BENCHMARK_SAVE")
	if path == "" {
		b.Skip("set SSUI_ANALYSIS_BENCHMARK_SAVE to benchmark a real save")
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := AnalyzeSave(context.Background(), path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadSaveSummary(b *testing.B) {
	path := os.Getenv("SSUI_ANALYSIS_BENCHMARK_SAVE")
	if path == "" {
		b.Skip("set SSUI_ANALYSIS_BENCHMARK_SAVE to benchmark a real save")
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := ReadSaveSummary(path); err != nil {
			b.Fatal(err)
		}
	}
}
