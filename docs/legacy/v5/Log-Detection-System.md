!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

The Log Detection System monitors and analyzes your server logs in real-time, identifying important events.

## Built-in Detection Patterns

The detector comes with several pre-configured detection patterns that monitor common server events. These patterns are divided into two categories: simple keyword patterns and more complex regex-based patterns.

### Keyword Patterns

These simple patterns trigger when specific keywords appear in the logs:

| Keyword | Event Type | Description |
|---------|------------|-------------|
| `Ready` | `EventServerReady` | Triggered when the server is ready to accept connections |
| `Unloading 1 Unused Serialized files` | `EventServerStarting` | Indicates the server is in the startup process |
| `EXCEPTION` | `EventServerError` | Captures server errors and exceptions |
| `Initialize engine version` | `EventServerRunning` | Shows the server **process** is now running |

### Regex Patterns

More complex patterns that use regular expressions to capture detailed information:

#### Player Events

| Pattern | Event Type | Description |
|---------|------------|-------------|
| `Client\s+(.+)\s+\((\d+)\)\s+is\s+ready!` | `EventPlayerReady` | Detects when a player has fully loaded into the game and is ready to play |
| `Client:?\s+(.+?)\s+\((\d+)\)\.\s+Receiving` | `EventPlayerConnecting` | Identifies a player in the process of connecting to the server |
| `Client\s+disconnected:\s+\d+\s+\|\s+(.+)\s+connectTime:\s+\d+,\d+s,\s+ClientId:\s+(\d+)` | `EventPlayerDisconnect` | Triggered when a player disconnects from the server |

#### Server Events

| Pattern | Event Type | Description |
|---------|------------|-------------|
| `World Saved:\s.*,\sBackupIndex:\s(\d+)` | `EventWorldSaved` | Detects when the game world has been saved, including the backup index |
| `(?m)^\s*>\s*\d{2}:\d{2}:\d{2}:.*Exception.*\|>\s+\d{2}:\d{2}:\d{2}:.*StackTrace` | `EventException` | Captures detailed exception and stack trace information |
| `\d{2}:\d{2}:\d{2}: Changed setting '(.+?)' from '(.+?)' to '(.+?)'` | `EventSettingsChanged` | Tracks changes to server settings, including the setting name and old/new values |
| `RocketNet Succesfully hosted with Address: (.+?) Port: (\d+)` | `EventServerHosted` | Captures server hosting information including IP address and port |
| `Started new game in world (.+)` | `EventNewGameStarted` | Detects when a new game has started in a specific world |

## Custom Detections

Refer to [Custom Detections](Custom-Detections-&-The-Custom-Detections-Manager.md)
#### Features

- **Easy-to-use interface**: Toggle between Keyword and Regex detection modes
- **Real-time feedback**: Instant validation of your detection patterns
- **Centralized management**: View all your custom detections in one place
- **Quick**: Add and Remove unwanted detections with a single click while the server is running

## Technical Implementation

Under the hood, the detection system uses the following core components:

### Event Structure

```go
type Event struct {
    Type          EventType
    Message       string
    RawLog        string
    Timestamp     string
    PlayerInfo    *PlayerInfo     // Optional
    BackupInfo    *BackupInfo     // Optional
    ExceptionInfo *ExceptionInfo  // Optional
}
```

### Custom Pattern Structure

```go
type CustomPattern struct {
	Pattern     *regexp.Regexp
	EventType   EventType
	MessageTmpl string
	IsRegex     bool
	Keyword     string
}
```

### Detection Process Flow

1. Log messages are internally processed
2. Simple keyword patterns are checked first
3. Complex regex patterns are processed next
4. Custom patterns (both keyword and regex) are processed last
5. When matches are found, the appropriate event is triggered with relevant details

## The Custom Detections API

The detection manager is powered by a robust API that allows for programmatic management of your custom detection patterns. This is useful for further automation, backup/restore operations, or integration with other tools.

**For detailed API documentation, refer to the [API Reference](API.md#custom-detections).**

### Combining with Discord Notifications

Detections integrate seamlessly with the Discord system. Detection messages will be forwarded to your configured Discord channel, allowing for real-time monitoring from anywhere.
