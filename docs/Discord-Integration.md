# Discord

SSUI maintains a shared **server hub** with status, players and optional votes. Administrators open private controls from the same message. You do not need a separate command channel or the old reaction panel.

## Connect the bot

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications), then obtain its bot token from the Bot page.
2. Invite it to your Discord server with the `bot` and `applications.commands` scopes.
3. Create a hub channel. Give the bot View Channel, Send Messages, Embed Links and Read Message History there; Attach Files is needed for downloads. Members need access to use application commands.
4. Create a trusted administrator role. With Discord Developer Mode enabled, copy its **role ID** and the hub's **channel ID**.
5. Open SSUI's Discord configuration and enter the values below, enable Discord, then save and reload the backend. Check the Discord subsystem log if the hub does not appear.

| Configuration field | Enter |
|---|---|
| `discordToken` | Bot token |
| `discordAdminRoleID` | Trusted admin role ID |
| `statusPanelChannelID` | Hub channel ID |
| `eventLogChannelID` | Optional private channel for parsed game events |
| `logChannelID` | Optional private channel for verbose logs |
| `isDiscordEnabled` | `true` |

Channel names are not IDs. Yes, `server-hub` looks perfectly clear to a human. Discord would like its number anyway.

!!! caution
    The bot token is a credential. Keep it out of screenshots, issues and repositories. Reset an exposed token in Discord and replace it in SSUI.

## Verify access

With a normal member account, run `/status` in the hub and check that admin actions are denied. With an account holding the configured role, open **Server Admin Actions** and confirm that the private menu appears.

!!! important
    No configured admin role means no admin access. Discord's Administrator permission does not bypass this role check. All slash commands except `/status` require the configured role and run in the configured hub, including `/help` and backup downloads.

Role checks are repeated on confirmations and submissions. If a role is revoked while a menu is open, the old menu does not preserve access. Votes and rotating join codes use channel membership instead; only expose them to the intended community.

## Operate the server

Click **Server Admin Actions** for Start, Stop, Restart, Update, backups, announcements, commands and bans. Stop, Restart, Update and Restore require confirmation. Save first when current progress matters; ordinary Stop does not guarantee a final save. [Shutdown behavior](Server-Operations.md#save-and-stop).

Update stops the game, runs SteamCMD and leaves it stopped. Restore stops it, restores the selected filename and starts it after success. Failed steps prevent later steps. Admin menus and download replies are visible only to the requester; recipients can still share downloaded files.

### Slash commands

| Command | Purpose |
|---|---|
| `/start` | Start the Stationeers server |
| `/stop` | Confirm and stop the Stationeers server |
| `/restart` | Confirm and restart the Stationeers server |
| `/status` | Show current hub status; no admin role required |
| `/update` | Confirm, stop, update through SteamCMD and leave stopped |
| `/command command:<text>` | Send a server-console command through SSCM |
| `/announce message:<text>` | Broadcast an announcement to in-game players |
| `/list [limit]` | Browse backups; default is all, ten per page |
| `/restore name:<filename>` | Restore the selected backup |
| `/download [name]` | Upload the selected backup to Discord; defaults to the newest |
| `/bansteamid steamid:<id>` | Add a Steam ID to the ban list; restart required by the game |
| `/unbansteamid steamid:<id>` | Remove a Steam ID from the ban list; restart required by the game |
| `/help` | Open the private admin action menu |

All slash commands run in the configured hub. Every command except `/status` requires the configured admin role, including `/help`, `/list` and `/download`. Slash commands and hub buttons use the same action functions.


## Read the hub

The hub refreshes every 15 seconds and on relevant events. It shows game state, players, version, uptime, next restart and available statistics from the latest archived save. Those statistics are not live simulation measurements.

The bot reuses its own hub message and does not clear unrelated channel history. On a fresh process it searches the newest 100 messages for its panel, so a quiet dedicated channel helps it find the old one.

For parsed connections, saves and detections, configure an event channel. Verbose logs belong in a separate private channel. Mute it unless your preferred notification sound is Discord, continuously.

## Enable community votes

Restart and restore voting are off by default. Enable the desired option in configuration, then open **Vote Menu** in the hub. A restart vote can start a stopped game; a restore vote selects from the three newest safe backups and restarts after a successful restore.

!!! warning
    Discord voters are not matched to in-game players. Channel access determines participation. A restore vote replaces the live save; review membership, thresholds and backup quality before enabling it.

<details>
<summary>Vote targets, cooldowns and temporary panels</summary>


Each new vote creates its own public vote panel in the hub channel, after the main hub message. It shows votes received, votes required, remaining votes, the closing time and, for restores, the pinned backup filename. Members vote through its button; acknowledgement of an individual vote stays private.

The panel is edited in place as votes arrive. Once voting ends, its button is disabled and the same message shows whether the vote passed, expired or was cancelled. A passed vote then updates that message with execution success or failure. There are no separate public "vote passed" or execution-error messages in the hub; detailed events still go to the configured event log.

Results are removed 15 minutes after voting ends. Execution updates do not restart that timer. If Discord rejects deletion, SSUI logs the failure and retries twice at one-minute intervals. A backend reload cancels open votes and closes their panels; already accepted server actions finish normally. Old vote buttons cannot join or start another vote. Vote state and deletion timers are held in RAM, so a full SSUI process exit does not preserve automatic cleanup; any leftover panel is inert and may need manual removal.

Voting is optional and disabled by default. When enabled, members who can see and interact with the status panel can open its **Vote Menu** and vote for:

- **Restart:** restart a running server, or start it if it is stopped.
- **Restore:** choose one of the three newest safe backups, stop the server if necessary, restore that exact pinned save, and start the server again after the vote passes.

The target is calculated when a vote starts: the configured percentage of connected players, rounded up, with the configured minimum as a floor. It stays fixed for that vote. Discord voters are not matched to in-game player identities: access to the status-panel channel is what determines who can participate. Configure that channel's Discord roles accordingly.

Only one vote of each kind can be active. A passed restart or restore vote cancels the other active vote, and passed votes enter their configured cooldown, even if execution subsequently fails. Expired votes also enter cooldown; a competing vote cancelled by a passed vote gets its own cooldown. If another Discord action is busy when the deciding vote arrives, that vote is not added and no cooldown is consumed; retry after the action finishes. The defaults are:

| Setting | Restart | Restore |
|---|---:|---:|
| Enabled | No | No |
| Vote duration | 5 minutes | 5 minutes |
| Required percentage | 60% | 100% |
| Minimum votes | 1 | 2 |
| Cooldown after pass or expiry | 30 minutes | 60 minutes |

The corresponding configuration fields are `discordRestartVoteEnabled`, `discordRestoreVoteEnabled`, `discordVoteDurationMinutes`, `discordRestartVoteThreshold`, `discordRestartVoteMinimum`, `discordRestartVoteCooldownMinutes`, `discordRestoreVoteThreshold`, `discordRestoreVoteMinimum`, and `discordRestoreVoteCooldownMinutes`.



</details>

## Rotate the game join code

Enable `rotateServerPassword` with Discord and the hub configured. SSUI generates a six-digit game password each time Stationeers starts; members retrieve it through the hub. Anyone with access to that control can obtain the current code. A copied code stops working after the next rotation, not when someone leaves your Discord.

## Troubleshoot

| Symptom | Check |
|---|---|
| Bot offline | Enabled setting, current token, outbound Discord access and backend log |
| No hub or updates | Numeric channel ID and View/Send/Embed/Read History permissions |
| Commands unavailable | Invite scopes, application-command access, registration errors |
| Admin action denied | Correct hub, configured role ID and current member role |
| Vote menu missing | Voting enabled and hub successfully refreshed |
| Download too large | Use web UI/filesystem above the 10 MiB Discord limit |
| Old button rejected | Open a fresh hub menu or run `/list` again |

Commands and announcements require SSCM. A write accepted by SSUI is not an execution acknowledgement from Stationeers. [Game commands](Allowed-SSCM-commands.md).

<details>
<summary>Upgrading from old Discord panels</summary>

The old `controlChannelID` and `controlPanelChannelID` settings are no longer used. An existing `statusPanelChannelID` becomes the hub. If you configured only the old admin channels, select a hub explicitly. Old reaction panels are inert and can be removed manually.

Legacy status/connection/save channel migrations are listed in [Configuration](Configuration.md#discord-settings).

</details>

<details>
<summary>Private dialogs, operation locking and download limits</summary>


Click **Server Admin Actions**, then choose Start, Stop, Restart, Update, Manage backups, Announcement, Console command, Ban or Unban. Text actions open an input form. Stop, Restart, Update and Restore require confirmation. Start and Download do not.

The backup browser shows ten files per page, with save time and available cached metadata. Select a backup for details, Download or Restore. It stores the actual filenames behind short, single-use dialog tokens, so long nested filenames work without being placed in Discord custom IDs. A selected file never changes when the inventory is reordered.

Admin interactions and downloads stay ephemeral (only visible to the requester). Menus update in place where possible. Actions log their kind and initiating Discord user ID to the event channel, without copying console text into that event. In-game announcements are intentionally visible to players.

Long-running work is acknowledged before disk or game operations. Discord's interaction response token lasts 15 minutes; an action can finish after that window, but its private response may no longer be editable. The hub retains a generic latest-action result and the event log records completion/failure. An expired interaction never triggers a public download fallback.

Discord admin operations and passed votes share an execution slot. Conflicting requests are rejected, not queued. The slot survives a bot reload while accepted work completes. This is a Discord-side guard; it does not serialize multi-step Discord actions against Web UI, CLI or automatic game-manager operations.

Update stops the server, runs SteamCMD, and leaves it stopped. Restore checks availability before stopping, restores the selected filename, then starts the server. Stop or restore failure prevents the subsequent steps. There is no extra fixed five-second delay after a confirmed shutdown.

### Backup selection and downloads

The status panel receives the latest available backup summary from BMv4's RAM inventory and analysis notifications. This is separate from how commands select a file.

Backup commands use `/restore name:<filename>` and `/download [name]`. Omit the download name to capture the newest backup once. Lists, download buttons and restore-vote menus carry the actual filename from display through execution; list reordering cannot select a different save. Missing names fail rather than falling back.

Old public download buttons and numeric selections are rejected. Open the hub menu or run `/list` again. The admin browser supports long nested filenames through user-bound tokens. The existing community restore-vote menu still embeds filenames directly and omits newly offered names that exceed its component limit.

Restore checks filename and availability before stopping the server, then checks again under the operation lock. An external deletion or storage failure after preflight can still cause a failure after the stop. A storage failure after stopping leaves the server stopped and reports the failure privately.

`/download` uploads into the private interaction response. SSUI keeps its conservative 10 MiB cap and rejects oversized files before reading their contents. At most two Discord downloads are prepared at once. For larger saves, use the Web UI/API or retrieve the archive directly.

Backup files may contain server/world information. Private Discord visibility is not encryption or DRM; recipients can still save or share downloaded files.


</details>

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
