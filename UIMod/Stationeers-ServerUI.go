package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var cmd *exec.Cmd
var mu sync.Mutex
var outputChannel chan string

// Config holds the server configuration
type Config struct {
	Server struct {
		ExePath  string `xml:"exePath"`
		Settings string `xml:"settings"`
	} `xml:"server"`
	SaveFileName string `xml:"saveFileName"`
}

// LoadConfig loads the configuration from an XML file
func LoadConfig() (*Config, error) {
	configPath := "./config.xml"
	xmlFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = xml.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %v", err)
	}

	return &config, nil
}

func main() {
	outputChannel = make(chan string, 100)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/", serveUI)
	http.HandleFunc("/start", startServer)
	http.HandleFunc("/stop", stopServer)
	http.HandleFunc("/output", getOutput)
	http.HandleFunc("/backups", listBackups)
	http.HandleFunc("/restore", restoreBackup)
	http.HandleFunc("/config", handleConfig)   // Serve configuration form
	http.HandleFunc("/saveconfig", saveConfig) // Save configuration form
	http.ListenAndServe(":8080", nil)
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func startServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		fmt.Fprintf(w, "Server is already running.")
		return
	}

	config, err := LoadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf(`%s -LOAD %s -settings %s`, config.Server.ExePath, config.SaveFileName, config.Server.Settings))
	fmt.Println(fmt.Sprintf(`Load command: %s -LOAD %s -settings %s`, config.Server.ExePath, config.SaveFileName, config.Server.Settings))

	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(w, "Error creating StdoutPipe: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(w, "Error creating StderrPipe: %v", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "Error starting server: %v", err)
		return
	}

	go readPipe(stdout)
	go readPipe(stderr)

	fmt.Fprintf(w, "Server started.")
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	config, err := LoadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	htmlFile, err := os.ReadFile("./config.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading config.html: %v", err), http.StatusInternalServerError)
		return
	}

	htmlContent := string(htmlFile)

	// Split the settings string into a map for easier access
	settingsMap := make(map[string]string)
	settings := strings.Split(config.Server.Settings, " ")
	for i := 0; i < len(settings)-1; i += 2 {
		settingsMap[settings[i]] = settings[i+1]
	}

	// Replace placeholders with actual values
	replacements := map[string]string{
		"{{ExePath}}":          config.Server.ExePath,
		"{{StartLocalHost}}":   settingsMap["StartLocalHost"],
		"{{ServerVisible}}":    settingsMap["ServerVisible"],
		"{{GamePort}}":         settingsMap["GamePort"],
		"{{UpdatePort}}":       settingsMap["UpdatePort"],
		"{{AutoSave}}":         settingsMap["AutoSave"],
		"{{SaveInterval}}":     settingsMap["SaveInterval"],
		"{{LocalIpAddress}}":   settingsMap["LocalIpAddress"],
		"{{ServerPassword}}":   settingsMap["ServerPassword"],
		"{{AdminPassword}}":    settingsMap["AdminPassword"],
		"{{ServerMaxPlayers}}": settingsMap["ServerMaxPlayers"],
		"{{ServerName}}":       settingsMap["ServerName"],
		"{{AdditionalParams}}": getAdditionalParams(settings),
		"{{SaveFileName}}":     config.SaveFileName,
	}

	for placeholder, value := range replacements {
		htmlContent = strings.ReplaceAll(htmlContent, placeholder, value)
	}

	fmt.Fprint(w, htmlContent)
}

func getAdditionalParams(settings []string) string {
	// List of known parameters
	knownParams := map[string]bool{
		"StartLocalHost":   true,
		"ServerVisible":    true,
		"GamePort":         true,
		"UpdatePort":       true,
		"AutoSave":         true,
		"SaveInterval":     true,
		"LocalIpAddress":   true,
		"ServerPassword":   true,
		"AdminPassword":    true,
		"ServerMaxPlayers": true,
		"ServerName":       true,
	}

	var additionalParams []string
	for i := 0; i < len(settings)-1; i += 2 {
		if !knownParams[settings[i]] {
			additionalParams = append(additionalParams, settings[i]+" "+settings[i+1])
		}
	}

	return strings.Join(additionalParams, " ")
}

// SaveConfig saves the updated configuration to the XML file
func saveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		settings := []string{
			"StartLocalHost", r.FormValue("StartLocalHost"),
			"ServerVisible", r.FormValue("ServerVisible"),
			"GamePort", r.FormValue("GamePort"),
			"UpdatePort", r.FormValue("UpdatePort"),
			"AutoSave", r.FormValue("AutoSave"),
			"SaveInterval", r.FormValue("SaveInterval"),
			"LocalIpAddress", r.FormValue("LocalIpAddress"),
			"ServerPassword", r.FormValue("ServerPassword"),
			"AdminPassword", r.FormValue("AdminPassword"),
			"ServerMaxPlayers", r.FormValue("ServerMaxPlayers"),
			"ServerName", r.FormValue("ServerName"),
		}

		// Append additional parameters if any
		additionalParams := r.FormValue("AdditionalParams")
		if additionalParams != "" {
			settings = append(settings, strings.Split(additionalParams, " ")...)
		}

		settingsStr := strings.Join(settings, " ")

		config := Config{
			Server: struct {
				ExePath  string `xml:"exePath"`
				Settings string `xml:"settings"`
			}{
				ExePath: "../rocketstation_DedicatedServer.exe",
				// hardcoded ../rocketstation_DedicatedServer.exe for now, otherwise this is a security risk because an attacker could run a malicious exe and or command on the server
				// explaination: if the exepath is set to powershell.exe again, and matching parameters are set, the server would be able to run arbitrary code on the server.
				// this is very much an RCE vulnerability, and thus should be avoided at all costs.
				// with the hardcoded exepath, the server will not be able to run arbitrary code, but it will still be able to run the server with the given parameters.
				//more info: https://github.com/JacksonTheMaster/StationeersServerUI/issues/
				Settings: settingsStr,
			},
			SaveFileName: r.FormValue("saveFileName"),
		}

		configPath := "./config.xml"
		file, err := os.Create(configPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating config file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		encoder := xml.NewEncoder(file)
		encoder.Indent("", "  ")
		if err := encoder.Encode(config); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding config: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func readPipe(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		outputChannel <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		outputChannel <- fmt.Sprintf("Error reading pipe: %v", err)
	}
	close(outputChannel)
}

func stopServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		fmt.Fprintf(w, "Server is not running.")
		return
	}

	// Find the PID of the process by name
	pid, err := getPIDByName("rocketstation_DedicatedServer")
	if err != nil {
		fmt.Fprintf(w, "Error finding process: %v", err)
		return
	}

	// Terminate the process by PID
	err = exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
	if err != nil {
		fmt.Fprintf(w, "Error stopping server: %v", err)
		return
	}

	cmd = nil
	fmt.Fprintf(w, "Server stopped.")
}

func getPIDByName(name string) (int, error) {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`(Get-Process -Name "%s" | Select-Object -ExpandProperty Id) -join ","`, name))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(lines) > 0 {
		pid, err := strconv.Atoi(lines[0])
		if err != nil {
			return 0, err
		}
		return pid, nil
	}

	return 0, fmt.Errorf("process %s not found", name)
}

func listBackups(w http.ResponseWriter, r *http.Request) {
	config, err := LoadConfig()
	if err != nil {
		http.Error(w, "Error loading config", http.StatusInternalServerError)
		return
	}

	basePath := "../saves/" + config.SaveFileName + "/backup"
	files, err := os.ReadDir(basePath)
	if err != nil {
		http.Error(w, "Unable to read backups directory", http.StatusInternalServerError)
		return
	}

	backupDetails := make(map[int]time.Time)
	backupIndices := make(map[int]bool)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		backupIndex := parseBackupIndex(fileName)
		if backupIndex != -1 {
			if !backupIndices[backupIndex] {
				backupIndices[backupIndex] = true
				fullPath := filepath.Join(basePath, fileName)
				info, err := os.Stat(fullPath)
				if err != nil {
					http.Error(w, "Error getting file info", http.StatusInternalServerError)
					return
				}
				backupDetails[backupIndex] = info.ModTime()
			}
		}
	}

	var sortedBackups []struct {
		index   int
		modTime time.Time
	}

	for index, modTime := range backupDetails {
		sortedBackups = append(sortedBackups, struct {
			index   int
			modTime time.Time
		}{index, modTime})
	}

	sort.Slice(sortedBackups, func(i, j int) bool {
		return sortedBackups[i].index > sortedBackups[j].index // Sort by descending index
	})

	var output []string
	for _, backup := range sortedBackups {
		creationTime := backup.modTime.Format("02.01.2006 15:04:05")
		output = append(output, fmt.Sprintf("BackupIndex: %d, Created: %s", backup.index, creationTime))
	}

	if len(output) == 0 {
		fmt.Fprint(w, "No valid backup files found. Is the directory specified?")
		return
	}

	response := strings.Join(output, "\n")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, response)
}

// Helper function to sort backup details by index
func sortedKeys(backupDetails map[int]time.Time) []struct {
	index   int
	modTime time.Time
} {
	var sorted []struct {
		index   int
		modTime time.Time
	}
	for index, modTime := range backupDetails {
		sorted = append(sorted, struct {
			index   int
			modTime time.Time
		}{index, modTime})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].index < sorted[j].index
	})
	return sorted
}

func parseBackupIndex(fileName string) int {
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) > 1 {
		index, err := strconv.Atoi(matches[1])
		if err == nil {
			return index
		}
	}
	return -1
}

func restoreBackup(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "Index parameter is required", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index parameter", http.StatusBadRequest)
		return
	}

	config, err := LoadConfig()
	backupDir := "../saves/" + config.SaveFileName + "/backup"
	saveDir := "../saves/" + config.SaveFileName
	files := []struct {
		backupName string
		destName   string
	}{
		{fmt.Sprintf("world_meta(%d).xml", index), "world_meta.xml"},
		{fmt.Sprintf("world(%d).xml", index), "world.xml"},
		{fmt.Sprintf("world(%d).bin", index), "world.bin"},
	}

	// Create a map to store successful restore operations
	restoredFiles := make(map[string]string)

	// First, try to restore all files
	for _, file := range files {
		backupFile := filepath.Join(backupDir, file.backupName)
		destFile := filepath.Join(saveDir, file.destName)

		err := copyFile(backupFile, destFile)
		if err != nil {
			// Revert any successful operations if an error occurs
			revertRestore(restoredFiles, saveDir, backupDir)
			http.Error(w, fmt.Sprintf("Error restoring file %s: %v", file.backupName, err), http.StatusInternalServerError)
			return
		}
		restoredFiles[destFile] = backupFile
	}

	fmt.Fprintf(w, "Backup %d restored successfully.", index)
}

// copyFile copies a file from src to dst. If dst already exists, it will be overwritten. This is the inteded behavior of the restoreBackup function. We overwrite the destination files with the backup files
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// revertRestore reverts the file restore operation if an error occurs
func revertRestore(restoredFiles map[string]string, saveDir, backupDir string) {
	for destFile, backupFile := range restoredFiles {
		err := os.Remove(destFile)
		if err != nil {
			fmt.Printf("Error removing file %s: %v\n", destFile, err)
		} else {
			err = copyFile(backupFile, destFile)
			if err != nil {
				fmt.Printf("Error restoring file %s: %v\n", destFile, err)
			}
		}
	}
}

func getOutput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for output := range outputChannel {
		fmt.Fprintf(w, "data: %s\n\n", output)
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}
