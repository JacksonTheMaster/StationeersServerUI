package gamemgr

import (
	"runtime"
	"syscall"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

func StartIsGameServerRunningCheck() {
	go func() {
		for {
			if InternalIsServerRunning() {
				config.SetIsGameServerRunning(true)
			} else {
				config.SetIsGameServerRunning(false)
				SetServerState(ServerStateStopped)
			}
			time.Sleep(4 * time.Second)
		}
	}()
}

// InternalIsServerRunning checks if the server process is running.
// Safe to call standalone as it manages its own locking.
func InternalIsServerRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	status := internalIsServerRunningNoLock()
	config.SetIsGameServerRunning(status)
	return status
}

// internalIsServerRunningNoLock checks if the server process is running.
// Caller M U S T hold mu.Lock().
func internalIsServerRunningNoLock() bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}

	if runtime.GOOS == "windows" {
		select {
		case <-processExited:
			cmd = nil
			clearGameServerUUID()
			return false
		default:
			// Process is still running
			return true
		}
	}

	if runtime.GOOS == "linux" {
		// On Unix-like systems, use Signal(0)
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			logger.Core.Debug("Signal(0) failed, assuming process is dead: " + err.Error())
			cmd = nil
			clearGameServerUUID()
			return false
		}
		return true
	}

	logger.Core.Warn("Failed to check if server is running, assuming it's dead")
	return false
}
