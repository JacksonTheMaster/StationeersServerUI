// logstream.go
package detectionmgr

import (
	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/ssestream"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/discordbot"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

/*
Real-time Log Processing Pipeline
- Bridges internal SSE stream to detection system
- Performs log enrichment and distribution:
  - Adds logs to the Discord integrations log buffer if enabled
  - Feeds messages to Detector
*/

// StartLogStream starts processing logs directly from the internal SSE manager
func StreamLogs(detector *Detector) {
	logChan := ssestream.ConsoleStreamManager.AddInternalSubscriber()

	go func() {
		logger.Detection.Debug("Connected to internal log stream.")
		for logMessage := range logChan {
			if config.GetIsDiscordEnabled() {
				discordbot.PassLogStreamToDiscordLogBuffer(logMessage)
			}
			ProcessLog(detector, logMessage)
		}
	}()
}
