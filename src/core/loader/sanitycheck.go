package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

var containerCheckWG = sync.WaitGroup{}

func runSanityCheck() error {

	if runtime.GOOS == "windows" {
		return nil
	}

	if config.GetNoSanityCheck() {
		return nil
	}

	IsInsideContainer(&containerCheckWG)
	containerCheckWG.Wait()

	// Check if running as root (UID 0)
	if os.Geteuid() == 0 {
		// Check if running inside a container
		if !config.GetIsDockerContainer() {
			return fmt.Errorf("root: SSUI should not be run as root")
		}
	}

	// Get the current executable path from /proc/self/exe
	exePath, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}
	// Get the directory path of the executable
	dirPath := filepath.Dir(exePath)
	// Change the working directory to the executable's directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if cwd != dirPath && !isGoRunBuildDir(dirPath) {
		err = os.Chdir(dirPath)
		if err != nil {
			return err
		}
	}

	// Check if current working directory is writable
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(workDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("cannot write to working directory, please make sure your user has write permissions in %s: %w", workDir, err)
	}
	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to clean up sanity check writetest file: %w", err)
	}

	// Check if steamcmd package is installed  (requires further testing, disabled for now)
	//cmd := exec.Command("dpkg-query", "-W", "-f='${Status}'", "steamcmd")
	//output, err := cmd.CombinedOutput()
	//if err == nil && strings.Contains(string(output), "install ok installed") {
	//	return fmt.Errorf("steamcmd apt package is installed, it is not recommended to run SSUI when the apt steamcmd package is installed. Please uninstall the steamcmd package or have a look at our Docker image and try again")
	//}

	return nil
}

// isGoRunBuildDir reports whether executableDir belongs to a temporary or
// cached go run build. Go 1.24 began caching go run executables in the Go
// build cache, so they are no longer guaranteed to live below /tmp.
func isGoRunBuildDir(executableDir string) bool {
	cleanDir := filepath.ToSlash(filepath.Clean(executableDir))
	return strings.Contains(cleanDir, "/tmp") ||
		strings.Contains(cleanDir, "/go-build/") ||
		strings.HasSuffix(cleanDir, "/go-build")
}
