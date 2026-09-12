# Logs and detections

**Console** shows Stationeers output. **Events** turns recognized lines into saves, connections and other readable activity. **Logs** shows SSUI's own subsystem messages. Start with the stream for the component that failed.

Built-in detections also drive connected-player and server-state information. If Stationeers changes its log format, that interpretation can lag behind the actual game. Check the raw console before assuming an empty player list is accurate.

## Add a custom event

Open **Detection Manager**, or `/detectionmanager`. Custom detections publish events; they do not run game or operating-system commands.

1. Copy a real console line that identifies the condition.
2. Choose `keyword` for a case-sensitive literal match, or `regex` to capture variable values.
3. Enter the pattern and message, then add it.
4. Observe or trigger the condition and check Events. If Discord event output is configured, check that channel too.

| Type | Pattern | Message |
|---|---|---|
| `keyword` | `Unsupported shader` | `Unity reported an unsupported shader.` |
| `regex` | `Player (.+) has reached level (\d+)` | `Player {1} reached level {2}` |

The regex example is illustrative log text, not a claim that Stationeers has that event. For `Player Alex has reached level 12`, it produces `Player Alex reached level 12`.

Use Go-compatible regex syntax. Numbered capture groups become `{1}`, `{2}` and so on. Test both matching and unrelated lines. `.*` will match impressively often and tell you almost nothing.

Definitions are stored at `SSUI/config/customdetections.json`. Include it when moving your configuration. Add/remove changes apply while SSUI is running.

## Diagnose a detection

| Symptom | Check |
|---|---|
| Nothing fires | Exact case, spaces and punctuation; confirm the line exists in Console |
| Regex rejected | Regex syntax; API JSON needs an additional level of backslash escaping |
| Missing capture text | Pattern defines and matches the requested group |
| Too many events | Narrow the pattern to a distinctive fragment |
| UI event but no Discord event | Discord enabled, event channel ID and bot permissions |

For a broken built-in detection, report the raw line, SSUI version and game build. A screenshot of the last error in a cascade is less useful than the first error in text. [Collect diagnostics](Support-Packages.md).

## API and implementation

`/api/v3/streams/events` is the authenticated SSE stream. Custom definitions use GET/POST/DELETE `/api/v3/detections`. Request details are in [API](API.md#settings-and-maintenance).

The implementation lives in `src/managers/detectionmgr`: keyword and regex handlers parse console text, publish structured events and update relevant runtime state. Raw console output remains available independently of whether a pattern matches.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
