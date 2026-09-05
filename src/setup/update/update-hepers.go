package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

var githubReleasesURL = "https://api.github.com/repos/SteamServerUI/StationeersServerUI/releases?per_page=100"
var releaseHTTPClient = &http.Client{Timeout: 30 * time.Second}
var downloadHTTPClient = &http.Client{Timeout: 30 * time.Minute}

// parseVersion parses a version string (e.g., "4.6.10") into a Version struct and tries to handle a few culprits too
func parseVersion(v string) (Version, error) {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}

	var ver Version
	_, err := fmt.Sscanf(v, "%d.%d.%d", &ver.Major, &ver.Minor, &ver.Patch)
	if err != nil {
		return Version{}, fmt.Errorf("no valid X.Y.Z in tag: %s", v)
	}
	return ver, nil
}

// shouldUpdate determines if an update should proceed, returning reason if not
func shouldUpdate(current, latest Version, isInUpdateableState, majorUpdateApproved bool) (string, bool) {
	// Check if already up-to-date or older
	if latest.Major < current.Major ||
		(latest.Major == current.Major && latest.Minor < current.Minor) ||
		(latest.Major == current.Major && latest.Minor == current.Minor && latest.Patch <= current.Patch) {
		return "up-to-date", false
	}

	// Check if it’s a major update and not allowed
	if current.Major != latest.Major && !config.GetAllowMajorUpdates() && !majorUpdateApproved {
		return "major-update", false
	}

	if !isInUpdateableState {
		return "not-in-updateable-state", false
	}

	return "", true
}

// getLatestRelease fetches the most recent release (or prerelease) from GitHub API
func getLatestRelease() (*githubRelease, error) {
	resp, err := releaseHTTPClient.Get(githubReleasesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from GitHub API: %s", resp.Status)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %v", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Find the newest release allowed by the configured release channel.
	var latestRelease *githubRelease
	var latestVersion Version
	for i := range releases {
		release := &releases[i]
		if release.Draft || (release.Prerelease && !config.GetAllowPrereleaseUpdates()) {
			continue
		}
		version, err := parseVersion(release.TagName)
		if err != nil {
			logger.Install.Warn(fmt.Sprintf("Skipping invalid version tag %s: %v", release.TagName, err))
			continue
		}
		if latestRelease == nil || isReleaseNewerVersion(version, latestVersion) {
			latestVersion = version
			latestRelease = release
		}
	}

	if latestRelease == nil {
		return nil, fmt.Errorf("no suitable releases found")
	}

	return latestRelease, nil
}

func IsMajorUpdate(version string) bool {
	current, err := parseVersion(config.GetVersion())
	if err != nil {
		return false
	}
	latest, err := parseVersion(version)
	if err != nil {
		return false
	}
	return current.Major != latest.Major
}

// isNewerVersion compares two versions to determine if the first is newer
func isReleaseNewerVersion(v1, v2 Version) bool {
	if v1.Major != v2.Major {
		return v1.Major > v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor > v2.Minor
	}
	if v1.Patch == v2.Patch {
		return false
	}
	return v1.Patch > v2.Patch
}
