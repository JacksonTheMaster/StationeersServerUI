package backupmgr

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMissingArchiveIsHandledAgain(t *testing.T) {
	for _, mode := range []string{"running", "restart", "reload", "reload without manifest", "cleanup before detection"} {
		t.Run(mode, func(t *testing.T) {
			cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
			name := "040926_172441_auto.save"
			source := filepath.Join(cfg.BackupDir, name)
			if err := copyFile(analysisFixture(t), source); err != nil {
				t.Fatal(err)
			}
			m := NewBackupManager(cfg)
			if err := m.handleNewBackup(source); err != nil {
				t.Fatal(err)
			}
			archive := filepath.Join(cfg.SafeBackupDir, name)
			if err := os.Remove(archive); err != nil {
				t.Fatal(err)
			}
			if mode == "cleanup before detection" {
				if err := m.Cleanup(); err == nil {
					t.Fatal("cleanup did not notice the missing archive")
				}
				if m.retired[name] {
					t.Fatal("cleanup claimed an already-missing archive as its own deletion")
				}
			}
			if mode == "reload without manifest" {
				if err := os.Remove(filepath.Join(cfg.SafeBackupDir, manifestFilename)); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "restart" || strings.HasPrefix(mode, "reload") {
				m.Shutdown()
				next := NewBackupManager(cfg)
				if strings.HasPrefix(mode, "reload") {
					inheritDetectorState(next, m)
				}
				m = next
				if err := loadInventory(m); err != nil {
					t.Fatal(err)
				}
			}
			defer m.Shutdown()
			now := time.Now()
			if err := pollBackupsAt(m, now); err != nil {
				t.Fatal(err)
			}
			if len(m.pending) != 0 {
				t.Fatal("repair skipped the stability interval")
			}
			if saves, err := m.ListBackups(0); err != nil || len(saves) != 0 {
				t.Fatalf("missing archive still listed: %+v, %v", saves, err)
			}
			if err := pollBackupsAt(m, now.Add(m.config.WaitTime)); err != nil {
				t.Fatal(err)
			}
			ready, identity := nextAutosave(m)
			if ready != name {
				t.Fatalf("missing archive was not queued: %q", ready)
			}
			if err := handleBackup(m, ready, identity); err != nil {
				t.Fatal(err)
			}
			if !analysisReady(m.records[name]) || m.records[name].Analysis.Players != 3 {
				t.Fatal("repair did not run the complete handling pipeline")
			}
			original, _ := os.ReadFile(source)
			repaired, err := os.ReadFile(archive)
			if err != nil || string(original) != string(repaired) {
				t.Fatal("repair changed the archive bytes")
			}
		})
	}
}

func TestRetentionIntentSurvivesInterruptedManifestUpdate(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	name := "040926_172441_auto.save"
	source := filepath.Join(cfg.BackupDir, name)
	if err := copyFile(analysisFixture(t), source); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	if err := m.handleNewBackup(source); err != nil {
		t.Fatal(err)
	}
	// Emulate the durable deletion intent followed by a crash after unlink,
	// before the removed inventory record is written to the manifest.
	m.retired[name] = true
	m.revision++
	if err := saveManifest(m); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(cfg.SafeBackupDir, name)); err != nil {
		t.Fatal(err)
	}
	next := NewBackupManager(cfg)
	if err := loadInventory(next); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, at := range []time.Time{now, now.Add(next.config.WaitTime)} {
		if err := pollBackupsAt(next, at); err != nil {
			t.Fatal(err)
		}
	}
	if len(next.pending) != 0 || !next.retired[name] {
		t.Fatal("intentional deletion was mistaken for a missing archive")
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := pollBackupsAt(next, now.Add(2*next.config.WaitTime)); err != nil {
		t.Fatal(err)
	}
	if len(next.retired) != 0 || len(next.handled) != 0 {
		t.Fatal("retention tombstones outlived the source")
	}
}

func TestArchiveRepairDoesNotOverwriteChangedFile(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	source := filepath.Join(cfg.BackupDir, "changed.save")
	if err := copyFile(analysisFixture(t), source); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	if err := m.handleNewBackup(source); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(cfg.SafeBackupDir, "changed.save")
	if err := os.WriteFile(archive, []byte("externally replaced"), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, at := range []time.Time{now, now.Add(m.config.WaitTime)} {
		if err := pollBackupsAt(m, at); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.pending) != 0 {
		t.Fatal("queued an existing archive for replacement")
	}
	data, _ := os.ReadFile(archive)
	if string(data) != "externally replaced" {
		t.Fatal("overwrote the user's archive")
	}
}

func TestStartRetriesUnreadableManifest(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: filepath.Dir(path), WaitTime: 10 * time.Millisecond}
	manifestPath := filepath.Join(cfg.SafeBackupDir, manifestFilename)
	if err := os.Mkdir(manifestPath, 0755); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	defer m.Shutdown()
	started := make(chan error, 1)
	go func() { started <- m.Start("retry-test") }()
	select {
	case err := <-started:
		t.Fatalf("gave up on unreadable storage: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if err := os.Remove(manifestPath); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-started:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("did not recover when storage became readable")
	}
	waitForBackup(t, m, filepath.Base(path), true)
}

func TestUnavailableArchiveFolderPreservesDetectorState(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: filepath.Join(t.TempDir(), "safe")}
	source := filepath.Join(cfg.BackupDir, "kept.save")
	if err := copyFile(analysisFixture(t), source); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(cfg.SafeBackupDir, 0755); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	if err := m.handleNewBackup(source); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(cfg.SafeBackupDir, cfg.SafeBackupDir+".offline"); err != nil {
		t.Fatal(err)
	}
	if err := pollBackups(m); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected unavailable destination: %v", err)
	}
	if _, exists := m.records["kept.save"]; !exists || !m.handled["kept.save"] || len(m.pending) != 0 {
		t.Fatal("unavailable storage was mistaken for a missing archive")
	}
}

func TestBusyHandlingDoesNotBlockSourceObservation(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	source := filepath.Join(cfg.BackupDir, "new.save")
	if err := copyFile(analysisFixture(t), source); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	if err := loadInventory(m); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	polled := make(chan error, 1)
	go func() { polled <- pollBackups(m) }()
	select {
	case err := <-polled:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("source observation waited for handling")
	}
	if _, exists := m.observed["new.save"]; !exists {
		t.Fatal("busy handling lost the new source observation")
	}
}

func TestStartDoesNotRetryConfigurationOrFutureManifest(t *testing.T) {
	for _, kind := range []string{"overlap", "future manifest"} {
		t.Run(kind, func(t *testing.T) {
			cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: time.Hour}
			if kind == "overlap" {
				cfg.SafeBackupDir = cfg.BackupDir
			} else if err := os.WriteFile(filepath.Join(cfg.SafeBackupDir, manifestFilename), []byte(`{"version":999,"backups":{}}`), 0600); err != nil {
				t.Fatal(err)
			}
			m := NewBackupManager(cfg)
			defer m.Shutdown()
			started := make(chan error, 1)
			go func() { started <- m.Start("invalid-start") }()
			select {
			case err := <-started:
				if err == nil {
					t.Fatal("accepted invalid startup state")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("retried a non-storage error")
			}
		})
	}
}

func TestShutdownInterruptsStartupRetry(t *testing.T) {
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: time.Hour}
	if err := os.Mkdir(filepath.Join(cfg.SafeBackupDir, manifestFilename), 0755); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(cfg)
	started := make(chan error, 1)
	go func() { started <- m.Start("cancel-test") }()
	time.Sleep(20 * time.Millisecond)
	stopped := make(chan struct{})
	go func() { m.Shutdown(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown waited for the retry interval")
	}
	if err := <-started; !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation error: %v", err)
	}
}

func TestUnchangedInventoryDoesNotRewriteManifest(t *testing.T) {
	path := analysisFixture(t)
	cfg := BackupConfig{SafeBackupDir: filepath.Dir(path)}
	m := NewBackupManager(cfg)
	primeBackupInventory(t, m)
	if err := saveManifest(m); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(cfg.SafeBackupDir, manifestFilename)
	stamp := time.Unix(1700000000, 0)
	if err := os.Chtimes(manifestPath, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	next := NewBackupManager(cfg)
	if err := loadInventory(next); err != nil {
		t.Fatal(err)
	}
	next.Shutdown()
	stat, err := os.Stat(manifestPath)
	if err != nil || !stat.ModTime().Equal(stamp) {
		t.Fatal("unchanged startup rewrote the manifest")
	}
}

func TestDeepScanFailureKeepsSummaryWithoutPartialCounts(t *testing.T) {
	world := `<WorldData><ThingSaveData type="HumanSaveData"><State>Alive</State></ThingSaveData><ThingSaveData><PrefabName>` + strings.Repeat("x", 70000) + `</PrefabName></ThingSaveData></WorldData>`
	path := writeAnalysisSave(t, `<WorldMetaData><WorldName>Europa</WorldName><DaysPast>67</DaysPast></WorldMetaData>`, world)
	analysis, err := AnalyzeSave(context.Background(), path)
	if err == nil || analysis.WorldName != "Europa" || analysis.DaysPlayed != 67 || analysis.Players != 0 {
		t.Fatalf("failed scan lost summary or exposed partial counts: %+v, %v", analysis, err)
	}
	identity, err := identifySave(path)
	if err != nil {
		t.Fatal(err)
	}
	record, err := scanCopiedBackup(context.Background(), path, identity)
	if err != nil || !record.SummaryReady || record.Analysis.DaysPlayed != 67 || record.ScanError == "" || record.Analysis.Players != 0 {
		t.Fatalf("copy analysis failed to keep usable metadata: %+v, %v", record, err)
	}
}

func TestWorldXMLLimitIs1000MiB(t *testing.T) {
	if maxWorldSize != 1000*1024*1024 {
		t.Fatalf("unexpected limit: %d", maxWorldSize)
	}
	path := filepath.Join(t.TempDir(), "oversized.save")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	meta, err := writer.Create(worldMetaFilename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := meta.Write([]byte(`<WorldMetaData/>`)); err != nil {
		t.Fatal(err)
	}
	// A declared oversized member must be rejected before decompression.
	if _, err := writer.CreateRaw(&zip.FileHeader{Name: worldFilename, Method: zip.Store, UncompressedSize64: maxWorldSize + 1}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	identity, err := identifySave(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBackupSave(context.Background(), path, identity); err == nil || !strings.Contains(err.Error(), "oversized world.xml") {
		t.Fatalf("validation accepted oversized world: %v", err)
	}
	if _, err := AnalyzeSave(context.Background(), path); err == nil || !strings.Contains(err.Error(), "safety limit") {
		t.Fatalf("analysis accepted oversized world: %v", err)
	}
}
