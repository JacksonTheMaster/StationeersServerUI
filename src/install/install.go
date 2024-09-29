package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Install performs the entire installation process and ensures the server waits for it to complete
func Install(wg *sync.WaitGroup) {
	defer wg.Done() // Signal that installation is complete

	// Check and download the UIMod folder contents
	CheckAndDownloadUIMod()

	// Check for Blacklist.txt and create it if it doesn't exist
	checkAndCreateBlacklist()

	InstallAndRunSteamCMD()
}

func CheckAndDownloadUIMod() {
	workingDir := "./UIMod/"

	// Check if the directory exists
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		fmt.Println("Folder ./UIMod does not exist. Creating it...")

		// Create the folder
		err := os.MkdirAll(workingDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating folder: %v\n", err)
			return
		}

		// List of files to download
		files := map[string]string{
			"apiinfo.html":       "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/apiinfo.html",
			"config.html":        "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/config.html",
			"furtherconfig.html": "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/furtherconfig.html",
			"config.json":        "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/config.json",
			"config.xml":         "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/config.xml",
			"index.html":         "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/index.html",
			"script.js":          "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/script.js",
			"stationeers.png":    "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/stationeers.png",
			"style.css":          "https://raw.githubusercontent.com/JacksonTheMaster/StationeersServerUI/main/UIMod/style.css",
		}

		// Download each file
		for fileName, url := range files {
			err := downloadFile(workingDir+fileName, url)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", fileName, err)
				return
			}
			fmt.Printf("Downloaded %s successfully\n", fileName)
		}

		fmt.Println("All files downloaded successfully.")
	} else {
		fmt.Println("Folder ./UIMod already exists. Skipping download.")
	}
}

// checkAndCreateBlacklist ensures Blacklist.txt exists in the root directory
func checkAndCreateBlacklist() {
	blacklistFile := "./Blacklist.txt"

	// Check if Blacklist.txt exists
	if _, err := os.Stat(blacklistFile); os.IsNotExist(err) {
		// Create an empty Blacklist.txt file
		file, err := os.Create(blacklistFile)
		if err != nil {
			fmt.Printf("Error creating Blacklist.txt: %v\n", err)
			return
		}
		defer file.Close()

		fmt.Println("Created Blacklist.txt.")
	} else {
		fmt.Println("Blacklist.txt already exists. Skipping creation.")
	}
}

// downloadFile downloads a file from the given URL and saves it to the given filepath
func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
