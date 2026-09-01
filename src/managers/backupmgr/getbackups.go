package backupmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// getBackupSaveFiles retrieves all backup save files from the safe backup directory
func (m *BackupManager) getBackupSaveFiles() ([]BackupSaveFile, error) {
	var saves []BackupSaveFile

	err := filepath.WalkDir(m.config.SafeBackupDir, func(path string, de os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() {
			// Process the save file
			filename := de.Name()

			// Skip invalid backup files
			if !isValidBackupFile(filename) {
				return nil
			}

			// Get the full path
			fullPath := filepath.Join(m.config.SafeBackupDir, filename)

			summary, err := ReadSaveSummary(fullPath)
			if err != nil {
				logger.Backup.Warnf("Skipping invalid backup file %s: %s", fullPath, err.Error())
				return nil
			}

			// Add the backup save file info to the list
			saves = append(saves, BackupSaveFile{
				SaveFile: fullPath,
				SaveTime: summary.SavedAt,
				Summary:  summary,
			})
		}
		return nil
	})

	// Handle errors
	if err != nil {
		// if the error contains no such file or directory, return nil but return a custom string intsted 	of the error
		if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "The system cannot find the file specified") {
			return nil, fmt.Errorf("save dir doesn't seem to exist (yet). Try starting the gameserver and click ↻ once it's up. If the Save folder exists and you still get this error, verify the 'Use New Terrain and Save System' setting. Detailed Error: %w", err)
		}
		return nil, fmt.Errorf("failed to handle safe backup dir: %w", err)
	}

	// Sort saves by save time ascending
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].SaveTime.Before(saves[j].SaveTime)
	})
	// Add the index to each save
	for i := range saves {
		saves[i].Index = i
	}

	return saves, nil
}
