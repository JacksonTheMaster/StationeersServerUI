package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func (l *Logger) writeToFile(logLine, subsystem string) {
	const maxRetries = 5
	const retryDelay = 100 * time.Millisecond

	// Files to write: combined log + subsystem-specific log
	logFiles := []string{
		config.GetLogFolder() + "ssui.log", // Combined log
		getSubsystemLogPath(subsystem),     // Subsystem log (e.g., logs/install.log)
	}

	for _, logFile := range logFiles {
		for attempt := 0; attempt < maxRetries; attempt++ {
			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm); err != nil {
				fmt.Printf("%s%s [ERROR/LOGGER] Failed to create log file %s: %v%s\n",
					colorRed, time.Now().Format("2006-01-02 15:04:05"), filepath.Dir(logFile), err, colorReset)
				return
			}

			// Open file
			file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer file.Close()
				if _, err := file.WriteString(logLine); err != nil {
					fmt.Printf("%s%s [ERROR/LOGGER] Failed to write to log file %s: %v%s\n",
						colorRed, time.Now().Format("2006-01-02 15:04:05"), logFile, err, colorReset)
				}
				break // Success, move to next file
			}

			// Retry on transient errors
			if os.IsNotExist(err) || os.IsPermission(err) {
				if attempt == maxRetries-1 {
					fmt.Printf("%s%s [ERROR/LOGGER] Gave up writing to log file %s after %d attempts: %v%s\n",
						colorRed, time.Now().Format("2006-01-02 15:04:05"), logFile, maxRetries, err, colorReset)
					break
				}
				time.Sleep(retryDelay)
				continue
			}

			// Non-retryable error
			fmt.Printf("%s%s [ERROR/LOGGER] Failed to open log file %s: %v%s\n",
				colorRed, time.Now().Format("2006-01-02 15:04:05"), logFile, err, colorReset)
			break
		}
	}
}

// getSubsystemLogPath generates path for subsystem-specific log file
func getSubsystemLogPath(subsystem string) string {
	dir := filepath.Dir(config.GetLogFolder())
	// Lowercase subsystem for cleaner filenames (e.g., install.log)
	filename := fmt.Sprintf("%s.log", strings.ToLower(subsystem))
	return filepath.Join(dir, filename)
}
