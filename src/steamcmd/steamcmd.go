package steamcmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

var steamMu sync.Mutex
var isUpdatingMu sync.Mutex

// ExtractorFunc is a type that represents a function for extracting archives.
// It takes an io.ReaderAt, the size of the content, and the destination directory.
type ExtractorFunc func(io.ReaderAt, int64, string) error

// Constants for repeated strings
const (
	SteamCMDLinuxURL   = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"
	SteamCMDWindowsURL = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip"
	SteamCMDLinuxDir   = "./steamcmd"
	SteamCMDWindowsDir = "C:\\SteamCMD"
)

// isIPv6Enabled returns a heuristic whether the system has non-loopback IPv6 addresses configured.
// Used for diagnostic warning on SteamCMD exit 2 (0x2).
func isIPv6Enabled() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() == nil && !ip.IsLoopback() {
				return true
			}
		}
	}
	return false
}

// InstallAndRunSteamCMD installs and runs SteamCMD based on the platform (Windows/Linux).
// It returns the exit status of the SteamCMD execution and any error encountered.
func InstallAndRunSteamCMD() (int, error) {
	if isUpdatingMu.TryLock() {
		// Successfully acquired the lock; we are not updating currently
		logger.Core.Debug("🔄 Locking isUpdatingMu for SteamCMD Update run...")
	} else {
		// already updating, return
		logger.Core.Warn("🔄 isUpdatingMu is currently locked, cannot update server using SteamCMD right now...")
		return -1, fmt.Errorf("already updating")
	}
	defer isUpdatingMu.Unlock()
	defer logger.Core.Debug("🔄 Unlocking isUpdatingMu after SteamCMD Update run...")

	if gamemgr.InternalIsServerRunning() {
		logger.Core.Warn("Server is running, stopping server first...")
		err := gamemgr.InternalStopServer()
		if err != nil {
			logger.Core.Error("Error stopping server before running Steamcmd: " + err.Error())
		}
	}
	logger.Core.Info("Running SteamCMD")

	switch runtime.GOOS {
	case "windows":
		return installSteamCMDWindows()
	case "linux":
		return installSteamCMDLinux()
	default:
		err := fmt.Errorf("SteamCMD installation is not supported on this OS")
		logger.Install.Error("❌ " + err.Error() + "\n")
		return -1, err
	}
}

// runSteamCMD runs the SteamCMD command to update the game and returns its exit status and any error.
func runSteamCMD(steamCMDDir string) (int, error) {
	if steamMu.TryLock() {
		// Successfully acquired the lock; no other func holds it
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD execution...")
	} else {
		// Another goroutine holds the lock; log and wait.
		logger.Core.Warn("🔄 SteamMu is currently locked, waiting for it to be unlocked and then continuing...")
		steamMu.Lock() // Block until steamMu becomes available, then snag it and lock it again
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD execution..")
	}
	defer steamMu.Unlock()
	defer logger.Core.Debug("🔄 Unlocking SteamMu after SteamCMD execution...")
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Install.Error("❌ Error getting current working directory: " + err.Error() + "\n")
		return -1, err
	}
	logger.Install.Debug("✅ Current working directory: " + currentDir + "\n")

	// Ensure permissions every time if we run on linux
	if runtime.GOOS != "windows" {
		if err := setExecutablePermissions(steamCMDDir); err != nil {
			logger.Install.Error("❌ Error setting executable permissions, your Steamcmd install might be broken: " + err.Error() + "\n")
			return -1, err
		}
	}

	// Build the initial SteamCMD command
	cmd := buildSteamCMDCommand(steamCMDDir, currentDir)

	// Set output to stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Apply Linux-specific HOME environment variable override
	if runtime.GOOS == "linux" {
		env := os.Environ()
		newEnv := make([]string, 0, len(env)+1)
		foundHome := false
		for _, e := range env {
			if !strings.HasPrefix(e, "HOME=") {
				newEnv = append(newEnv, e)
			} else {
				newEnv = append(newEnv, "HOME="+currentDir)
				foundHome = true
			}
		}
		if !foundHome {
			newEnv = append(newEnv, "HOME="+currentDir)
		}
		cmd.Env = newEnv
	}

	if config.GetLogLevel() == 10 {
		cmdString := strings.Join(cmd.Args, " ")
		logger.Install.Info("🕑 Running SteamCMD: " + cmdString)
	} else {
		logger.Install.Info("🕑 Running SteamCMD...")
	}

	// Retry loop: maximum 2 attempts, with retry only on exit status 8
	var exitCode int = -1
	var runErr error

	for attempt := 1; attempt <= 2; attempt++ {
		runErr = cmd.Run()

		if runErr == nil {
			// Success!
			logger.Install.Info("✅ SteamCMD executed successfully.\n")
			return 0, nil
		}

		// Check if it's an ExitError so we can inspect the code
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			logger.Install.Error("❌ SteamCMD exited unsuccessfully: " + runErr.Error() + "\n")

			if exitCode == 8 && attempt == 1 {
				logger.Install.Warn("⚠️ SteamCMD failed with exit status 8 on first attempt. Retrying once...")
				// Rebuild a fresh command for the retry
				cmd = buildSteamCMDCommand(steamCMDDir, currentDir)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				// Re-apply Linux env modifications
				if runtime.GOOS == "linux" {
					env := os.Environ()
					newEnv := make([]string, 0, len(env)+1)
					foundHome := false
					for _, e := range env {
						if !strings.HasPrefix(e, "HOME=") {
							newEnv = append(newEnv, e)
						} else {
							newEnv = append(newEnv, "HOME="+currentDir)
							foundHome = true
						}
					}
					if !foundHome {
						newEnv = append(newEnv, "HOME="+currentDir)
					}
					cmd.Env = newEnv
				}

				// Log the retry
				if config.GetLogLevel() == 10 {
					cmdString := strings.Join(cmd.Args, " ")
					logger.Install.Info("🕑 Retrying SteamCMD: " + cmdString)
				} else {
					logger.Install.Info("🕑 Retrying SteamCMD...")
				}

				continue // Go to next attempt
			}

			// If we get here: either not exit 8, or it was exit 8 on the second attempt
			if exitCode == 8 {
				logger.Install.Error("   ⚠️ Exit status 8 persisted after retry. Please restart SSUI and try again. If the issue persists, feel free to ask for help on the SSUI Discord server or GitHub issues page.")
			}

			// #131: specific diagnostic for exit 2 / 0x2 + IPv6
			if exitCode == 2 && isIPv6Enabled() {
				logger.Install.Warn("⚠️ SteamCMD failed with exit code 2 (0x2) and IPv6 is detected as enabled on this host. IPv6 is a frequent root cause for Steam exit code 0x2. Consider disabling IPv6 (temporarily) for the affected network interface, or configure the system to prefer IPv4 for SteamCMD.")
			}
		} else {
			// Not an ExitError (e.g., command not found, permission denied, etc.)
			logger.Install.Error("❌ Error running SteamCMD: " + runErr.Error() + "\n")
			exitCode = -1
		}

		// If we reach here, the command failed and we're not retrying
		break
	}

	// Final return after failure (with or without retry)
	return exitCode, runErr
}

// buildSteamCMDCommand constructs the SteamCMD command based on the OS.
func buildSteamCMDCommand(steamCMDDir, currentDir string) *exec.Cmd {
	//print the config.GameBranch and config.GameServerAppID
	logger.Install.Info("🔍 Game Branch: " + config.GetGameBranch())
	logger.Install.Debug("🔍 Game Server App ID: " + config.GetGameServerAppID())

	if runtime.GOOS == "windows" {
		return exec.Command(filepath.Join(steamCMDDir, "steamcmd.exe"), "+force_install_dir", currentDir, "+login", "anonymous", "+app_update", config.GetGameServerAppID(), "-beta", config.GetGameBranch(), "validate", "+quit")
	}
	return exec.Command(filepath.Join(steamCMDDir, "steamcmd.sh"), "+force_install_dir", currentDir, "+login", "anonymous", "+app_update", config.GetGameServerAppID(), "-beta", config.GetGameBranch(), "validate", "+quit")
}
