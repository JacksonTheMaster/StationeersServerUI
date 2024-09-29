package discord

import (
	"StationeersServerUI/src/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

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
- ` + "`!validate`" + `: Validates the server files if there is a game update available. (Currently Stable Branch only)
- ` + "`!help`" + `: Displays this help message.

Please stop the server before using update commands.
	`

	_, err := s.ChannelMessageSend(channelID, helpMessage)
	if err != nil {
		fmt.Println("Error sending help message:", err)
		SendMessageToControlChannel("Error sending help message.")
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
				s.ChannelMessageSend(channelID, "❌Invalid number provided. Use `!list:<number>` or `!list:all`.")
				return
			}
		}
	}

	// Step 1: Fetch the backup list from the server
	resp, err := http.Get("http://localhost:8080/backups")
	if err != nil {
		fmt.Println("Failed to fetch backup list:", err)
		s.ChannelMessageSend(channelID, "❌Failed to fetch backup list.")
		return
	}
	defer resp.Body.Close()

	// Step 2: Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read backup list response:", err)
		s.ChannelMessageSend(channelID, "❌Failed to read backup list.")
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

func handleRestoreCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(m.ChannelID, "❌Invalid restore command. Use `!restore:<index>`.")
		sendMessageToStatusChannel("⚠️Restore command received, but not able to restore Server.")
		return
	}
	SendCommandToAPI("/stop")
	indexStr := parts[1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌Invalid index provided for restore.")
		sendMessageToStatusChannel("⚠️Restore command received, but not able to restore Server.")
		return
	}

	url := fmt.Sprintf("http://localhost:8080/restore?index=%d", index)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌Failed to restore backup at index %d.", index))
		sendMessageToStatusChannel("⚠️Restore command received, but not able to restore Server.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅Backup %d restored successfully, Starting Server...", index))
	//sleep 5 sec to give the server time to start
	time.Sleep(5 * time.Second)
	SendCommandToAPI("/start")
}

func handleUpdateCommand(s *discordgo.Session, channelID string) {
	// Notify that the update process is starting
	s.ChannelMessageSend(channelID, "🕛Starting the server update process...")

	// PowerShell command to run SteamCMD
	powerShellScript := `
		cd C:\SteamCMD
		.\steamcmd +force_install_dir C:/SteamCMD/Stationeers/ +login anonymous +app_update 600760 -beta public +quit
	`

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-Command", powerShellScript)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting update command: %v\n", err)
		s.ChannelMessageSend(channelID, "❌Failed to start the update process.")
		return
	}

	// Wait for the process to complete
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error during update process: %v\n", err)
		s.ChannelMessageSend(channelID, "❌The update process encountered an error.")
	} else {
		// Notify that the update process has finished
		s.ChannelMessageSend(channelID, "✅Game Update process completed successfully. Server is up to date.")
	}
}

func handleValidateCommand(s *discordgo.Session, channelID string) {
	// Notify that the update process is starting
	s.ChannelMessageSend(channelID, "🕛Starting the server validate process...")

	// PowerShell command to run SteamCMD
	powerShellScript := `
		cd C:\SteamCMD
		.\steamcmd +force_install_dir C:/SteamCMD/Stationeers/ +login anonymous +app_update 600760 -beta public -validate +quit
	`

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-Command", powerShellScript)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting update command: %v\n", err)
		s.ChannelMessageSend(channelID, "❌Failed to start the validate process.")
		return
	}

	// Wait for the process to complete
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error during update process: %v\n", err)
		s.ChannelMessageSend(channelID, "❌The validate process encountered an error.")
	} else {
		// Notify that the update process has finished
		s.ChannelMessageSend(channelID, "✅Game validate process completed successfully. Server is valid, but custom changes are overwritten.")
	}
}

func handleBanCommand(s *discordgo.Session, channelID string, content string) {
	// Extract the SteamID from the command
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(channelID, "❌Invalid ban command. Use `!ban:<SteamID>`.")
		return
	}
	steamID := strings.TrimSpace(parts[1])

	// Read the current blacklist
	blacklist, err := readBlacklist(config.BlackListFilePath)
	if err != nil {
		s.ChannelMessageSend(channelID, "❌Error reading blacklist file.")
		return
	}

	// Check if the SteamID is already in the blacklist
	if strings.Contains(blacklist, steamID) {
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌SteamID %s is already banned.", steamID))
		return
	}

	// Add the SteamID to the blacklist
	blacklist = appendToBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	err = os.WriteFile(config.BlackListFilePath, []byte(blacklist), 0644)
	if err != nil {
		s.ChannelMessageSend(channelID, "❌Error writing to blacklist file.")
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("✅SteamID %s has been banned.", steamID))
}

func handleUnbanCommand(s *discordgo.Session, channelID string, content string) {
	// Extract the SteamID from the command
	parts := strings.Split(content, ":")
	if len(parts) != 2 {
		s.ChannelMessageSend(channelID, "❌Invalid unban command. Use `!unban:<SteamID>`.")
		return
	}
	steamID := strings.TrimSpace(parts[1])

	// Read the current blacklist
	blacklist, err := readBlacklist(config.BlackListFilePath)
	if err != nil {
		s.ChannelMessageSend(channelID, "❌Error reading blacklist file.")
		return
	}

	// Check if the SteamID is in the blacklist
	if !strings.Contains(blacklist, steamID) {
		s.ChannelMessageSend(channelID, fmt.Sprintf("✅SteamID %s is not banned.", steamID))
		return
	}

	// Remove the SteamID from the blacklist
	updatedBlacklist := removeFromBlacklist(blacklist, steamID)

	// Write the updated blacklist back to the file
	err = os.WriteFile(config.BlackListFilePath, []byte(updatedBlacklist), 0644)
	if err != nil {
		s.ChannelMessageSend(channelID, "❌Error writing to blacklist file.")
		return
	}

	s.ChannelMessageSend(channelID, fmt.Sprintf("✅SteamID %s has been unbanned.", steamID))
}
