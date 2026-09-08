package steamcmd

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/commandmgr"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
)

var (
	branches     = make(map[string]string)
	branchesLock sync.RWMutex          // Protects branches map for concurrent access
	stopPoller   = make(chan struct{}) // Channel to signal poller cancellation
)

// InitAppInfoPoller starts a goroutine that periodically fetches app info and stops any previous poller.
func AppInfoPoller() {
	// Signal previous poller to stop
	select {
	case stopPoller <- struct{}{}:
		// Previous poller was signaled to stop
	default:
		// No previous poller running
	}

	// if AutoGameServerUpdates is disabled, dont start the poller.
	if !config.GetAllowAutoGameServerUpdates() {
		return
	}
	// Start new poller
	go func() {
		for {
			select {
			case <-stopPoller:
				logger.Install.Debug("🛑 Previous app info poller stopped")
				return
			default:
				err := getAppInfo()
				if err != nil {
					logger.Install.Warn("❌ Failed to get Update info: " + err.Error())
				}
				select {
				case <-stopPoller:
					logger.Install.Debug("🛑 App info poller stopped")
					return
				case <-time.After(10 * time.Minute):
					// Continue to next iteration
				}
			}
		}
	}()
}

// getAppInfo fetches the branches and their build IDs for the specified app ID using SteamCMD
// and stores them in the package-level branches map.
func getAppInfo() error {

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Install.Error("❌ Error getting current working directory: " + err.Error() + "\n")
		return err
	}

	if steamMu.TryLock() {
		// Successfully acquired the lock; no other func holds it
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD AppInfo...")
	} else {
		// Another goroutine holds the lock; log and wait.
		logger.Core.Warn("🔄 SteamMu is currently locked, waiting for it to be unlocked and then continuing...")
		steamMu.Lock() // Block until steamMu becomes available, then snack it and lock it again
		logger.Core.Debug("🔄 Locking SteamMu for SteamCMD AppInfo...")
	}
	steamcmddir := SteamCMDLinuxDir
	executable := "steamcmd.sh"
	appid := config.GetGameServerAppID()

	if runtime.GOOS == "windows" {
		executable = "steamcmd.exe"
		steamcmddir = SteamCMDWindowsDir
	}

	// Build SteamCMD command with +app_info_update to ensure fresh data
	cmd := exec.Command(filepath.Join(steamcmddir, executable), "+login", "anonymous", "+app_info_update", "1", "+app_info_print", appid, "+quit")

	// Capture output instead of printing directly
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if runtime.GOOS == "linux" {
		env := os.Environ()
		// Replace or set HOME
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

	// Log the command
	//if config.GetLogLevel() == 10 {
	//	cmdString := strings.Join(cmd.Args, " ")
	//	logger.Install.Debug("🕑 Running SteamCMD for app info: " + cmdString)
	//}
	// Run the command
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Install.Errorf("❌ SteamCMD app info failed (code %d): %s\n", exitErr.ExitCode(), stderr.String())
			return fmt.Errorf("SteamCMD app info failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		logger.Install.Errorf("❌ Error running SteamCMD app info: %s\n", err.Error())
		return fmt.Errorf("failed to run SteamCMD app info: %w", err)
	}

	steamMu.Unlock()
	logger.Core.Debug("🔄 Unlocking SteamMu after SteamCMD AppInfo...")

	// Extract branches and build IDs
	newBranches, err := extractBranches(stdout.String())
	if err != nil {
		logger.Install.Debug("❌ Failed to extract branches: " + err.Error() + "\n")
		return err
	}

	// Update package-level branches map
	branchesLock.Lock()
	maps.Copy(branches, newBranches)
	branchesLock.Unlock()
	wasRunning := false
	currentBranch := config.GetGameBranch()
	if buildID, ok := branches[currentBranch]; ok {
		if config.GetCurrentBranchBuildID() != "" && config.GetCurrentBranchBuildID() != buildID {
			logger.Install.Info("❗New gameserver update detected!")
			if config.GetAllowAutoGameServerUpdates() {
				logger.Install.Info("🔍 Updating gameserver via SteamCMD...")
				if gamemgr.InternalIsServerRunning() {
					commandmgr.WriteCommand("announce Update found, stopping server in 60 seconds...")
					logger.Install.Info("❗Stopping server in 60 seconds...")
					time.Sleep(20 * time.Second)
					commandmgr.WriteCommand("announce Update found, stopping server in 40 seconds...")
					time.Sleep(10 * time.Second)
					commandmgr.WriteCommand("announce Update found, stopping server in 30 seconds...")
					time.Sleep(10 * time.Second)
					commandmgr.WriteCommand("announce Update found, saving and stopping server in 20 seconds...")
					time.Sleep(20 * time.Second)

					commandmgr.WriteCommand("SAVE")
					time.Sleep(5 * time.Second)
					commandmgr.WriteCommand("announce Update found, STOPPING SERVER NOW. World was Saved. ")
					time.Sleep(5 * time.Second)
					gamemgr.InternalStopServer()
					wasRunning = true
				}
				_, err := InstallAndRunSteamCMD()
				if err != nil {
					logger.Install.Error("❌ Failed to update gameserver: " + err.Error() + "\n")
				}
				if wasRunning {
					gamemgr.InternalStartServer()
				}
			}
		}
		config.SetCurrentBranchBuildID(buildID)
	}

	return nil
}

// extractBranches uses regex to extract branch names and their build IDs from SteamCMD output.
func extractBranches(output string) (map[string]string, error) {
	// Regex to match the entire branches section, handling nested braces
	pattern := regexp.MustCompile(`"branches"\s*\{([\s\S]*?)\}\s*\}`)
	branchSection := pattern.FindStringSubmatch(output)
	if len(branchSection) < 2 {
		return nil, fmt.Errorf("branches section not found in output")
	}

	// Regex to match individual branch blocks
	branchPattern := regexp.MustCompile(`"([^"]+)"\s*\{[\s\S]*?"buildid"\s*"(\d+)"[\s\S]*?\}`)
	matches := branchPattern.FindAllStringSubmatch(branchSection[1], -1)

	branches := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			branches[match[1]] = match[2]
		}
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("no branches with build IDs found")
	}

	return branches, nil
}
