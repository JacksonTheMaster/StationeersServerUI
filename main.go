package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
)

var cmd *exec.Cmd
var mu sync.Mutex
var outputChannel chan string

func main() {
	outputChannel = make(chan string, 100)

	http.HandleFunc("/", serveUI)
	http.HandleFunc("/start", startServer)
	http.HandleFunc("/stop", stopServer)
	http.HandleFunc("/output", getOutput)
	http.ListenAndServe(":8080", nil)
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./UIMod/index.html")
}

func startServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		fmt.Fprintf(w, "Server is already running.")
		return
	}

	cmd = exec.Command("powershell.exe", "-Command", "./rocketstation_DedicatedServer.exe -LOAD MarsProd Mars -settings StartLocalHost true ServerVisible true GamePort 27016 UpdatePort 27015 AutoSave true SaveInterval 300 LocalIpAddress 10.10.50.99 ServerPassword JMG AdminPassword JMG ServerMaxPlayers 4 ServerName SpaceInc")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

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

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "Error starting server: %v", err)
		return
	}

	go readPipe(stdout)
	go readPipe(stderr)

	fmt.Fprintf(w, "Server started.")
}

func readPipe(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		outputChannel <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		outputChannel <- fmt.Sprintf("Error reading pipe: %v", err)
	}
	close(outputChannel)
}

func stopServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		fmt.Fprintf(w, "Server is not running.")
		return
	}

	err := cmd.Process.Signal(syscall.SIGTERM) // Use SIGTERM to attempt a graceful termination
	if err != nil {
		fmt.Fprintf(w, "Error stopping server: %v", err)
		return
	}
	cmd = nil
	fmt.Fprintf(w, "Server stopped.")
}

func getOutput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for output := range outputChannel {
		fmt.Fprintf(w, "data: %s\n\n", output)
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}
