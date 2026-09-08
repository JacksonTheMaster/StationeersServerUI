package gamemgr

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

type Arg struct {
	Flag          string
	Value         string
	RequiresValue bool
	Condition     func() bool
	NoQuote       bool
}

func buildCommandArgs() []string {
	var argOrder []Arg

	argOrder = []Arg{
		{Flag: "-nographics", RequiresValue: false},
		{Flag: "-batchmode", RequiresValue: false},
		/* file start: (expects up to four optional args for:
		-worldid (Optional to LOAD, required to CREATE save, if start value is not found, tries to create map with worldid)
		-difficulty (Optional, defaults to "Normal" if not provided)
		-startcondition  (Optional, defaults to the default start condition for the world setting if not provided.)
		-startlocation (Optional, defaults to "DefaultStartLocation" if not provided.)
		*/
		{Flag: "-file", RequiresValue: false},
		{Flag: "start", Value: config.GetSaveName(), RequiresValue: true},
		{Flag: config.GetWorldID(), RequiresValue: false},
		{Flag: config.GetDifficulty(), RequiresValue: false, Condition: func() bool { return config.GetDifficulty() != "" }},
		{Flag: config.GetStartCondition(), RequiresValue: false, Condition: func() bool { return config.GetStartCondition() != "" }},
		{Flag: config.GetStartLocation(), RequiresValue: false, Condition: func() bool { return config.GetStartLocation() != "" }},
		// file start end
		{Flag: "-logFile", Value: "./debug.log", Condition: func() bool { return runtime.GOOS == "linux" }, RequiresValue: true},
		{Flag: "-settings", RequiresValue: false},
		{Flag: "StartLocalHost", Value: strconv.FormatBool(config.GetStartLocalHost()), RequiresValue: true},
		{Flag: "ServerVisible", Value: strconv.FormatBool(config.GetServerVisible()), RequiresValue: true},
		{Flag: "GamePort", Value: config.GetGamePort(), RequiresValue: true},
		{Flag: "UPNPEnabled", Value: strconv.FormatBool(config.GetUPNPEnabled()), RequiresValue: true},
		{Flag: "ServerName", Value: config.GetServerName(), RequiresValue: true},
		{Flag: "ServerPassword", Value: config.GetServerPassword(), Condition: func() bool { return config.GetServerPassword() != "" }, RequiresValue: true},
		{Flag: "ServerMaxPlayers", Value: config.GetServerMaxPlayers(), RequiresValue: true},
		{Flag: "AutoSave", Value: strconv.FormatBool(config.GetAutoSave()), RequiresValue: true},
		{Flag: "SaveInterval", Value: config.GetSaveInterval(), RequiresValue: true},
		{Flag: "ServerAuthSecret", Value: config.GetServerAuthSecret(), Condition: func() bool { return config.GetServerAuthSecret() != "" }, RequiresValue: true},
		{Flag: "UpdatePort", Value: config.GetUpdatePort(), RequiresValue: true},
		{Flag: "AutoPauseServer", Value: strconv.FormatBool(config.GetAutoPauseServer()), RequiresValue: true},
		{Flag: "UseSteamP2P", Value: strconv.FormatBool(config.GetUseSteamP2P()), RequiresValue: true},
		{Flag: "AdminPassword", Value: config.GetAdminPassword(), Condition: func() bool { return config.GetAdminPassword() != "" }, RequiresValue: true},
	}

	var args []string
	for _, arg := range argOrder {
		if arg.Condition != nil && !arg.Condition() {
			continue
		}
		if arg.RequiresValue && arg.Value == "" {
			continue
		}

		args = append(args, arg.Flag)

		if arg.Value != "" {
			args = append(args, arg.Value)
		}
	}

	if config.GetAdditionalParams() != "" {
		args = append(args, strings.Fields(config.GetAdditionalParams())...)
	}

	if config.GetLocalIpAddress() != "" {
		args = append(args, "LocalIpAddress")
		args = append(args, config.GetLocalIpAddress())
	}

	return args
}
