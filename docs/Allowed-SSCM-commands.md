# Game commands and SSCM

SSCM is the Stationeers Server Command Manager plugin. It carries commands from SSUI into the running game and is enabled by default. Scheduled restart warnings and the explicit restart save use it too.

## Send a command

1. Start the game with SSCM enabled and check that the plugin loads.
2. Use the game-command input in the web UI, or `/command command:<text>` from a Discord admin account in the hub.
3. Check the game console or resulting event for execution.

For example, SSUI itself uses `save` and `announce <message>`. Discord also provides `/announce message:<text>`.

!!! important
    Successful submission means SSUI wrote the command, not that Stationeers executed it. For a save, wait for the world-saved event before stopping the game.

The installed game defines command syntax. Use its help output for the exact build. SSUI's Go transport has no comprehensive command allowlist or game-syntax validator. The old page title “Allowed SSCM commands” should not be read as a security guarantee.

Discord limits console input to one line and 1,000 characters. The HTTP endpoint accepts POST JSON:

```json
{"command":"save"}
```

Send it to `/api/v3/sscm/commands` with an authenticated session or Bearer token. [API authentication](API.md#authentication).

## Delivery and failures

SSUI writes prefixed text to `BepInEx/plugins/SSCM/SSCM.socket`. This is a file-based transport, despite the filename. The plugin reads it inside Stationeers. Writes are serialized, but there is no end-to-end command acknowledgement or durable queue.

If a command does nothing, check `IsSSCMEnabled`, plugin loading, file permissions and the original game syntax. `/api/v3/sscm/status` reports the enabled setting; it does not prove the plugin consumed a command.

Only trusted admins should have this input. A text box being willing to accept a command is not an endorsement of what it does to your world.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
