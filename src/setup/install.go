package setup

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os" // Added for filepath.Dir
	fp "path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/steamcmd"
)

var downloadBranch string // Holds the branch to download from

// Install performs the entire installation process and ensures the server waits for it to complete
func Install(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done() // Signal that installation is complete

	// Step 0: Check for updates
	if err, _ := update.Update(true); err != nil {
		logger.Install.Error("❌Update check went sideways: " + err.Error())
	}

	// Step 1: Check and download the UIMod folder contents
	logger.Install.Debug("🔄Checking UIMod folder...")
	CheckAndDownloadUIMod()
	logger.Install.Debug("✅UIMod folder setup complete.")
	// Step 2: Check for Blacklist.txt and create it if it doesn't exist
	logger.Install.Info("🔄Checking for Blacklist.txt...")
	checkAndCreateBlacklist()
	logger.Install.Info("✅Blacklist.txt verified or created.")
	// Step 3: Install and run SteamCMD
	logger.Install.Info("🔄Installing and running SteamCMD...")
	if config.GetSkipSteamCMD() {
		logger.Install.Info("✅Skipping SteamCMD installation, SkipSteamCMD is true")
	} else {
		steamcmd.InstallAndRunSteamCMD()
	}
	logger.Install.Info("✅Setup complete!")
}

func CheckAndDownloadUIMod() {
	uiModDir := config.GetUIModFolder()
	configDir := config.GetUIModFolder() + "config/"
	tlsDir := config.GetUIModFolder() + "tls/"

	requiredDirs := []string{uiModDir, configDir}

	// Set branch
	if config.GetBranch() == "release" || config.GetBranch() == "Release" {
		downloadBranch = "main"
	} else {
		downloadBranch = config.GetBranch()
	}
	logger.Install.Debug("Using branch: " + downloadBranch)

	// Define file mappings
	files := map[string]string{
		// NOTE: Now empty as files are now embedded in the executable. Kept this structure for future use.

		// UI - commented out since files are embedded, left here for reference in case we need this funcitonality again
		// "ui/config.html":           "https://raw.githubusercontent.com/SteamServerUI/SteamServerUI/{branch}/UIMod/ui/config.html",
	}

	createRequiredDirs(requiredDirs)

	if len(files) == 0 {
		logger.Install.Debug("📁 File mappings empty - no additional files to download available")
		return
	}

	// Check if the directory exists
	if _, err := os.Stat(uiModDir); os.IsNotExist(err) {
		// Initial download
		//check if tlsDir exists, if not, set isFirstTimeSetup to true
		if _, err := os.Stat(tlsDir); os.IsNotExist(err) {
			config.SetIsFirstTimeSetup(true)
		} else {
			config.SetIsFirstTimeSetup(false)
		}
		downloadAllFiles(files)
	} else {
		// Directory exists
		config.SetIsFirstTimeSetup(false)
		logger.Install.Debug(fmt.Sprintf("IsUpdateEnabled: %v", config.GetIsUpdateEnabled()))
		logger.Install.Debug(fmt.Sprintf("IsFirstTimeSetup: %v", config.GetIsFirstTimeSetup()))
		if config.GetIsUpdateEnabled() {
			logger.Install.Info("🔍Validating UIMod files for updates...")
			if config.GetBranch() == "release" || config.GetBranch() == "Release" {
				downloadBranch = "main"
				updateFilesIfDifferent(files)
			} else {
				downloadBranch = config.GetBranch()
				updateFilesIfDifferent(files)
			}
		} else {
			logger.Install.Info("♻️Folder ./UIMod already exists. Updates disabled, skipping validation.")
		}
	}
}

// downloadAllFiles downloads all files in the provided map concurrently
func downloadAllFiles(files map[string]string) {
	var wg sync.WaitGroup
	for filepath, url := range files {
		wg.Add(1)
		go func(filepath, url string) {
			defer wg.Done()
			fileName := filepath[strings.LastIndex(filepath, "/")+1:]
			logger.Install.Info("Downloading " + fileName + "...")
			err := downloadFile(filepath, url)
			if err != nil {
				logger.Install.Error("❌Error downloading " + fileName + ": " + err.Error())
			} else {
				logger.Install.Info("✅Downloaded " + fileName + " successfully from branch " + downloadBranch)
			}
		}(filepath, url)
	}
	wg.Wait()
	logger.Install.Info("✅All files downloaded successfully.")
}

// updateFilesIfDifferent checks for differences and updates files if necessary using concurrency
func updateFilesIfDifferent(files map[string]string) {
	var wg sync.WaitGroup
	for filepath, url := range files {
		fileName := filepath[strings.LastIndex(filepath, "/")+1:]
		if fileName == "config.json" {
			continue // Skip updating config.json to preserve local changes
		}
		wg.Add(1)
		go func(filepath, url string) {
			defer wg.Done()
			checkAndUpdateFile(filepath, url)
		}(filepath, url)
	}
	wg.Wait()
	logger.Install.Info("✅File validation and update check complete.")
}

// checkAndUpdateFile checks if a file needs updating by comparing its SHA-1 hash with the remote ETag
func checkAndUpdateFile(filepath, url string) {
	fileName := filepath[strings.LastIndex(filepath, "/")+1:]

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// File doesn't exist locally, download it
		logger.Install.Info("Downloading " + fileName + "...")
		err := downloadFile(filepath, url)
		if err != nil {
			logger.Install.Error("❌Error downloading " + fileName + ": " + err.Error())
		} else {
			logger.Install.Info("✅Downloaded " + fileName + " successfully")
		}
	} else {
		// File exists, check if it needs updating
		localHash, err := computeGitBlobSHA1(filepath)
		if err != nil {
			logger.Install.Error("❌Error computing hash for " + fileName + ": " + err.Error())
			return
		}

		// Extract the necessary parts from URL to build the API call
		// Example URL: https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/index.html
		urlParts := strings.Split(url, "/")
		if len(urlParts) < 7 {
			logger.Install.Error("❌Invalid URL format: " + url)
			return
		}

		repoOwner := urlParts[3]
		repoName := urlParts[4]
		branch := urlParts[5]
		filePath := strings.Join(urlParts[6:], "/")

		remoteHash, err := getFileHash(repoOwner, repoName, branch, filePath)
		if err != nil {
			return
		}

		logger.Install.Debug("Local hash for " + fileName + ": " + localHash)
		logger.Install.Debug("Remote hash for " + fileName + ": " + remoteHash)

		if localHash != remoteHash {
			logger.Install.Info("🔄Updating " + fileName + " due to differences...")
			err := downloadFile(filepath, url)
			if err != nil {
				logger.Install.Error("❌Error updating " + fileName + ": " + err.Error())
			} else {
				logger.Install.Info("✅Updated " + fileName + " successfully from branch " + downloadBranch)
			}
		} else {
			logger.Install.Info("✅" + fileName + " is up-to-date.")
		}
	}
}

func computeGitBlobSHA1(filepath string) (string, error) {
	// Read the file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Create the blob header (format: "blob " + content length + null byte)
	header := fmt.Sprintf("blob %d\x00", len(content))

	// Combine header and content
	blobData := append([]byte(header), content...)

	// Compute SHA-1 hash
	hash := sha1.New()
	hash.Write(blobData)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getFileHash fetches the file hash from GitHub API
func getFileHash(repoOwner, repoName, branch, filePath string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		repoOwner, repoName, filePath, branch)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	// Optional: req.Header.Add("Authorization", "token YOUR_GITHUB_TOKEN")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err := handleRateLimitErrors(resp); err != nil {
			return "", err
		}
		logger.Install.Error(fmt.Sprintf("❌Request failed with status: %s", resp.Status))
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var fileInfo struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return "", err
	}

	return fileInfo.SHA, nil
}

// handleRateLimitErrors processes GitHub API rate limit errors
func handleRateLimitErrors(resp *http.Response) error {
	if resp.StatusCode != http.StatusForbidden {
		return nil
	}

	// Check secondary rate limit
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		return handleSecondaryRateLimit(retryAfter, resp.Status)
	}

	// Check primary rate limit
	if remaining := resp.Header.Get("x-ratelimit-remaining"); remaining == "0" {
		return handlePrimaryRateLimit(resp.Header.Get("x-ratelimit-reset"), resp.Status)
	}

	// Generic 403 error
	logger.Install.Error("❌Forbidden request, no specific rate limit info available")
	return fmt.Errorf("bad status: %s", resp.Status)
}

// handleSecondaryRateLimit processes secondary rate limit exceeded errors
func handleSecondaryRateLimit(retryAfter, status string) error {
	waitSeconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		logger.Install.Error(fmt.Sprintf("❌Failed to parse Retry-After header: %v", err))
		return fmt.Errorf("bad status: %s, failed to parse Retry-After: %v", status, err)
	}
	errMsg := fmt.Sprintf("Github API secondary rate limit exceeded. (too many requests in a short time). Retry after %d seconds", waitSeconds)
	logger.Install.Error("❌" + errMsg)
	return fmt.Errorf("bad status: %s, %s", status, errMsg)
}

// handlePrimaryRateLimit processes primary rate limit exceeded errors
func handlePrimaryRateLimit(reset, status string) error {
	resetInt, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		logger.Install.Error(fmt.Sprintf("❌Failed to parse x-ratelimit-reset header: %v", err))
		return fmt.Errorf("bad status: %s, failed to parse x-ratelimit-reset: %v", status, err)
	}
	resetTime := time.Unix(resetInt, 0).UTC()
	errMsg := fmt.Sprintf("Github ratelimit exceeded: hourly request quota of 60 calls reached. Resets at %s", resetTime.Format(time.RFC1123))
	logger.Install.Warn("🧱" + errMsg)
	return fmt.Errorf("bad status: %s, %s", status, errMsg)
}

// downloadFile downloads a file from the given URL to the specified filepath
func downloadFile(filepath, url string) error {
	client := &http.Client{Timeout: 45 * time.Second}
	return downloadFileWithClient(client, filepath, url)
}

func downloadFileWithClient(client *http.Client, filepath, url string) error {
	// Ensure the directory exists
	dir := fp.Dir(filepath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Fetch the data
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	// Write to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// checkAndCreateBlacklist ensures Blacklist.txt exists in the root directory
func checkAndCreateBlacklist() {
	blacklistFile := "./Blacklist.txt"

	// Check if Blacklist.txt exists
	if _, err := os.Stat(blacklistFile); os.IsNotExist(err) {
		// Create Blacklist.txt file with a dummy steamID64 so the gameserver doesn't fail reading this file, as it would not be the expected format if it was empty.
		perm := os.FileMode(0644) // Still works cross-platform
		err := os.WriteFile(blacklistFile, []byte("76561197960265728"), perm)
		if err != nil {
			logger.Install.Error("❌Error creating Blacklist.txt: " + err.Error())
			return
		}

		logger.Install.Info("✅Created Blacklist.txt with dummy steamID64.")
	} else {
		logger.Install.Debug("♻️Blacklist.txt already exists. Skipping creation.")
	}
}

func createRequiredDirs(requiredDirs []string) {
	// Create directories
	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			config.SetIsFirstTimeSetup(true)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				logger.Install.Error("❌Error creating folder: " + err.Error())
				return
			}
			logger.Install.Warn("⚠️Created folder: " + dir)
		}
	}
}
