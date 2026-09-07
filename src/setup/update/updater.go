package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
)

var ErrUpdateBusy = errors.New("an update operation is already in progress")
var ErrContainerManaged = errors.New("SSUI updates are managed by the container image; pull the new image and recreate the container")

type Candidate struct {
	Version        string `json:"version"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	Major          bool   `json:"major"`
	Prerelease     bool   `json:"prerelease"`
	AssetAvailable bool   `json:"assetAvailable"`
}

type Status struct {
	CurrentVersion    string      `json:"currentVersion"`
	Enabled           bool        `json:"enabled"`
	DevelopmentBuild  bool        `json:"developmentBuild"`
	UpdateAvailable   bool        `json:"updateAvailable"`
	Candidates        []Candidate `json:"candidates"`
	CompatibleStable  *Candidate  `json:"compatibleStable,omitempty"`
	LatestStable      *Candidate  `json:"latestStable,omitempty"`
	LatestPrerelease  *Candidate  `json:"latestPrerelease,omitempty"`
	LatestMajor       *Candidate  `json:"latestMajor,omitempty"`
	Automatic         *Candidate  `json:"automatic,omitempty"`
	Checking          bool        `json:"checking"`
	Installing        bool        `json:"installing"`
	InstallingVersion string      `json:"installingVersion,omitempty"`
	LastError         string      `json:"lastError,omitempty"`
	CheckedAt         *time.Time  `json:"checkedAt,omitempty"`
	ContainerManaged  bool        `json:"containerManaged"`
}

type InstallRequest struct {
	Version           string `json:"version"`
	ConfirmMajor      bool   `json:"confirmMajor"`
	ConfirmPrerelease bool   `json:"confirmPrerelease"`
}

func discover(ctx context.Context, manual bool) (Status, error) {
	containerManaged := config.GetIsDockerContainer()
	status := Status{CurrentVersion: config.GetVersion(), Enabled: config.GetIsUpdateEnabled() || containerManaged, ContainerManaged: containerManaged, Candidates: []Candidate{}}
	if !manual && !status.Enabled {
		return status, nil
	}
	if config.GetBranch() != "release" {
		status.DevelopmentBuild = true
		return status, nil
	}
	current, err := parseVersion(config.GetVersion())
	if err != nil {
		return status, fmt.Errorf("parse current version %q: %w", config.GetVersion(), err)
	}
	releases, err := fetchReleases(ctx)
	if err != nil {
		return status, fmt.Errorf("fetch releases: %w", err)
	}
	return buildStatus(status, current, releases, config.GetAllowMajorUpdates(), config.GetAllowPrereleaseUpdates()), nil
}

func buildStatus(status Status, current Version, releases []release, allowMajor, allowPrerelease bool) Status {
	for _, item := range releases {
		if compareVersions(item.Version, current) <= 0 {
			continue
		}
		_, hasAsset := findAsset(item)
		candidate := Candidate{Version: item.Tag, ReleaseNotes: item.URL, Major: item.Version.Major > current.Major, Prerelease: item.Prerelease, AssetAvailable: hasAsset}
		status.Candidates = append(status.Candidates, candidate)
		if !candidate.Prerelease && !candidate.Major && status.CompatibleStable == nil {
			status.CompatibleStable = candidatePointer(candidate)
		}
		if !candidate.Prerelease && status.LatestStable == nil {
			status.LatestStable = candidatePointer(candidate)
		}
		if candidate.Prerelease && status.LatestPrerelease == nil {
			status.LatestPrerelease = candidatePointer(candidate)
		}
		if candidate.Major && status.LatestMajor == nil {
			status.LatestMajor = candidatePointer(candidate)
		}
		if status.Automatic == nil && candidate.AssetAvailable &&
			(!candidate.Major || allowMajor) &&
			(!candidate.Prerelease || allowPrerelease) {
			status.Automatic = candidatePointer(candidate)
		}
	}
	status.UpdateAvailable = len(status.Candidates) > 0
	now := time.Now()
	status.CheckedAt = &now
	return status
}

func candidatePointer(candidate Candidate) *Candidate {
	copy := candidate
	return &copy
}

func install(ctx context.Context, request InstallRequest, manual bool) error {
	if config.GetIsDockerContainer() {
		return ErrContainerManaged
	}
	if request.Version == "" {
		return errors.New("no update version selected")
	}
	if config.GetBranch() != "release" {
		return errors.New("updates are not installed on development builds")
	}
	current, err := parseVersion(config.GetVersion())
	if err != nil {
		return fmt.Errorf("parse current version %q: %w", config.GetVersion(), err)
	}

	// Always fetch again. We install the exact release the user reviewed, not
	// whichever release happens to be newest when the request arrives.
	releases, err := fetchReleases(ctx)
	if err != nil {
		return fmt.Errorf("revalidate release: %w", err)
	}
	target, found := findRelease(releases, request.Version)
	if !found {
		return fmt.Errorf("release %q is no longer available", request.Version)
	}
	if compareVersions(target.Version, current) <= 0 {
		return fmt.Errorf("release %s is not newer than %s", target.Tag, config.GetVersion())
	}
	isMajor := target.Version.Major > current.Major
	if manual {
		if isMajor && !request.ConfirmMajor {
			return errors.New("major update confirmation is required")
		}
		if target.Prerelease && !request.ConfirmPrerelease {
			return errors.New("prerelease confirmation is required")
		}
	} else {
		if isMajor && !config.GetAllowMajorUpdates() {
			return errors.New("major updates are disabled")
		}
		if target.Prerelease && !config.GetAllowPrereleaseUpdates() {
			return errors.New("prerelease updates are disabled")
		}
	}
	asset, found := findAsset(target)
	if !found {
		return fmt.Errorf("release %s has no %s asset", target.Tag, runtime.GOOS)
	}

	filename := expectedExecutable(target.Tag)
	logger.Install.Infof("Updating from %s to %s...", config.GetVersion(), target.Tag)
	if err := downloadNewExecutable(ctx, filename, asset.URL); err != nil {
		return fmt.Errorf("download %s: %w", target.Tag, err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filename, 0755); err != nil {
			return fmt.Errorf("make %s executable: %w", filename, err)
		}
	}
	if config.GetIsGameServerRunning() {
		logger.Install.Info("Stopping game server before applying update...")
		if err := gamemgr.InternalStopServer(); err != nil {
			logger.Install.Warnf("Failed to stop the game server before update: %v. Proceeding anyway.", err)
		}
	}
	logger.Install.Info("Launching the new version...")
	if runtime.GOOS == "windows" {
		return runAndExit(filename)
	}
	if runtime.GOOS == "linux" {
		return runAndExitLinux(filename)
	}
	return fmt.Errorf("updates are not supported on %s", runtime.GOOS)
}

// Update keeps the old startup and CLI entrypoint working. Automated calls use
// the configured policy; the Web UI uses InstallVersion with explicit consent.
func Update(isInUpdateableState bool) (error, string) {
	status, err := RefreshStatus(context.Background(), false)
	if err != nil {
		return err, ""
	}
	if status.ContainerManaged {
		for _, candidate := range status.Candidates {
			if !candidate.AssetAvailable {
				continue
			}
			if isInUpdateableState {
				return ErrContainerManaged, candidate.Version
			}
			return nil, candidate.Version
		}
		return nil, ""
	}
	if status.Automatic == nil {
		return nil, ""
	}
	if !isInUpdateableState {
		return nil, status.Automatic.Version
	}
	request := InstallRequest{Version: status.Automatic.Version}
	if err := InstallVersion(context.Background(), request, false); err != nil {
		return err, status.Automatic.Version
	}
	return nil, ""
}

func downloadNewExecutable(ctx context.Context, filename, url string) error {
	tmpFile := filename + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		out.Close()
		return err
	}
	req.Header.Set("User-Agent", "StationeersServerUI-Updater")
	client := &http.Client{Timeout: 20 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		out.Close()
		return fmt.Errorf("download returned %s", resp.Status)
	}
	counter := &WriteCounter{Total: resp.ContentLength}
	written, err := io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		out.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if written == 0 {
		out.Close()
		return errors.New("downloaded file is empty")
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale download: %w", err)
	}
	if err := os.Rename(tmpFile, filename); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	logger.Install.Info("Downloaded " + filename)
	return nil
}
