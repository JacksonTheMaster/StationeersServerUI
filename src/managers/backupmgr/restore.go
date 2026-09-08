package backupmgr

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// RestoreBackup restores a named archive from the current inventory.
func (m *BackupManager) RestoreBackup(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path, err := backupFilePath(m, name)
	if err != nil {
		return err
	}
	logger.Backup.Infof("Restoring backup %s", name)
	return m.restoreBackupSave(path)
}

func (m *BackupManager) restoreBackupSave(backupFile string) error {

	restoredFiles := make(map[string]string)

	destFile := filepath.Join("./saves/"+m.config.WorldName, m.config.WorldName+".save")

	// This check was disabled since it was relatively unnecessary and didnt bring much benefit

	// Before restore, check if we have existing .save files in the root saves/WorldName dir
	//saveDir := filepath.Join("./saves/", m.config.WorldName)
	//files, err := os.ReadDir(saveDir)
	//if err != nil {
	//	return fmt.Errorf("failed to read save directory %s: %w", saveDir, err)
	//}

	//for _, file := range files {
	//	if file.IsDir() {
	//		continue
	//	}
	//	if strings.HasSuffix(file.Name(), ".save") {
	//		existingFile := filepath.Join(saveDir, file.Name())
	//		// Move existing .save file to SafeBackupDir with timestamp to avoid overwrites
	//		timestamp := time.Now().Format("2006-01-02_15-04-05")
	//		savedPreviousHeadSaveFilePath := filepath.Join(m.config.SafeBackupDir, fmt.Sprintf("%s_%s_%s", "pre-restore-HEAD-", timestamp, file.Name()))
	//		if err := os.Rename(existingFile, savedPreviousHeadSaveFilePath); err != nil {
	//			return fmt.Errorf("failed to move existing HEAD .save file %s to %s: %w", existingFile, savedPreviousHeadSaveFilePath, err)
	//		}
	//		logger.Backup.Info("Moved previous HEAD .save file to: " + savedPreviousHeadSaveFilePath)
	//	}
	//}

	// Create temp directory for mod time shenanigans (https://discordapp.com/channels/276525882049429515/392080751648178188/1407157281606336602)
	tempDir := filepath.Join("./saves", m.config.WorldName, "tmp")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temp directory %s: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	// Extract .save (zip) file to tempDir
	r, err := zip.OpenReader(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open zip reader for %s: %w", backupFile, err)
	}
	defer r.Close()

	// --- Safe extraction -------------------------------------------------
	for _, f := range r.File {
		// Sanitize the entry name – strip any leading / or .. components.
		entryName := filepath.Clean(f.Name)

		// Skip empty names or names that contain '..' after cleaning.
		if entryName == "." || entryName == ".." || strings.Contains(entryName, "..") {
			// This entry would escape the target directory; reject it.
			logger.Backup.Warn(fmt.Sprintf("Skipping potentially unsafe zip entry %q", f.Name))
			continue
		}

		destPath := filepath.Join(tempDir, entryName)
		// Ensure the destination is still inside tempDir.
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(tempDir)+string(os.PathSeparator)) {
			logger.Backup.Warn(fmt.Sprintf("Skipping zip entry that would escape extraction dir: %q", f.Name))
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				m.revertRestore(restoredFiles)
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			continue
		}

		// Create any missing parent directories.
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to create parent directory for %s: %w", destPath, err)
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to open file in zip %s: %w", f.Name, err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to extract file %s: %w", destPath, err)
		}
		rc.Close()
		outFile.Close()
	}
	// --------------------------------------------------------------------

	// Update world_meta.xml DateTime with current Windows file time using regex
	now := time.Now()
	metaFilePath := filepath.Join(tempDir, "world_meta.xml")
	if _, err := os.Stat(metaFilePath); err == nil {
		// Read world_meta.xml
		data, err := os.ReadFile(metaFilePath)
		if err != nil {
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to read world_meta.xml: %w", err)
		}

		// Calculate Windows file time
		const windowsEpochToUnixEpoch = 116444736000000000 // 100-ns intervals from 1601 to 1970
		windowsFileTime := now.UnixNano()/100 + windowsEpochToUnixEpoch

		re, err := regexp.Compile(`<DateTime>\d+</DateTime>`)
		if err != nil {
			m.revertRestore(restoredFiles)
			return fmt.Errorf("failed to compile DateTime regex: %w", err)
		}
		newDateTime := fmt.Sprintf("<DateTime>%d</DateTime>", windowsFileTime)
		updatedData := re.ReplaceAll(data, []byte(newDateTime))

		if !re.Match(data) {
			logger.Backup.Warn("Restore: DateTime element not found in world_meta.xml, proceeding without updating. Server might not load correct save.")
		} else {
			if err := os.WriteFile(metaFilePath, updatedData, 0644); err != nil {
				m.revertRestore(restoredFiles)
				return fmt.Errorf("failed to write updated world_meta.xml: %w", err)
			}
		}
	} else {
		logger.Backup.Warn("world_meta.xml not found in extracted files, proceeding without updating DateTime")
	}

	// Modify timestamps of extracted files to current system time
	if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return os.Chtimes(path, now, now)
	}); err != nil {
		m.revertRestore(restoredFiles)
		return fmt.Errorf("failed to modify timestamps in %s: %w", tempDir, err)
	}

	// Create new .save (zip) file at destFile with updated timestamps
	dest, err := os.Create(destFile)
	if err != nil {
		m.revertRestore(restoredFiles)
		return fmt.Errorf("failed to create destination .save file %s: %w", destFile, err)
	}
	defer dest.Close()

	w := zip.NewWriter(dest)
	defer w.Close()

	if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		// Create zip entry with current system timestamp
		fw, err := w.CreateHeader(&zip.FileHeader{
			Name:     relPath,
			Method:   zip.Deflate,
			Modified: now,
		})
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", relPath, err)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(fw, srcFile); err != nil {
			return fmt.Errorf("failed to write file %s to zip: %w", relPath, err)
		}
		return nil
	}); err != nil {
		m.revertRestore(restoredFiles)
		return fmt.Errorf("failed to restore .save file %s: %w", backupFile, err)
	}
	restoredFiles[destFile] = backupFile
	logger.Backup.Debug(fmt.Sprintf("%v", restoredFiles))
	return nil // restore and mod time shenanigans successful, no need to return an error
}

// revertRestore undoes a failed restore operation
func (m *BackupManager) revertRestore(restoredFiles map[string]string) {
	for destFile, backupFile := range restoredFiles {
		if err := os.Remove(destFile); err == nil {
			_ = copyFile(backupFile, destFile)
		}
	}
}
