package loader

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/backupmgr"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
)

// HandleShutdownSignals gives the managed game process and backup manager a
// chance to stop before Docker or the service manager tears down SSUI.
func HandleShutdownSignals() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Main.Infof("Received %s, shutting down...", sig)

		done := make(chan struct{})
		go func() {
			if config.GetIsGameServerRunning() {
				if err := gamemgr.InternalStopServer(); err != nil {
					logger.Main.Warnf("Could not stop the game server during shutdown: %v", err)
				}
			}
			if manager := backupmgr.CurrentBackupManager(); manager != nil {
				manager.Shutdown()
			}
			close(done)
		}()

		select {
		case <-done:
			logger.Main.Info("Shutdown complete")
			os.Exit(0)
		case second := <-signals:
			logger.Main.Warnf("Received %s during shutdown, exiting immediately", second)
			os.Exit(1)
		case <-time.After(25 * time.Second):
			logger.Main.Error("Graceful shutdown timed out")
			os.Exit(1)
		}
	}()
}
