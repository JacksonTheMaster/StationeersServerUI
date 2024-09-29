package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// InstallAndRunSteamCMD installs and runs SteamCMD based on the platform (Windows/Linux)
func InstallAndRunSteamCMD() {
	if runtime.GOOS == "windows" {
		installSteamCMDWindows()
	} else if runtime.GOOS == "linux" {
		installSteamCMDLinux()
	} else {
		fmt.Println("SteamCMD installation is not supported on this OS.")
		return
	}
}

// installSteamCMDWindows downloads and installs SteamCMD on Windows
func installSteamCMDWindows() {
	steamCMDDir := "C:\\SteamCMD"

	// Check if SteamCMD is already installed
	if _, err := os.Stat(steamCMDDir); os.IsNotExist(err) {
		fmt.Println("SteamCMD not found, downloading...")

		// Create SteamCMD directory
		err := os.MkdirAll(steamCMDDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating SteamCMD directory: %v\n", err)
			return
		}

		// Download SteamCMD
		downloadURL := "https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip"
		resp, err := http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Error downloading SteamCMD: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Read zip content
		zipContent, err := io.ReadAll(resp.Body)
		zipReader := bytes.NewReader(zipContent)

		if err != nil {
			fmt.Printf("Error reading SteamCMD zip: %v\n", err)
			return
		}

		// Unzip to C:\SteamCMD
		err = unzip(zipReader, zipReader.Size(), steamCMDDir)
		if err != nil {
			fmt.Printf("Error extracting SteamCMD zip: %v\n", err)
			return
		}

		fmt.Println("SteamCMD installed successfully.")
	}

	// Run SteamCMD
	runSteamCMD(steamCMDDir)
}

// installSteamCMDLinux downloads and installs SteamCMD on Linux
func installSteamCMDLinux() {
	steamCMDDir := "./steamcmd"

	// Check if SteamCMD is already installed
	if _, err := os.Stat(steamCMDDir); os.IsNotExist(err) {
		fmt.Println("SteamCMD not found, downloading...")

		// Create SteamCMD directory
		err := os.MkdirAll(steamCMDDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating SteamCMD directory: %v\n", err)
			return
		}

		// Download SteamCMD for Linux
		downloadURL := "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"
		resp, err := http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Error downloading SteamCMD: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Read tar.gz content
		err = untar(steamCMDDir, resp.Body)
		if err != nil {
			fmt.Printf("Error extracting SteamCMD tar.gz: %v\n", err)
			return
		}

		// Ensure executable permissions
		err = os.Chmod(filepath.Join(steamCMDDir, "steamcmd.sh"), 0755)
		if err != nil {
			fmt.Printf("Error setting SteamCMD executable permissions: %v\n", err)
			return
		}

		fmt.Println("SteamCMD installed successfully.")
	}

	// Run SteamCMD
	runSteamCMD(steamCMDDir)
}

// runSteamCMD runs the SteamCMD command to update the game
func runSteamCMD(steamCMDDir string) {
	//auto update at startup? Maybe make this a config option?
	//if config.AutoUpdate {
	//	fmt.Println("Auto updating...")
	//	runSteamCMD(steamCMDDir)
	//}
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		return
	}

	// Construct SteamCMD command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(steamCMDDir, "steamcmd.exe"), "+force_install_dir", currentDir, "+login", "anonymous", "+app_update", "600760", "+quit")
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command(filepath.Join(steamCMDDir, "steamcmd.sh"), "+force_install_dir", currentDir, "+login", "anonymous", "+app_update", "600760", "+quit")
	}

	// Set output to stdout
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	fmt.Println("Running SteamCMD...")
	err = cmd.Run() // <-- Here we just assign to the existing err variable
	if err != nil {
		fmt.Printf("Error running SteamCMD: %v\n", err)
		return
	}

	fmt.Println("SteamCMD command executed successfully.")
}

// unzip extracts a zip archive
// unzip extracts a zip archive
func unzip(zipReader io.ReaderAt, size int64, dest string) error {
	reader, err := zip.NewReader(zipReader, size)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create the file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// untar extracts a tar.gz archive
func untar(dest string, r io.Reader) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}
