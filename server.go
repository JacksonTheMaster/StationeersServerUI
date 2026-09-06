// Package main is the entry point for StationeersServerUI, a tool for managing Stationeers servers.
// It coordinates setup, configuration, logging, resource loading, and a web-based UI.
//
// The server initializes by running the setup process, loading resources, and starting a web server.
// Key functionality is provided by the following subpackages:
//   - src/config: Manages server configuration.
//   - src/configchanger: Handles configuration changes.
//   - src/gamemgr: Manages process management.
//   - src/loader: Handles resource loading and detection initialization.
//   - src/logger: Provides logging utilities.
//   - src/setup: Performs initial server setup.
//   - src/web: Runs the web-based user interface.
//   - src/discordbot: Handles Discord bot functionality.
//   - src/backupmgr: Manages backups of the server's world.
//   - src/detectionmgr: Manages event detection and processing.
//   - src/setup: Performs initial server setup.
//   - src/ssestream: Handles Server-Sent Events (SSE) streaming.
//   - src/security: Handles security-related tasks.
//
// For detailed documentation, see the subpackages or the project Wiki on GitHub.
package main

import (
	"embed"
	"os"
	"sync"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/cli"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/loader"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/web"
)

//go:embed SSUI/onboard_bundled
var v1uiFS embed.FS

func main() {
	var wg sync.WaitGroup
	logger.ConfigureConsole()
	loader.ParseFlags()
	loader.HandleSanityCheckFlag()
	loader.SanityCheck()
	if migrated, err := setup.MigrateLegacyRuntimeFolder(); err != nil {
		logger.Main.Error("Failed to migrate UIMod to SSUI: " + err.Error())
		os.Exit(1)
	} else if migrated {
		logger.Main.Info("Copied UIMod data to SSUI. The old UIMod folder was left untouched.")
	}
	logger.Main.Info("Initializing resources...")
	loader.InitVirtFS(v1uiFS)
	logger.Install.Info("Starting setup...")
	loader.ReloadConfig() // Load the config file before starting the setup process
	loader.HandleFlags()
	setup.Install(&wg)
	wg.Wait()
	logger.Main.Debug("Initializing Backend...")
	loader.InitBackend()
	logger.Main.Debug("Starting webserver...")
	web.StartWebServer(&wg)
	logger.Main.Debug("Initializing after start tasks...")
	loader.AfterStartComplete()
	logger.Main.Debug("Initializing SSUICLI...")
	cli.StartConsole(&wg)
	wg.Wait()
}
