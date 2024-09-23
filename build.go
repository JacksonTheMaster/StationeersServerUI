// build.go
package main

import (
	"StationeersServerUI/src/config"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// Load the config to access Version and Branch
	config.LoadConfig("./UIMod/config.json")

	// Prepare the output file name with version and branch
	outputName := fmt.Sprintf("StationeersServerControl%s_%s.exe", config.Version, config.Branch)

	// Run the go build command with the custom output name
	cmd := exec.Command("go", "build", "-o", outputName, "./src")

	// Capture any output or errors
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Build failed: %s\nOutput: %s", err, string(cmdOutput))
	}

	fmt.Printf("Build successful! Output: %s\n", outputName)
}
