// sseutils.go
package ssestream

import (
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

// Global managers for SSE streams
var (
	ConsoleStreamManager    = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	EventStreamManager      = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	DebugLogStreamManager   = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	InfoLogStreamManager    = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	WarnLogStreamManager    = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	ErrorLogStreamManager   = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
	BackendLogStreamManager = NewSSEManager(config.GetMaxSSEConnections(), config.GetSSEMessageBufferSize())
)

// BroadcastConsoleOutput sends log to all connected console log clients
func BroadcastConsoleOutput(message string) {
	ConsoleStreamManager.Broadcast(message)
}

// BroadcastDetectionEvent sends an event to all connected clients
func BroadcastDetectionEvent(message string) {
	EventStreamManager.Broadcast(message)
}

// BroadcastDebugLog sends an event to all connected clients
func BroadcastDebugLog(message string) {
	DebugLogStreamManager.Broadcast(message)
}

// BroadcastInfoLog sends an event to all connected clients
func BroadcastInfoLog(message string) {
	InfoLogStreamManager.Broadcast(message)
}

// BroadcastWarnLog sends an event to all connected clients
func BroadcastWarnLog(message string) {
	WarnLogStreamManager.Broadcast(message)
}

// BroadcastErrorLog sends an event to all connected clients
func BroadcastErrorLog(message string) {
	ErrorLogStreamManager.Broadcast(message)
}

// BroadcastInternalLog sends an event to all connected clients
func BroadcastBackendLog(message string) {
	BackendLogStreamManager.Broadcast(message)
}
