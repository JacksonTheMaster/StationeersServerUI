package backupmgr

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateBackupSave(t *testing.T) {
	for _, tc := range []struct {
		name, meta, world string
		valid             bool
	}{
		{"complete", "<WorldMetaData></WorldMetaData>", "<WorldData><Things/></WorldData>", true},
		{"game text with raw controls", "<WorldMetaData/>", "<WorldData><Text>label\x0bvalue\x01</Text></WorldData>", true},
		{"game text with control references", "<WorldMetaData/>", "<WorldData><Text>label&#xB;value&#11;</Text></WorldData>", true},
		{"controls do not repair truncation", "<WorldMetaData/>", "<WorldData><Text>label\x0b</Text>", false},
		{"empty roots", "<WorldMetaData/>", "<WorldData/>", true},
		{"trailing whitespace", "<WorldMetaData/>\n", "<WorldData/>\n", true},
		{"incomplete world", "<WorldMetaData/>", "<WorldData><Things/>", false},
		{"incomplete metadata", "<WorldMetaData>", "<WorldData/>", false},
		{"bad nesting", "<WorldMetaData/>", "<WorldData><Things></WorldData>", false},
		{"extra root", "<WorldMetaData/>", "<WorldData/><WorldData/>", false},
		{"wrong root", "<WorldMetaData/>", "<Other/>", false},
		{"trailing garbage", "<WorldMetaData/>", "<WorldData/>broken", false},
		{"missing root", "<WorldMetaData/>", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeAnalysisSave(t, tc.meta, tc.world)
			identity, err := identifySave(path)
			if err != nil {
				t.Fatal(err)
			}
			err = validateBackupSave(context.Background(), path, identity)
			if (err == nil) != tc.valid {
				t.Fatalf("validation = %v, want valid=%v", err, tc.valid)
			}
		})
	}
}

func TestValidateBackupSaveRejectsBrokenArchives(t *testing.T) {
	for _, kind := range []string{"missing member", "duplicate member", "truncated zip", "bad crc"} {
		t.Run(kind, func(t *testing.T) {
			var data bytes.Buffer
			writer := zip.NewWriter(&data)
			members := []string{worldMetaFilename, worldFilename}
			if kind == "missing member" {
				members = members[:1]
			}
			if kind == "duplicate member" {
				members = append(members, worldFilename)
			}
			for _, name := range members {
				entry, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
				if err != nil {
					t.Fatal(err)
				}
				body := "<WorldData>original</WorldData>"
				if name == worldMetaFilename {
					body = "<WorldMetaData/>"
				}
				if _, err := entry.Write([]byte(body)); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			contents := data.Bytes()
			if kind == "truncated zip" {
				contents = contents[:len(contents)-22]
			}
			if kind == "bad crc" {
				contents = bytes.Replace(contents, []byte("original"), []byte("modified"), 1)
			}
			path := filepath.Join(t.TempDir(), "broken.save")
			if err := os.WriteFile(path, contents, 0600); err != nil {
				t.Fatal(err)
			}
			identity, err := identifySave(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateBackupSave(context.Background(), path, identity); err == nil {
				t.Fatal("accepted broken archive")
			}
		})
	}
}

func TestPollBackupsRequiresTwoUnchangedObservations(t *testing.T) {
	root := t.TempDir()
	m := NewBackupManager(BackupConfig{BackupDir: root, SafeBackupDir: t.TempDir()})
	path := filepath.Join(root, "130226_173454_auto.save")
	if err := copyFile(analysisFixture(t), path); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	poll := func(at time.Time) {
		t.Helper()
		if err := pollBackups(m, at); err != nil {
			t.Fatal(err)
		}
	}
	poll(now)
	if len(m.pending) != 0 {
		t.Fatal("queued on first observation")
	}
	poll(now.Add(44 * time.Second))
	if len(m.pending) != 0 {
		t.Fatal("queued before a full interval")
	}
	stamp := now.Add(time.Hour)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	poll(now.Add(45 * time.Second))
	if len(m.pending) != 0 {
		t.Fatal("queued changed file")
	}
	poll(now.Add(90 * time.Second))
	name, identity := nextAutosave(m)
	if name == "" {
		t.Fatal("did not queue stable save")
	}
	if err := handleBackup(m, name, identity); err != nil {
		t.Fatal(err)
	}
	// Dated filenames remain archival keys even if a source is modified.
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatal(err)
	}
	poll(now.Add(135 * time.Second))
	if len(m.pending) != 0 {
		t.Fatal("queued an already archived name")
	}
	m.config.BackupDir = filepath.Join(root, "offline")
	if err := pollBackups(m, now.Add(180*time.Second)); err == nil {
		t.Fatal("expected scan error")
	}
	if !m.handled[name] {
		t.Fatal("scan failure discarded processed state")
	}
	m.config.BackupDir = root
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	poll(now.Add(225 * time.Second))
	if len(m.handled) != 0 {
		t.Fatal("removed sources still consume processed state")
	}
}

func TestBackupValidationCancellationAndChangedIdentity(t *testing.T) {
	path := analysisFixture(t)
	identity, err := identifySave(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := validateBackupSave(ctx, path, identity); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	identity.size++
	if err := validateBackupSave(context.Background(), path, identity); err == nil {
		t.Fatal("accepted changed file")
	}
}

func TestBackupManagerReloadCancelsInitialization(t *testing.T) {
	m := NewBackupManager(BackupConfig{BackupDir: filepath.Join(t.TempDir(), "missing"), SafeBackupDir: t.TempDir()})
	result := make(chan error, 1)
	go func() { result <- m.Start("test") }()
	m.Shutdown()
	select {
	case err := <-result:
		// Initialization retains its legacy formatted cancellation message.
		if err != nil && !strings.Contains(err.Error(), context.Canceled.Error()) {
			t.Fatalf("got %v, want cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reload did not cancel initialization")
	}
	m.Shutdown() // Repeated shutdown must be safe.
}

func TestBackupManagerPollingUsesExistingCopy(t *testing.T) {
	root, safe := t.TempDir(), t.TempDir()
	m := NewBackupManager(BackupConfig{BackupDir: root, SafeBackupDir: safe, WaitTime: 10 * time.Millisecond,
		RetentionPolicy: RetentionPolicy{CleanupInterval: time.Hour}})
	if err := m.Start("test"); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()
	if err := copyFile(analysisFixture(t), filepath.Join(root, "new.save")); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		// Listing shares the copy lock and cannot observe a half-written output.
		saves, err := m.ListBackups(0)
		if err == nil && len(saves) == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("polling did not hand save to copy operation")
}

func TestScanBackupFilesFollowsConfiguredRootOnly(t *testing.T) {
	root := t.TempDir()
	actual := t.TempDir()
	name := "130226_173454_auto.save"
	if err := copyFile(analysisFixture(t), filepath.Join(actual, name)); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "autosave")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(actual, filepath.Join(actual, "loop")); err != nil {
		t.Fatal(err)
	}
	files, err := scanBackupFiles(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected exactly one file through root link, got %d", len(files))
	}
	if _, exists := files[filepath.Join(link, name)]; !exists {
		t.Fatal("snapshot no longer uses configured paths")
	}
}
