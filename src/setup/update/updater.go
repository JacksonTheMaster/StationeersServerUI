package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
)

// githubRelease represents the structure of a GitHub release response
type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	Assets     []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Version holds semantic version components
type Version struct {
	Major int
	Minor int
	Patch int
}

func CheckForUpdates() (err error, newVersion string) {
	return update(false, false, "")
}

func ApplyAvailableUpdate() (err error, newVersion string) {
	return update(true, false, "")
}

// ApplyVersion installs the release an administrator saw and approved in the UI.
// A one-time major approval does not change the persistent AllowMajorUpdates setting.
func ApplyVersion(version string, majorUpdateApproved bool) (err error, newVersion string) {
	return update(true, majorUpdateApproved, version)
}

func update(isInUpdateableState, majorUpdateApproved bool, expectedVersion string) (err error, newVersion string) {
	if !config.GetIsUpdateEnabled() {
		logger.Install.Warn("⚠️ Update check is disabled. Skipping update check. Change 'IsUpdateEnabled' in config.json to true to re-enable update checks.")
		return nil, ""
	}

	if config.GetBranch() != "release" {
		logger.Install.Warn("⚠️ You are running a development build. Skipping update check.")
		return nil, ""
	}

	if config.GetAllowPrereleaseUpdates() {
		logger.Install.Info("🕵️ Querying GitHub API for the latest (pre)release...")
	} else {
		logger.Install.Info("🕵️ Querying GitHub API for the latest stable release...")
	}
	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch latest release: %v", err), ""
	}
	if expectedVersion != "" && latestRelease.TagName != expectedVersion {
		return fmt.Errorf("available update changed from %s to %s; review it before installing", expectedVersion, latestRelease.TagName), latestRelease.TagName
	}

	// Parse current and latest versions
	currentVer, err := parseVersion(config.GetVersion())
	if err != nil {
		return fmt.Errorf("❌ Failed to parse current version %s: %v", config.GetVersion(), err), ""
	}
	latestVer, err := parseVersion(latestRelease.TagName)
	if err != nil {
		return fmt.Errorf("❌ Failed to parse latest version %s: %v", latestRelease.TagName, err), ""
	}

	logger.Install.Debug(fmt.Sprintf("Current version: %s, Latest version: %s", config.GetVersion(), latestRelease.TagName))

	// Check if we should update
	updateReason, shouldUpdate := shouldUpdate(currentVer, latestVer, isInUpdateableState, majorUpdateApproved)
	if !shouldUpdate {
		switch updateReason {
		case "up-to-date":
			logger.Install.Info("🎉 No update needed: you’re already on the latest version.")
		case "major-update":
			logger.Install.Warn(fmt.Sprintf("⚠️ Update found: Latest version %s is a major update from %s. Major Updates include Breaking changes in this project. Read the release notes and backup your Server folder before updating. Enable 'AllowMajorUpdates' in config to proceed.", latestRelease.TagName, config.Version))
			return nil, latestRelease.TagName
		case "not-in-updateable-state":
			logger.Install.Debug("⚠️ Update found but SSUI is not in an updatable state.")
			return nil, latestRelease.TagName
		}
		return nil, ""
	}

	// Proceed with update
	expectedExt := ".exe"
	if runtime.GOOS != "windows" {
		expectedExt = ".x86_64"
	}
	expectedExe := fmt.Sprintf("StationeersServerControl%s%s", latestRelease.TagName, expectedExt)

	// Find the asset
	var downloadURL string
	for _, asset := range latestRelease.Assets {
		if asset.Name == expectedExe {
			downloadURL = asset.URL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("❌ No matching asset found for %s", expectedExe), latestRelease.TagName
	}

	// Download and replace
	logger.Install.Info(fmt.Sprintf("📡 Updating from %s to %s...", config.GetVersion(), latestRelease.TagName))
	if err := downloadNewExecutable(expectedExe, downloadURL); err != nil {
		logger.Install.Warn(fmt.Sprintf("⚠️ Update failed: %v. Keeping version %s.", err, config.GetVersion()))
		return err, ""
	}

	// Set executable permissions on Linux
	if runtime.GOOS != "windows" {
		if err := os.Chmod(expectedExe, 0755); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Update failed: couldn’t make %s executable: %v. Keeping version %s.", expectedExe, err, config.GetVersion()))
			return err, ""
		}
	}

	// Stop the game server before launching the new version to prevent detached processes
	if config.GetIsGameServerRunning() {
		logger.Install.Info("🛑 Stopping game server before applying update...")
		if err := gamemgr.InternalStopServer(); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Failed to stop game server before update: %v. Proceeding anyway.", err))
		}
	}

	// Launch the new executable and exit
	logger.Install.Info("🚀 Launching the new version and retiring the old one...")
	if runtime.GOOS == "windows" {
		if err := runAndExit(expectedExe); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Update failed: couldn’t launch %s: %v. Keeping version %s.", expectedExe, err, config.GetVersion()))
			return err, ""
		}
	}
	if runtime.GOOS == "linux" {
		if err := runAndExitLinux(expectedExe); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Update failed: couldn’t launch %s: %v. Keeping version %s.", expectedExe, err, config.GetVersion()))
			return err, ""
		}
	}

	return nil, ""
}

// downloadNewExecutable downloads the new executable with a progress bar
func downloadNewExecutable(filename, url string) error {
	// Use a temp file to avoid partial downloads
	tmpFile := filename + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile) // Clean up .tmp on any failure after creation

	// Download from GitHub
	resp, err := downloadHTTPClient.Get(url)
	if err != nil {
		out.Close()
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out.Close()
		return fmt.Errorf("bad response from download: %s", resp.Status)
	}

	// Show progress
	counter := &WriteCounter{Total: resp.ContentLength}
	written, err := io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		out.Close()
		return fmt.Errorf("failed to write download to file: %v", err)
	}
	if written == 0 {
		out.Close()
		return fmt.Errorf("downloaded file is empty")
	}

	// Explicitly close the file before renaming
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// The destination can be left behind by an earlier failed launch. The
	// running executable always has an older version and is not this file.
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to replace stale download %s: %v", filename, err)
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %v", filename, err)
	}

	logger.Install.Info("✅ Downloaded " + filename)
	return nil
}
