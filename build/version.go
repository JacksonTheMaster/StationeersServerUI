//go:build ignore
// +build ignore

// version.go is a helper script to sync the backends version to package.json so the Electron auto updater can find its latest update on GitHub.
// Gets called only when the user runs the "Build: Full Project (Prep a release)" task.

// The source of truth for the current version is the backends version defined in config.go

package main

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func main() {
	err := incrementUIVersion()
	if err != nil {
		panic(err)
	}
}

func incrementUIVersion() error {
	backendVersion := config.Version

	// Define the path to package.json
	packagePath := filepath.Join(".", "frontend", "package.json")

	// Read the package.json file
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	// Create a regex to match the version field (e.g., "version": "v1.2.3")
	// The regex captures the entire line, including quotes and commas
	re, err := regexp.Compile(`"version"\s*:\s*"v\d+\.\d+\.\d+"`)
	if err != nil {
		return err
	}

	// Prepare the replacement string
	replacement := `"version": "v` + backendVersion + `"`

	// Perform the replacement
	updatedData := re.ReplaceAllString(string(data), replacement)

	// Write the updated data back to package.json
	if err := os.WriteFile(packagePath, []byte(updatedData), 0644); err != nil {
		return err
	}

	return nil
}
