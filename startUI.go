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
				fmt.Println("Process started successfully.")
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
	fmt.Println("Tasklist output:", output)
	return strings.Contains(output, processName)
}

func startProcess(exePath, workingDir string) error {
	cmd := exec.Command(exePath)
	cmd.Dir = workingDir

	// Debugging: Print the command to be executed
	fmt.Println("Executing command:", cmd.String())

	// Start the process
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting process: %v", err)
	}
	return nil
}
