package backupmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func checkBackupFolders(cfg BackupConfig) error {
	if cfg.BackupDir == "" || cfg.SafeBackupDir == "" {
		return fmt.Errorf("autosave and safe backup folders must both be configured")
	}
	source, err := resolveBackupFolder(cfg.BackupDir)
	if err != nil {
		return err
	}
	safe, err := resolveBackupFolder(cfg.SafeBackupDir)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		source = strings.ToLower(source)
		safe = strings.ToLower(safe)
	}
	toSafe, safeErr := filepath.Rel(source, safe)
	toSource, sourceErr := filepath.Rel(safe, source)
	if safeErr == nil && filepath.IsLocal(toSafe) || sourceErr == nil && filepath.IsLocal(toSource) {
		return fmt.Errorf("autosave and safe backup folders must not overlap: %s and %s", cfg.BackupDir, cfg.SafeBackupDir)
	}
	return nil
}

// Resolve existing parents too, so a not-yet-created folder beneath a symlink
// cannot hide an overlap. The game may create the actual autosave folder later.
func resolveBackupFolder(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(absolute)
		if err == nil {
			stat, err := os.Stat(resolved)
			if err != nil {
				return "", err
			}
			if !stat.IsDir() {
				return "", fmt.Errorf("backup folder is not a directory: %s", absolute)
			}
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(absolute)
		if parent == absolute {
			return "", err
		}
		missing = append(missing, filepath.Base(absolute))
		absolute = parent
	}
}
