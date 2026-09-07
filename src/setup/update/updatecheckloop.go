package update

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

var updateState = struct {
	sync.RWMutex
	status Status
}{status: Status{Candidates: []Candidate{}}}

var updateOperation sync.Mutex

func StatusSnapshot() Status {
	updateState.RLock()
	status := copyStatus(updateState.status)
	updateState.RUnlock()
	if status.CurrentVersion == "" {
		status.CurrentVersion = config.GetVersion()
		status.Enabled = config.GetIsUpdateEnabled() || config.GetIsDockerContainer()
		status.ContainerManaged = config.GetIsDockerContainer()
	}
	return status
}

func RefreshStatus(ctx context.Context, manual bool) (Status, error) {
	if !updateOperation.TryLock() {
		return StatusSnapshot(), ErrUpdateBusy
	}
	defer updateOperation.Unlock()
	setChecking(true)
	status, err := discover(ctx, manual)
	if err != nil {
		status = StatusSnapshot()
		status.Checking = false
		status.LastError = err.Error()
		storeStatus(status)
		return status, err
	}
	status.Checking = false
	storeStatus(status)
	return copyStatus(status), nil
}

func InstallVersion(ctx context.Context, request InstallRequest, manual bool) error {
	if config.GetIsDockerContainer() {
		return ErrContainerManaged
	}
	if !updateOperation.TryLock() {
		return ErrUpdateBusy
	}
	defer updateOperation.Unlock()
	beginInstall(request.Version)
	err := install(ctx, request, manual)
	finishInstall(err)
	return err
}

func QueueInstall(request InstallRequest) error {
	if config.GetIsDockerContainer() {
		return ErrContainerManaged
	}
	if err := validateApproval(request, StatusSnapshot()); err != nil {
		return err
	}
	if !updateOperation.TryLock() {
		return ErrUpdateBusy
	}
	beginInstall(request.Version)
	go func() {
		defer updateOperation.Unlock()
		err := install(context.Background(), request, true)
		finishInstall(err)
		if err != nil {
			logger.Install.Error("Manual update failed: " + err.Error())
		}
	}()
	return nil
}

func validateApproval(request InstallRequest, status Status) error {
	if request.Version == "" {
		return errors.New("no update version selected")
	}
	for _, candidate := range status.Candidates {
		if candidate.Version != request.Version {
			continue
		}
		if !candidate.AssetAvailable {
			return errors.New("this release has no download for the current platform")
		}
		if candidate.Major && !request.ConfirmMajor {
			return errors.New("major update confirmation is required")
		}
		if candidate.Prerelease && !request.ConfirmPrerelease {
			return errors.New("prerelease confirmation is required")
		}
		return nil
	}
	return errors.New("refresh update information before installing this version")
}

func beginInstall(version string) {
	status := StatusSnapshot()
	status.Installing = true
	status.InstallingVersion = version
	status.LastError = ""
	storeStatus(status)
}

func finishInstall(err error) {
	status := StatusSnapshot()
	status.Installing = false
	status.InstallingVersion = ""
	if err != nil {
		status.LastError = err.Error()
	}
	storeStatus(status)
}

func setChecking(checking bool) {
	status := StatusSnapshot()
	status.Checking = checking
	if checking {
		status.LastError = ""
	}
	storeStatus(status)
}

func storeStatus(status Status) {
	updateState.Lock()
	updateState.status = copyStatus(status)
	updateState.Unlock()
}

func copyStatus(status Status) Status {
	copy := status
	copy.Candidates = append([]Candidate(nil), status.Candidates...)
	copy.CompatibleStable = cloneCandidate(status.CompatibleStable)
	copy.LatestStable = cloneCandidate(status.LatestStable)
	copy.LatestPrerelease = cloneCandidate(status.LatestPrerelease)
	copy.LatestMajor = cloneCandidate(status.LatestMajor)
	copy.Automatic = cloneCandidate(status.Automatic)
	return copy
}

func cloneCandidate(candidate *Candidate) *Candidate {
	if candidate == nil {
		return nil
	}
	copy := *candidate
	return &copy
}

func StartUpdateCheckLoop() {
	lastContainerNotice := ""
	for _, candidate := range StatusSnapshot().Candidates {
		if candidate.AssetAvailable {
			lastContainerNotice = candidate.Version
			break
		}
	}
	for {
		status, err := RefreshStatus(context.Background(), false)
		if err != nil {
			logger.Install.Warn("Automatic SSUI update check failed: " + err.Error())
		} else if status.ContainerManaged {
			for _, candidate := range status.Candidates {
				if candidate.AssetAvailable && candidate.Version != lastContainerNotice {
					logger.Install.Infof("SSUI %s is available. This installation runs in a container; pull the new image and recreate the container to update.", candidate.Version)
					lastContainerNotice = candidate.Version
					break
				}
			}
		}
		delay := 6 * time.Hour
		if !config.GetIsUpdateEnabled() && !config.GetIsDockerContainer() {
			delay = 30 * time.Minute
		}
		time.Sleep(delay)
	}
}
