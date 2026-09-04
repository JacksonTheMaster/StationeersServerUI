package discordbot

import (
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"

	"github.com/bwmarrin/discordgo"
)

var discordRuntimeMutex sync.Mutex
var discordRuntimeStop chan struct{}

// InitializeDiscordBot starts or restarts the Discord bot and connects it to the Discord API.
func InitializeDiscordBot() {
	discordRuntimeMutex.Lock()
	defer discordRuntimeMutex.Unlock()
	stopDiscordRuntime()
	if !config.GetIsDiscordEnabled() {
		return
	}
	prepareDiscordRuntimeState()

	// Clean up previous session
	if previous := config.GetDiscordSession(); previous != nil {
		logger.Discord.Debug("Previous Discord session found, closing it...")
		previous.Close()
	}
	config.ConfigMu.Lock()
	config.DiscordSession = nil
	config.ConfigMu.Unlock()
	statusPanelMutex.Lock()
	statusPanelMessageID = ""
	statusPanelMutex.Unlock()
	if BufferFlushTicker != nil {
		BufferFlushTicker.Stop()
	}

	// Create new session
	session, err := discordgo.New("Bot " + config.GetDiscordToken())
	if err != nil {
		logger.Discord.Error("Error creating Discord session: " + err.Error())
		return
	}

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuilds

	logger.Discord.Info("Starting Discord integration...")
	//logger.Discord.Debug("Discord token: " + config.GetDiscordToken())
	logger.Discord.Debug("DiscordAdminRoleID: " + config.GetDiscordAdminRoleID())
	if config.GetDiscordAdminRoleID() == "" {
		logger.Discord.Warn("No Discord Admin Role ID configured; admin actions are disabled")
	}
	logger.Discord.Debug("EventLogChannelID: " + config.GetEventLogChannelID())
	logger.Discord.Debug("StatusPanelChannelID: " + config.GetStatusPanelChannelID())
	logger.Discord.Debug("LogChannelID: " + config.GetLogChannelID())

	// Register handlers before opening the gateway, including on reconnects.
	session.AddHandler(listenToSlashCommands)
	session.AddHandler(handlePanelButtonInteraction)
	session.AddHandler(handleAdminInteraction)
	// Open session first
	err = session.Open()
	if err != nil {
		logger.Discord.Error("Error opening Discord connection: " + err.Error())
		return
	}
	config.ConfigMu.Lock()
	config.DiscordSession = session
	config.ConfigMu.Unlock()
	syncApplicationEmojis(session)
	initializeDiscordBackupSummary()

	registerSlashCommands(session)

	logger.Discord.Info("Bot is now running.")
	SendMessageToEventLogChannel("🤖 SSUI Version " + config.GetVersion() + " connected to Discord.")
	sendServerStatusPanel() // Send server status panel to Discord
	UpdateBotStatusWithMessage("StationeersServerUI v" + config.GetVersion())
	// Start buffer flush ticker
	BufferFlushTicker = time.NewTicker(5 * time.Second)
	flushTicker := BufferFlushTicker
	hubTicker := time.NewTicker(15 * time.Second)
	done := make(chan struct{})
	discordRuntimeStop = done
	go func() {
		defer flushTicker.Stop()
		defer hubTicker.Stop()
		for {
			select {
			case <-done:
				return
			case <-flushTicker.C:
				flushLogBufferToDiscord()
			case <-hubTicker.C:
				refreshStatusPanel()
			}
		}
	}()
}

// Caller holds discordRuntimeMutex. Stopping a ticker alone would leave its
// goroutine waiting forever; the stop channel also releases the receive loop.
func stopDiscordRuntime() {
	if discordRuntimeStop != nil {
		close(discordRuntimeStop)
		discordRuntimeStop = nil
	}
}

// Updates the bot status with a string message
func UpdateBotStatusWithMessage(message string) {
	session := config.GetDiscordSession()
	if session == nil {
		return
	}
	err := session.UpdateGameStatus(0, message)
	if err != nil {
		logger.Discord.Error("Error updating bot status: " + err.Error())
	}
}
