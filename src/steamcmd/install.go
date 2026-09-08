package steamcmd

import (
	"os"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

func installSteamCMD(platform string, steamCMDDir string, downloadURL string, extractFunc ExtractorFunc) (int, error) {
	// Check if SteamCMD is already installed
	if _, err := os.Stat(steamCMDDir); os.IsNotExist(err) {
		logger.Install.Warn("⚠️ SteamCMD not found for " + platform + ", downloading...\n")

		// Create SteamCMD directory
		if err := createSteamCMDDirectory(steamCMDDir); err != nil {
			logger.Install.Error("❌ Error creating SteamCMD directory: " + err.Error() + "\n")
			return -1, err
		}

		// Ensure cleanup on failure
		success := false
		defer func() {
			if !success {
				logger.Install.Warn("⚠️ Cleaning up due to failure...\n")
				os.RemoveAll(steamCMDDir)
			}
		}()

		// Install required libraries
		if err := installRequiredLibraries(); err != nil {
			logger.Install.Error("❌ Error installing required libraries: " + err.Error() + "\n")
			return -1, err
		}

		// Download and extract SteamCMD
		if err := downloadAndExtractSteamCMD(downloadURL, steamCMDDir, extractFunc); err != nil {
			logger.Install.Error("❌ " + err.Error() + "\n")
			return -1, err
		}

		// Set executable permissions for SteamCMD files
		if err := setExecutablePermissions(steamCMDDir); err != nil {
			logger.Install.Error("❌ Error setting executable permissions: " + err.Error() + "\n")
			return -1, err
		}

		// Verify the steamcmd binary
		if err := verifySteamCMDBinary(steamCMDDir); err != nil {
			logger.Install.Error("❌ " + err.Error() + "\n")
			return -1, err
		}

		// Mark installation as successful
		success = true
		logger.Install.Info("✅ SteamCMD installed successfully.\n")
	} else {
		logger.Install.Info("✅ SteamCMD is already installed.")
	}

	// Run SteamCMD and return its exit status and error
	return runSteamCMD(steamCMDDir)
}

// installSteamCMDLinux downloads and installs SteamCMD on Linux.
func installSteamCMDLinux() (int, error) {
	return installSteamCMD("Linux", SteamCMDLinuxDir, SteamCMDLinuxURL, untarWrapper)
}

// installSteamCMDWindows downloads and installs SteamCMD on Windows.
func installSteamCMDWindows() (int, error) {
	return installSteamCMD("Windows", SteamCMDWindowsDir, SteamCMDWindowsURL, Unzip)
}
