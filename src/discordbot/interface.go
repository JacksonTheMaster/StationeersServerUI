package discordbot

import (
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"

	"github.com/bwmarrin/discordgo"
)

// InitializeDiscordBot starts or restarts the Discord bot and connects it to the Discord API.
func InitializeDiscordBot() {
	var err error
	prepareDiscordRuntimeState()

	// Clean up previous session
	if config.DiscordSession != nil {
		logger.Discord.Debug("Previous Discord session found, closing it...")
		config.DiscordSession.Close()
	}
	if BufferFlushTicker != nil {
		BufferFlushTicker.Stop()
	}

	// Create new session
	config.DiscordSession, err = discordgo.New("Bot " + config.GetDiscordToken())
	if err != nil {
		logger.Discord.Error("Error creating Discord session: " + err.Error())
		return
	}

	// Set intents
	config.DiscordSession.Identify.Intents = discordgo.IntentsGuildMessageReactions

	logger.Discord.Info("Starting Discord integration...")
	//logger.Discord.Debug("Discord token: " + config.GetDiscordToken())
	logger.Discord.Debug("ControlChannelID: " + config.GetControlChannelID())
	logger.Discord.Debug("EventLogChannelID: " + config.GetEventLogChannelID())
	logger.Discord.Debug("StatusPanelChannelID: " + config.GetStatusPanelChannelID())
	logger.Discord.Debug("LogChannelID: " + config.GetLogChannelID())

	// Open session first
	err = config.DiscordSession.Open()
	if err != nil {
		logger.Discord.Error("Error opening Discord connection: " + err.Error())
		return
	}
	syncApplicationEmojis(config.DiscordSession)
	initializeDiscordBackupSummary()

	// Register handlers and commands after session is open
	config.DiscordSession.AddHandler(listenToDiscordReactions)
	config.DiscordSession.AddHandler(listenToSlashCommands)
	config.DiscordSession.AddHandler(handlePanelButtonInteraction)    // Handle button interactions (server info + players panel)
	config.DiscordSession.AddHandler(handleDownloadButtonInteraction) // Handle download button interactions
	registerSlashCommands(config.DiscordSession)

	logger.Discord.Info("Bot is now running.")
	SendMessageToEventLogChannel("🤖 SSUI Version " + config.GetVersion() + " connected to Discord.")
	sendControlPanel()      // Send control panel message to Discord
	sendServerStatusPanel() // Send server status panel to Discord
	UpdateBotStatusWithMessage("StationeersServerUI v" + config.GetVersion())
	// Start buffer flush ticker
	BufferFlushTicker = time.NewTicker(5 * time.Second)
	go func() {
		for range BufferFlushTicker.C {
			flushLogBufferToDiscord()
		}
	}()

	select {} // Keep it running
}

// Updates the bot status with a string message
func UpdateBotStatusWithMessage(message string) {
	err := config.DiscordSession.UpdateGameStatus(0, message)
	if err != nil {
		logger.Discord.Error("Error updating bot status: " + err.Error())
	}
}
