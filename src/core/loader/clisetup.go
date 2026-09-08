// clisetup.go - Interactive CLI setup track for first-time configuration
package loader

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/mattn/go-isatty"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorPurple = "\033[35m"
	colorBold   = "\033[1m"
)

// promptStyle adds consistent styling to prompts
func promptStyle(s string) string {
	return colorPurple + colorBold + s + colorReset
}

// IsInteractiveTerminal checks if we're running in an interactive terminal
func IsInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// PromptSetupTrackChoice asks the user which setup track they want to use
// Returns "cli", "web", or "skip"
func PromptSetupTrackChoice() string {
	if !IsInteractiveTerminal() {
		// Non-interactive, default to web silently
		return "web"
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(promptStyle("                      " + colorBold + "SSUI First-Time Setup" + colorReset))
	fmt.Println(promptStyle(" How would you like to configure your server?                              "))
	fmt.Println(promptStyle("                                                                            "))
	fmt.Println(promptStyle("   [1] " + colorGreen + "CLI Setup" + colorReset + " - Configure right here in the terminal"))
	fmt.Println(promptStyle("   [2] " + colorCyan + "Web Setup" + colorReset + " - Use the browser-based setup wizard"))
	fmt.Println(promptStyle("   [3] " + colorYellow + "Skip" + colorReset + "      - Use defaults and configure later"))
	fmt.Println()
	fmt.Print(promptStyle("\n  Enter choice [1/2/3]: "))

	input, err := reader.ReadString('\n')
	if err != nil {
		return "web"
	}

	input = strings.TrimSpace(input)
	switch input {
	case "1", "cli", "CLI":
		return "cli"
	case "2", "web", "Web", "WEB":
		return "web"
	case "3", "skip", "Skip", "SKIP":
		// Mark setup as complete so it doesn't prompt on web UI either
		config.SetIsFirstTimeSetup(false)
		return "skip"
	default:
		fmt.Println(colorCyan + "  Invalid choice, defaulting to Web Setup..." + colorReset)
		return "web"
	}
}
