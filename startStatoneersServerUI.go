package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r3labs/sse"
)

var (
	discordToken      = "MTI3NTA1Mjk5Mjg0ODIwMzc3OA.GXBztW.UAa7ijUAsbu5hOtswa6IxXZn_d-QRH_bnpfFBw"
	controlChannelID  = "1275055797616771123"
	statusChannelID   = "1276701394543313038"
	logChannelID      = "1275067875647819830"
	saveChannelID     = "1276705219518140416"
	FILEPATH          = "C:/SteamCMD/Stationeers/Blacklist.txt"
	discordSession    *discordgo.Session
	logMessageBuffer  string
	maxBufferSize     = 1000
	bufferFlushTicker *time.Ticker
	serverCmd         *exec.Cmd
	serverStdin       io.WriteCloser
)

func main() {
	go startDiscordBot()
	go startLogStream()

	processName := "Stationeers-ServerUI.exe"
	workingDir := "./UIMod/"
	exePath := "./" + processName

	for {
		if !isProcessRunning(processName) {
			fmt.Printf("Process %s not running. Starting it...\n", processName)

			// Start the process
			if err := startProcess(exePath, workingDir); err != nil {
				fmt.Printf("Failed to start process: %v\n", err)
			} else {
				fmt.Println("UI and API started successfully at http://localhost:8080.")
			}
		} else {
			//fmt.Printf("Process %s is running.\n", processName)
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}
}

func startDiscordBot() {
	var err error
	discordSession, err = discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	discordSession.AddHandler(messageCreate)

	err = discordSession.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is now running.")
	// Start the buffer flush ticker to send the remaining buffer every 5 seconds
	bufferFlushTicker = time.NewTicker(5 * time.Second)
	sendMessageToStatusChannel("Bot reconnected to Discord. Good Morning, Stationeers!")
	go func() {
		for range bufferFlushTicker.C {
			flushLogBufferToDiscord()
		}
	}()

	select {} // Keep the program running
}

func startLogStream() {
	client := sse.NewClient("http://localhost:8080/output")
	client.Headers["Content-Type"] = "text/event-stream"
	client.Headers["Connection"] = "keep-alive"
	client.Headers["Cache-Control"] = "no-cache"

	for {
		// Attempt to connect to the SSE stream
		fmt.Println("Attempting to connect to SSE stream...")

		err := client.SubscribeRaw(func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				logMessage := string(msg.Data)
				addToLogBuffer(logMessage)
			}
		})

		if err != nil {
			fmt.Printf("Error subscribing to SSE stream: %v\n", err)
			fmt.Println("Reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue // Retry connection
		}

		// If the connection is successful, block until an error occurs
		// The error handling and reconnection logic should be inside the SubscribeRaw callback
		break
	}
}

func addToLogBuffer(logMessage string) {
	logMessageBuffer += logMessage + "\n" // Add the log message to the buffer with a newline
	checkForKeywords(logMessage)
	// If the buffer exceeds the max size, send it to Discord
	if len(logMessageBuffer) >= maxBufferSize {
		flushLogBufferToDiscord()
	}
}

func checkForKeywords(logMessage string) {
	// List of keywords to detect and their corresponding messages
	keywordActions := map[string]string{
		"Ready": "Attention! Server is ready to connect!",
		// "No clients connected.": "No clients connected. Server is going into idle mode!",
	}

	// Iterate through the keywordActions map
	for keyword, actionMessage := range keywordActions {
		if strings.Contains(logMessage, keyword) {
			sendMessageToStatusChannel(actionMessage)
		}
	}

	// Detect more complex patterns using regex
	complexPatterns := []struct {
		pattern *regexp.Regexp
		handler func(matches []string)
	}{
		{
			// Example: "Client Jacksonthemaster (76561198334231312) is ready!"
			pattern: regexp.MustCompile(`Client\s+(.+)\s+\((\d+)\)\s+is\s+ready!`),
			handler: func(matches []string) {
				username := matches[1]
				steamID := matches[2]
				message := fmt.Sprintf("Client %s (Steam ID: %s) is ready!", username, steamID)
				sendMessageToStatusChannel(message)
			},
		},
		{
			// Example: "Client disconnected: 135108291984612402 | Jacksonthemaster"
			pattern: regexp.MustCompile(`Client\s+disconnected:\s+\d+\s+\|\s+(.+)\s+connectTime:\s+\d+,\d+s,\s+ClientId:\s+(\d+)`),
			handler: func(matches []string) {
				username := matches[1]
				message := fmt.Sprintf("Client %s disconnected.", username)
				sendMessageToStatusChannel(message)
			},
		},
		{
			// Enhanced "World Saved" pattern: "World Saved: C:/SteamCMD/Stationeers/saves/EuropaProd, BackupIndex: 1057"
			pattern: regexp.MustCompile(`World Saved:\s.*,\sBackupIndex:\s(\d+)`),
			handler: func(matches []string) {
				backupIndex := matches[1]
				currentTime := time.Now().UTC().Format(time.RFC3339)
				message := fmt.Sprintf("World Saved: BackupIndex: %s UTCTime: %s", backupIndex, currentTime)
				sendMessageToSavesChannel(message)
			},
		},
		// Add more complex patterns and handlers here
	}

	for _, cp := range complexPatterns {
		if matches := cp.pattern.FindStringSubmatch(logMessage); matches != nil {
			cp.handler(matches)
		}
	}
}

func sendMessageToControlChannel(message string) {
	if discordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := discordSession.ChannelMessageSend(controlChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to control channel:", err)
	} else {
		fmt.Println("Sent message to control channel:", message)
	}
}

func sendMessageToStatusChannel(message string) {
	if discordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := discordSession.ChannelMessageSend(statusChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to status channel:", err)
	} else {
		fmt.Println("Sent message to status channel:", message)
	}
}

func sendMessageToSavesChannel(message string) {
	if discordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := discordSession.ChannelMessageSend(saveChannelID, message)
	if err != nil {
		fmt.Println("Error sending message to saves channel:", err)
	} else {
		fmt.Println("Sent message to saves channel:", message)
	}
}

func flushLogBufferToDiscord() {
	if len(logMessageBuffer) == 0 {
		return // No messages to send
	}

	if discordSession == nil {
		fmt.Println("Discord session is not initialized")
		return
	}

	_, err := discordSession.ChannelMessageSend(logChannelID, logMessageBuffer)
	if err != nil {
		fmt.Println("Error sending log to Discord:", err)
	} else {
		//fmt.Println("Flushed log buffer to Discord.")
		logMessageBuffer = "" // Clear the buffer after sending
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.ChannelID != controlChannelID {
		return
	}

	content := strings.TrimSpace(m.Content)

	switch {
	case strings.HasPrefix(content, "!start"):
		sendCommandToAPI("/start")
		s.ChannelMessageSend(m.ChannelID, "Server is starting...")
		sendMessageToStatusChannel("Start command received from @Server Controller, Server is trying to start...")

	case strings.HasPrefix(content, "!stop"):
		sendCommandToAPI("/stop")
		s.ChannelMessageSend(m.ChannelID, "Server is stopping...")
		sendMessageToStatusChannel("Stop command received from @Server Controller, Server is stopping...")

	case strings.HasPrefix(content, "!restore"):
		handleRestoreCommand(s, m, content)

	case strings.HasPrefix(content, "!list"):
		handleListCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!update"):
		handleUpdateCommand(s, m.ChannelID)

	case strings.HasPrefix(content, "!serverrun"):
		handleServerRunCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!help"):
		handleHelpCommand(s, m.ChannelID)

	case strings.HasPrefix(content, "!ban"):
		handleBanCommand(s, m.ChannelID, content)

	case strings.HasPrefix(content, "!unban"):
		handleUnbanCommand(s, m.ChannelID, content)
	default:
		// Optionally handle unrecognized commands or ignore them
	}
}

func handleHelpCommand(s *discordgo.Session, channelID string) {
	helpMessage := `
**Available Commands:**
- ` + "`!start`" + `: Starts the server.
- ` + "`!stop`" + `: Stops the server.
- ` + "`!restore:<index>`" + `: Restores a backup at the specified index. Usage: ` + "`!restore:1`" + `.
- ` + "`!list:<number/all>`" + `: Lists the most recent backups. Use ` + "`!list:all`" + ` to list all backups or ` + "`!list:<number>`" + ` to specify how many to list.
- ` + "`!ban:<SteamID>`" + `: Bans a player by their SteamID. Usage: ` + "`!ban:76561198334231312`" + `.
- ` + "`!unban:<SteamID>`" + `: Unbans a player by their SteamID. Usage: ` + "`!unban:76561198334231312`" + `.
- ` + "`!update`" + `: Updates the server files if there is a game update available. (Currently Stable Branch only)
- ` + "`!help`" + `: Displays this help message.

Please stop the server before using restore or update commands.
	`

	_, err := s.ChannelMessageSend(channelID, helpMessage)
	if err != nil {
		fmt.Println("Error sending help message:", err)
	} else {
		fmt.Println("Help message sent to control channel.")
	}
}

func handleListCommand(s *discordgo.Session, channelID string, content string) {
	fmt.Println("!list command received, fetching backup list...")

	// Extract the "top" number or "all" option from the command
	parts := strings.Split(content, ":")
	top := 5 // Default to 5
	var err error
	if len(parts) == 2 {
		if parts[1] == "all" {
			top = -1 // No limit
		} else {
			top, err = strconv.Atoi(parts[1])
			if err != nil || top < 1 {
				s.ChannelMessageSend(channelID, "Invalid number provided. Use `!list:<number>` or `!list:all`.")
				return
			}
		}
	}

	// Step 1: Fetch the backup list from the server
	resp, err := http.Get("http://localhost:8080/backups")
	if err != nil {
		fmt.Println("Failed to fetch backup list:", err)
		s.ChannelMessageSend(channelID, "Failed to fetch backup list.")
		return
	}
	defer resp.Body.Close()

	// Step 2: Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read backup list response:", err)
		s.ChannelMessageSend(channelID, "Failed to read backup list.")
		return
	}

	// Step 3: Output the raw backup list data for debugging
	//fmt.Println("Raw backup list data:", string(body))

	// Step 4: Parse the backup list data into a formatted string
	backupList := parseBackupList(string(body))
	//fmt.Println("Formatted backup list:\n", backupList)

	// Step 5: Split the backup list into individual lines
	lines := strings.Split(backupList, "\n")

	// Step 6: Send each line as a separate message, respecting the "top" limit
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		if top > 0 && count >= top {
			break // Stop if we've reached the "top" limit
		}
		fmt.Println("Sending line to Discord:", line)
		message, err := s.ChannelMessageSend(channelID, line)
		if err != nil {
			fmt.Println("Error sending line to Discord:", err)
		} else {
			fmt.Println("Successfully sent line to Discord. Message ID:", message.ID)
		}
		count++

		// Optional: Add a small delay to avoid hitting rate limits
		time.Sleep(500 * time.Millisecond)
	}
}

func handleUpdateCommand(s *discordgo.Session, channelID string) {
	// Notify that the update process is starting
	s.ChannelMessageSend(channelID, "Starting the server update process...")

	// PowerShell command to run SteamCMD
	powerShellScript := `
		cd C:\SteamCMD
		.\steamcmd +force_install_dir C:/SteamCMD/Stationeers/ +login anonymous +app_update 600760 -beta public validate +quit
	`

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-Command", powerShellScript)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting update command: %v\n", err)
		s.ChannelMessageSend(channelID, "Failed to start the update process.")
		return
	}

	// Wait for the process to complete
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error during update process: %v\n", err)
		s.ChannelMessageSend(channelID, "The update process encountered an error.")
	} else {
		// Notify that the update process has finished
		s.ChannelMessageSend(channelID, "Game Update process completed successfully. Server is up to date.")
	}
}

func handleRestoreCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(m.ChannelID, "Invalid restore command. Use `!restore:<index>`.")
		return
	}

	indexStr := parts[1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid index provided for restore.")
		return
	}

	url := fmt.Sprintf("http://localhost:8080/restore?index=%d", index)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to restore backup at index %d.", index))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Backup %d restored successfully.", index))
}

func handleBanCommand(s *discordgo.Session, channelID string, content string) {
	// Extract the SteamID from the command
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(channelID, "Invalid ban command. Use `!ban:<SteamID>`.")
		return
	}
	steamID := strings.TrimSpace(parts[1])

	// Read the current blacklist
	blacklist, err := readBlacklist(FILEPATH)
	if err != nil {
		s.ChannelMessageSend(channelID, "Error reading blacklist file.")
		return
	}

	// Check if the SteamID is already in the blacklist
	if strings.Contains(blacklist, steamID) {
		s.ChannelMessageSend(channelID, fmt.Sprintf("SteamID %s is already banned.", steamID))
		return
	}

	// Add the SteamID to the blacklist
	blacklist = appendToBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	err = os.WriteFile(FILEPATH, []byte(blacklist), 0644)
	if err != nil {
		s.ChannelMessageSend(channelID, "Error writing to blacklist file.")
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("SteamID %s has been banned.", steamID))
}

func handleUnbanCommand(s *discordgo.Session, channelID string, content string) {
	// Extract the SteamID from the command
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(channelID, "Invalid unban command. Use `!unban:<SteamID>`.")
		return
	}
	steamID := strings.TrimSpace(parts[1])

	// Read the current blacklist
	blacklist, err := readBlacklist(FILEPATH)
	if err != nil {
		s.ChannelMessageSend(channelID, "Error reading blacklist file.")
		return
	}

	// Check if the SteamID is in the blacklist
	if !strings.Contains(blacklist, steamID) {
		s.ChannelMessageSend(channelID, fmt.Sprintf("SteamID %s is not banned.", steamID))
		return
	}

	// Remove the SteamID from the blacklist
	updatedBlacklist := removeFromBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	err = os.WriteFile(FILEPATH, []byte(updatedBlacklist), 0644)
	if err != nil {
		s.ChannelMessageSend(channelID, "Error writing to blacklist file.")
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("SteamID %s has been unbanned.", steamID))
}

func readBlacklist(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func appendToBlacklist(blacklist, steamID string) string {
	if len(blacklist) > 0 && !strings.HasSuffix(blacklist, ",") {
		blacklist += ","
	}
	return blacklist + steamID
}

func removeFromBlacklist(blacklist, steamID string) string {
	entries := strings.Split(blacklist, ",")
	var updatedEntries []string
	for _, entry := range entries {
		if strings.TrimSpace(entry) != steamID {
			updatedEntries = append(updatedEntries, entry)
		}
	}
	return strings.Join(updatedEntries, ",")
}

func parseBackupList(rawData string) string {
	lines := strings.Split(rawData, "\n")
	var formattedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) == 2 {
			formattedLines = append(formattedLines, fmt.Sprintf("**%s** - %s", parts[0], parts[1]))
		}
	}
	return strings.Join(formattedLines, "\n")
}

func sendCommandToAPI(endpoint string) {
	url := "http://localhost:8080" + endpoint
	if _, err := http.Get(url); err != nil {
		fmt.Printf("Failed to send %s command: %v\n", endpoint, err)
	}
}

func isProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/fi", fmt.Sprintf(`imagename eq %s`, processName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error checking process: %v\n", err)
		return false
	}
	return strings.Contains(string(out), processName)
}

func handleServerRunCommand(s *discordgo.Session, channelID string, content string) {
	// Extract the command string from the message
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(channelID, "Invalid server run command. Use `!serverrun:<command>`.")
		return
	}
	command := strings.TrimSpace(parts[1])

	// Send the extracted command to the server
	err := sendCommandToServer(command)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send command to server: %v", err))
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("Command sent to server: %s", command))
}

func sendCommandToServer(command string) error {
	if serverStdin == nil {
		sendMessageToControlChannel("server stdin is not initialized")
		return fmt.Errorf("server stdin is not initialized")
	}

	_, err := serverStdin.Write([]byte(command + "\n"))
	if err != nil {
		sendMessageToControlChannel(fmt.Sprintf("failed to send command to server: %v", err))
		return fmt.Errorf("failed to send command to server: %v", err)
	}

	return nil
}

func startProcess(exePath, workingDir string) error {
	cmd := exec.Command(exePath)
	cmd.Dir = workingDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %v", err)
	}

	serverStdin = stdin
	serverCmd = cmd

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting Server UI: %v", err)
	}

	fmt.Println("Ensure Windows Firewall allows incoming connections on game port and update port (27015 and 27016 by default).")
	return nil
}
