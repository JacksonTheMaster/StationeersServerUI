package setup

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/steamcmd"
)

// BepInEx version: 5.4.23.2 or v5-lts
// SSCM version: 1.0.0

var installMutex sync.Mutex

func CheckAndDownloadSSCM() {
	SSCMPluginDir := config.GetSSCMPluginDir()
	sscmDir := config.GetSSCMWebDir()

	requiredDirs := []string{SSCMPluginDir, sscmDir}

	// Set branch
	if config.GetBranch() == "release" || config.GetBranch() == "Release" {
		downloadBranch = "main"
	} else {
		downloadBranch = config.GetBranch()
	}
	logger.Install.Debug("Using branch for SSCM: " + downloadBranch)

	// Define file mappings
	files := map[string]string{
		SSCMPluginDir + "SSCM.dll": fmt.Sprintf("https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/%s/sscm/SSCM.dll", downloadBranch),
	}

	// Check if the directory exists
	if _, err := os.Stat(SSCMPluginDir); os.IsNotExist(err) {
		logger.Install.Warn("⚠️SSCM dir does not exist. Creating it...")

		// Create directories
		for _, dir := range requiredDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					logger.Install.Error("❌Error creating folder: " + err.Error())
					return
				}
				logger.Install.Warn("⚠️Created folder: " + dir)
			}
		}

		// Initial download
		config.SetIsFirstTimeSetup(true)
		downloadAllFiles(files)
	} else {
		// Directory exists
		config.SetIsFirstTimeSetup(false)
		logger.Install.Debug(fmt.Sprintf("IsUpdateEnabled: %v", config.GetIsUpdateEnabled()))
		logger.Install.Debug(fmt.Sprintf("IsFirstTimeSetup: %v", config.GetIsFirstTimeSetup()))
		if config.GetIsUpdateEnabled() {
			logger.Install.Info("🔍Validating SSCM files for updates...")
			if config.GetBranch() == "release" || config.GetBranch() == "Release" {
				downloadBranch = "main"
				updateFilesIfDifferent(files)
			} else {
				downloadBranch = config.GetBranch()
				updateFilesIfDifferent(files)
			}
		} else {
			logger.Install.Info("♻️Folder SSCM already exists. Updates disabled, skipping validation.")
		}
	}
}

func CheckAndInstallBepInEx() error {
	// Ensure thread safety
	installMutex.Lock()
	defer installMutex.Unlock()

	logger.Install.Info("Checking for BepInEx installation...")

	// Check if BepInEx is already installed
	if _, err := os.Stat("BepInEx"); err == nil {
		logger.Install.Info("BepInEx folder already exists, skipping installation")
		return nil
	}

	// Determine the URL based on platform
	var url string
	if runtime.GOOS == "windows" {
		url = "https://github.com/BepInEx/BepInEx/releases/download/v5.4.23.2/BepInEx_win_x64_5.4.23.2.zip"
		logger.Install.Info("Detected Windows platform, using Windows BepInEx package")
	} else {
		url = "https://github.com/BepInEx/BepInEx/releases/download/v5.4.23.2/BepInEx_linux_x64_5.4.23.2.zip"
		logger.Install.Info("Detected non-Windows platform, using Linux BepInEx package")
	}

	// Download and install BepInEx
	if err := downloadAndInstallBepInEx(url); err != nil {
		logger.Install.Error(fmt.Sprintf("❌Failed to install BepInEx: %v", err))
		return err
	}

	logger.Install.Info("✅BepInEx installed successfully")
	return nil
}

// downloadAndInstallBepInEx downloads the BepInEx zip and extracts it to the current directory
func downloadAndInstallBepInEx(url string) error {
	// Create a temporary file to store the downloaded zip
	tempFile, err := os.CreateTemp("", "bepinex_*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the zip file when done

	// Download the BepInEx zip file
	logger.Install.Info("📥Downloading BepInEx from: " + url)
	client := &http.Client{Timeout: 45 * time.Second}
	err = downloadFileWithClient(client, tempFile.Name(), url)
	if err != nil {
		return fmt.Errorf("failed to download BepInEx: %w", err)
	}

	// Get file info for the zip
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Reopen the file for reading
	tempFile.Close()
	zipFile, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open downloaded zip: %w", err)
	}
	defer zipFile.Close()

	// Extract the zip file to the current directory
	logger.Install.Info("📦Extracting BepInEx to current directory")
	err = steamcmd.Unzip(zipFile, fileInfo.Size(), ".")
	if err != nil {
		return fmt.Errorf("failed to extract BepInEx: %w", err)
	}

	// Clean up changelog.txt if it exists
	if _, err := os.Stat("changelog.txt"); err == nil {
		logger.Install.Debug("🗑️Removing changelog.txt")
		if err := os.Remove("changelog.txt"); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️Failed to remove changelog.txt: %v", err))
		}
	}

	if runtime.GOOS == "linux" {
		if _, err := os.Stat("./run_bepinex.sh"); err == nil {
			err = os.Remove("./run_bepinex.sh")
			if err != nil {
				logger.Install.Warn(fmt.Sprintf("⚠️Failed to remove run_bepinex.sh: %v", err))
			}
		}
	}

	return nil
}

func InstallSSCM() {
	logger.Install.Info("🕑Installing SSCM...")

	if err := CheckAndInstallBepInEx(); err != nil {
		config.SetIsSSCMEnabled(false)
		logger.Install.Warn("⚠️SSCM was disabled because BepInEx could not be installed. SSUI will continue without SSCM.")
		return
	}
	CheckAndDownloadSSCM()

	// Enable SSCM
	config.SetIsSSCMEnabled(true)

	logger.Install.Info("✅SSCM enabled")
}
