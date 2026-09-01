package steamcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/modding"
)

// DownloadWorkshopItems downloads all installed workshop mods using SteamCMD
func UpdateWorkshopItems() ([]string, error) {
	var logs []string
	workshopHandles := modding.GetModWorkshopHandles()
	if len(workshopHandles) == 0 {
		logger.Install.Debug("ℹ️  No workshop items to download")
		logs = append(logs, "No workshop items to download")
		return logs, nil
	}

	//logger.Install.Debugf("Workshop handles to update: %v", workshopHandles)
	logs2, err := DownloadWorkshopItems(workshopHandles)
	logs = append(logs, logs2...)
	return logs, err
}

func DownloadWorkshopItems(workshopHandles []string) ([]string, error) {
	var logs []string
	validatedHandles := make([]string, 0, len(workshopHandles))
	for _, handle := range workshopHandles {
		workshopID, err := strconv.ParseUint(handle, 10, 64)
		if err != nil || workshopID == 0 {
			return logs, fmt.Errorf("invalid Steam Workshop ID %q: must be a positive integer", handle)
		}
		validatedHandles = append(validatedHandles, strconv.FormatUint(workshopID, 10))
	}
	workshopHandles = validatedHandles

	logger.Install.Infof("🔄 Downloading %d workshop items...", len(workshopHandles))
	logs = append(logs, fmt.Sprintf("Downloading %d workshop items...", len(workshopHandles)))

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Install.Error("❌ Error getting current working directory: " + err.Error())
		logs = append(logs, "Error getting current working directory: "+err.Error())
		return logs, err
	}

	// Acquire lock for SteamCMD access
	if steamMu.TryLock() {
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD Workshop Downloads...")
	} else {
		logger.Core.Warn("🔄 SteamMu is currently locked, waiting for it to be unlocked and then continuing...")
		steamMu.Lock()
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD Workshop Downloads...")
	}
	defer func() {
		steamMu.Unlock()
		logger.Core.Debug("🔄 Unlocking SteamMu after SteamCMD Workshop Downloads...")
	}()

	steamcmddir := SteamCMDLinuxDir
	executable := "steamcmd.sh"

	if runtime.GOOS == "windows" {
		executable = "steamcmd.exe"
		steamcmddir = SteamCMDWindowsDir
	}

	steamcmdPath := filepath.Join(steamcmddir, executable)
	if _, err := os.Stat(steamcmdPath); err != nil {
		err := fmt.Errorf("SteamCMD executable not found at %s (is SteamCMD disabled?)", steamcmdPath)
		logger.Install.Error("❌ " + err.Error())
		logs = append(logs, err.Error())
		return logs, err
	}

	// Download each workshop item
	var downloadedHandles []string
	var failedHandles []string
	for i, appID := range workshopHandles {
		logger.Install.Infof("📦 Downloading workshop item %d/%d: %s", i+1, len(workshopHandles), appID)

		// Build SteamCMD command
		cmd := exec.Command(
			steamcmdPath,
			"+force_install_dir", "../",
			"+login", "anonymous",
			"+workshop_download_item", "544550", appID,
			"validate",
			"+quit",
		)

		// Capture output
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Set up environment for Linux
		if runtime.GOOS == "linux" {
			env := os.Environ()
			newEnv := make([]string, 0, len(env)+1)
			foundHome := false
			for _, e := range env {
				if !strings.HasPrefix(e, "HOME=") {
					newEnv = append(newEnv, e)
				} else {
					newEnv = append(newEnv, "HOME="+currentDir)
					foundHome = true
				}
			}
			if !foundHome {
				newEnv = append(newEnv, "HOME="+currentDir)
			}
			cmd.Env = newEnv
		}

		// Run the command
		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				logger.Install.Warnf("⚠️  SteamCMD workshop download failed for %s (code %d): %s", appID, exitErr.ExitCode(), stderr.String())
				logs = append(logs, fmt.Sprintf("SteamCMD workshop download failed for %s (code %d): %s", appID, exitErr.ExitCode(), stderr.String()))
			} else {
				logger.Install.Warnf("⚠️  Error running SteamCMD for workshop item %s: %s", appID, err.Error())
				logs = append(logs, fmt.Sprintf("Error running SteamCMD for workshop item %s: %s", appID, err.Error()))
			}
			failedHandles = append(failedHandles, appID)
			continue // Continue with next workshop item even if this one fails
		}

		logger.Install.Debugf("✅ Successfully downloaded workshop item: %s", appID)
		logs = append(logs, fmt.Sprintf("Successfully downloaded workshop item: %s", appID))
		downloadedHandles = append(downloadedHandles, appID)
	}

	if len(failedHandles) > 0 {
		logger.Install.Warn("⚠️ Workshop items download completed with failures")
		logs = append(logs, "Workshop items download completed with failures")
	} else {
		logger.Install.Info("✅ Workshop items download complete")
		logs = append(logs, "Workshop items download complete")
	}

	// Copy only items whose SteamCMD download succeeded.
	if len(downloadedHandles) > 0 {
		logs2, err := copyDownloadedItemsToMods(downloadedHandles)
		logs = append(logs, logs2...)
		if err != nil {
			logger.Install.Error("❌ Error copying workshop items to mods directory: " + err.Error())
			logs = append(logs, "Error copying workshop items to mods directory: "+err.Error())
			return logs, err
		}
	}

	if len(failedHandles) > 0 {
		err := fmt.Errorf("failed to download %d workshop item(s): %s", len(failedHandles), strings.Join(failedHandles, ", "))
		logger.Install.Error("❌ " + err.Error())
		logs = append(logs, err.Error())
		return logs, err
	}

	return logs, nil
}

// copyDownloadedItemsToMods copies downloaded workshop items from the Steam directory to ./mods
func copyDownloadedItemsToMods(workshopHandles []string) ([]string, error) {
	var logs []string
	var missingHandles []string
	// Determine the steam content directory based on OS
	var steamContentDir string
	if runtime.GOOS == "windows" {
		steamContentDir = SteamCMDWindowsDir
		// Windows SteamCMD dir is C:\SteamCMD, so workshop content is at C:\SteamCMD\steamapps\workshop\content\544550
		steamContentDir = filepath.Join(steamContentDir, "steamapps", "workshop", "content", "544550")
	} else {
		// Linux: ./steamapps/workshop/content/544550
		steamContentDir = filepath.Join(".", "steamapps", "workshop", "content", "544550")
	}

	// Ensure mods directory exists
	modsDir := "./mods"
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return logs, fmt.Errorf("failed to create mods directory: %w", err)
	}

	logger.Install.Infof("📂 Copying %d workshop items to mods directory...", len(workshopHandles))
	logs = append(logs, fmt.Sprintf("Copying %d workshop items to mods directory...", len(workshopHandles)))

	// Copy each workshop item
	for i, appID := range workshopHandles {
		logger.Install.Infof("📋 Processing workshop item %d/%d: %s", i+1, len(workshopHandles), appID)

		// Source path: steamapps/workshop/content/544550/{appID}
		srcPath := filepath.Join(steamContentDir, appID)

		// Check if source directory exists
		srcInfo, err := os.Stat(srcPath)
		if err != nil || !srcInfo.IsDir() {
			logger.Install.Errorf("❌ Workshop item not found at expected path: %s (skipping)", srcPath)
			logs = append(logs, fmt.Sprintf("Workshop item not found at expected path: %s (skipping)", srcPath))
			missingHandles = append(missingHandles, appID)
			continue
		}

		// Destination path: ./mods/Workshop_{appID}
		destPath := filepath.Join(modsDir, fmt.Sprintf("Workshop_%s", appID))

		// Remove existing destination directory if it exists
		if _, err := os.Stat(destPath); err == nil {
			logger.Install.Debugf("🗑️  Removing existing directory: %s", destPath)
			if err := os.RemoveAll(destPath); err != nil {
				logger.Install.Warnf("⚠️  Failed to remove existing directory %s: %s (continuing anyway)", destPath, err.Error())
				logs = append(logs, fmt.Sprintf("Failed to remove existing directory %s: %s (continuing anyway)", destPath, err.Error()))
			}
		}

		// Copy the entire directory
		if err := copyDir(srcPath, destPath); err != nil {
			logger.Install.Warnf("⚠️  Failed to copy workshop item %s: %s (skipping)", appID, err.Error())
			logs = append(logs, fmt.Sprintf("Failed to copy workshop item %s: %s (skipping)", appID, err.Error()))
			continue
		}

		logger.Install.Debugf("✅ Successfully copied workshop item to: %s", destPath)
		logs = append(logs, fmt.Sprintf("Successfully copied workshop item to: %s", destPath))
	}

	logger.Install.Info("✅ Workshop items copy complete")
	logs = append(logs, "Workshop items copy complete")
	if len(missingHandles) > 0 {
		return logs, fmt.Errorf("workshop item(s) missing after download: %s", strings.Join(missingHandles, ", "))
	}
	return logs, nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
