package logger

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/ssestream"
)

// Logger instances
var (
	Main         = &Logger{suffix: SYS_MAIN}
	Web          = &Logger{suffix: SYS_WEB}
	Discord      = &Logger{suffix: SYS_DISCORD}
	Backup       = &Logger{suffix: SYS_BACKUP}
	Detection    = &Logger{suffix: SYS_DETECT}
	Core         = &Logger{suffix: SYS_CORE}
	Config       = &Logger{suffix: SYS_CONFIG}
	Install      = &Logger{suffix: SYS_INSTALL}
	SSE          = &Logger{suffix: SYS_SSE}
	Security     = &Logger{suffix: SYS_SECURITY}
	Localization = &Logger{suffix: SYS_LOCALIZATION}
	Advertiser   = &Logger{suffix: SYS_ADVERTISER}
	Modding      = &Logger{suffix: SYS_MODDING}
)

// Severity Levels
const (
	DEBUG = 10 // Fine-grained debugging
	INFO  = 20 // Normal operations
	WARN  = 30 // Potential issues
	ERROR = 40 // Critical errors
	CLEAN = 50 // Just the message, no timestamp or severity
)

// Subsystems
const (
	SYS_MAIN         = "MAIN"
	SYS_WEB          = "WEB"
	SYS_DISCORD      = "DISCORD"
	SYS_BACKUP       = "BACKUP"
	SYS_DETECT       = "DETECT"
	SYS_CORE         = "CORE"
	SYS_CONFIG       = "CONFIG"
	SYS_INSTALL      = "INSTALL"
	SYS_SSE          = "SSE"
	SYS_SECURITY     = "SECURITY"
	SYS_LOCALIZATION = "LOCALIZATION"
	SYS_ADVERTISER   = "ADVERTISER"
	SYS_MODDING      = "MODDING"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

// Subsystem color map (distinct colors, cohesive vibe)
var subsystemColors = map[string]string{
	SYS_MAIN:         colorBlue,    // Calm, default system
	SYS_WEB:          colorCyan,    // Clean, UI-related
	SYS_DISCORD:      colorMagenta, // Flashy, chatty subsystem
	SYS_BACKUP:       colorGreen,   // Safe, reliable vibe
	SYS_DETECT:       colorYellow,  // Attention-grabbing for detection
	SYS_CORE:         colorMagenta, // Critical, stands out
	SYS_CONFIG:       colorYellow,  // Warning-like, config tweaks
	SYS_INSTALL:      colorBlue,    // Matches MAIN, setup-related
	SYS_SSE:          colorCyan,    // Matches WEB, streaming vibe
	SYS_SECURITY:     colorRed,     // Screams "pay attention"
	SYS_LOCALIZATION: colorCyan,    // Matches WEB, localization-related
	SYS_ADVERTISER:   colorYellow,  // Matches Config, advanced feature
	SYS_MODDING:      colorCyan,    //
}

// Global channels and mutex for all loggers
var (
	globalLogChan     chan logEntry
	globalConsoleChan chan string
	globalOnce        sync.Once
)

type Logger struct {
	mu     sync.Mutex
	suffix string // Subsystem identifier (e.g., "DISCORD")
}

type logEntry struct {
	severity    int
	suffix      string // Log type (e.g., "INFO", "DEBUG")
	color       string
	message     string
	consoleLine string
	fileLine    string
	logger      *Logger // Reference to the logger for file writing
}

// Init initializes the global logging and console output goroutines
func (l *Logger) Init() {
	globalOnce.Do(func() {
		globalLogChan = make(chan logEntry, 1000) // Buffered global channel for log processing
		globalConsoleChan = make(chan string, 50) // Buffered global channel for console output
		go processLogs()
		go processConsoleOutput()
	})
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...any) {
	l.log(logEntry{DEBUG, "DEBUG", colorReset, fmt.Sprintf(format, args...), "", "", l})
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...any) {
	l.log(logEntry{INFO, "INFO", colorReset, fmt.Sprintf(format, args...), "", "", l})
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...any) {
	l.log(logEntry{WARN, "WARN", colorYellow, fmt.Sprintf(format, args...), "", "", l})
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...any) {
	l.log(logEntry{ERROR, "ERROR", colorRed, fmt.Sprintf(format, args...), "", "", l})
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(logEntry{DEBUG, "DEBUG", colorReset, message, "", "", l})
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(logEntry{INFO, "INFO", colorReset, message, "", "", l})
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(logEntry{WARN, "WARN", colorYellow, message, "", "", l})
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(logEntry{ERROR, "ERROR", colorRed, message, "", "", l})
}

// Error logs an error message
func (l *Logger) Clean(message string) {
	l.log(logEntry{CLEAN, "", "", message, "", "", l})
}

// Errorf logs a formatted error message
func (l *Logger) Cleanf(format string, args ...any) {
	l.log(logEntry{CLEAN, "", "", fmt.Sprintf(format, args...), "", "", l})
}

// log sends the log entry to the global channel
func (l *Logger) log(entry logEntry) {
	l.mu.Lock()
	if !l.shouldLog(entry.severity) {
		l.mu.Unlock()
		return
	}

	// Initialize global channels if not already done
	l.Init()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	// Handle CLEAN severity separately
	if entry.severity == CLEAN {
		entry.consoleLine = fmt.Sprintf("%s\n", entry.message)
		entry.fileLine = fmt.Sprintf("%s\n", entry.message)
	} else {
		// Use subsystem color by default, override with severity color if set
		entryColor := subsystemColors[l.suffix]
		if entry.color != colorReset {
			entryColor = entry.color
		}
		entry.consoleLine = fmt.Sprintf("%s%s [%s/%s] %s%s\n", entryColor, timestamp, l.suffix, entry.suffix, entry.message, colorReset)
		entry.fileLine = fmt.Sprintf("%s [%s/%s] %s\n", timestamp, l.suffix, entry.suffix, entry.message)
	}
	l.mu.Unlock()

	// Send to global log channel (non-blocking unless channel is full)
	select {
	case globalLogChan <- entry:
	default:
		// Channel full, log to stderr
		fmt.Fprintf(os.Stderr, "%s%s [ERROR/LOGGER] Log channel full, dropping: %s%s\n",
			colorRed, timestamp, entry.fileLine, colorReset)
	}
}

// processLogs handles SSE broadcasts and file output for all loggers
func processLogs() {
	for entry := range globalLogChan {
		// Broadcast to SSE streams
		if entry.severity >= DEBUG {
			ssestream.BroadcastDebugLog(entry.fileLine)
		}
		if entry.severity == INFO {
			ssestream.BroadcastInfoLog(entry.fileLine)
		}
		if entry.severity == WARN {
			ssestream.BroadcastWarnLog(entry.fileLine)
		}
		if entry.severity == ERROR {
			ssestream.BroadcastErrorLog(entry.fileLine)
		}
		ssestream.BroadcastBackendLog(entry.fileLine)

		// File output if enabled
		if config.GetCreateSSUILogFile() {
			entry.logger.writeToFile(entry.fileLine, entry.logger.suffix)
		}

		// Send to global console channel
		select {
		case globalConsoleChan <- entry.consoleLine:
		default:
			if runtime.GOOS == "windows" {
				ssestream.BroadcastErrorLog("ATTENTION: WINDOWS-RELATED ISSUE: THE TERMINAL WHERE SSUI IS RUNNING IS NO LONGER ACCEPTING MESSAGES. PLEASE CHECK THE TERMINAL AND PRESS ENTER TO FREE THE BUFFER.")
				ssestream.BroadcastConsoleOutput("ATTENTION: WINDOWS-RELATED ISSUE: THE TERMINAL WHERE SSUI IS RUNNING IS NO LONGER ACCEPTING MESSAGES. PLEASE CHECK THE TERMINAL AND PRESS ENTER TO FREE THE BUFFER.")
			}
		}
	}
}

// processConsoleOutput handles console output for all loggers
func processConsoleOutput() {
	for consoleLine := range globalConsoleChan {
		fmt.Print(consoleLine)
	}
}

// shouldLog checks severity and subsystem filters
func (l *Logger) shouldLog(severity int) bool {
	if len(config.GetSubsystemFilters()) > 0 {
		allowed := false
		for _, sub := range config.GetSubsystemFilters() {
			if sub == l.suffix {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	return severity >= config.GetLogLevel()
}
