package dashboard

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Panel System

type Panel int // holds current active panel
type ConfigSection int

// ConfigItem represents a single editable config item
type ConfigItem struct {
	Key         string
	Label       string
	Value       string
	Type        string // "string", "bool", "int"
	Section     ConfigSection
	Description string
}

const (
	PanelStatus  Panel = iota // Server status overview (default)
	PanelSSUILog              // SSUI backend logs
	PanelPlayers              // Connected players list
	PanelConfig               // Configuration editor
	PanelCount                // Used for cycling
)

var panelNames = map[Panel]string{
	PanelStatus:  "Status",
	PanelSSUILog: "Logs",
	PanelPlayers: "Players",
	PanelConfig:  "Config",
}

var panelIcons = map[Panel]string{
	PanelStatus:  "📊",
	PanelSSUILog: "📜",
	PanelPlayers: "👥",
	PanelConfig:  "⚙",
}

const (
	ConfigSectionBasic ConfigSection = iota
	ConfigSectionNetwork
	ConfigSectionAdvanced
	ConfigSectionWorldGen
	ConfigSectionCount
)

var configSectionNames = map[ConfigSection]string{
	ConfigSectionBasic:    "Basic Settings",
	ConfigSectionNetwork:  "Network Settings",
	ConfigSectionAdvanced: "Advanced Settings",
	ConfigSectionWorldGen: "World Generation",
}

// Model - Dashboard State

// Model represents the dashboard state
type Model struct {

	// Window and Layout

	width  int
	height int

	// Navigation

	activePanel Panel

	// Components

	logViewport viewport.Model // For SSUI backend logs
	help        help.Model
	keys        keyMap

	// Server State

	serverRunning    bool
	serverUptime     time.Duration
	serverStartTime  time.Time
	serverName       string
	saveName         string
	worldID          string
	gamePort         string
	updatePort       string
	maxPlayers       int
	connectedPlayers map[string]string // SteamID -> PlayerName

	// Game Info

	gameVersion string
	gameBranch  string
	buildID     string

	// SSUI Info

	ssuiVersion string
	goRuntime   string
	isDocker    bool

	// Feature Flags

	discordEnabled     bool
	autoSaveEnabled    bool
	autoRestartEnabled bool
	autoRestartTimer   string
	upnpEnabled        bool
	authEnabled        bool
	bepInExEnabled     bool
	saveInterval       string
	autoStartEnabled   bool

	// Backup Info

	backupKeepNewestCount        int
	backupDailyRetentionDays     int
	backupWeeklyRetentionWeeks   int
	backupMonthlyRetentionMonths int

	// Config Panel State

	configItems         []ConfigItem           // All config items
	configSectionOpen   map[ConfigSection]bool // Which sections are expanded
	configSelectedIndex int                    // Currently selected config item
	configEditing       bool                   // Whether we're editing a value
	configEditValue     string                 // Current edit buffer
	configCursorPos     int                    // Cursor position in edit
	configHasChanges    bool                   // Unsaved changes flag
	configStatusMsg     string                 // Status message (e.g., "Saved!")
	configStatusTick    int                    // Countdown for status message

	// Internal State

	err          error
	quitting     bool
	startTime    time.Time
	tickCount    int
	lastRefresh  time.Time
	showFullHelp bool
}

// Key Bindings

type keyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Start    key.Binding
	Stop     key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
	Enter    key.Binding // For config editing
	Space    key.Binding // Toggle sections/bools
	Save     key.Binding // Save config changes
}

// ShortHelp returns the short help text (displayed in footer)
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Start, k.Stop, k.Refresh, k.Help, k.Quit}
}

// FullHelp returns the full help text (displayed when '?' is pressed)
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.Up, k.Down},
		{k.PageUp, k.PageDown, k.Home, k.End},
		{k.Start, k.Stop, k.Refresh},
		{k.Enter, k.Space, k.Save},
		{k.Help, k.Quit},
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up/scroll"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down/scroll"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "bottom"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start server"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop server"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit/confirm"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save config"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "exit"),
		),
	}
}

// Model Initialization

// NewModel creates a new dashboard model with initial values
func NewModel() Model {
	vp := viewport.New(80, 20)
	vp.SetContent("Initializing log viewer...")

	h := help.New()
	h.ShowAll = false

	// Initialize config sections as collapsed
	sectionOpen := make(map[ConfigSection]bool)
	for i := range ConfigSectionCount {
		sectionOpen[i] = false
	}
	// Open Basic section by default
	sectionOpen[ConfigSectionBasic] = true

	return Model{
		activePanel:       PanelStatus, // Default
		logViewport:       vp,
		help:              h,
		keys:              defaultKeyMap(),
		startTime:         time.Now(),
		lastRefresh:       time.Now(),
		connectedPlayers:  make(map[string]string),
		ssuiVersion:       "...",
		goRuntime:         "...",
		configSectionOpen: sectionOpen,
		configItems:       buildConfigItems(), // Populate config vals
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.SetWindowTitle("SSUI Dashboard"),
	)
}

// Messages

// tickMsg is sent periodically to update the dashboard
type tickMsg time.Time

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// statusUpdateMsg is sent when server status is updated
type statusUpdateMsg struct {
	// Server state
	running    bool
	uptime     time.Duration
	startTime  time.Time
	serverName string
	saveName   string
	worldID    string
	gamePort   string
	updatePort string
	maxPlayers int
	players    map[string]string

	// Game info
	gameVersion string
	gameBranch  string
	buildID     string

	// SSUI info
	version   string
	goRuntime string
	isDocker  bool

	// Features
	discordEnabled     bool
	autoSaveEnabled    bool
	autoRestartEnabled bool
	autoRestartTimer   string
	upnpEnabled        bool
	authEnabled        bool
	bepInExEnabled     bool
	saveInterval       string
	autoStartEnabled   bool

	// Backup
	backupKeepNewestCount        int
	backupDailyRetentionDays     int
	backupWeeklyRetentionWeeks   int
	backupMonthlyRetentionMonths int
}

// logUpdateMsg is sent when new SSUI logs are available
type logUpdateMsg struct {
	logs []string
}

// serverActionMsg is sent after a server action (start/stop)
type serverActionMsg struct {
	action string
	err    error
}

// Helpers

// formatUptime formats duration to a human-readable string
func formatUptime(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}
	d = d.Round(time.Second)

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// formatUptimeShort returns a shorter uptime format
func formatUptimeShort(d time.Duration) string {
	if d == 0 {
		return "--:--:--"
	}
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// parseMaxPlayers converts the string max players to int
func parseMaxPlayers(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
