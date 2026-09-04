package backupmgr

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func primeBackupInventory(t *testing.T, m *BackupManager) {
	t.Helper()
	if err := loadInventory(m); err != nil {
		t.Fatal(err)
	}
	for name := range m.records {
		if _, err := analyzeBackup(context.Background(), m, name); err != nil {
			t.Fatal(err)
		}
	}
}

func waitForBackup(t *testing.T, m *BackupManager, name string, analyzed bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m.stateMu.RLock()
		record, exists := m.records[name]
		m.stateMu.RUnlock()
		if exists && (!analyzed || record.ScanVersion == analysisVersion) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("backup %s did not reach analyzed=%v", name, analyzed)
}

func TestManifestRestoresAnalysisAndRuntimeReadsUseRAM(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{SafeBackupDir: filepath.Dir(path)}
	m := NewBackupManager(cfg)
	primeBackupInventory(t, m)
	if err := saveManifest(m); err != nil {
		t.Fatal(err)
	}
	next := NewBackupManager(cfg)
	if err := loadInventory(next); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, path+".hidden"); err != nil {
		t.Fatal(err)
	}
	// Runtime reads must not reopen either the ZIP or the manifest.
	if err := os.Remove(filepath.Join(cfg.SafeBackupDir, manifestFilename)); err != nil {
		t.Fatal(err)
	}
	saves, err := next.ListBackups(0)
	if err != nil || len(saves) != 1 || saves[0].Summary.DaysPlayed != 67 {
		t.Fatalf("list: %+v, %v", saves, err)
	}
	analysis, err := next.AnalyzeBackup(context.Background(), 0)
	if err != nil || analysis.Players != 3 || analysis.Furnaces != 2 {
		t.Fatalf("analysis: %+v, %v", analysis, err)
	}
}

func TestBackfillStartsWithoutAutosaveFolder(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{BackupDir: filepath.Join(t.TempDir(), "not-created-yet"), SafeBackupDir: filepath.Dir(path), WaitTime: 10 * time.Millisecond}
	m := NewBackupManager(cfg)
	if err := m.Start("test"); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()
	waitForBackup(t, m, filepath.Base(path), true)
	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(path, filepath.Join(cfg.BackupDir, "130226_173454_auto.save")); err != nil {
		t.Fatal(err)
	}
	waitForBackup(t, m, "130226_173454_auto.save", true)
}

func TestReloadKeepsPendingObservation(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	name := "130226_173454_auto.save"
	if err := copyFile(analysisFixture(t), filepath.Join(cfg.BackupDir, name)); err != nil {
		t.Fatal(err)
	}
	old := NewBackupManager(cfg)
	if err := loadInventory(old); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := pollBackups(old, now); err != nil {
		t.Fatal(err)
	}
	old.Shutdown()
	next := NewBackupManager(cfg)
	inheritDetectorState(next, old)
	if err := loadInventory(next); err != nil {
		t.Fatal(err)
	}
	if err := pollBackups(next, now.Add(45*time.Second)); err != nil {
		t.Fatal(err)
	}
	ready, identity := nextAutosave(next)
	if ready != name {
		t.Fatal("reload lost the first observation")
	}
	if err := handleBackup(next, ready, identity); err != nil {
		t.Fatal(err)
	}
	if next.records[name].ScanVersion != analysisVersion {
		t.Fatal("handling did not analyze the copied save")
	}
	other := NewBackupManager(BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: cfg.SafeBackupDir})
	inheritDetectorState(other, old)
	if len(other.observed) != 0 {
		t.Fatal("state crossed source folders")
	}
}

func TestRetentionDoesNotResurrectAutosaveAfterRestart(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	name := "130226_173454_auto.save"
	source := filepath.Join(cfg.BackupDir, name)
	if err := copyFile(analysisFixture(t), source); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	if err := m.handleNewBackup(source); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	assertFileExists(t, filepath.Join(cfg.SafeBackupDir, name), false)
	assertFileExists(t, source, true)
	m.Shutdown()
	next := NewBackupManager(cfg)
	if err := loadInventory(next); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := pollBackups(next, now); err != nil {
		t.Fatal(err)
	}
	if err := pollBackups(next, now.Add(45*time.Second)); err != nil {
		t.Fatal(err)
	}
	if len(next.pending) != 0 || !next.handled[name] {
		t.Fatal("retention-deleted archive was queued again")
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := pollBackups(next, now.Add(90*time.Second)); err != nil {
		t.Fatal(err)
	}
	if len(next.handled) != 0 {
		t.Fatal("source tombstones grew without bound")
	}
}

func TestManifestFailurePreventsRetentionDeletion(t *testing.T) {
	path := analysisFixture(t)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
	primeBackupInventory(t, m)
	// A directory at the manifest path makes publishing fail on both supported OSes.
	if err := os.Mkdir(filepath.Join(m.config.SafeBackupDir, manifestFilename), 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.cleanSafeBackupDir(); err == nil {
		t.Fatal("retention ignored persistence failure")
	}
	assertFileExists(t, path, true)
}

func TestManifestRecoveryAndFutureVersionProtection(t *testing.T) {
	for _, test := range []struct {
		name, data string
		future     bool
	}{
		{"broken", "{truncated", false},
		{"future", `{"version":999,"new-format":{}}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := analysisFixture(t)
			manifestPath := filepath.Join(filepath.Dir(path), manifestFilename)
			if err := os.WriteFile(manifestPath, []byte(test.data), 0600); err != nil {
				t.Fatal(err)
			}
			m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
			err := loadInventory(m)
			if test.future {
				if err == nil {
					t.Fatal("accepted a newer manifest version")
				}
				m.Shutdown()
				data, err := os.ReadFile(manifestPath)
				if err != nil || string(data) != test.data {
					t.Fatal("future manifest was overwritten")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			quarantined, _ := filepath.Glob(manifestPath + ".invalid-*")
			if len(quarantined) != 1 {
				t.Fatal("broken manifest not preserved")
			}
			primeBackupInventory(t, m)
			if err := saveManifest(m); err != nil {
				t.Fatal(err)
			}
			manifest, err := readManifest(manifestPath)
			if err != nil || len(manifest.Backups) != 1 {
				t.Fatalf("rebuilt manifest: %+v, %v", manifest, err)
			}
		})
	}
}

func TestStartupRetriesStaleAndFailedAnalysis(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{SafeBackupDir: filepath.Dir(path), BackupDir: t.TempDir()}
	m := NewBackupManager(cfg)
	primeBackupInventory(t, m)
	name := filepath.Base(path)
	record := m.records[name]
	record.ScanVersion = analysisVersion + 1
	record.ScanError = "previous interrupted scan"
	m.records[name] = record
	m.revision++
	if err := saveManifest(m); err != nil {
		t.Fatal(err)
	}
	next := NewBackupManager(cfg)
	if err := next.Start("test"); err != nil {
		t.Fatal(err)
	}
	waitForBackup(t, next, name, true)
	next.Shutdown()
	if next.records[name].ScanError != "" {
		t.Fatal("successful retry kept old failure")
	}
}

func TestFailedScanKeepsSummaryAndDoesNotBusyRetry(t *testing.T) {
	path := writeAnalysisSave(t, `<WorldMetaData><DaysPast>12</DaysPast></WorldMetaData>`, `<WorldData><Things>`)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path), BackupDir: t.TempDir()})
	if err := m.Start("test"); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		m.stateMu.RLock()
		record := m.records[filepath.Base(path)]
		revision := m.revision
		m.stateMu.RUnlock()
		if record.ScanError != "" {
			if !record.SummaryReady || record.Analysis.DaysPlayed != 12 || record.ScanVersion != 0 {
				t.Fatalf("failed analysis state: %+v", record)
			}
			time.Sleep(40 * time.Millisecond)
			m.stateMu.RLock()
			unchanged := revision == m.revision
			m.stateMu.RUnlock()
			if !unchanged {
				t.Fatal("failed archive was continuously retried")
			}
			assertFileExists(t, path, true)
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("failed scan never recorded")
}

func TestBackfillDoesNotHoldListOrDetectorLocks(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{SafeBackupDir: filepath.Dir(path), BackupDir: t.TempDir(), WaitTime: 10 * time.Millisecond}
	m := NewBackupManager(cfg)
	m.scanGate <- struct{}{} // Hold the expensive-work slot, not the inventory.
	if err := m.Start("test"); err != nil {
		t.Fatal(err)
	}
	name := "130226_173454_auto.save"
	if err := copyFile(path, filepath.Join(cfg.BackupDir, name)); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if saves, err := m.ListBackups(0); err != nil || len(saves) != 1 {
			t.Fatalf("list blocked by pending scan: %v", err)
		}
		m.stateMu.RLock()
		_, ready := m.pending[name]
		m.stateMu.RUnlock()
		if ready {
			break
		}
		time.Sleep(time.Millisecond)
	}
	m.stateMu.RLock()
	_, ready := m.pending[name]
	m.stateMu.RUnlock()
	if !ready {
		t.Fatal("detector stopped during backfill")
	}
	<-m.scanGate
	waitForBackup(t, m, name, true)
	m.Shutdown()
}

func TestPendingScanCancelsAndCanResume(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{SafeBackupDir: filepath.Dir(path), BackupDir: t.TempDir()}
	m := NewBackupManager(cfg)
	if err := loadInventory(m); err != nil {
		t.Fatal(err)
	}
	m.scanGate <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { _, err := m.AnalyzeBackup(ctx, 0); result <- err }()
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("scan wait ignored cancellation")
	}
	<-m.scanGate
	m.Shutdown()
	next := NewBackupManager(cfg)
	if err := next.Start("test"); err != nil {
		t.Fatal(err)
	}
	waitForBackup(t, next, filepath.Base(path), true)
	next.Shutdown()
}

func TestConcurrentListAnalysisAndPersistence(t *testing.T) {
	path := analysisFixture(t)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path), BackupDir: t.TempDir()})
	if err := m.Start("test"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if _, err := m.ListBackups(0); err != nil {
					t.Error(err)
				}
				if _, err := m.AnalyzeBackup(context.Background(), 0); err != nil {
					t.Error(err)
				}
				if err := saveManifest(m); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	m.Shutdown()
}

func makeLargeManifest() backupManifest {
	manifest := backupManifest{Version: manifestVersion, Backups: make(map[string]backupRecord, 4000)}
	for i := 0; i < 4000; i++ {
		stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * 5 * time.Minute)
		manifest.Backups[stamp.Format("020106_150405_auto.save")] = backupRecord{
			Size: 20000000, ModifiedNS: stamp.UnixNano(), SummaryReady: true, ScanVersion: analysisVersion,
			Analysis: SaveAnalysis{SaveSummary: SaveSummary{ID: fmt.Sprintf("world-%032d", i), GameVersion: "0.2.6136.26812", SavedAt: stamp, DaysPlayed: 100, WorldName: "Long running world", WorldFileName: "autosave", Things: 123456, Rooms: 1234, Atmospheres: 55555, ArchiveSize: 20000000, WorldXMLSize: 200000000}, Players: 10, PlayersAlive: 5, Furnaces: 300},
		}
	}
	return manifest
}

func TestManifest4000Records(t *testing.T) {
	manifest := makeLargeManifest()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 4<<20 {
		t.Fatalf("4000 aggregate records unexpectedly need %d bytes", len(data))
	}
	runtime.GC()
	runtime.GC() // Discard JSON's pooled encode buffers before measuring retained data.
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	var restored backupManifest
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	runtime.GC()
	runtime.ReadMemStats(&after)
	retained := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	if retained > 8<<20 {
		t.Fatalf("unexpected inventory growth: %d bytes", retained)
	}
	t.Logf("4000 backups: manifest %.2f MiB, decode %s, retained heap delta %.2f MiB", float64(len(data))/(1<<20), elapsed, float64(retained)/(1<<20))
	if len(restored.Backups) != 4000 {
		t.Fatal("lost records")
	}
	runtime.KeepAlive(manifest)
	runtime.KeepAlive(data)
	runtime.KeepAlive(restored)
}

// Opt in with an actual game save. All writes stay in t.TempDir().
func TestHandleConfiguredSave(t *testing.T) {
	path := os.Getenv("SSUI_ANALYSIS_TEST_SAVE")
	if path == "" {
		t.Skip("set SSUI_ANALYSIS_TEST_SAVE to test a real save")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(BackupConfig{BackupDir: filepath.Dir(path), SafeBackupDir: t.TempDir()})
	start := time.Now()
	if err := m.handleNewBackup(path); err != nil {
		t.Fatal(err)
	}
	t.Logf("complete handling: %s", time.Since(start))
	record := m.records[filepath.Base(path)]
	if record.ScanVersion != analysisVersion {
		t.Fatal("real save did not finish analysis")
	}
	t.Logf("archive %.2f MiB, world.xml %.2f MiB, things %d, players %d, furnaces %d", float64(record.Size)/(1<<20), float64(record.Analysis.WorldXMLSize)/(1<<20), record.Analysis.Things, record.Analysis.Players, record.Analysis.Furnaces)
	hash := func(path string) string {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		digest := sha256.New()
		if _, err := io.Copy(digest, file); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%x", digest.Sum(nil))
	}
	if hash(path) != hash(filepath.Join(m.config.SafeBackupDir, filepath.Base(path))) {
		t.Fatal("archived bytes differ from the game save")
	}
	m.Shutdown()
}

func BenchmarkManifest4000(b *testing.B) {
	manifest := makeLargeManifest()
	data, err := json.Marshal(manifest)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var decoded backupManifest
			if err := json.Unmarshal(data, &decoded); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := json.Marshal(manifest); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestCopiedSaveSurvivesCanceledAnalysis(t *testing.T) {
	path := analysisFixture(t)
	identity, err := identifySave(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	record, err := scanCopiedBackup(ctx, path, identity)
	if err != nil {
		t.Fatalf("discarded a good copy on reload: %v", err)
	}
	if record.ScanError != context.Canceled.Error() || !record.SummaryReady || record.ScanVersion != 0 || record.Analysis.DaysPlayed != 67 {
		t.Fatalf("bad interrupted handling state: %+v", record)
	}
}

func TestNestedBackupDownloadAndRestore(t *testing.T) {
	t.Chdir(t.TempDir())
	safe := t.TempDir()
	if err := os.Mkdir(filepath.Join(safe, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(safe, "nested", "130226_173454_auto.save")
	if err := copyFile(analysisFixture(t), path); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(BackupConfig{WorldName: "restored", SafeBackupDir: safe})
	primeBackupInventory(t, m)
	saves, err := m.ListBackups(0)
	if err != nil || len(saves) != 1 || saves[0].SaveFile != path {
		t.Fatalf("nested list: %+v, %v", saves, err)
	}
	download, err := m.GetBackupFileData(0)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(original) != sha256.Sum256(download.Data) {
		t.Fatal("download bytes changed")
	}
	if err := m.RestoreBackupFile(path); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join("saves", "restored", "restored.save")
	identity, err := identifySave(restored)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBackupSave(context.Background(), restored, identity); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeSave(context.Background(), restored)
	if err != nil || analysis.Players != 3 {
		t.Fatalf("restored analysis: %+v, %v", analysis, err)
	}
}
