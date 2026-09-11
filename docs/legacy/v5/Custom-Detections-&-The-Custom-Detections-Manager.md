!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

_The Custom Detection System is now fully stable and integrated, allowing users to define and manage custom detection patterns for log processing._

The Custom Detections allow you to create even more powerful monitoring patterns that trigger notifications when specific events occur in your Stationeers server logs. This feature helps server administrators identify issues, track important events, and automate responses to common scenarios.

## The Custom Detections Manager

The Custom Detection Manager on `/detectionmanager` provides an intuitive interface for creating, viewing, and managing your detection patterns. 

### Features

- **Easy-to-use interface**: Toggle between Keyword and Regex detection modes
- **Real-time feedback**: Instant validation of your detection patterns
- **Centralized management**: View all your custom detections in one place
- **Quick**: Add and Remove unwanted detections with a single click while the server is running
- **User-Defined Detections:** Create custom patterns using keywords or regex.
- **UI Management:** Manage detections through a dedicated tab in the web interface.
- **REST API:** Programmatically interact with detections via RESTful endpoints.
- **Thread Safety:** Ensures safe concurrent access to detection configurations.
- **Real-Time Events:** Detection events are broadcasted via Server-Sent Events (SSE).

### Detection Types

The manager supports two detection methods:

#### 1. Keyword Detection
Simple string matching for exact text matches (case-sensitive). This is perfect for straightforward pattern matching when you know precisely what text to look for.

**Example:**
- **Pattern:** `Unsupported shader`
- **Message:** `Unity detected an unsupported shader. This may cause unexpected behavior.`

This detection will trigger whenever the exact phrase "Unsupported shader" appears in your server logs.

#### 2. Regex Detection
Advanced pattern matching using regular expressions. This provides powerful flexibility for detecting complex patterns and extracting variables.

**Fictional Elevator Example:**
- **Pattern:** `Player (.+) has reached level (\d+)`
- **Message:** `Player {1} has reached level {2}`

This detection will trigger when text matching the pattern appears, and the captured groups (player name and level) will be inserted into your message.

### Creating Effective Detections

1. **Select the appropriate detection type** based on your needs:
   - Use **Keyword** for exact text matching
   - Use **Regex** for complex patterns or when you need to extract variables

2. **Craft your pattern carefully**:
   - For keywords, use the exact text as it appears in logs
   - For regex, test your patterns with a tool like [Regex101](https://regex101.com/)
   - Consider case sensitivity and special characters

3. **Create meaningful messages**:
   - Include relevant context in your notification messages
   - For regex detections, use `{1}`, `{2}`, etc. to reference captured groups
   - Keep messages clear and actionable

## The Custom Detections API

The detection manager is powered by a robust API that allows for programmatic management of your detection patterns. This is useful for automation, backup/restore operations, or integration with other tools.

**For detailed API documentation, refer to the [API Reference](API.md#custom-detections).**

## Best Practices

1. **Start simple and refine**: Begin with basic patterns and refine them as needed
2. **Test thoroughly**: Verify your patterns match what you expect before deploying
3. **Be specific**: Overly general patterns may trigger too frequently
4. **Document your detections**: Keep track of what each detection is monitoring and why
5. **Review regularly**: Periodically audit your detection list to remove outdated patterns

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Pattern not triggering | Verify the exact format in your logs matches your pattern |
| Too many notifications | Make your pattern more specific or consider filtering certain events |
| Invalid regex errors | Use a regex testing tool to validate your pattern syntax |
| Message variables not appearing | Ensure your capture groups are correctly referenced in the message |

### Combining with Discord Notifications

Custom detections integrate seamlessly with the Discord notification system. When enabled, your detection messages will be forwarded to your configured Discord channel, allowing for real-time monitoring from anywhere.
