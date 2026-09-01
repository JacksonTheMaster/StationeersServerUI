package dashboard

import (
	"runtime"
	"strings"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/detectionmgr"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/managers/gamemgr"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model - handles all messages and key presses
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key presses

		// When editing config, ONLY handle editing-related keys
		if m.configEditing && m.activePanel == PanelConfig {
			switch msg.Type {
			case tea.KeyEnter:
				// Confirm edit - save the new value
				item, isHeader, _ := m.getConfigItemAtIndex(m.configSelectedIndex)
				if item != nil && !isHeader {
					for i := range m.configItems {
						if m.configItems[i].Key == item.Key {
							m.configItems[i].Value = m.configEditValue
							break
						}
					}
					m.configHasChanges = true
				}
				m.configEditing = false
				m.configEditValue = ""
			case tea.KeyEscape:
				// Cancel editing
				m.configEditing = false
				m.configEditValue = ""
			case tea.KeyBackspace:
				if len(m.configEditValue) > 0 {
					m.configEditValue = m.configEditValue[:len(m.configEditValue)-1]
				}
			case tea.KeySpace:
				// Insert a space character when editing
				m.configEditValue += " "
			case tea.KeyRunes:
				m.configEditValue += string(msg.Runes)
			}
			return m, nil // Don't process other keybinds while editing
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Tab):
			// Cycle forward through panels
			m.activePanel = (m.activePanel + 1) % PanelCount

		case key.Matches(msg, m.keys.ShiftTab):
			// Cycle backward through panels
			m.activePanel = (m.activePanel - 1 + PanelCount) % PanelCount

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.showFullHelp = m.help.ShowAll

		case key.Matches(msg, m.keys.Up):
			if m.activePanel == PanelSSUILog {
				m.logViewport.LineUp(3)
			} else if m.activePanel == PanelConfig && !m.configEditing {
				if m.configSelectedIndex > 0 {
					m.configSelectedIndex--
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.activePanel == PanelSSUILog {
				m.logViewport.LineDown(3)
			} else if m.activePanel == PanelConfig && !m.configEditing {
				if m.configSelectedIndex < m.getTotalConfigItems()-1 {
					m.configSelectedIndex++
				}
			}

		case key.Matches(msg, m.keys.PageUp):
			if m.activePanel == PanelSSUILog {
				m.logViewport.HalfPageUp()
			}

		case key.Matches(msg, m.keys.PageDown):
			if m.activePanel == PanelSSUILog {
				m.logViewport.HalfPageDown()
			}

		case key.Matches(msg, m.keys.Home):
			if m.activePanel == PanelSSUILog {
				m.logViewport.GotoTop()
			}

		case key.Matches(msg, m.keys.End):
			if m.activePanel == PanelSSUILog {
				m.logViewport.GotoBottom()
			}

		case key.Matches(msg, m.keys.Start):
			if !m.serverRunning {
				return m, startServerCmd()
			}

		case key.Matches(msg, m.keys.Stop):
			if m.serverRunning {
				return m, stopServerCmd()
			}

		case key.Matches(msg, m.keys.Refresh):
			// Force refresh of all data
			return m, tea.Batch(fetchStatusCmd(), fetchLogsCmd())

		case key.Matches(msg, m.keys.Enter):
			if m.activePanel == PanelConfig {
				if m.configEditing {
					// Confirm edit - save the new value
					item, isHeader, _ := m.getConfigItemAtIndex(m.configSelectedIndex)
					if item != nil && !isHeader {
						// Update the item value
						for i := range m.configItems {
							if m.configItems[i].Key == item.Key {
								m.configItems[i].Value = m.configEditValue
								break
							}
						}
						m.configHasChanges = true
					}
					m.configEditing = false
					m.configEditValue = ""
				} else {
					// Start editing
					item, isHeader, section := m.getConfigItemAtIndex(m.configSelectedIndex)
					if isHeader {
						// Toggle section open/closed
						m.configSectionOpen[section] = !m.configSectionOpen[section]
					} else if item != nil {
						if item.Type == "bool" {
							// Toggle bool directly
							if item.Value == "true" {
								item.Value = "false"
							} else {
								item.Value = "true"
							}
							// Update in configItems
							for i := range m.configItems {
								if m.configItems[i].Key == item.Key {
									m.configItems[i].Value = item.Value
									break
								}
							}
							m.configHasChanges = true
						} else {
							// Start text editing
							m.configEditing = true
							m.configEditValue = item.Value
						}
					}
				}
			}

		case key.Matches(msg, m.keys.Space):
			if m.activePanel == PanelConfig && !m.configEditing {
				item, isHeader, section := m.getConfigItemAtIndex(m.configSelectedIndex)
				if isHeader {
					// Toggle section open/closed
					m.configSectionOpen[section] = !m.configSectionOpen[section]
				} else if item != nil && item.Type == "bool" {
					// Toggle bool value
					newVal := "true"
					if item.Value == "true" {
						newVal = "false"
					}
					for i := range m.configItems {
						if m.configItems[i].Key == item.Key {
							m.configItems[i].Value = newVal
							break
						}
					}
					m.configHasChanges = true
				}
			}

		case key.Matches(msg, m.keys.Save):
			if m.activePanel == PanelConfig && m.configHasChanges && !m.configEditing {
				// Save all config changes
				err := saveAllConfigChanges(m.configItems)
				if err != nil {
					m.configStatusMsg = "✗ " + err.Error()
				} else {
					m.configStatusMsg = "✓ Saved & reloaded"
					m.configHasChanges = false
					// Reload config items to reflect any changes
					m.configItems = buildConfigItems()
				}
				m.configStatusTick = 30 // Show for ~3 seconds
			}
		}

	case tea.WindowSizeMsg:
		// Window was resized
		m.width = msg.Width
		m.height = msg.Height

		// Calculate content area height
		headerHeight := 6 // Status bar + tabs
		footerHeight := 2 // Footer
		panelPadding := 6 // Border + padding

		contentHeight := m.height - headerHeight - footerHeight - panelPadding
		if contentHeight < 5 {
			contentHeight = 5
		}
		contentWidth := m.width - 8 // Account for panel borders and padding
		if contentWidth < 20 {
			contentWidth = 20
		}

		// Update viewport
		m.logViewport.Width = contentWidth
		m.logViewport.Height = contentHeight
		m.help.Width = m.width

	case tickMsg:
		// Periodic update - refresh data every second
		m.tickCount++
		cmds = append(cmds, tickCmd()) // Schedule next tick
		cmds = append(cmds, fetchStatusCmd())
		cmds = append(cmds, fetchLogsCmd())

		// Decrement config status message timer
		if m.configStatusTick > 0 {
			m.configStatusTick--
			if m.configStatusTick == 0 {
				m.configStatusMsg = ""
			}
		}

	case statusUpdateMsg:
		// Update all status fields from fetched data
		m.serverRunning = msg.running
		m.serverUptime = msg.uptime
		m.serverStartTime = msg.startTime
		m.serverName = msg.serverName
		m.saveName = msg.saveName
		m.worldID = msg.worldID
		m.gamePort = msg.gamePort
		m.updatePort = msg.updatePort
		m.maxPlayers = msg.maxPlayers
		m.connectedPlayers = msg.players

		// Game info
		m.gameVersion = msg.gameVersion
		m.gameBranch = msg.gameBranch
		m.buildID = msg.buildID

		// SSUI info
		m.ssuiVersion = msg.version
		m.goRuntime = msg.goRuntime
		m.isDocker = msg.isDocker

		// Features
		m.discordEnabled = msg.discordEnabled
		m.autoSaveEnabled = msg.autoSaveEnabled
		m.autoRestartEnabled = msg.autoRestartEnabled
		m.autoRestartTimer = msg.autoRestartTimer
		m.upnpEnabled = msg.upnpEnabled
		m.authEnabled = msg.authEnabled
		m.bepInExEnabled = msg.bepInExEnabled
		m.saveInterval = msg.saveInterval
		m.autoStartEnabled = msg.autoStartEnabled

		// Backup info
		m.backupKeepNewestCount = msg.backupKeepNewestCount
		m.backupDailyRetentionDays = msg.backupDailyRetentionDays
		m.backupWeeklyRetentionWeeks = msg.backupWeeklyRetentionWeeks
		m.backupMonthlyRetentionMonths = msg.backupMonthlyRetentionMonths

		m.lastRefresh = msg.startTime

	case logUpdateMsg:
		// Update log viewport content
		if len(msg.logs) > 0 {
			content := strings.Join(msg.logs, "")
			m.logViewport.SetContent(content)
			m.logViewport.GotoBottom()
		}

	case serverActionMsg:
		// Handle server action results
		if msg.err != nil {
			m.err = msg.err
		}
		// Refresh status after action
		cmds = append(cmds, fetchStatusCmd())
	}

	return m, tea.Batch(cmds...)
}

// fetchStatusCmd returns a command that fetches current server status from real sources
func fetchStatusCmd() tea.Cmd {
	return func() tea.Msg {
		// Get real data from config and managers
		running := config.GetIsGameServerRunning()
		uptime := gamemgr.GetServerUptime()
		startTime := gamemgr.GetServerStartTime()

		// Get connected players
		detector := detectionmgr.GetDetector()
		players := detectionmgr.GetPlayers(detector)

		// Parse auto restart timer - check if it's enabled (not empty or "0")
		autoRestartTimer := config.GetAutoRestartServerTimer()
		autoRestartEnabled := autoRestartTimer != "" && autoRestartTimer != "0"

		return statusUpdateMsg{
			// Server state
			running:    running,
			uptime:     uptime,
			startTime:  startTime,
			serverName: config.GetServerName(),
			saveName:   config.GetSaveName(),
			worldID:    config.GetWorldID(),
			gamePort:   config.GetGamePort(),
			updatePort: config.GetUpdatePort(),
			maxPlayers: parseMaxPlayers(config.GetServerMaxPlayers()),
			players:    players,

			// Game info
			gameVersion: config.GetExtractedGameVersion(),
			gameBranch:  config.GetGameBranch(),
			buildID:     config.GetCurrentBranchBuildID(),

			// SSUI info
			version:   config.GetVersion(),
			goRuntime: runtime.GOOS + "/" + runtime.GOARCH,
			isDocker:  config.GetIsDockerContainer(),

			// Features
			discordEnabled:     config.GetIsDiscordEnabled(),
			autoSaveEnabled:    config.GetAutoSave(),
			autoRestartEnabled: autoRestartEnabled,
			autoRestartTimer:   autoRestartTimer,
			upnpEnabled:        config.GetUPNPEnabled(),
			authEnabled:        config.GetAuthEnabled(),
			bepInExEnabled:     config.GetIsSSCMEnabled(), // SSCM is the BepInEx integration
			saveInterval:       config.GetSaveInterval(),
			autoStartEnabled:   config.GetAutoStartServerOnStartup(),

			// Backup
			backupKeepNewestCount:        config.GetBackupKeepNewestCount(),
			backupDailyRetentionDays:     config.GetBackupDailyRetentionDays(),
			backupWeeklyRetentionWeeks:   config.GetBackupWeeklyRetentionWeeks(),
			backupMonthlyRetentionMonths: config.GetBackupMonthlyRetentionMonths(),
		}
	}
}

// fetchLogsCmd returns a command that fetches SSUI log entries
func fetchLogsCmd() tea.Cmd {
	return func() tea.Msg {
		logs := GetLogBuffer()

		if len(logs) == 0 {
			logs = []string{
				"Waiting for log entries...\n",
				"\n",
				"Backend activity will appear here.\n",
			}
		}

		return logUpdateMsg{logs: logs}
	}
}

// startServerCmd starts the game server
func startServerCmd() tea.Cmd {
	return func() tea.Msg {
		err := gamemgr.InternalStartServer()
		return serverActionMsg{action: "start", err: err}
	}
}

// stopServerCmd stops the game server
func stopServerCmd() tea.Cmd {
	return func() tea.Msg {
		err := gamemgr.InternalStopServer()
		return serverActionMsg{action: "stop", err: err}
	}
}
