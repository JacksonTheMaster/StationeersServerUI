package main

import (
	"StationeersServerUI/src/api"
	"StationeersServerUI/src/config"
	discord "StationeersServerUI/src/discord"
	"StationeersServerUI/src/install"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/r3labs/sse"
)

const (
	// ANSI color codes for styling terminal output
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

func main() {
	var wg sync.WaitGroup

	fmt.Println(string(colorCyan), "Starting installation...", string(colorReset))

	// Start the installation process and wait for it to complete
	wg.Add(1)
	go install.Install(&wg)

	// Wait for the installation to finish before starting the rest of the server
	wg.Wait()

	fmt.Println(string(colorGreen), "Installation complete!", string(colorReset))

	// Check if the branch is not "Prod" and enable pprof if it's not
	if config.Branch != "Release" {
		go func() {
			fmt.Println(string(colorMagenta), "Starting pprof server on localhost:6060...", string(colorReset))
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				fmt.Printf(string(colorRed)+"Error starting pprof server: %v\n"+string(colorReset), err)
			}
		}()
	}

	workingDir := "./UIMod/"
	configFilePath := workingDir + "config.json"

	fmt.Println(string(colorBlue), "Loading configuration from", configFilePath, string(colorReset))
	config.LoadConfig(configFilePath)

	// If Discord is enabled, start the Discord bot
	if config.IsDiscordEnabled {
		fmt.Println(string(colorGreen), "Starting Discord bot...", string(colorReset))
		go discord.StartDiscordBot()
	}

	go startLogStream()

	fmt.Println(string(colorBlue), "Starting API services...", string(colorReset))
	go api.StartAPI()
	go api.StartBackupCleanupRoutine()

	fs := http.FileServer(http.Dir("./UIMod"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", api.ServeUI)
	http.HandleFunc("/start", api.StartServer)
	http.HandleFunc("/stop", api.StopServer)
	http.HandleFunc("/output", api.GetOutput)
	http.HandleFunc("/backups", api.ListBackups)
	http.HandleFunc("/restore", api.RestoreBackup)
	http.HandleFunc("/config", api.HandleConfig)
	http.HandleFunc("/saveconfig", api.SaveConfig)
	http.HandleFunc("/futherconfig", api.HandleConfigJSON)
	http.HandleFunc("/saveconfigasjson", api.SaveConfigJSON)

	fmt.Println(string(colorYellow), "Starting the HTTP server on port 8080...", string(colorReset))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(string(colorRed)+"Error starting HTTP server: %v\n"+string(colorReset), err)
	} else {
		fmt.Println(string(colorGreen), "Server started successfully!", string(colorReset))
		fmt.Println(string(colorGreen), "UI available at: http://127.0.0.1:8080", string(colorReset))
	}
}

func startLogStream() {
	client := sse.NewClient("http://localhost:8080/output")
	client.Headers["Content-Type"] = "text/event-stream"
	client.Headers["Connection"] = "keep-alive"
	client.Headers["Cache-Control"] = "no-cache"

	for {
		// Attempt to connect to the SSE stream
		fmt.Println(string(colorYellow), "Attempting to connect to SSE stream...", string(colorReset))

		err := client.SubscribeRaw(func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				logMessage := string(msg.Data)
				discord.AddToLogBuffer(logMessage)
				fmt.Println(string(colorGreen), "Received SSE log message:", logMessage, string(colorReset))
			}
		})

		if err != nil {
			fmt.Printf(string(colorRed)+"Error subscribing to SSE stream: %v\n"+string(colorReset), err)
			fmt.Println(string(colorYellow), "Reconnecting in 5 seconds...", string(colorReset))
			time.Sleep(5 * time.Second)
			continue // Retry connection
		}

		// Instead of breaking the loop, we let it continue handling incoming messages.
		// The reconnection logic will handle errors or disconnects automatically.
	}
}
