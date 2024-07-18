package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	processName := "Stationeers-ServerUI.exe"
	workingDir := "./UIMod/"
	exePath := "./" + processName

	for {
		if !isProcessRunning(processName) {
			fmt.Printf("Process %s not running. Starting it...\n", processName)

			// Start the process
			err := startProcess(exePath, workingDir)
			if err != nil {
				fmt.Printf("Failed to start process: %v\n", err)
			} else {
				fmt.Println("UI and API started successfully at http://localhost:8080 What a relief!") // link should be clickable, but istn in console
			}
		} else {
			fmt.Printf("Process %s is running.\n", processName)
		}

		// Wait for a while before checking again
		time.Sleep(5 * time.Second)
	}
}

func isProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/fi", fmt.Sprintf(`imagename eq %s`, processName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error checking process: %v\n", err)
		return false
	}
	output := string(out)
	//fmt.Println("Tasklist output:", output)
	return strings.Contains(output, processName)
}

func startProcess(exePath, workingDir string) error {
	cmd := exec.Command(exePath)
	cmd.Dir = workingDir

	// Debugging: Print the command to be executed
	fmt.Println("Executing Server UI Startup command:", cmd.String())

	// Start the process
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting Server UI: %v", err)
	}
	fmt.Println("Make sure to configure the Windows Firewall to allow incoming connections on game port and update port (27015 and 27016 by default)")
	return nil
}
