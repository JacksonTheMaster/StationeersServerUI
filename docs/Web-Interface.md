# Web interface

Open `https://<server-address>:8443` and sign in. The dashboard is the place for daily operation; configuration and backups each have their own workspace. Controls you cannot use are disabled or replaced with a permission notice rather than failing mysteriously.

The established interface lives at `/`. An experimental Svelte interface also exists at `/app`, but it does not yet contain every control. Use the main interface for administering a real server.

## Dashboard

The top status strip answers the questions that matter before touching anything: is Stationeers running, how long has it been running, who is connected and when was the newest backup?

The main actions start and stop the game, open configuration and run a game update. Below them:

- **Game Log** shows raw Stationeers output.
- **Events** shows useful activity parsed from that output.
- **Backend Log** shows what SSUI itself is doing.
- **Recent Backups** gives quick access to the newest archives.
- **Connected Players** shows the current player list when log detection has established it.

Starting a process, loading a world and becoming joinable are different stages. If the status sits in the middle, read the Game Log before clicking Start again. Two servers do not make the first one load faster.

The command input sends commands to **Stationeers through SSCM**. Commands such as `supportmode` and `reloadbackend` belong in [SSCLI](SSUICLI-Commands.md), the terminal where SSUI runs. They are two different consoles attached to two different programs because one console would apparently have been too peaceful.

## Backups

Open **Backups** or go directly to `/backups`. Newest archives appear first. A row can show the save's timestamp, Stationeers version, days played and analysis statistics.

Those values describe that archived save, not the live world. Download keeps the original archive; Restore replaces the active world after confirmation. Read [backups and restores](Backup-System.md) before using Restore for the first time.

If you do not have `backups.view`, the page shows no archive data and explains that access is missing. Viewing, analyzing, downloading and restoring are separate permissions.

## Configuration

Select **Edit Config** or open `/config`. Settings are grouped by purpose rather than by their order in `config.json`:

- game, save and world generation;
- network and visibility;
- autosaves, automatic restarts and updates;
- backup retention;
- logging and diagnostics;
- Discord and modding; and
- **People & Access** for accounts, groups and API tokens.

**Show Expert Settings** reveals less common controls in a separate tab. Hiding that tab later does not erase its values.

Users with `settings.view` can read configuration. Saving requires `settings.manage`; fields are not leaked to users who have neither permission. A successful change or refusal appears through the same notification system used across the dashboard, configuration, backups and setup screens.

Game launch settings take effect after restarting Stationeers. Listener settings such as the SSUI HTTPS port require an SSUI restart. The [configuration reference](Configuration.md) lists exact keys, defaults and environment variables when the labels alone are not enough.

## People & Access

Inside configuration, this tab lets every user change their own password. Administrators with the appropriate permissions can create people, assign access-group presets, fine-tune permissions and disable or remove accounts. Users allowed to manage tokens can create scoped API keys which never exceed their own access.

See [manage people and access](Security-Considerations.md#manage-people-and-access) before building a custom group. The owner preset is excellent at owning things and therefore a poor default for everybody you have ever met.

## Detection Manager

Open `/detectionmanager` to turn a distinctive game-log line into a named event using a keyword or regular expression. This page requires `detections.manage`. Custom detections publish events; they do not execute shell or game commands. [Logs and detections](Log-Detection-System.md) includes examples.

## Appearance and background activity

Themes and background modes are stored by the browser, so two administrators can choose different appearances without fighting over server configuration. Animated modes pause when the page is not in focus, and reduced-motion/static options are available for browsers without suitable graphics acceleration or for anyone who would simply like the planet to hold still.

Changing pages within the interface keeps the active background state instead of restarting the orbit. Login and setup use a lightweight static background so public assets and startup time stay limited.

## When a page stops updating

Live console, event and log panels use persistent browser connections. If they freeze:

1. Check whether SSUI is still running.
2. Refresh and sign in again if the session expired.
3. Test direct HTTPS access if a reverse proxy is involved.
4. Check proxy buffering and long-lived timeouts for the API stream routes.

For a page that never opens, use [Web UI troubleshooting](Troubleshooting.md#web-ui-does-not-open).

---

[Documentation home](index.md) · [Server operations](Server-Operations.md) · [Access and HTTPS](Security-Considerations.md)
