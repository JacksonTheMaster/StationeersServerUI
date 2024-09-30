package install

import (
	"StationeersServerUI/src/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Install performs the entire installation process and ensures the server waits for it to complete
func Install(wg *sync.WaitGroup) {
	defer wg.Done()             // Signal that installation is complete
	time.Sleep(1 * time.Second) // Small pause for effect

	// Step 1: Check and download the UIMod folder contents
	fmt.Println("🔄 Checking UIMod folder contents...")
	CheckAndDownloadUIMod()
	fmt.Println("✅ UIMod folder setup complete.")
	time.Sleep(1 * time.Second)

	// Step 2: Check for Blacklist.txt and create it if it doesn't exist
	fmt.Println("🔄 Checking for Blacklist.txt...")
	checkAndCreateBlacklist()
	fmt.Println("✅ Blacklist.txt verified or created.")
	time.Sleep(1 * time.Second)

	// Step 3: Install and run SteamCMD
	fmt.Println("🔄 Installing and running SteamCMD...")
	InstallAndRunSteamCMD()
	fmt.Println("Thank you for using this Software! 🙏")
}

func CheckAndDownloadUIMod() {
	workingDir := "./UIMod/"

	// Check if the directory exists
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		fmt.Println("⚠️ Folder ./UIMod does not exist. Creating it...")

		// Create the folder
		err := os.MkdirAll(workingDir, os.ModePerm)
		if err != nil {
			fmt.Printf("❌ Error creating folder: %v\n", err)
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
		//set the first time setup flag to true
		config.IsFirstTimeSetup = true
		// Download each file
		for fileName, url := range files {
			err := downloadFile(workingDir+fileName, url)
			if err != nil {
				fmt.Printf("❌ Error downloading %s: %v\n", fileName, err)
				return
			}
			fmt.Printf("✅ Downloaded %s successfully\n", fileName)
		}

		fmt.Println("✅ All files downloaded successfully.")
	} else {
		fmt.Println("♻️ Folder ./UIMod already exists. Skipping download.")
		config.IsFirstTimeSetup = false
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
			fmt.Printf("❌ Error creating Blacklist.txt: %v\n", err)
			return
		}
		defer file.Close()

		fmt.Println("✅ Created Blacklist.txt.")
	} else {
		fmt.Println("♻️ Blacklist.txt already exists. Skipping creation.")
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
