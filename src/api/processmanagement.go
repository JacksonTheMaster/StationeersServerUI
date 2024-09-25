// processmanagement.go
package api

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
)

func StartServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		fmt.Fprintf(w, "Server is already running.")
		return
	}

	config, err := loadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	cmd = exec.Command(config.Server.ExePath, "-LOAD", config.SaveFileName, "-settings", config.Server.Settings)
	fmt.Printf(`Load command: %s -LOAD %s -settings %s\n`, config.Server.ExePath, config.SaveFileName, config.Server.Settings)

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(w, "Error creating StdoutPipe: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(w, "Error creating StderrPipe: %v", err)
		return
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "Error starting server: %v", err)
		return
	}

	// Start reading stdout and stderr
	go readPipe(stdout)
	go readPipe(stderr)

	fmt.Fprintf(w, "Server started.")
}

func readPipe(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		output := scanner.Text()
		clientsMu.Lock()
		for _, clientChan := range clients {
			clientChan <- output
		}
		clientsMu.Unlock()
	}
	if err := scanner.Err(); err != nil {
		output := fmt.Sprintf("Error reading pipe: %v", err)
		clientsMu.Lock()
		for _, clientChan := range clients {
			clientChan <- output
		}
		clientsMu.Unlock()
	}
}

func GetOutput(w http.ResponseWriter, r *http.Request) {
	// Create a new channel for this client
	clientChan := make(chan string)

	// Register the client
	clientsMu.Lock()
	clients = append(clients, clientChan)
	clientsMu.Unlock()

	// Ensure the channel is removed when the client disconnects
	defer func() {
		clientsMu.Lock()
		for i, ch := range clients {
			if ch == clientChan {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		clientsMu.Unlock()
		close(clientChan)
	}()

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Write data to the client as it comes in
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for msg := range clientChan {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

func StopServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		fmt.Fprintf(w, "Server is not running.")
		return
	}

	// Attempt a graceful shutdown
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		// If sending SIGTERM fails, attempt to kill the process
		fmt.Fprintf(w, "(Known issue) Error sending SIGTERM to server: %v. Killing the process.\n", err)

		err = cmd.Process.Kill()
		if err != nil {
			fmt.Fprintf(w, "Error killing server process: %v", err)
			return
		}
	}

	// Wait for the process to exit
	err = cmd.Wait()
	if err != nil && !strings.Contains(err.Error(), "exit status") {
		fmt.Fprintf(w, "Error waiting for process: %v", err)
		return
	}

	cmd = nil
	fmt.Fprintf(w, "Server stopped.")
}
