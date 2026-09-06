package setup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const legacyRuntimeFolder = "UIMod"
const runtimeFolder = "SSUI"

// MigrateLegacyRuntimeFolder copies the old UIMod data once. The old folder is
// deliberately kept around so a failed v6 upgrade can still be rolled back.
func MigrateLegacyRuntimeFolder() (bool, error) {
	return migrateRuntimeFolder(legacyRuntimeFolder, runtimeFolder)
}

func migrateRuntimeFolder(oldRoot, newRoot string) (bool, error) {
	oldInfo, err := os.Stat(oldRoot)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect old runtime folder: %w", err)
	}
	if !oldInfo.IsDir() {
		return false, fmt.Errorf("old runtime path %s is not a directory", oldRoot)
	}

	hasData, err := hasRuntimeData(newRoot)
	if err != nil {
		return false, err
	}
	marker := filepath.Join(newRoot, ".uimod-migration-incomplete")
	_, markerErr := os.Stat(marker)
	resume := markerErr == nil
	if markerErr != nil && !os.IsNotExist(markerErr) {
		return false, fmt.Errorf("inspect migration marker: %w", markerErr)
	}
	if hasData && !resume {
		return false, nil
	}
	if err := os.MkdirAll(newRoot, 0755); err != nil {
		return false, fmt.Errorf("create runtime folder: %w", err)
	}
	if err := os.WriteFile(marker, []byte("v6 UIMod migration in progress\n"), 0600); err != nil {
		return false, fmt.Errorf("create migration marker: %w", err)
	}

	entries, err := os.ReadDir(oldRoot)
	if err != nil {
		return false, fmt.Errorf("read old runtime folder: %w", err)
	}
	for _, entry := range entries {
		// These are compiled into the executable. Copying an old build over the
		// current source tree is both wasteful and surprisingly easy to get wrong.
		if entry.Name() == "onboard_bundled" || entry.Name() == "tests" {
			continue
		}
		source := filepath.Join(oldRoot, entry.Name())
		destination := filepath.Join(newRoot, entry.Name())
		if err := copyRuntimeEntry(source, destination); err != nil {
			return false, fmt.Errorf("copy %s: %w", source, err)
		}
	}
	if err := os.Remove(marker); err != nil {
		return false, fmt.Errorf("finish runtime migration: %w", err)
	}
	return true, nil
}

func hasRuntimeData(root string) (bool, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read runtime folder: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() != "onboard_bundled" && entry.Name() != "tests" && entry.Name() != ".uimod-migration-incomplete" {
			return true, nil
		}
	}
	return false, nil
}

func copyRuntimeEntry(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symbolic links are not migrated automatically")
	}
	if info.IsDir() {
		if err := os.MkdirAll(destination, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyRuntimeEntry(filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported file type")
	}
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".migrate-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	}
	_, copyErr := io.Copy(temp, input)
	if copyErr == nil {
		copyErr = temp.Sync()
	}
	closeErr := temp.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tempName, destination)
}
