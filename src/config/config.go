package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	DiscordToken            string `json:"discordToken"`
	ControlChannelID        string `json:"controlChannelID"`
	StatusChannelID         string `json:"statusChannelID"`
	ConnectionListChannelID string `json:"connectionListChannelID"`
	LogChannelID            string `json:"logChannelID"`
	SaveChannelID           string `json:"saveChannelID"`
	ControlPanelChannelID   string `json:"controlPanelChannelID"`
	BlackListFilePath       string `json:"blackListFilePath"`
	IsDiscordEnabled        bool   `json:"isDiscordEnabled"`
	ErrorChannelID          string `json:"errorChannelID"`
}

var (
	DiscordToken              string
	ControlChannelID          string
	StatusChannelID           string
	LogChannelID              string
	ErrorChannelID            string
	ConnectionListChannelID   string
	SaveChannelID             string
	BlackListFilePath         string
	DiscordSession            *discordgo.Session
	LogMessageBuffer          string
	MaxBufferSize             = 1000
	BufferFlushTicker         *time.Ticker
	ConnectedPlayers          = make(map[string]string) // SteamID -> Username
	ConnectedPlayersMessageID string
	ControlMessageID          string
	ExceptionMessageID        string
	BackupRestoreMessageID    string
	ControlPanelChannelID     string
	IsDiscordEnabled          bool
	Version                   = "2.4.0"
	Branch                    = "Installer"
)

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	//print all the values to console
	fmt.Println("DiscordToken:", config.DiscordToken)
	fmt.Println("ControlChannelID:", config.ControlChannelID)
	fmt.Println("StatusChannelID:", config.StatusChannelID)
	fmt.Println("ConnectionListChannelID:", config.ConnectionListChannelID)
	fmt.Println("LogChannelID:", config.LogChannelID)
	fmt.Println("SaveChannelID:", config.SaveChannelID)
	fmt.Println("BlackListFilePath:", config.BlackListFilePath)
	fmt.Println("IsDiscordEnabled:", config.IsDiscordEnabled)
	fmt.Println("ErrorChannelID:", config.ErrorChannelID)
	DiscordToken = config.DiscordToken
	ControlChannelID = config.ControlChannelID
	StatusChannelID = config.StatusChannelID
	LogChannelID = config.LogChannelID
	ConnectionListChannelID = config.ConnectionListChannelID
	SaveChannelID = config.SaveChannelID
	BlackListFilePath = config.BlackListFilePath
	ControlPanelChannelID = config.ControlPanelChannelID
	IsDiscordEnabled = config.IsDiscordEnabled
	ErrorChannelID = config.ErrorChannelID
	return &config, nil
}
