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

func TestPollBackupsRetriesAndDeduplicates(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	old := filepath.Join(root, "old.save")
	if err := os.WriteFile(old, []byte("startup baseline"), 0600); err != nil {
		t.Fatal(err)
	}
	handled, err := scanBackupFiles(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "new.save")
	if err := os.WriteFile(path, []byte("still writing"), 0600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	fail := true
	copyBackup := func(got string) error {
		if got != path {
			t.Fatalf("unexpected copy: %s", got)
		}
		calls++
		if fail {
			return errors.New("temporary copy failure")
		}
		return nil
	}
	poll := func() {
		t.Helper()
		if err := pollBackups(ctx, root, handled, copyBackup); err != nil {
			t.Fatal(err)
		}
	}
	poll()
	if calls != 0 {
		t.Fatal("copied incomplete save")
	}
	if err := copyFile(analysisFixture(t), path); err != nil {
		t.Fatal(err)
	}
	poll()
	if calls != 1 {
		t.Fatal("did not try completed save")
	}
	fail = false
	poll()
	poll()
	if calls != 2 {
		t.Fatalf("retry/dedup: got %d copy attempts", calls)
	}
	stamp := time.Now().Add(time.Hour)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	poll()
	if calls != 3 {
		t.Fatal("did not detect modified save")
	}
	if err := pollBackups(ctx, filepath.Join(root, "offline"), handled, copyBackup); err == nil {
		t.Fatal("expected scan error")
	}
	if len(handled) != 2 {
		t.Fatal("scan error discarded baseline")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	poll()
	if _, exists := handled[path]; exists {
		t.Fatal("deleted file still tracked")
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
		if err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
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
