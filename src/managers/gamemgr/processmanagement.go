// processmanagement.go
package gamemgr

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

var (
	cmd           *exec.Cmd
	mu            sync.Mutex
	logDone       chan struct{}
	err           error
	processExited chan struct{}
	// autoRestartDone is defined in autorestart.go
)

func InternalStartServer() error {
	mu.Lock()
	defer mu.Unlock()

	if internalIsServerRunningNoLock() {
		return fmt.Errorf("server is already running")
	}

	// Rotate password if enabled (sets new random password before building args)
	rotatePasswordIfEnabled()

	args := buildCommandArgs()

	logger.Core.Info("=== GAMESERVER STARTING ===")

	if config.GetIsSSCMEnabled() && runtime.GOOS == "linux" {

		var envVars []string
		// Set up SSCM (BepInEx/Doorstop) environment
		envVars, err = SetupBepInExEnvironment()
		if err != nil {
			return fmt.Errorf("failed to set up SSCM environment: %v", err)
		}
		// Create command after environment is set
		cmd = exec.Command(config.GetExePath(), args...)
		// Set the environment for the command
		if envVars != nil {
			cmd.Env = envVars
			logger.Core.Info("BepInEx/Doorstop environment configured for server process")
		}
		logger.Core.Info("• Executable: " + config.GetExePath() + " (with SSCM)")
		logger.Core.Info("• Arguments: " + strings.Join(args, " "))
	}

	if !config.GetIsSSCMEnabled() && runtime.GOOS == "linux" {
		// Use ExePath directly as the command
		cmd = exec.Command(config.GetExePath(), args...)
		logger.Core.Info("• Executable: " + config.GetExePath())
		logger.Core.Info("• Arguments: " + strings.Join(args, " "))
	}

	if runtime.GOOS == "windows" {

		// On Windows, set the command to use the executable path and arguments
		cmd = exec.Command(config.GetExePath(), args...)
		logger.Core.Info("• Executable: " + config.GetExePath())
		logger.Core.Debug("Switching to pipes for logs as we are on Windows!")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("error creating StdoutPipe: %v", err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("error creating StderrPipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("error starting server: %v", err)
		}
		logger.Core.Info("• Arguments: " + strings.Join(args, " "))
		logger.Core.Debug("Server process started with PID:" + strconv.Itoa(cmd.Process.Pid))
		logger.Core.Debug("Created pipes")

		// Start reading stdout and stderr pipes on Windows
		go readPipe(stdout)
		go readPipe(stderr)

		// Monitor process exit
		processExited = make(chan struct{})
		go func() {
			err := cmd.Wait()
			if err != nil {
				logger.Core.Debug("Process exited with error: " + err.Error())
			} else {
				logger.Core.Debug("Process exited successfully")
			}
			close(processExited)
		}()
	} else {

		logger.Core.Debug("Switching to log file for logs as we are on Linux! Hail the Penguin!")

		if logDone != nil {
			close(logDone) // Close any existing channel
		}
		logDone = make(chan struct{})
		// On Linux, start the command without pipes since we're using the log file
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("error starting server: %v", err)
		}
		logger.Core.Debug("Server process started with PID:" + strconv.Itoa(cmd.Process.Pid))

		// check if debug.log file exists, if not, create it
		if _, err := os.Stat("./debug.log"); os.IsNotExist(err) {
			file, err := os.OpenFile("./debug.log", os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating debug.log file: %v", err)
			}
			defer file.Close()
		}
		// Start tailing the debug.log file on Linux
		go tailLogFile("./debug.log")
	}
	// create a UUID for this specific run
	SetServerState(ServerStateStarting)
	createGameServerUUID()
	setServerStartTime()
	config.SetIsGameServerRunning(true)

	// Start auto-restart goroutine if AutoRestartServerTimer is set greater than 0
	if config.GetAutoRestartServerTimer() != "0" {
		if autoRestartDone != nil {
			close(autoRestartDone)
		}
		autoRestartDone = make(chan struct{})
		go startAutoRestart(config.GetAutoRestartServerTimer(), autoRestartDone)
		logger.Core.Info("New Auto-restart scheduled: " + config.GetAutoRestartServerTimer())
	}

	return nil
}

func InternalStopServer() error {
	mu.Lock()
	defer mu.Unlock()

	if !internalIsServerRunningNoLock() {
		SetServerState(ServerStateStopped)
		return fmt.Errorf("server not running")
	}
	SetServerState(ServerStateStopping)

	// Stop auto-restart goroutine
	if autoRestartDone != nil {
		close(autoRestartDone)
		autoRestartDone = nil
	}

	// Process is running, stop it
	isWindows := runtime.GOOS == "windows"
	var killErr error

	if isWindows {
		// On Windows, terminate the process (no graceful shutdown)
		killErr = cmd.Process.Kill()
		// Wait for the processExited channel to confirm exit
		if processExited != nil {
			select {
			case <-processExited:
				logger.Core.Debug("processExited channel confirmed server shutdown")
			case <-time.After(2 * time.Second):
				logger.Core.Warn("Timeout waiting for processExited confirmation")
			}
		}
	} else {
		// On Linux/Unix, send SIGTERM for graceful shutdown
		if termErr := cmd.Process.Signal(syscall.SIGTERM); termErr != nil {
			logger.Core.Debug("SIGTERM failed: " + termErr.Error())
			killErr = cmd.Process.Kill() // Fallback to Kill if SIGTERM fails
		} else {
			// Wait for graceful shutdown
			waitErrChan := make(chan error, 1)
			go func() {
				waitErrChan <- cmd.Wait()
			}()

			select {
			case waitErr := <-waitErrChan:
				if waitErr != nil && !strings.Contains(waitErr.Error(), "exit status") {
					logger.Core.Debug("Wait error after SIGTERM: " + waitErr.Error())
				}
			case <-time.After(10 * time.Second): // Increased timeout
				logger.Core.Warn("Timeout waiting for graceful shutdown, sending SIGKILL")
				killErr = cmd.Process.Kill() // Fallback to SIGKILL
				select {
				case waitErr := <-waitErrChan:
					if waitErr != nil && !strings.Contains(waitErr.Error(), "exit status") {
						logger.Core.Debug("Wait error after SIGKILL: " + waitErr.Error())
					}
				case <-time.After(2 * time.Second): // Additional wait for SIGKILL
					return fmt.Errorf("timeout waiting for process to exit after SIGKILL")
				}
			}
		}

		// Stop log tailing (Linux only)
		if logDone != nil {
			close(logDone)
			logDone = nil
		}
	}

	if killErr != nil {
		return fmt.Errorf("error stopping server: %v", killErr)
	}

	// Process is confirmed stopped, clear cmd
	cmd = nil
	config.SetIsGameServerRunning(false)
	SetServerState(ServerStateStopped)
	clearServerStartTime()
	clearGameServerUUID()
	return nil
}
