package update

import (
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

type Status struct {
	Available bool
	Version   string
	Major     bool
	State     string
	Error     string
}

var updateOperation sync.Mutex

var UpdateInfo struct {
	sync.RWMutex

	Available bool
	Version   string
	Major     bool
	State     string
	Error     string
}

func init() {
	UpdateInfo.State = "idle"
}

func TryStartOperation() bool {
	return updateOperation.TryLock()
}

func FinishOperation() {
	updateOperation.Unlock()
}

func GetStatus() Status {
	UpdateInfo.RLock()
	defer UpdateInfo.RUnlock()
	return Status{
		Available: UpdateInfo.Available,
		Version:   UpdateInfo.Version,
		Major:     UpdateInfo.Major,
		State:     UpdateInfo.State,
		Error:     UpdateInfo.Error,
	}
}

func SetApplying(version string) {
	UpdateInfo.Lock()
	defer UpdateInfo.Unlock()
	UpdateInfo.Available = true
	UpdateInfo.Version = version
	UpdateInfo.Major = IsMajorUpdate(version)
	UpdateInfo.State = "installing"
	UpdateInfo.Error = ""
}

func SetCheckResult(err error, version string) {
	UpdateInfo.Lock()
	defer UpdateInfo.Unlock()
	if err != nil {
		UpdateInfo.State = "idle"
		UpdateInfo.Error = err.Error()
		return
	}
	UpdateInfo.Available = version != ""
	UpdateInfo.Version = version
	UpdateInfo.Major = IsMajorUpdate(version)
	UpdateInfo.State = "idle"
	UpdateInfo.Error = ""
}

func SetUpdateFailed(version string, err error) {
	UpdateInfo.Lock()
	defer UpdateInfo.Unlock()
	UpdateInfo.Available = version != ""
	UpdateInfo.Version = version
	UpdateInfo.Major = IsMajorUpdate(version)
	UpdateInfo.State = "failed"
	UpdateInfo.Error = err.Error()
}

// StartUpdateCheckLoop checks for updates every six hours without blocking UI status requests.
func StartUpdateCheckLoop() {
	for {
		if TryStartOperation() {
			err, newVersion := CheckForUpdates()
			if err != nil {
				logger.Install.Warn("⚠️ Automatic SSUI Update check failed: " + err.Error())
			}
			SetCheckResult(err, newVersion)
			FinishOperation()
		}
		time.Sleep(6 * time.Hour)
	}
}
