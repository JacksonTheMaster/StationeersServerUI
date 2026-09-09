package loader

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/security"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// Define flags matching the config variable names
var backendEndpointPortFlag string
var gamePortFlag string
var gameBranchFlag string
var logLevelFlag int
var isDebugModeFlag bool
var createSSUILogFileFlag bool
var recoveryPasswordFlag string
var devModeFlag bool
var skipSteamCMDFlag bool
var sanityCheckFlag bool
var advertiserOverrideFlag string

// ParseFlags parses command-line arguments ONCE at startup (called from func main)
func ParseFlags() {
	flag.StringVar(&backendEndpointPortFlag, "BackendEndpointPort", "", "Override the backend endpoint port (e.g., 8080)")
	flag.StringVar(&gamePortFlag, "GamePort", "", " Override the game endpoint port (e.g., 27018)")
	flag.StringVar(&backendEndpointPortFlag, "p", "", "(Alias) Override the backend endpoint port (e.g., 8080)")
	flag.StringVar(&gameBranchFlag, "GameBranch", "", "Override the game branch (e.g., beta)")
	flag.StringVar(&gameBranchFlag, "b", "", "(Alias) Override the game branch (e.g., beta)")
	flag.StringVar(&recoveryPasswordFlag, "RecoveryPassword", "", "Adds a 'recovery' user (expects password as argument)")
	flag.StringVar(&recoveryPasswordFlag, "r", "", "(Alias) Adds a 'recovery' user (expects password as argument)")
	flag.BoolVar(&devModeFlag, "dev", false, "Enable dev mode: Auth, and enables cli-console. For development only.")
	flag.IntVar(&logLevelFlag, "LogLevel", 0, "Override the log level (e.g., 10)")
	flag.IntVar(&logLevelFlag, "ll", 0, "(Alias) Override the log level (e.g., 10)")
	flag.BoolVar(&isDebugModeFlag, "IsDebugMode", false, "Enable debug mode")
	flag.BoolVar(&isDebugModeFlag, "debug", false, "(Alias) Enable debug mode")
	flag.BoolVar(&createSSUILogFileFlag, "LogToFiles", false, "Create log files for SSUI")
	flag.BoolVar(&createSSUILogFileFlag, "lf", false, "(Alias) Create log files for SSUI")
	flag.BoolVar(&skipSteamCMDFlag, "NoSteamCMD", false, "Skips SteamCMD installation")
	flag.BoolVar(&sanityCheckFlag, "NoSanityCheck", false, "Skips the sanity check. Not recommended.")
	flag.StringVar(&advertiserOverrideFlag, "AdvertiserOverride", "", "Override the advertised server IP. For this, the ServerVisible setting must be set to false. Use \"auto\" for automatic public IP detection, an IPv4 address, or a DNS hostname (to allow server advertisement if you are behind a reverse proxy)")

	// Parse command-line flags
	flag.Parse()
}

// HandleCmdArgs handles command-line arguments ONCE at startup (called from func main) and applies them using the config setters.
// Because this is using the config rather than adding features to it, it is a part of the loader package.
func HandleFlags() {

	if devModeFlag {
		config.SetAuthEnabled(true)
		config.SetIsFirstTimeSetup(false)
		config.SetIsConsoleEnabled(true)
		logger.Main.Info("Dev mode enabled: console enabled; development owner will be configured after identity startup")
	}

	if skipSteamCMDFlag {
		config.SetSkipSteamCMD(true)
	}

	if backendEndpointPortFlag != "" {
		oldPort := config.GetSSUIWebPort()
		config.SetSSUIWebPort(backendEndpointPortFlag)
		logger.Main.Info(fmt.Sprintf("Overriding SetSSUIWebPort from command line: Before=%s, Now=%s", oldPort, backendEndpointPortFlag))
	}
	if gamePortFlag != "" {
		oldPort := config.GetGamePort()
		config.SetGamePort(gamePortFlag)
		logger.Main.Info(fmt.Sprintf("Overriding GamePort from command line: Before=%s, Now=%s", oldPort, gamePortFlag))
	}

	if gameBranchFlag != "" {
		oldBranch := config.GetGameBranch()
		config.SetGameBranch(gameBranchFlag)
		logger.Main.Info(fmt.Sprintf("Overriding GameBranch from command line: Before=%s, Now=%s", oldBranch, gameBranchFlag))
	}

	if logLevelFlag != 0 {
		oldLevel := config.GetLogLevel()
		config.SetLogLevel(logLevelFlag)
		logger.Main.Info(fmt.Sprintf("Overriding LogLevel from command line: Before=%d, Now=%d", oldLevel, logLevelFlag))
	}

	if isDebugModeFlag {
		oldDebug := config.GetIsDebugMode()
		config.SetIsDebugMode(true)
		config.SetLogLevel(10)
		logger.Main.Info(fmt.Sprintf("Overriding IsDebugMode from command line: Before=%t, Now=true", oldDebug))
	}

	if advertiserOverrideFlag != "" {
		oldAdvertiserOverride := config.GetAdvertiserOverride()

		if advertiserOverrideFlag == oldAdvertiserOverride {
			logger.Advertiser.Info(fmt.Sprintf("Advertised Server IP is already set to %s", advertiserOverrideFlag))
			return
		}
		config.SetAdvertiserOverride(advertiserOverrideFlag)
		logger.Advertiser.Info(fmt.Sprintf("Overriding Advertised Server IP from command line: Before=%s, Now=%s", oldAdvertiserOverride, advertiserOverrideFlag))
	}

	if createSSUILogFileFlag {
		oldCreateSSUILogFile := config.GetCreateSSUILogFile()
		config.SetCreateSSUILogFile(true)
		logger.Main.Info(fmt.Sprintf("Overriding CreateSSUILogFile from command line: Before=%t, Now=true", oldCreateSSUILogFile))
	}
}

// HandleIdentityFlags runs after the identity store has been initialized.
func HandleIdentityFlags() {
	password := strings.TrimSpace(recoveryPasswordFlag)
	if devModeFlag {
		if _, err := security.EnableDevelopmentOwner(time.Now()); err != nil {
			logger.Security.Error(fmt.Sprintf("Failed to configure development owner: %v", err))
		} else {
			logger.Security.Warn("Development owner enabled: admin:admin with full access. For development only.")
		}
	}
	if recoveryPasswordFlag == "" {
		removed, err := security.RemoveRecoveryOwner(time.Now())
		if errors.Is(err, security.ErrRecoveryOwnerStillRequired) {
			logger.Security.Warn("Temporary recovery account is still the only owner. Create another owner before restarting SSUI.")
		} else if err != nil {
			logger.Security.Error(fmt.Sprintf("Failed to remove temporary recovery account: %v", err))
		} else if removed {
			logger.Security.Info("Removed temporary recovery account.")
		}
		return
	}
	if password == "" {
		logger.Security.Error("Recovery flag provided but password is empty. Skipping owner recovery.")
		return
	}
	if _, err := security.RecoverOwner("recovery", password, time.Now()); err != nil {
		logger.Security.Error(fmt.Sprintf("Failed to recover owner account: %v", err))
		return
	}
	logger.Security.Warn("Recovered owner account 'recovery'. Existing sessions and API tokens were revoked.")
}

// HandleSanityCheckFlag has special handling to allow usage directly at startup before other systems are initialized.
func HandleSanityCheckFlag() {
	if sanityCheckFlag {
		config.NoSanityCheck = true
		logger.Main.Warn("Sanity check flag enabled, skipping sanity check. Not recommended.")
		logger.Main.Info("Sleeping for 5 seconds to remind you again to not use this flag in production.")
		time.Sleep(5 * time.Second)
	}
}
