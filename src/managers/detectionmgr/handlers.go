// handlers.go
package detectionmgr

import (
	"fmt"
	"strings"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/ssestream"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/discordbot"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

/*
Event Handler Subsystem
- Defines default handling logic for detected events
- Formats and routes event notifications to:
  - Terminal output with ANSI coloring
  - SSE stream for web UI
*/

var lastWorldSavedTime time.Time // zero value means never saved

// DefaultHandlers returns a map of event types to default handlers
func DefaultHandlers() map[EventType]Handler {
	return map[EventType]Handler{

		EventCustomDetection: func(event Event) {
			message := fmt.Sprintf("🎮 [Custom Detection] %s", event.Message)
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},

		EventServerReady: func(event Event) {
			message := "🎮 [Gameserver] 🔔 Server is ready to connect!"
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventServerStarting: func(event Event) {
			message := "🎮 [Gameserver] 🕑 Server is starting up..."
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventServerError: func(event Event) {
			message := "🎮 [Gameserver] ⚠️ Server error detected"
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventSettingsChanged: func(event Event) {
			message := fmt.Sprintf("🎮 [Gameserver] ⚙️ %s", event.Message)
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventServerHosted: func(event Event) {
			message := fmt.Sprintf("🎮 [Gameserver] 🌐 %s", event.Message)
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventNewGameStarted: func(event Event) {
			message := fmt.Sprintf("🎮 [Gameserver] 🎲 %s", event.Message)
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventVersionExtracted: func(event Event) {
			message := fmt.Sprintf("🎮 [Gameserver] 📦 Version %s detected", event.Message)
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventServerRunning: func(event Event) {
			message := "🎮 [Gameserver] ✅ Server process has started!"
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventGameManagerReady: func(event Event) {
			message := "🎮 [Gameserver] 🗺️ Game manager initialized; loading map..."
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventSessionStarting: func(event Event) {
			message := "🎮 [Gameserver] 🌐 Starting multiplayer session..."
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventSessionRegistered: func(event Event) {
			message := "🎮 [Gameserver] ✅ Session registered; server is running."
			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},
		EventPlayerConnecting: func(event Event) {
			if event.PlayerInfo != nil {
				message := fmt.Sprintf("🎮 [Gameserver] 🔄 Player %s (SteamID: %s) is connecting...",
					event.PlayerInfo.Username, event.PlayerInfo.SteamID)
				logger.Detection.Info(message)
				ssestream.BroadcastDetectionEvent(message)
				discordbot.SendMessageToEventLogChannel(message)
			}
		},
		EventPlayerReady: func(event Event) {
			if event.PlayerInfo != nil {
				message := fmt.Sprintf("🎮 [Gameserver] ✅ Player %s (SteamID: %s) is ready!",
					event.PlayerInfo.Username, event.PlayerInfo.SteamID)
				logger.Detection.Info(message)
				ssestream.BroadcastDetectionEvent(message)
				discordbot.SendMessageToEventLogChannel(message)
			}
		},
		EventPlayerDisconnect: func(event Event) {
			if event.PlayerInfo != nil {
				message := fmt.Sprintf("🎮 [Gameserver] 👋 Player %s disconnected",
					event.PlayerInfo.Username)
				logger.Detection.Info(message)
				ssestream.BroadcastDetectionEvent(message)
				discordbot.SendMessageToEventLogChannel(message)
			}
		},
		EventWorldSaved: func(event Event) {
			const debounceDuration = 15 * time.Second // since SSCM triggers a HEAD save after an autosave is detected by the Backup Manager, we debounce save messages here to prevent spamming and user confusion.

			now := time.Now()

			// Check if we handled a world save recently
			if now.Sub(lastWorldSavedTime) < debounceDuration {
				return
			}

			lastWorldSavedTime = now

			timeStr := event.Timestamp
			message := fmt.Sprintf("🎮 [Gameserver] 💾 World Saved: ServerTime: %s", timeStr)

			logger.Detection.Info(message)
			ssestream.BroadcastDetectionEvent(message)
			discordbot.SendMessageToEventLogChannel(message)
		},

		// not sure if this Detector still works, so I changed this to SendMessageToEventLogChannel instead of SendMessageToErrorChannel.
		EventException: func(event Event) {
			// Initial alert message
			alertMessage := "🎮 [Gameserver] 🚨 Exception detected!"
			logger.Detection.Info(alertMessage)
			ssestream.BroadcastDetectionEvent(alertMessage)
			discordbot.SendMessageToEventLogChannel(alertMessage)

			if event.ExceptionInfo != nil && len(event.ExceptionInfo.StackTrace) > 0 {
				// Format stack trace as a single-line string for SSE compatibility
				stackTrace := strings.ReplaceAll(event.ExceptionInfo.StackTrace, "\n", " | ")
				message := fmt.Sprintf("Exception Details: Stack Trace: %s", stackTrace)

				logger.Detection.Info(message)
				ssestream.BroadcastDetectionEvent(message)
				discordbot.SendMessageToEventLogChannel(message)
			}
		},
	}
}

// RegisterDefaultHandlers registers all default handlers with a detector
func RegisterDefaultHandlers(detector *Detector) {
	for eventType, handler := range DefaultHandlers() {
		detector.RegisterHandler(eventType, handler)
	}
}
