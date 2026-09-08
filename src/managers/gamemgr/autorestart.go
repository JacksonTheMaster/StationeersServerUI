package gamemgr

import (
	"fmt"
	"strconv"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/commandmgr"
)

var (
	autoRestartDone chan struct{}
	// other local vars are defined in processmanagement.go
)

// sendAutoRestartWarnings sends in-game restart warnings via the `announce` command (SSCM).
// Duration is controlled by AutoRestartCountdown config (default 60s). Addresses #171 + announce integration.
func sendAutoRestartWarnings() {
	if !config.GetIsSSCMEnabled() {
		return
	}
	countStr := config.GetAutoRestartCountdown()
	lead, err := strconv.Atoi(countStr)
	if err != nil || lead < 5 {
		lead = 60
	}

	// Initial warning
	commandmgr.WriteCommand(fmt.Sprintf("announce Attention, server is restarting in %d seconds!", lead))

	// Step down ~every 10s
	remaining := lead
	for remaining > 10 {
		sleep := 10
		if remaining-10 < 10 && remaining-10 > 0 {
			sleep = remaining - 10
		}
		time.Sleep(time.Duration(sleep) * time.Second)
		remaining -= sleep
		if remaining > 0 {
			commandmgr.WriteCommand(fmt.Sprintf("announce Attention, server is restarting in %d seconds!", remaining))
		}
	}

	// Final save + short countdown (approx lands near 10s/5s)
	time.Sleep(5 * time.Second)
	commandmgr.WriteCommand("announce Attention, server is restarting in 10 seconds, saving world now!")
	commandmgr.WriteCommand("save")
	time.Sleep(5 * time.Second)
	commandmgr.WriteCommand("announce Attention, server is restarting in 5 seconds!")
	time.Sleep(5 * time.Second)
}

// startAutoRestart runs a goroutine that restarts the server either after a specified duration in minutes
// or at a specific time of day (HH:MM) every day.
func startAutoRestart(schedule string, done chan struct{}) {
	// Try parsing as a time in HH:MM format
	if t, err := time.Parse("15:04", schedule); err == nil {
		// Valid HH:MM format, schedule daily restart
		setNextDailyRestartTime(t)
		go scheduleDailyRestart(t, done)
		return
	}

	// Try parsing as a time in HH:MMAM/PM format
	if t, err := time.Parse("03:04PM", schedule); err == nil {
		// Valid HH:MMAM/PM format, schedule daily restart
		setNextDailyRestartTime(t)
		go scheduleDailyRestart(t, done)
		return
	}

	// Fallback to parsing as minutes duration
	minutesInt, err := strconv.Atoi(schedule)
	if err != nil {
		logger.Core.Error("Invalid AutoRestartServerTimer format: " + schedule)
		return
	}
	if minutesInt <= 0 {
		logger.Core.Error("AutoRestartServerTimer must be a positive number of minutes or valid HH:MM or HH:MMAM/PM time")
		return
	}

	config.SetNextAutoRestartTime(time.Now().Add(time.Duration(minutesInt) * time.Minute))

	ticker := time.NewTicker(time.Duration(minutesInt) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if !internalIsServerRunningNoLock() {
				mu.Unlock()
				logger.Core.Info("Auto-restart skipped: server is not running")
				return
			}
			mu.Unlock()

			sendAutoRestartWarnings()
			logger.Core.Info("Auto-restart triggered: stopping server")
			if err := InternalStopServer(); err != nil {
				logger.Core.Error("Auto-restart failed to stop server: " + err.Error())
				return
			}

			logger.Core.Info("Auto-restart: waiting 5 seconds before restarting")
			time.Sleep(5 * time.Second)

			logger.Core.Info("Auto-restart: starting server")
			if err := InternalStartServer(); err != nil {
				logger.Core.Error("Auto-restart failed to start server: " + err.Error())
				return
			}
		case <-done:
			config.SetNextAutoRestartTime(time.Time{})
			return
		}
	}
}

// scheduleDailyRestart schedules a server restart at the specified time of day (HH:MM) every day.
func scheduleDailyRestart(t time.Time, done chan struct{}) {
	// Extract hour and minute from the parsed time
	hour, min := t.Hour(), t.Minute()

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
		if now.After(next) || now.Equal(next) {
			// If the time is in the past or now, schedule for tomorrow
			next = next.Add(24 * time.Hour)
		}
		duration := next.Sub(now)

		// Wait until the next restart time or until interrupted
		timer := time.NewTimer(duration)
		select {
		case <-timer.C:
			mu.Lock()
			if !internalIsServerRunningNoLock() {
				mu.Unlock()
				logger.Core.Info("Auto-restart skipped: server is not running")
				// Schedule next day
				setNextDailyRestartTime(t)
				continue
			}
			mu.Unlock()

			sendAutoRestartWarnings()
			logger.Core.Info("Daily auto-restart triggered: stopping server")
			if err := InternalStopServer(); err != nil {
				logger.Core.Error("Daily auto-restart failed to stop server: " + err.Error())
				// Schedule next day
				setNextDailyRestartTime(t)
				continue
			}

			logger.Core.Debug("Daily auto-restart: waiting 5 seconds before restarting")
			time.Sleep(5 * time.Second)

			logger.Core.Info("Daily auto-restart: starting server")
			if err := InternalStartServer(); err != nil {
				logger.Core.Error("Daily auto-restart failed to start server: " + err.Error())
				continue
			}
		case <-done:
			timer.Stop()
			config.SetNextAutoRestartTime(time.Time{})
			return
		}
	}
}

// setNextDailyRestartTime calculates and stores the next daily restart time.
func setNextDailyRestartTime(t time.Time) {
	hour, min := t.Hour(), t.Minute()
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
	if now.After(next) || now.Equal(next) {
		next = next.Add(24 * time.Hour)
	}
	config.SetNextAutoRestartTime(next)
}
