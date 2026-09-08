// ssemanager.go
package ssestream

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

// The SSE blocking issue is NOT related to the backend; the API handles 200 clients per channel fine.
// The issue lies in the JS frontend, where browsers cap HTTP/1.1 SSE streams at ~6 concurrent connections per domain.
// See RFC 2616 (HTTP/1.1), Section 8.1.4: https://datatracker.ietf.org/doc/html/rfc2616#section-8.1.4
// SSE spec (WHATWG HTML): https://html.spec.whatwg.org/multipage/server-sent-events.html#server-sent-events
// Workaround: Use HTTP/2 (RFC 7540, Section 5.1.2) for up to 100 streams: https://datatracker.ietf.org/doc/html/rfc7540#section-5.1.2
// I spent way too much time on this thinking it would be the backend blocking requests. It's not.

// Client represents a connected SSE client
type Client struct {
	messages chan string
	lastSeen time.Time
}

// SSEManager manages Server-Sent Event streams
type SSEManager struct {
	clients            map[*Client]bool
	clientsMu          sync.RWMutex
	maxClients         int
	maxBuffer          int
	kinematicDropCount int
	lastKinematicLog   time.Time
	dropMu             sync.Mutex
}

// NewSSEManager creates a new SSE stream manager
func NewSSEManager(maxClients, maxBuffer int) *SSEManager {
	return &SSEManager{
		clients:    make(map[*Client]bool),
		maxClients: maxClients,
		maxBuffer:  maxBuffer,
	}
}

// CreateStreamHandler creates an HTTP handler for SSE streaming
func (m *SSEManager) CreateStreamHandler(streamType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Default to the server's origin if none provided
			serverOrigin := r.Host
			if !strings.HasPrefix(serverOrigin, "http") {
				if r.TLS != nil {
					serverOrigin = "https://" + serverOrigin
				} else {
					serverOrigin = "http://" + serverOrigin
				}
			}
			w.Header().Set("Access-Control-Allow-Origin", serverOrigin)
		}

		// Allow credentials
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Ensure the response writer supports flushing
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Check maximum client limit
		m.clientsMu.Lock()
		if len(m.clients) >= m.maxClients {
			m.clientsMu.Unlock()
			http.Error(w, "Too many clients", http.StatusServiceUnavailable)
			return
		}

		// Create a new client
		client := &Client{
			messages: make(chan string, m.maxBuffer),
			lastSeen: time.Now(),
		}
		m.clients[client] = true
		m.clientsMu.Unlock()

		// Send initial connection event
		_, err := fmt.Fprintf(w, "data: %s Stream Connected\n\n", streamType)
		if err != nil {
			m.removeClient(client)
			//logger.SSE.Error(" ⚠️ Failed to send initial message: " + err.Error())
			return
		}
		flusher.Flush()

		// Handle client disconnection
		notify := r.Context().Done()

		// Start streaming messages
		m.streamMessages(w, flusher, client, streamType, notify)
	}
}

// streamMessages handles sending messages to a specific client
func (m *SSEManager) streamMessages(
	w http.ResponseWriter,
	flusher http.Flusher,
	client *Client,
	streamType string,
	notify <-chan struct{},
) {
	defer m.removeClient(client)

	for {
		select {
		case msg := <-client.messages:
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				//logger.SSE.Error(" ❌ Failed to send message: " + err.Error())
				return
			}
			flusher.Flush()

		case <-notify:
			return
		}
	}
}

// excludeClutterLogs checks if a message should be dropped due to kinematic warnings. This is a workaround "fix" for a bug in the gameserver.
func (m *SSEManager) excludeClutterLogs(message string) bool {
	if config.GetLogClutterToConsole() {
		return false
	}
	dropMessages := map[string]bool{
		"Setting linear velocity of a kinematic body is not supported":  true,
		"Setting angular velocity of a kinematic body is not supported": true,
		"WARNING: Shader":                           true,
		"ERROR: Shader":                             true,
		"No mesh data available":                    true,
		"The image effect Main Camera":              true,
		"Unsupported shader":                        true,
		"The shader":                                true,
		"memorysetup":                               true,
		"Microsoft Media Foundation video decoding": true,
		"The referenced script on this Behaviour":   true,
		"Fallback handler could not load library":   true,
	}

	// Check if message contains any of the drop messages
	for dropMsg := range dropMessages {
		if strings.Contains(message, dropMsg) {
			m.dropMu.Lock()
			defer m.dropMu.Unlock()

			m.kinematicDropCount++
			now := time.Now()

			// Log only if it's been more than a minute since last log and we have messages to report
			if m.kinematicDropCount > 0 && now.Sub(m.lastKinematicLog) >= time.Minute {
				//logger.SSE.Info(fmt.Sprintf("🗑️ Detected and Dropped %d unhelpful game server log messages. (Workaround for Gameserver Bug)", m.kinematicDropCount))
				m.lastKinematicLog = now
				m.kinematicDropCount = 0 // Reset count after logging
			}
			return true
		}
	}

	return false
}

// Broadcast sends a message to all clients with a non-blocking approach
func (m *SSEManager) Broadcast(message string) {
	// Check if message should be dropped
	if m.excludeClutterLogs(message) {
		return
	}

	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	for client := range m.clients {
		select {
		case client.messages <- message:
			// Message sent successfully
		default:
			// Client channel is full, log and skip
			//logger.SSE.Warn("⏳ Message dropped for slow client")
		}
	}
}

func (m *SSEManager) AddInternalSubscriber() chan string {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	client := &Client{
		messages: make(chan string, m.maxBuffer),
		lastSeen: time.Now(),
	}
	m.clients[client] = true
	return client.messages
}

func (m *SSEManager) RemoveInternalSubscriber(messages chan string) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()
	for client := range m.clients {
		if client.messages == messages {
			delete(m.clients, client)
			close(client.messages)
			return
		}
	}
}

// removeClient safely removes a client from the manager
func (m *SSEManager) removeClient(client *Client) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	delete(m.clients, client)
	close(client.messages)
}
