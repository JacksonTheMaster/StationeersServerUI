package api

import (
	"net/http"
	"os/exec"
	"sync"
)

var cmd *exec.Cmd
var mu sync.Mutex
var outputChannel chan string
var clients []chan string
var clientsMu sync.Mutex

type Config struct {
	Server struct {
		ExePath  string `xml:"exePath"`
		Settings string `xml:"settings"`
	} `xml:"server"`
	SaveFileName string `xml:"saveFileName"`
}

func StartAPI() {
	outputChannel = make(chan string, 100)
	go watchBackupDir()
}

func ServeUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./UIMod/index.html")
}
