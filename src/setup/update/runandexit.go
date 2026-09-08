package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
)

// runAndExit launches the new executable and terminates the current process
func runAndExit(newExe string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(newExe)
	if err != nil {
		return fmt.Errorf("❌ Couldn’t resolve path to %s: %v", newExe, err)
	}

	// Prepare the new process
	cmd := exec.Command(absPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set SysProcAttr based on OS using the OS-specific implementation
	setSysProcAttr(cmd)

	// Start the new process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("❌ Failed to start new executable: %v", err)
	}

	// Exit gracefully
	logger.Install.Warn("✨ New version’s live! Catch you on the flip side!")
	time.Sleep(500 * time.Millisecond) // Dramatic pause
	os.Exit(0)
	return nil
}

func runAndExitLinux(newExe string) error {
	absPath, err := filepath.Abs(newExe)
	if err != nil {
		return fmt.Errorf("❌ Couldn’t resolve path to %s: %v", newExe, err)
	}

	// Use syscall.Exec to replace the current process
	logger.Install.Warn("✨ New version’s live! Catch you on the flip side!")
	time.Sleep(500 * time.Millisecond)

	// Replace the current process with the new executable
	err = syscall.Exec(absPath, []string{absPath}, os.Environ())
	if err != nil {
		return fmt.Errorf("❌ Failed to exec new executable: %v", err)
	}

	// This line is never reached if Exec succeeds
	return nil
}

func RestartMySelf() { // Stop the game server before restarting to prevent detached processes
	if config.GetIsGameServerRunning() {
		logger.Install.Info("🛑 Stopping game server before restart...")
		if err := gamemgr.InternalStopServer(); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Failed to stop game server before restart: %v. Proceeding anyway.", err))
		}
	}
	currentExe, err := os.Executable()
	if err != nil {
		logger.Install.Warn(fmt.Sprintf("⚠️ Restart failed: couldn’t get current executable path: %v. Keeping version %s.", err, config.GetVersion()))
		return
	}

	if runtime.GOOS == "windows" {
		if err := runAndExit(currentExe); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Restart failed: couldn’t launch %s: %v. Keeping version %s.", currentExe, err, config.GetVersion()))
			return
		}
	}
	if runtime.GOOS == "linux" {
		if err := runAndExitLinux(currentExe); err != nil {
			logger.Install.Warn(fmt.Sprintf("⚠️ Restart failed: couldn’t launch %s: %v. Keeping version %s.", currentExe, err, config.GetVersion()))
			return
		}
	}
}
