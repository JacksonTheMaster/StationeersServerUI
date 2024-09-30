package main

import (
	"StationeersServerUI/src/api"
	"StationeersServerUI/src/config"
	discord "StationeersServerUI/src/discord"
	"StationeersServerUI/src/install"
	"fmt"
	"net/http"
	"os"
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

	fmt.Println(string(colorCyan), "Starting checks...", string(colorReset))

	// Start the installation process and wait for it to complete
	wg.Add(1)
	go install.Install(&wg)

	// Wait for the installation to finish before starting the rest of the server
	wg.Wait()

	fmt.Println(string(colorGreen), "Installation complete!", string(colorReset))

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
	go api.WatchBackupDir()

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
	http.HandleFunc("/furtherconfig", api.HandleConfigJSON)
	http.HandleFunc("/saveconfigasjson", api.SaveConfigJSON)

	fmt.Println(string(colorYellow), "Starting the HTTP server on port 8080...", string(colorReset))
	fmt.Println(string(colorGreen), "UI available at: http://127.0.0.1:8080", string(colorReset))
	fmt.Println(string(colorMagenta), "For first time Setup, follow the instructions on:", string(colorReset))
	fmt.Println(string(colorMagenta), "https://github.com/jacksonthemaster/StationeersServerUI/blob/main/readme.md#first-time-setup", string(colorReset))
	fmt.Println(string(colorMagenta), "Or just copy your save folder to /Saves and edit the save file name from the UI (Config Page)", string(colorReset))
	if config.Branch != "Release" {
		fmt.Println(string(colorRed), "⚠️Starting pprof server on /debug/pprof", string(colorReset))
	}
	// Start the HTTP server and check for errors
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Printf(string(colorRed)+"Error starting HTTP server: %v\n"+string(colorReset), err)
		os.Exit(1)
	}

}

func startLogStream() {
	client := sse.NewClient("http://localhost:8080/output")
	client.Headers["Content-Type"] = "text/event-stream"
	client.Headers["Connection"] = "keep-alive"
	client.Headers["Cache-Control"] = "no-cache"

	retryDelay := 5 * time.Second // Retry every 5 seconds

	go func() {
		for {
			fmt.Println(string(colorYellow), "Attempting to connect to SSE stream...", string(colorReset))

			err := client.SubscribeRaw(func(msg *sse.Event) {
				if len(msg.Data) > 0 {
					logMessage := string(msg.Data)
					discord.AddToLogBuffer(logMessage)
					//fmt.Println(string(colorGreen), "Serverlog:", logMessage, string(colorReset))
					//dont spam the console with the server log
				}
			})

			if err != nil {
				// Instead of logging errors repeatedly, retry silently until the endpoint is available
				fmt.Println(string(colorYellow), "SSE stream not available yet, retrying in 5 seconds...", string(colorReset))
				time.Sleep(retryDelay)
				continue
			}

			// Successfully connected, break the loop and handle messages
			fmt.Println(string(colorGreen), "Connected to SSE stream.", string(colorReset))
			return
		}
	}()
}
