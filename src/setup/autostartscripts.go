package setup

import (
	"io"
	"io/fs"
	"os"
	"runtime"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

func SetupAutostartScripts() {
	scriptFS, err := fs.Sub(config.V1UIFS, "SSUI/onboard_bundled/scripts")
	if err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		script, err := scriptFS.Open("autostart.ps1")
		if err != nil {
			return
		}
		defer script.Close()
		data, err := io.ReadAll(script)
		if err != nil {
			return
		}
		err = os.WriteFile("autostart.ps1", data, 0755)
		if err != nil {
			return
		}
	}
	if runtime.GOOS == "linux" {
		script, err := scriptFS.Open("autostart.sh")
		if err != nil {
			return
		}
		defer script.Close()
		data, err := io.ReadAll(script)
		if err != nil {
			return
		}
		err = os.WriteFile("autostart.service", data, 0755)
		if err != nil {
			return
		}
	}
}
