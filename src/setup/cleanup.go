package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

func CleanUpOldUIModFolderFiles() error {
	uiModFolder := config.GetUIModFolder()
	customdetectionsSourceFile := filepath.Join(uiModFolder, "detectionmanager", "customdetections.json")
	customdetectionsDestinationFile := config.GetCustomDetectionsFilePath()
	oldUiFolder := filepath.Join(uiModFolder, "ui") // used to test if we need clean up from a structure before v5.5 (since we now have embedded assets)

	//if uiModFolder doesn't contain a folder called UI, return early as there is nothing to clean up
	if _, err := os.Stat(oldUiFolder); os.IsNotExist(err) {
		return nil
	}

	// Copy customdetections.json to the destination path
	if _, err := os.Stat(customdetectionsSourceFile); err == nil {
		// Ensure destination directory exists
		destDir := filepath.Dir(customdetectionsDestinationFile)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Read source file
		data, err := os.ReadFile(customdetectionsSourceFile)
		if err != nil {
			return fmt.Errorf("failed to read source file: %w", err)
		}

		// Write to destination file
		if err := os.WriteFile(customdetectionsDestinationFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write destination file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		logger.Core.Error("Error moving customdetections.json file to new location: " + err.Error())
	}

	// List of folders to remove
	foldersToRemove := []string{
		filepath.Join(uiModFolder, "detectionmanager"),
		filepath.Join(uiModFolder, "ui"),
		filepath.Join(uiModFolder, "twoboxform"),
		filepath.Join(uiModFolder, "assets"),
	}

	// Remove specified folders if they exist
	for _, folder := range foldersToRemove {
		if _, err := os.Stat(folder); err == nil {
			if err := os.RemoveAll(folder); err != nil {
				return fmt.Errorf("failed to remove folder %s: %w", folder, err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking folder %s: %w", folder, err)
		}
	}

	return nil
}

func CleanUpOldExecutables() error {
	// Exit early if update is disabled to allow running old versions if needed
	if !config.GetIsUpdateEnabled() || config.GetIsDebugMode() {
		return nil
	}
	currentBackendVersion := config.GetVersion()
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^StationeersServerUI_v(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)_(?:windows|linux)_[A-Za-z0-9]+(?:\.exe)?$`),
		regexp.MustCompile(`^StationeersServerControlv(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)(?:\.exe|\.x86_64)$`),
	}

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Read only the root directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Process each entry in the root directory
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), "_old") {
			continue
		}

		fileVersion := ""
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(info.Name()); len(matches) == 2 {
				fileVersion = matches[1]
				break
			}
		}
		if fileVersion == "" {
			continue
		}

		// Skip if the version matches the current backend version
		if fileVersion == currentBackendVersion {
			continue
		}

		// Generate new filename with _old prefix
		newName := "_old" + info.Name()
		newPath := filepath.Join(dir, newName)

		// Rename the file
		path := filepath.Join(dir, info.Name())
		err = os.Rename(path, newPath)
		if err != nil {
			return fmt.Errorf("failed to rename %s to %s: %w", path, newName, err)
		}
		logger.Install.Info(fmt.Sprintf("Old Executable cleanup: Renamed %s to %s", path, newName))
	}

	return nil
}
