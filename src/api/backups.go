package api

import (
	"StationeersServerUI/src/discord"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func ListBackups(w http.ResponseWriter, r *http.Request) {
	config, err := loadConfig()
	if err != nil {
		http.Error(w, "Error loading config", http.StatusInternalServerError)
		return
	}

	// Read from the Safebackups folder
	safeBackupDir := "./saves/" + config.SaveFileName + "/Safebackups"
	files, err := os.ReadDir(safeBackupDir)
	if err != nil {
		http.Error(w, "Unable to read Safebackups directory", http.StatusInternalServerError)
		return
	}

	backupDetails := make(map[int]time.Time)
	backupIndices := make(map[int]bool)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		backupIndex := parseBackupIndex(fileName)
		if backupIndex != -1 {
			if !backupIndices[backupIndex] {
				backupIndices[backupIndex] = true
				fullPath := filepath.Join(safeBackupDir, fileName)
				info, err := os.Stat(fullPath)
				if err != nil {
					http.Error(w, "Error getting file info", http.StatusInternalServerError)
					return
				}
				backupDetails[backupIndex] = info.ModTime()
			}
		}
	}

	var sortedBackups []struct {
		index   int
		modTime time.Time
	}

	for index, modTime := range backupDetails {
		sortedBackups = append(sortedBackups, struct {
			index   int
			modTime time.Time
		}{index, modTime})
	}

	sort.Slice(sortedBackups, func(i, j int) bool {
		return sortedBackups[i].index > sortedBackups[j].index // Sort by descending index
	})

	var output []string
	for _, backup := range sortedBackups {
		creationTime := backup.modTime.Format("02.01.2006 15:04:05")
		output = append(output, fmt.Sprintf("BackupIndex: %d, Created: %s", backup.index, creationTime))
	}

	if len(output) == 0 {
		fmt.Fprint(w, "No valid backup files found. Is the directory specified?")
		return
	}

	response := strings.Join(output, "\n")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, response)
}

// Helper function to sort backup details by index
func SortedKeys(backupDetails map[int]time.Time) []struct {
	index   int
	modTime time.Time
} {
	var sorted []struct {
		index   int
		modTime time.Time
	}
	for index, modTime := range backupDetails {
		sorted = append(sorted, struct {
			index   int
			modTime time.Time
		}{index, modTime})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].index < sorted[j].index
	})
	return sorted
}

func parseBackupIndex(fileName string) int {
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) > 1 {
		index, err := strconv.Atoi(matches[1])
		if err == nil {
			return index
		}
	}
	return -1
}

func RestoreBackup(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "Index parameter is required", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index parameter", http.StatusBadRequest)
		return
	}

	config, err := loadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	// Use the Safebackups folder for restoring
	safeBackupDir := "./saves/" + config.SaveFileName + "/Safebackups"
	saveDir := "./saves/" + config.SaveFileName
	files := []struct {
		backupName    string
		backupNameAlt string // Alternative name with _AutoSave
		destName      string
	}{
		{fmt.Sprintf("world_meta(%d).xml", index), fmt.Sprintf("world_meta(%d)_AutoSave.xml", index), "world_meta.xml"},
		{fmt.Sprintf("world(%d).xml", index), fmt.Sprintf("world(%d)_AutoSave.xml", index), "world.xml"},
		{fmt.Sprintf("world(%d).bin", index), fmt.Sprintf("world(%d)_AutoSave.bin", index), "world.bin"},
	}

	// Create a map to store successful restore operations
	restoredFiles := make(map[string]string)

	// First, try to restore all files
	for _, file := range files {
		backupFile := filepath.Join(safeBackupDir, file.backupName)
		destFile := filepath.Join(saveDir, file.destName)

		err := copyFile(backupFile, destFile)
		if err != nil {
			// Try alternative file name with _AutoSave suffix
			backupFileAlt := filepath.Join(safeBackupDir, file.backupNameAlt)
			errAlt := copyFile(backupFileAlt, destFile)
			if errAlt != nil {
				// Revert any successful operations if an error occurs
				revertRestore(restoredFiles, saveDir, safeBackupDir)
				http.Error(w, fmt.Sprintf("Error restoring file %s and %s: %v", file.backupName, file.backupNameAlt, err), http.StatusInternalServerError)
				return
			}
			backupFile = backupFileAlt
		}
		restoredFiles[destFile] = backupFile
	}

	fmt.Fprintf(w, "Backup %d restored successfully.", index)
}

// copyFile copies a file from src to dst. If dst already exists, it will be overwritten. This is the inteded behavior of the restoreBackup function. We overwrite the destination files with the backup files
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// revertRestore reverts the file restore operation if an error occurs
func revertRestore(restoredFiles map[string]string, saveDir, backupDir string) {
	for destFile, backupFile := range restoredFiles {
		err := os.Remove(destFile)
		if err != nil {
			fmt.Printf("Error removing file %s: %v\n", destFile, err)
		} else {
			err = copyFile(backupFile, destFile)
			if err != nil {
				fmt.Printf("Error restoring file %s: %v\n", destFile, err)
			}
		}
	}
}

func watchBackupDir() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	backupDir := "./saves/" + config.SaveFileName + "/backup"
	safeBackupDir := "./saves/" + config.SaveFileName + "/Safebackups"

	// Ensure the safe backup directory exists
	if err := os.MkdirAll(safeBackupDir, os.ModePerm); err != nil {
		fmt.Println("Error creating safe backup directory:", err)
		return
	}

	err = watcher.Add(backupDir)
	if err != nil {
		fmt.Println("Error watching backup directory:", backupDir, err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("New backup file detected:", event.Name)
				go copyBackupToSafeLocation(event.Name, safeBackupDir)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error watching backup directory:", err)
		}
	}
}

// Copy the detected backup file to a safe location
func copyBackupToSafeLocation(srcFilePath string, safeBackupDir string) {
	go func() {
		// Introduce a 1-minute asynchronous wait before trying to copy the file
		time.Sleep(30 * time.Second)

		fileName := filepath.Base(srcFilePath)
		dstFilePath := filepath.Join(safeBackupDir, fileName)

		// Read the file content without explicitly opening it
		data, err := os.ReadFile(srcFilePath)
		if err != nil {
			fmt.Println("Error reading backup file:", err)
			return
		}

		// Write the file content to the destination
		err = os.WriteFile(dstFilePath, data, 0644)
		if err != nil {
			fmt.Println("Error copying backup to safe location:", err)
			return
		}

		fmt.Println("Backup successfully copied to safe location:", dstFilePath)
		discord.SendMessageToSavesChannel(fmt.Sprintf("Backup file %s copied to safe location.", dstFilePath))
	}()
}

func CleanUpBackups(backupDir, safeBackupDir string) {
	// Cleanup the backup folder
	err := cleanBackupFolder(backupDir, time.Hour*24) // Only retain backups from the current day in the backup folder
	if err != nil {
		fmt.Printf("Error cleaning backup folder: %v\n", err)
	}

	// Cleanup the Safebackups folder with custom retention rules
	err = cleanSafebackupsFolder(safeBackupDir)
	if err != nil {
		fmt.Printf("Error cleaning Safebackups folder: %v\n", err)
	}
}

// Cleanup the backup folder, keeping only files from the current day
func cleanBackupFolder(backupDir string, maxAge time.Duration) error {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fullPath := filepath.Join(backupDir, file.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		// Delete files older than 24 hours (current day)
		if now.Sub(info.ModTime()) > maxAge {
			err = os.Remove(fullPath)
			if err != nil {
				fmt.Printf("Error removing file %s: %v\n", fullPath, err)
			}
		}
	}
	return nil
}

// Cleanup the Safebackups folder according to your custom retention rules
func cleanSafebackupsFolder(safeBackupDir string) error {
	files, err := os.ReadDir(safeBackupDir)
	if err != nil {
		return err
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}

	var backups []fileInfo
	now := time.Now()

	// Collect file information
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fullPath := filepath.Join(safeBackupDir, file.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			return err
		}
		backups = append(backups, fileInfo{name: file.Name(), modTime: info.ModTime()})
	}

	// Sort backups by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	// Tracking the last kept backups for each retention period
	var lastKept15Min, lastKeptHour, lastKeptDay time.Time

	for _, backup := range backups {
		age := now.Sub(backup.modTime)

		// Keep backups younger than 24 hours
		if age < time.Hour*24 {
			continue
		}

		// Retain backups older than 24 hours but within 48 hours every 15 minutes
		if age < time.Hour*48 {
			if lastKept15Min.IsZero() || backup.modTime.Sub(lastKept15Min) > time.Minute*15 {
				lastKept15Min = backup.modTime
				continue
			}
		}

		// Retain backups older than 48 hours but younger than 7 days every hour
		if age < time.Hour*24*7 {
			if lastKeptHour.IsZero() || backup.modTime.Sub(lastKeptHour) > time.Hour {
				lastKeptHour = backup.modTime
				continue
			}
		}

		// Retain one backup per day for backups older than 7 days
		if age >= time.Hour*24*7 {
			if lastKeptDay.IsZero() || backup.modTime.Sub(lastKeptDay) > time.Hour*24 {
				lastKeptDay = backup.modTime
				continue
			}
		}

		// If the file doesn't meet any retention criteria, delete it
		fullPath := filepath.Join(safeBackupDir, backup.name)
		err := os.Remove(fullPath)
		if err != nil {
			fmt.Printf("Error removing backup file %s: %v\n", fullPath, err)
		} else {
			fmt.Printf("Removed old backup file: %s\n", fullPath)
		}
	}

	return nil
}

func StartBackupCleanupRoutine(backupDir, safeBackupDir string) {
	ticker := time.NewTicker(24 * time.Hour) // Run cleanup every 24 hours
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Starting backup cleanup...")
			CleanUpBackups(backupDir, safeBackupDir)
		}
	}
}