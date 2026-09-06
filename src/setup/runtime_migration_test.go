package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeMigrationKeepsUIModAndCopiesUserData(t *testing.T) {
	root := t.TempDir()
	oldRoot := filepath.Join(root, "UIMod")
	newRoot := filepath.Join(root, "SSUI")
	writeMigrationTestFile(t, filepath.Join(oldRoot, "config", "config.json"), "old config")
	writeMigrationTestFile(t, filepath.Join(oldRoot, "onboard_bundled", "old.html"), "old ui")
	writeMigrationTestFile(t, filepath.Join(newRoot, "onboard_bundled", "current.html"), "current ui")

	migrated, err := migrateRuntimeFolder(oldRoot, newRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !migrated {
		t.Fatal("migration did not run")
	}
	assertMigrationTestFile(t, filepath.Join(oldRoot, "config", "config.json"), "old config")
	assertMigrationTestFile(t, filepath.Join(newRoot, "config", "config.json"), "old config")
	assertMigrationTestFile(t, filepath.Join(newRoot, "onboard_bundled", "current.html"), "current ui")
	if _, err := os.Stat(filepath.Join(newRoot, "onboard_bundled", "old.html")); !os.IsNotExist(err) {
		t.Fatalf("old bundled UI was copied: %v", err)
	}
}

func TestRuntimeMigrationDoesNotMergeIntoExistingData(t *testing.T) {
	root := t.TempDir()
	oldRoot := filepath.Join(root, "UIMod")
	newRoot := filepath.Join(root, "SSUI")
	writeMigrationTestFile(t, filepath.Join(oldRoot, "config", "config.json"), "old")
	writeMigrationTestFile(t, filepath.Join(newRoot, "config", "config.json"), "new")

	migrated, err := migrateRuntimeFolder(oldRoot, newRoot)
	if err != nil {
		t.Fatal(err)
	}
	if migrated {
		t.Fatal("migration merged two runtime folders")
	}
	assertMigrationTestFile(t, filepath.Join(newRoot, "config", "config.json"), "new")
}

func writeMigrationTestFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
}

func assertMigrationTestFile(t *testing.T, path, expected string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatalf("%s = %q, want %q", path, data, expected)
	}
}
