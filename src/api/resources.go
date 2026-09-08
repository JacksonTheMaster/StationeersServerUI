package api

import "time"

type Message struct {
	Message string `json:"message"`
}

type ServerStatus struct {
	Running       bool       `json:"running"`
	State         string     `json:"state"`
	StartedAt     *time.Time `json:"startedAt"`
	UptimeSeconds int64      `json:"uptimeSeconds"`
	ServerID      string     `json:"serverId"`
	WorldID       string     `json:"worldId"`
}

type Player struct {
	Username string `json:"username"`
	SteamID  string `json:"steamId"`
}

type SSCMStatus struct {
	Enabled bool `json:"enabled"`
}

type CommandReceipt struct {
	Command string `json:"command"`
	State   string `json:"state"`
}

type RestartReceipt struct {
	Message string `json:"message"`
	State   string `json:"state"`
}
