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
	BlackListFilePath       string `json:"blackListFilePath"`
}

var (
	//discordToken      = "MTI3NTA1Mjk5Mjg0ODIwMzc3OA.GXBztW.UAa7ijUAsbu5hOtswa6IxXZn_d-QRH_bnpfFBw"
	//controlChannelID  = "1275055797616771123"
	//statusChannelID   = "1276701394543313038"
	//logChannelID      = "1275067875647819830"
	//saveChannelID     = "1276705219518140416"
	//blackListFilePath = "C:/SteamCMD/Stationeers/Blacklist.txt"
	DiscordToken              string
	ControlChannelID          string
	StatusChannelID           string
	LogChannelID              string
	ConnectionListChannelID   string
	SaveChannelID             string
	BlackListFilePath         string
	DiscordSession            *discordgo.Session
	LogMessageBuffer          string
	MaxBufferSize             = 1000
	BufferFlushTicker         *time.Ticker
	ConnectedPlayers          = make(map[string]string) // SteamID -> Username
	ConnectedPlayersMessageID string
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
	return &config, nil
}
