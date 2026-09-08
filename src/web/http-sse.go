package web

import (
	"net/http"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/ssestream"
)

// handler for the /console endpoint
func GetLogOutput(w http.ResponseWriter, r *http.Request) {
	StartConsoleStream()(w, r)
}

// handler for the /console endpoint
func GetEventOutput(w http.ResponseWriter, r *http.Request) {
	StartDetectionEventStream()(w, r)
}

func GetDebugLogOutput(w http.ResponseWriter, r *http.Request) {
	StartDebugLogStream()(w, r)
}

func GetInfoLogOutput(w http.ResponseWriter, r *http.Request) {
	StartInfoLogStream()(w, r)
}

func GetWarnLogOutput(w http.ResponseWriter, r *http.Request) {
	StartWarnLogStream()(w, r)
}

func GetErrorLogOutput(w http.ResponseWriter, r *http.Request) {
	StartErrorLogStream()(w, r)
}

func GetBackendLogOutput(w http.ResponseWriter, r *http.Request) {
	StartBackendLogStream()(w, r)
}

// StartConsoleStream creates an HTTP handler for console log SSE streaming
func StartConsoleStream() http.HandlerFunc {
	return ssestream.ConsoleStreamManager.CreateStreamHandler("Console")
}

// StartDetectionEventStream creates an HTTP handler for detection event SSE streaming
func StartDetectionEventStream() http.HandlerFunc {
	return ssestream.EventStreamManager.CreateStreamHandler("Event")
}

func StartDebugLogStream() http.HandlerFunc {
	return ssestream.DebugLogStreamManager.CreateStreamHandler("Debug Log")
}

func StartInfoLogStream() http.HandlerFunc {
	return ssestream.InfoLogStreamManager.CreateStreamHandler("Info Log")
}

func StartWarnLogStream() http.HandlerFunc {
	return ssestream.WarnLogStreamManager.CreateStreamHandler("Warn Log")
}

func StartErrorLogStream() http.HandlerFunc {
	return ssestream.ErrorLogStreamManager.CreateStreamHandler("Error Log")
}

func StartBackendLogStream() http.HandlerFunc {
	return ssestream.BackendLogStreamManager.CreateStreamHandler("Full Backend Log")
}
