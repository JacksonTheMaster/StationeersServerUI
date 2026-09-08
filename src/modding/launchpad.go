package modding

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/setup/update"
)

// InstallSLP downloads the latest StationeersLaunchPad-server zip from GitHub
// and extracts it into BepInEx/plugins/StationeersLaunchPad
// Returns: (installed version tag or "", error)
func InstallSLP() (string, error) {
	const repoOwner = "StationeersLaunchPad"
	const repoName = "StationeersLaunchPad"
	const slpAssetPattern = "StationeersLaunchPad-server-"

	// Prepare target folder
	pluginsDir := "BepInEx/plugins"
	slpDir := filepath.Join(pluginsDir, "StationeersLaunchPad")

	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", repoOwner, repoName)
	logger.Install.Info("📡 Fetching latest Stationeers Launch Pad release...")

	resp, err := http.Get(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var releases []struct {
		TagName    string `json:"tag_name"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to parse GitHub releases: %w", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found in %s/%s", repoOwner, repoName)
	}

	// Find newest non-prerelease release with server zip
	var selectedRelease *struct {
		TagName string
		URL     string
	}

	for _, rel := range releases {
		if rel.Prerelease {
			continue // skip prereleases for now (can be made configurable later)
		}

		for _, asset := range rel.Assets {
			if strings.HasPrefix(asset.Name, slpAssetPattern) &&
				strings.HasSuffix(asset.Name, ".zip") {
				selectedRelease = &struct {
					TagName string
					URL     string
				}{rel.TagName, asset.URL}
				break
			}
		}
		if selectedRelease != nil {
			break
		}
	}

	if selectedRelease == nil {
		return "", fmt.Errorf("no suitable StationeersLaunchPad-server-*.zip found in latest releases")
	}

	zipName := fmt.Sprintf("StationeersLaunchPad-server-%s.zip", selectedRelease.TagName)
	downloadURL := selectedRelease.URL

	logger.Install.Info(fmt.Sprintf("Found SLP %s → downloading %s...", selectedRelease.TagName, zipName))

	// Download to temp file
	tmpZip := zipName + ".tmp"
	if err := downloadFile(tmpZip, downloadURL); err != nil {
		return "", fmt.Errorf("failed to download SLP zip: %w", err)
	}
	defer os.Remove(tmpZip)

	// Clean previous installation if exists
	if err := os.RemoveAll(slpDir); err != nil {
		logger.Install.Warn(fmt.Sprintf("Could not clean old SLP folder: %v", err))
	}

	// Make sure parent exists
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create BepInEx/plugins directory: %w", err)
	}

	// Extract
	logger.Install.Info("📦 Extracting Stationeers Launch Pad...")
	if err := unzipTo(tmpZip, slpDir, func(name string) bool {
		// Only process files inside the StationeersLaunchPad folder
		return strings.HasPrefix(name, "StationeersLaunchPad/")
	}); err != nil {
		return "", fmt.Errorf("failed to extract SLP: %w", err)
	}

	logger.Install.Info(fmt.Sprintf("✅ Stationeers Launch Pad %s installed to %s", selectedRelease.TagName, slpDir))
	logger.Install.Info("💡 SLP contains its own auto-updater - future updates should happen automatically.")
	config.SetIsStationeersLaunchPadEnabled(true)

	return selectedRelease.TagName, nil
}

// UninstallSLP removes the SLP folder from BepInEx/plugins
// Returns: ("success" or "failed", error)
func UninstallSLP() (string, error) {
	pluginsDir := "BepInEx/plugins"
	slpDir := filepath.Join(pluginsDir, "StationeersLaunchPad")

	// stat the folder to see if it exists, if not skip removal
	if _, err := os.Stat(slpDir); os.IsNotExist(err) {
		logger.Install.Info("SLP is not installed; nothing to uninstall")
		config.SetIsStationeersLaunchPadEnabled(false)
		return "not_installed", nil
	}

	if err := os.RemoveAll(slpDir); err != nil {
		logger.Install.Error("Failed to remove SLP folder: " + err.Error())
		return "failed", fmt.Errorf("failed to remove SLP folder: %w", err)
	}

	// stat the current directory to see if mod files exist, if not skip removal
	if _, err := os.Stat(filepath.Join(".", "mods")); os.IsNotExist(err) {
		logger.Install.Info("No mods folder found; skipping mods removal")
	} else {

		// remove the ./mods folder and the modconfig.xml file too
		if err := os.RemoveAll(filepath.Join(".", "mods")); err != nil {
			logger.Install.Error("Failed to remove the mods folder: " + err.Error())
			return "failed", fmt.Errorf("failed to remove mods folder: %w", err)
		}
	}

	// stat the modconfig.xml file to see if it exists, if not skip removal
	if _, err := os.Stat(filepath.Join(".", "modconfig.xml")); os.IsNotExist(err) {
		logger.Install.Info("No modconfig.xml file found; skipping its removal")
	} else {

		if err := os.Remove(filepath.Join(".", "modconfig.xml")); err != nil {
			logger.Install.Error("Failed to remove modconfig.xml file: " + err.Error())
			return "failed", fmt.Errorf("failed to remove modconfig.xml file: %w", err)
		}
	}

	config.SetIsStationeersLaunchPadEnabled(false)
	logger.Install.Info("SLP uninstalled successfully")

	return "success", nil
}

// ReinstallSLP removes only the SLP plugin directory (preserving mods and modconfig.xml),
// then downloads and installs the latest version.
// Returns: (installed version tag or "", error)
func ReinstallSLP() (string, error) {
	pluginsDir := "BepInEx/plugins"
	slpDir := filepath.Join(pluginsDir, "StationeersLaunchPad")

	// Remove only the SLP plugin directory, keep mods & modconfig.xml
	if _, err := os.Stat(slpDir); err == nil {
		logger.Install.Info("🔄 Removing existing SLP installation for reinstall...")
		if err := os.RemoveAll(slpDir); err != nil {
			return "", fmt.Errorf("failed to remove SLP folder for reinstall: %w", err)
		}
	} else {
		logger.Install.Info("SLP not currently installed; proceeding with fresh install")
	}

	// Now install fresh
	return InstallSLP()
}

func downloadFile(destPath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	counter := &update.WriteCounter{Total: resp.ContentLength}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	return nil
}

// unzipTo extracts zip contents into destDir
// Only extracts files where shouldExtract returns true
func unzipTo(zipPath, destDir string, shouldExtract func(fileName string) bool) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if !shouldExtract(f.Name) {
			continue
		}

		// Remove the leading "StationeersLaunchPad/" from the path
		// so BepInEx/plugins/StationeersLaunchPad/core.dll etc.
		relPath := strings.TrimPrefix(f.Name, "StationeersLaunchPad/")
		if relPath == f.Name { // safety check
			logger.Install.Warn("Unexpected file outside StationeersLaunchPad/: " + f.Name)
			continue
		}

		// Build full target path
		fpath := filepath.Join(destDir, relPath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
