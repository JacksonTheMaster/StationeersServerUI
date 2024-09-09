package main

import (
	"StationeersServerUI/src/config"
	discord "StationeersServerUI/src/discord"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/r3labs/sse"
)

func main() {
	processName := "Stationeers-ServerUI.exe"
	workingDir := "./UIMod/"
	exePath := "./" + processName

	configFilePath := workingDir + "config.json"
	config.LoadConfig(configFilePath)

	go discord.StartDiscordBot()
	go startLogStream()

	for {
		if !isProcessRunning(processName) {
			fmt.Printf("Process %s not running. Starting it...\n", processName)

			// Start the process
			if err := startProcess(exePath, workingDir); err != nil {
				fmt.Printf("Failed to start process: %v\n", err)
			} else {
				fmt.Println("UI and API started successfully at http://localhost:8080.")
			}
		} else {
			fmt.Printf("Process %s is running.\n", processName)
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}
}

func startLogStream() {
	client := sse.NewClient("http://localhost:8080/output")
	client.Headers["Content-Type"] = "text/event-stream"
	client.Headers["Connection"] = "keep-alive"
	client.Headers["Cache-Control"] = "no-cache"

	for {
		// Attempt to connect to the SSE stream
		fmt.Println("Attempting to connect to SSE stream...")

		err := client.SubscribeRaw(func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				logMessage := string(msg.Data)
				discord.AddToLogBuffer(logMessage)
			}
		})

		if err != nil {
			fmt.Printf("Error subscribing to SSE stream: %v\n", err)
			fmt.Println("Reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue // Retry connection
		}

		// If the connection is successful, block until an error occurs
		// The error handling and reconnection logic should be inside the SubscribeRaw callback
		break
	}
}

func isProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/fi", fmt.Sprintf(`imagename eq %s`, processName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error checking process: %v\n", err)
		return false
	}
	return strings.Contains(string(out), processName)
}

func startProcess(exePath, workingDir string) error {
	cmd := exec.Command(exePath)
	cmd.Dir = workingDir

	//stdin, err := cmd.StdinPipe()
	//if err != nil {
	//	return fmt.Errorf("error creating stdin pipe: %v", err)
	//}

	//serverStdin = stdin
	//serverCmd = cmd

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting Server UI: %v", err)
	}

	fmt.Println("Ensure Windows Firewall allows incoming connections on game port and update port (27015 and 27016 by default).")
	return nil
}
