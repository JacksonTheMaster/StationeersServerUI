package backupmgr

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupDownloadByteLimit(t *testing.T) {
	dir := t.TempDir()
	payload := bytes.Repeat([]byte("x"), 2048)
	if err := os.WriteFile(filepath.Join(dir, "limited.save"), payload, 0600); err != nil {
		t.Fatal(err)
	}
	m := NewBackupManager(BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: dir})
	for _, limit := range []int64{0, 2048, 2049} {
		data, err := ReadBackupFileData(m, "limited.save", limit)
		if err != nil || !bytes.Equal(data.Data, payload) {
			t.Fatalf("limit %d: %v", limit, err)
		}
	}
	if data, err := ReadBackupFileData(m, "limited.save", 2047); !errors.Is(err, ErrBackupTooLarge) || data != nil {
		t.Fatalf("oversized download: %v %v", data, err)
	}
	if _, err := ReadBackupFileData(m, "limited.save", -1); err == nil {
		t.Fatal("negative limit accepted")
	}
	if _, err := ReadBackupFileData(m, "../limited.save", 2048); !errors.Is(err, ErrInvalidBackupName) {
		t.Fatalf("limit bypassed name validation: %v", err)
	}
}
