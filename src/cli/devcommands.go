package cli

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/discordbot"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/localization"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/modding"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/steamcmd"
)

// COMMAND HANDLERS WITH COMMANDS USEFUL FOR DEVELOPMENT AND DEBUGGING

func downloadWorkshopItem(args []string) error {
	var workshopHandles []string
	if len(args) == 0 {
		workshopHandles = []string{"3672138641"} // blueprint mod
	} else {
		workshopHandles = args
	}
	logger.Core.Info(fmt.Sprintf("Downloading workshop items: %v", workshopHandles))
	_, err := steamcmd.DownloadWorkshopItems(workshopHandles)
	if err != nil {
		logger.Core.Error("Error downloading workshop items: " + err.Error())
	}
	return nil
}

func listworkshophandles() {
	handles := modding.GetModWorkshopHandles()
	if len(handles) == 0 {
		logger.Core.Info("No mods with Workshop handles found.")
		return
	}
	logger.Core.Info(fmt.Sprintf("Installed Mod Workshop Handles: (%d):", len(handles)))
	logger.Modding.Info(fmt.Sprintf("%v", handles))

}

func listmods() {
	mods := modding.GetModList()
	if len(mods) == 0 {
		logger.Core.Info("No mods installed.")
		return
	}
	logger.Core.Info(fmt.Sprintf("Installed Mods (%d):", len(mods)))
	for _, mod := range mods {

		// print mod details in one logger call but with /n for new lines
		logger.Modding.Info("Mod Details:\n" +
			fmt.Sprintf("Modname: %s\n", mod.Name) +
			fmt.Sprintf("  Version: %s\n", mod.Version) +
			fmt.Sprintf("  Author: %s\n", mod.Author) +
			fmt.Sprintf("  Workshop Handle: %s\n", mod.WorkshopHandle))
	}
}

func testLocalization() {
	currentLanguageSetting := config.GetLanguageSetting()
	s := localization.GetString("UIText_StartButton")
	logger.Core.Info("Start Server Button text (current language: " + currentLanguageSetting + "): " + s)
}

func dumpHeapProfile() {
	runtime.GC()
	if _, err := os.Stat("heap.pprof"); err == nil {
		if err := os.Remove("heap.pprof"); err != nil {
			logger.Main.Errorf("could not remove old heap profile: %v", err)
			return
		}
	}
	f, err := os.Create("heap.pprof")
	if err != nil {
		logger.Main.Errorf("could not create heap profile file: %v", err)
		return
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		logger.Main.Errorf("could not write heap profile: %v", err)
		return
	}

	logger.Main.Info("Heap profile written to heap.pprof")
}

func testServerStatusPanelDiscord() {
	players := map[string]string{
		"76561198334231312": "JacksonTheMaster",
		"76561198012262058": "Sebastian - TheNovice",
		"76561197995322389": "Non Action Man",
	}
	discordbot.UpdateStatusPanelPlayerConnected("ThisDataDoesntMatter", "ThisDataDoesntMatter", time.Now(), players)
}
