package modding

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// ProcessModPackageUpload handles the upload and extraction of a mod package zip file
func ProcessModPackageUpload(r io.Reader) error {
	logger.Modding.Info("Starting mod package upload process")

	const maxZipSize = 500 * 1024 * 1024 // 500 MB
	sizeLimitedReader := io.LimitReader(r, maxZipSize+1)

	// Read the entire zip file into memory
	zipBytes, err := io.ReadAll(sizeLimitedReader)
	if err != nil {
		logger.Modding.Errorf("Failed to read mod package zip file, filesize might exceed 500mb: %v", err)
		return fmt.Errorf("failed to read mod package zip file, filesize might exceed 500mb: %w", err)
	}

	if len(zipBytes) == 0 {
		logger.Modding.Error("Received empty mod package")
		return fmt.Errorf("mod package is empty")
	}

	logger.Modding.Debugf("Received Modpackage: %d bytes", len(zipBytes))

	// Create temporary file with timestamp
	timestamp := time.Now().Unix()
	tempFilename := fmt.Sprintf("tmp-uploaded-modpackage-%d.zip", timestamp)
	tempFilepath := filepath.Join(".", tempFilename)

	logger.Modding.Debugf("Saving temporary mod package: %s", tempFilename)
	if err := os.WriteFile(tempFilepath, zipBytes, 0644); err != nil {
		logger.Modding.Errorf("Failed to write temporary mod package: %v", err)
		return fmt.Errorf("failed to save temporary mod package: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFilepath); err != nil && !os.IsNotExist(err) {
			logger.Modding.Warnf("Failed to clean up temporary mod package: %v", err)
		}
	}()

	// Clear ./mods directory if it exists
	modsDir := filepath.Join(".", "mods")
	if err := clearDirectory(modsDir); err != nil {
		logger.Modding.Warnf("Ran into an issue while clearing mods directory: %v", err)
	}

	// Remove modconfig.xml if it exists
	modconfigPath := filepath.Join(".", "modconfig.xml")
	if err := os.Remove(modconfigPath); err != nil && !os.IsNotExist(err) {
		logger.Modding.Warnf("Failed to remove existing modconfig.xml: %v", err)
	}
	if !os.IsNotExist(err) {
		logger.Modding.Debug("Removed existing modconfig.xml")
	}

	// Extract the zip file to current working directory
	if err := extractZip(tempFilepath, "."); err != nil {
		logger.Modding.Errorf("Failed to extract mod package: %v", err)
		return fmt.Errorf("failed to extract mod package: %w", err)
	}

	// Call ImportModPackage with the zip bytes
	if err := ImportModPackage(zipBytes); err != nil {
		logger.Modding.Errorf("ImportModPackage failed: %v", err)
		return fmt.Errorf("import mod package failed: %w", err)
	}

	logger.Modding.Info("Mod package upload process completed successfully")
	return nil
}

// clearDirectory removes all files and subdirectories in a directory
func clearDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to clear
		}
		return err
	}

	logger.Modding.Debugf("Clearing directory: %s", dirPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	logger.Modding.Debug("Directory cleared successfully")
	return nil
}

// extractZip extracts all files from a zip archive to the destination directory
func extractZip(zipPath string, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	logger.Modding.Debugf("Extracting %d files from zip", len(reader.File))

	for i, file := range reader.File {
		filePath := filepath.Join(destDir, file.Name)

		// Prevent path traversal attacks
		if !filepath.IsLocal(filepath.Join(filepath.Dir(filePath), filepath.Base(filePath))) {
			return fmt.Errorf("invalid file path in archive: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in archive: %w", err)
			}

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				rc.Close()
				return fmt.Errorf("failed to create output file: %w", err)
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				rc.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}

			outFile.Close()
			rc.Close()

			if (i+1)%10 == 0 || i == len(reader.File)-1 {
				logger.Modding.Debugf("Extracted %d/%d files", i+1, len(reader.File))
			}
		}
	}

	return nil
}

func ImportModPackage(zipData []byte) error {
	if len(zipData) == 0 {
		return fmt.Errorf("empty zip data")
	}

	// Validate it's a valid zip file
	if _, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData))); err != nil {
		return fmt.Errorf("invalid zip file: %w", err)
	}

	return nil
}
