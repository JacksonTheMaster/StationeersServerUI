package discord

import (
	"fmt"
	"net/http"
)

func SendCommandToAPI(endpoint string) {
	url := "http://localhost:8080" + endpoint
	if _, err := http.Get(url); err != nil {
		fmt.Printf("Failed to send %s command: %v\n", endpoint, err)
	}
}
