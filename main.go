package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
)

var cmd *exec.Cmd
var mu sync.Mutex

func main() {
	http.HandleFunc("/", serveUI)
	http.HandleFunc("/start", startServer)
	http.HandleFunc("/stop", stopServer)
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
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "Error starting server: %v", err)
		return
	}
	fmt.Fprintf(w, "Server started.")
}

func stopServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		fmt.Fprintf(w, "Server is not running.")
		return
	}

	err := cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		fmt.Fprintf(w, "Error stopping server: %v", err)
		return
	}
	cmd = nil
	fmt.Fprintf(w, "Server stopped.")
}
