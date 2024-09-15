package main

import (
	"StationeersServerUI/src/api"
	"StationeersServerUI/src/config"
	discord "StationeersServerUI/src/discord"
	"fmt"
	"net/http"
	"time"

	"github.com/r3labs/sse"
)

func main() {
	workingDir := "./UIMod/"

	configFilePath := workingDir + "config.json"

	config.LoadConfig(configFilePath)
	//if config.IsDiscordEnabled true start discord bot else skip
	if config.IsDiscordEnabled {
		go discord.StartDiscordBot()
	}
	go startLogStream()
	go api.StartAPI()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./UIMod/static/"))))
	http.HandleFunc("/", api.ServeUI)
	http.HandleFunc("/start", api.StartServer)
	http.HandleFunc("/stop", api.StopServer)
	http.HandleFunc("/output", api.GetOutput)
	http.HandleFunc("/backups", api.ListBackups)
	http.HandleFunc("/restore", api.RestoreBackup)
	http.HandleFunc("/config", api.HandleConfig)   // Serve configuration form
	http.HandleFunc("/saveconfig", api.SaveConfig) // Save configuration form
	http.ListenAndServe(":8080", nil)
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
