package gamemgr

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/ssestream"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/google/uuid"
)

const defaultLogFolderMode = 0755
const defaultLogFileMode = 0644

var (
	gameServerLogFileFolder string
	gameServerLogFilePath   string
)

func setGameServerLogFilePath(serverUUID uuid.UUID) {
	gameServerLogFileFolder = config.GetLogFolder()
	logFileName := fmt.Sprintf("serverlog_%s_%s.log", time.Now().Format("200601021504"), serverUUID.String())
	gameServerLogFilePath = path.Join(gameServerLogFileFolder, logFileName)
}

func clearGameServerLogFilePath() {
	gameServerLogFileFolder = ""
	gameServerLogFilePath = ""
}

// readPipe for Windows
func readPipe(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	logger.Core.Debug("Started reading pipe")
	for scanner.Scan() {
		output := scanner.Text()
		ssestream.BroadcastConsoleOutput(output)
		logToFile(output)
	}
	if err := scanner.Err(); err != nil {
		logger.Core.Debug("Pipe error: " + err.Error())
		ssestream.BroadcastConsoleOutput(fmt.Sprintf("Error reading pipe: %v", err))
	}
	logger.Core.Debug("Pipe closed")
}

// tailLogFile uses tail to read the log file because using the gameserver's output in pipes to read the serverlog doesn't work on Linux with the Stationeers gameserver.
// I didn't manage to implement proper file tailing (tail behavior) here in go, so I opted to just use the actual tail.. This is a workaround for a workaround.

func tailLogFile(logFilePath string) {
	//if we somehow end up running THIS on windows, hard error and shutdown as the whole point of this software is to read the logs and do stuff with them.
	if runtime.GOOS == "windows" {
		logger.Core.Error("[MAJOR ISSUE DETECTED] Windows detected while trying to read log files the Linux way, skipping. You might wanna check your environment, as this should not happen.")
		ssestream.BroadcastConsoleOutput("[MAJOR ISSUE DETECTED] Windows detected while trying to read log files the Linux way, skipping. You might wanna check your environment, as this should not happen.")
		logger.Core.Error("[MAJOR ISSUE DETECTED] Shutting down...")
		ssestream.BroadcastConsoleOutput("[MAJOR ISSUE DETECTED] Shutting down...")
		os.Exit(1)
	}

	// Wait and retry until the log file exists
	for i := range 10 { // Retry up to 10 times
		if _, err := os.Stat(logFilePath); err == nil {
			break // File exists, proceed
		}
		logger.Core.Debug("Log file " + logFilePath + " not found, retrying in 1s (" + strconv.Itoa(i+1) + "/10)")
		time.Sleep(1 * time.Second)
	}

	// If file still doesn't exist, give up and report
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		logger.Core.Debug("Log file " + logFilePath + " still not found after retries")
		ssestream.BroadcastConsoleOutput(fmt.Sprintf("Log file %s not found after retries", logFilePath))
		return
	}

	// Start tail -F (robust against rotation)
	cmd := exec.Command("tail", "-F", logFilePath)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Core.Debug("Error creating stdout pipe for tail: " + err.Error())
		ssestream.BroadcastConsoleOutput(fmt.Sprintf("Error starting tail -F: %v", err))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		logger.Core.Debug("Error starting tail -F: " + err.Error())
		ssestream.BroadcastConsoleOutput(fmt.Sprintf("Error starting tail -F: %v", err))
		return
	}

	// Clean up when done
	defer func() {
		cmd.Process.Kill() // Kill tail when logDone triggers
		if err := cmd.Wait(); err != nil {
			logger.Core.Debug("Tail process exited with: " + err.Error())
		}
	}()

	scanner := bufio.NewScanner(pipe)
	logger.Core.Debug("Started tailing log file with tail -F")

	// Goroutine to read and broadcast tail output
	go func() {
		defer pipe.Close() // Close pipe when goroutine exits
		for scanner.Scan() {
			output := scanner.Text()
			ssestream.BroadcastConsoleOutput(output)
			logToFile(output)
		}
		if err := scanner.Err(); err != nil {

			logger.Core.Debug("Error reading tail -F output: " + err.Error())

			ssestream.BroadcastConsoleOutput(fmt.Sprintf("Error reading tail -F output: %v", err))
		}
	}()

	// Wait for logDone signal to stop
	<-logDone

	logger.Core.Debug("Received logDone signal, stopping tail -F")

}

func logToFile(message string) {
	if config.GetCreateGameServerLogFile() {
		logFileFolder := gameServerLogFileFolder
		if logFileFolder == "" {
			logFileFolder = config.GetLogFolder()
		}
		// Check if the log folder exists, if not create it
		_, err := os.Stat(logFileFolder)
		if os.IsNotExist(err) {
			err := os.Mkdir(logFileFolder, defaultLogFolderMode)
			if err != nil {
				logger.Core.Error("Error creating log folder: " + err.Error())
				return
			}
		} else if err != nil {
			logger.Core.Error("Error checking log folder: " + err.Error())
			return
		}

		if GameServerUUID != uuid.Nil {
			// append the log file to the log folder
			file, err := os.OpenFile(gameServerLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, defaultLogFileMode)
			if err != nil {
				logger.Core.Error("Error opening log file: " + err.Error())
				return
			}
			defer file.Close()

			_, err = file.WriteString(message + "\n")
			if err != nil {
				logger.Core.Error("Error writing to log file: " + err.Error())
				return
			}
		} else {
			logger.Core.Error("Game Server UUID not set, cannot log to file")
			return
		}
	}
}
