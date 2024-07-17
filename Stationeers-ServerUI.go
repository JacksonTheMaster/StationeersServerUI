package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var cmd *exec.Cmd
var mu sync.Mutex
var outputChannel chan string

func main() {
	outputChannel = make(chan string, 100)

	http.HandleFunc("/", serveUI)
	http.HandleFunc("/start", startServer)
	http.HandleFunc("/stop", stopServer)
	http.HandleFunc("/output", getOutput)
	http.HandleFunc("/backups", listBackups)
	http.HandleFunc("/restore", restoreBackup)
	http.ListenAndServe(":8080", nil)
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./UIMod/index.html")
}

func startServer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		fmt.Fprintf(w, "Server is already running.")
		return
	}

	cmd = exec.Command("powershell.exe", "-Command", "./rocketstation_DedicatedServer.exe -LOAD MarsProd Mars -settings StartLocalHost true ServerVisible true GamePort 27016 UpdatePort 27015 AutoSave true SaveInterval 300 LocalIpAddress 10.10.50.99 ServerPassword JMG AdminPassword JMG ServerMaxPlayers 4 ServerName SpaceInc")
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
	basePath := "./saves/MarsProd/backup"
	files, err := os.ReadDir(basePath)
	if err != nil {
		http.Error(w, "Unable to read backups directory", http.StatusInternalServerError)
		return
	}

	var backups []string
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
			}
		}
	}

	for index := range backupIndices {
		backups = append(backups, fmt.Sprintf("Index: %d", index))
	}

	if len(backups) == 0 {
		fmt.Fprint(w, "No valid backup files found.")
		return
	}

	response := strings.Join(backups, "\n")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, response)
}
func parseBackupIndex(fileName string) int {
	// Example file names: world_meta(50).xml
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

	backupDir := "./saves/MarsProd/backup"
	saveDir := "./saves/MarsProd"
	files := []struct {
		backupName string
		destName   string
	}{
		{fmt.Sprintf("world_meta(%d).xml", index), "world_meta.xml"},
		{fmt.Sprintf("world(%d).xml", index), "world.xml"},
		{fmt.Sprintf("world(%d).bin", index), "world.bin"},
	}

	for _, file := range files {
		backupFile := filepath.Join(backupDir, file.backupName)
		destFile := filepath.Join(saveDir, file.destName)

		err := os.Rename(backupFile, destFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error restoring file %s: %v", file.backupName, err), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "Backup %d restored successfully.", index)
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
