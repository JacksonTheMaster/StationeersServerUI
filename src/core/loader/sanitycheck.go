package loader

import (
	"errors"
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
	if config.GetNoSanityCheck() {
		return nil
	}

	if runtime.GOOS != "windows" {
		IsInsideContainer(&containerCheckWG)
		containerCheckWG.Wait()

		if os.Geteuid() == 0 && !config.GetIsDockerContainer() {
			return fmt.Errorf("root: SSUI should not be run as root")
		}

		exePath, err := os.Readlink("/proc/self/exe")
		if err != nil {
			return err
		}
		dirPath := filepath.Dir(exePath)
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if cwd != dirPath && !isGoRunBuildDir(dirPath) {
			if err := os.Chdir(dirPath); err != nil {
				return err
			}
		}
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if err := checkDirectoryWrite(workDir); err != nil {
		return fmt.Errorf("cannot write to working directory %s: %w", workDir, err)
	}
	if err := checkTreeAccess(config.GetSSUIFolder(), true); err != nil {
		return fmt.Errorf("SSUI data directory access check failed: %w", err)
	}
	// A separately managed server only needs readable saves. Inside the image,
	// Stationeers itself runs below /app and must be able to update them.
	if err := checkTreeAccess("./saves", config.GetIsDockerContainer()); err != nil {
		return fmt.Errorf("save directory access check failed: %w", err)
	}

	// Check if steamcmd package is installed  (requires further testing, disabled for now)
	//cmd := exec.Command("dpkg-query", "-W", "-f='${Status}'", "steamcmd")
	//output, err := cmd.CombinedOutput()
	//if err == nil && strings.Contains(string(output), "install ok installed") {
	//	return fmt.Errorf("steamcmd apt package is installed, it is not recommended to run SSUI when the apt steamcmd package is installed. Please uninstall the steamcmd package or have a look at our Docker image and try again")
	//}

	return nil
}

func checkTreeAccess(root string, writable bool) error {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("cannot read %s: %w", path, walkErr)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			directory, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("cannot read directory %s: %w", path, err)
			}
			if err := directory.Close(); err != nil {
				return fmt.Errorf("cannot close directory %s: %w", path, err)
			}
			if writable {
				if err := checkDirectoryWrite(path); err != nil {
					return fmt.Errorf("cannot write directory %s: %w", path, err)
				}
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("cannot close file %s: %w", path, err)
		}
		if writable {
			file, err = os.OpenFile(path, os.O_WRONLY, 0)
			if err != nil {
				return fmt.Errorf("cannot write file %s: %w", path, err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("cannot close file %s: %w", path, err)
			}
		}
		return nil
	})
}

func checkDirectoryWrite(path string) error {
	file, err := os.CreateTemp(path, ".ssui-write-test-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
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
