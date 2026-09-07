package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

const releasesURL = "https://api.github.com/repos/SteamServerUI/StationeersServerUI/releases?per_page=100"

var versionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	URL        string        `json:"html_url"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease []string
}

type release struct {
	Tag        string
	URL        string
	Version    Version
	Prerelease bool
	Assets     []githubAsset
}

func parseVersion(value string) (Version, error) {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return Version{}, fmt.Errorf("invalid semantic version %q", value)
	}
	major, _ := strconv.Atoi(match[1])
	minor, _ := strconv.Atoi(match[2])
	patch, _ := strconv.Atoi(match[3])
	version := Version{Major: major, Minor: minor, Patch: patch}
	if match[4] != "" {
		version.Prerelease = strings.Split(match[4], ".")
		for _, part := range version.Prerelease {
			if len(part) > 1 && part[0] == '0' {
				if _, err := strconv.Atoi(part); err == nil {
					return Version{}, fmt.Errorf("invalid numeric prerelease identifier %q", part)
				}
			}
		}
	}
	return version, nil
}

func versionString(version Version) string {
	value := fmt.Sprintf("v%d.%d.%d", version.Major, version.Minor, version.Patch)
	if len(version.Prerelease) > 0 {
		value += "-" + strings.Join(version.Prerelease, ".")
	}
	return value
}

func compareVersions(left, right Version) int {
	leftNumbers := []int{left.Major, left.Minor, left.Patch}
	rightNumbers := []int{right.Major, right.Minor, right.Patch}
	for i := range leftNumbers {
		if leftNumbers[i] < rightNumbers[i] {
			return -1
		}
		if leftNumbers[i] > rightNumbers[i] {
			return 1
		}
	}
	if len(left.Prerelease) == 0 && len(right.Prerelease) == 0 {
		return 0
	}
	if len(left.Prerelease) == 0 {
		return 1
	}
	if len(right.Prerelease) == 0 {
		return -1
	}
	for i := 0; i < len(left.Prerelease) && i < len(right.Prerelease); i++ {
		leftPart, leftNumeric := numericIdentifier(left.Prerelease[i])
		rightPart, rightNumeric := numericIdentifier(right.Prerelease[i])
		switch {
		case leftNumeric && rightNumeric:
			if leftPart < rightPart {
				return -1
			}
			if leftPart > rightPart {
				return 1
			}
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		default:
			if left.Prerelease[i] < right.Prerelease[i] {
				return -1
			}
			if left.Prerelease[i] > right.Prerelease[i] {
				return 1
			}
		}
	}
	if len(left.Prerelease) < len(right.Prerelease) {
		return -1
	}
	if len(left.Prerelease) > len(right.Prerelease) {
		return 1
	}
	return 0
}

func numericIdentifier(value string) (int, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	number, err := strconv.Atoi(value)
	return number, err == nil
}

func fetchReleases(ctx context.Context) ([]release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "StationeersServerUI-Updater")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned %s", resp.Status)
	}
	var raw []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode GitHub releases: %w", err)
	}
	releases := make([]release, 0, len(raw))
	for _, item := range raw {
		if item.Draft {
			continue
		}
		version, err := parseVersion(item.TagName)
		if err != nil {
			logger.Install.Warnf("Skipping release %q: %v", item.TagName, err)
			continue
		}
		semanticPrerelease := len(version.Prerelease) > 0
		if semanticPrerelease != item.Prerelease {
			logger.Install.Warnf("Skipping release %q: tag and GitHub prerelease state disagree", item.TagName)
			continue
		}
		releases = append(releases, release{Tag: item.TagName, URL: item.URL, Version: version, Prerelease: semanticPrerelease, Assets: item.Assets})
	}
	sort.SliceStable(releases, func(i, j int) bool { return compareVersions(releases[i].Version, releases[j].Version) > 0 })
	return releases, nil
}

func expectedExecutable(tag string) string {
	extension := ".exe"
	if runtime.GOOS != "windows" {
		extension = ".x86_64"
	}
	return "StationeersServerControl" + tag + extension
}

func findAsset(item release) (githubAsset, bool) {
	expected := expectedExecutable(item.Tag)
	for _, asset := range item.Assets {
		if asset.Name == expected && asset.URL != "" {
			return asset, true
		}
	}
	return githubAsset{}, false
}

func findRelease(releases []release, tag string) (release, bool) {
	for _, item := range releases {
		if item.Tag == tag {
			return item, true
		}
	}
	return release{}, false
}
