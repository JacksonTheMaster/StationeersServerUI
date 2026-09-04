package backupmgr

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCanceledManagerCannotStartRequestAnalysis(t *testing.T) {
	path := analysisFixture(t)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
	if err := loadInventory(m); err != nil {
		t.Fatal(err)
	}
	m.Shutdown()
	// Use an uncanceled request context to exercise the manager's own boundary,
	// without depending on scheduling of context.AfterFunc's callback.
	if _, err := analyzeBackup(context.Background(), m, filepath.Base(path)); !errors.Is(err, context.Canceled) {
		t.Fatalf("stopped manager started analysis: %v", err)
	}
}

func TestBackupAnalysisAcceptsNilRequestContext(t *testing.T) {
	path := analysisFixture(t)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
	primeBackupInventory(t, m)
	analysis, err := m.AnalyzeBackup(nil, 0)
	if err != nil || analysis.Players != 3 {
		t.Fatalf("nil-context analysis: %+v, %v", analysis, err)
	}
}

func TestBackfillPrefersRecentDatesAcrossYearBoundary(t *testing.T) {
	m := NewBackupManager(BackupConfig{})
	m.records["311225_120000_auto.save"] = backupRecord{}
	m.records["010126_120000_auto.save"] = backupRecord{}
	backlog := pendingAnalyses(m)
	if len(backlog) != 2 || backlog[0] != "010126_120000_auto.save" {
		t.Fatalf("backfill order: %v", backlog)
	}
}

func TestUnreadableManifestIsNotQuarantined(t *testing.T) {
	safe := t.TempDir()
	path := filepath.Join(safe, manifestFilename)
	// Reading a directory fails independently of privileges, on Linux and Windows.
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(BackupConfig{SafeBackupDir: safe})
	if err := loadInventory(m); err == nil {
		t.Fatal("expected a storage read error")
	}
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		t.Fatalf("storage error moved the original manifest path: %v", err)
	}
	quarantined, _ := filepath.Glob(path + ".invalid-*")
	if len(quarantined) != 0 {
		t.Fatal("storage error was treated as corrupt JSON")
	}
}

func TestNewerManifestCanChangeFieldTypes(t *testing.T) {
	safe := t.TempDir()
	path := filepath.Join(safe, manifestFilename)
	original := []byte(`{"version":999,"backups":[],"handled":[]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(BackupConfig{SafeBackupDir: safe})
	if err := loadInventory(m); err == nil {
		t.Fatal("accepted an unknown schema")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(original) {
		t.Fatal("modified a newer manifest")
	}
}

func TestRetentionDoesNotDeleteExternallyChangedArchive(t *testing.T) {
	safe := t.TempDir()
	path := writeBackupSave(t, safe, "old.save", time.Now().Add(-48*time.Hour))
	m := NewBackupManager(BackupConfig{SafeBackupDir: safe})
	primeBackupInventory(t, m)
	if err := os.WriteFile(path, []byte("a different archive imported by the user"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.cleanSafeBackupDir(); err == nil {
		t.Fatal("retention ignored an archive identity change")
	}
	assertFileExists(t, path, true)
}

func TestSourceCleanupKeepsLastCopyAfterArchiveDisappears(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(map[bool]string{false: "running", true: "restarted"}[restart], func(t *testing.T) {
			source, safe := t.TempDir(), t.TempDir()
			stamp := time.Now().Add(-48 * time.Hour)
			original := writeBackupSave(t, source, "old.save", stamp)
			if err := os.Chtimes(original, stamp, stamp); err != nil {
				t.Fatal(err)
			}
			cfg := BackupConfig{BackupDir: source, SafeBackupDir: safe, RetentionPolicy: RetentionPolicy{KeepNewestCount: 1}}
			m := NewBackupManager(cfg)
			if err := m.handleNewBackup(original); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(safe, "old.save")); err != nil {
				t.Fatal(err)
			}
			if restart {
				m.Shutdown()
				m = NewBackupManager(cfg)
				if err := loadInventory(m); err != nil {
					t.Fatal(err)
				}
			}
			_ = m.Cleanup() // It may report the missing archive, but cannot remove the source.
			assertFileExists(t, original, true)
		})
	}
}

func TestBackupFoldersCannotOverlap(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		name, source, safe string
		valid              bool
	}{
		{"siblings", filepath.Join(root, "autosave"), filepath.Join(root, "Safebackups"), true},
		{"same", root, root, false},
		{"destination in source", root, filepath.Join(root, "Safebackups"), false},
		{"source in destination", filepath.Join(root, "autosave"), root, false},
		{"similar prefix", filepath.Join(root, "save"), filepath.Join(root, "save-backups"), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := checkBackupFolders(BackupConfig{BackupDir: test.source, SafeBackupDir: test.safe})
			if (err == nil) != test.valid {
				t.Fatalf("folder validation: %v", err)
			}
		})
	}
	t.Run("symlink parent", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "linked")
		if err := os.Symlink(root, link); err != nil {
			t.Skipf("cannot create symlinks: %v", err)
		}
		if err := checkBackupFolders(BackupConfig{BackupDir: root, SafeBackupDir: filepath.Join(link, "not-created-yet")}); err == nil {
			t.Fatal("symlink concealed overlapping directories")
		}
	})
}
