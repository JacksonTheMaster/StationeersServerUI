# API v3 reference

SSUI v6 exposes one API below `/api/v3`. The old v2 routes no longer exist. This is an intentional breaking change: clients need to move to v3 instead of carrying the old route and response quirks forward.

Default base URL:

```text
https://<ssui-host>:8443/api/v3
```

## Response contract

Normal JSON responses use one envelope:

```json
{
  "data": {}
}
```

Errors use the same shape on every JSON route:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "A useful description"
  }
}
```

`error.details` can contain additional structured information. Successful deletes and logout can return `204 No Content`. Backup downloads return the save file directly and live streams use Server-Sent Events, so neither uses a JSON envelope.

State-changing operations use `POST`, `PUT`, `PATCH` or `DELETE`. Do not call them through link previews or health checks.

## Authentication

Authentication is mandatory in v6. `authEnabled` from an old configuration does not disable it.

### Browser sessions

```http
POST /api/v3/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}
```

Login returns the user, permissions, expiry and CSRF value. It also sets:

- `SSUISession`: secure, HTTP-only, same-site-strict session cookie.
- `SSUICSRF`: secure, same-site-strict CSRF cookie read by the bundled UI.

Browser requests which change state must send the CSRF value as `X-SSUI-CSRF`. Sessions expire after 24 hours without use and have an absolute lifetime of 30 days. Logout revokes the current session:

```http
POST /api/v3/auth/logout
X-SSUI-CSRF: <csrf value>
```

Read the current identity and permissions with `GET /api/v3/auth/session`.

### Personal access tokens

API clients should use a personal access token instead of copying a browser cookie:

```http
Authorization: Bearer <token>
```

Create a token with `POST /api/v3/auth/tokens`. The clear secret is returned once; SSUI stores only its hash. Token scopes must be permissions the owner already has. Tokens can have an optional expiry and can be revoked independently. CSRF does not apply to Bearer authentication.

### First owner

When no v5 user can be migrated, the limited first-owner setup flow is available until an owner exists. Use `/setup`, or call:

```http
POST /api/v3/auth/setup/bootstrap
Content-Type: application/json

{
  "username": "admin",
  "password": "a-long-unique-password"
}
```

`GET /api/v3/auth/setup` reports whether bootstrap is still required. Existing human v5 users are migrated as owners. Old JWT sessions and `apikey-*` entries are deliberately discarded.

## Identity and capabilities

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v3` | Backend version and permissions available to the caller. |
| POST | `/api/v3/auth/password` | Change the signed-in user's own password; requires a browser session. |
| GET | `/api/v3/auth/users` | List users. |
| POST | `/api/v3/auth/users` | Create a user. |
| PATCH | `/api/v3/auth/users/{id}` | Change username, password, enabled state or groups. |
| DELETE | `/api/v3/auth/users/{id}` | Delete a user and revoke its credentials. |
| GET/POST | `/api/v3/auth/groups` | List or create permission groups. |
| PUT/DELETE | `/api/v3/auth/groups/{id}` | Replace or delete a custom group. |
| GET/POST | `/api/v3/auth/tokens` | List or create personal access tokens. |
| DELETE | `/api/v3/auth/tokens/{id}` | Revoke a token. |
| GET | `/api/v3/auth/sessions` | List visible sessions. |
| DELETE | `/api/v3/auth/sessions/{id}` | Revoke a session. |
| GET | `/api/v3/auth/audit` | Read the bounded identity audit history. |

The built-in Owner group cannot be weakened or removed. Permissions are evaluated on every request, so group changes affect existing sessions immediately.

## Server and console

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/v3/server/start` | Start Stationeers. |
| POST | `/api/v3/server/stop` | Stop Stationeers. |
| GET | `/api/v3/server/status` | Read process state, start time and numeric uptime. |
| GET | `/api/v3/server/players` | List connected players as `username` and `steamId` objects. |
| GET | `/api/v3/server/connectivity` | Read the latest Spacecat gameserver reachability result. |
| POST | `/api/v3/server/connectivity/check` | Run a new external UDP reachability check while Stationeers is stopped. |
| POST | `/api/v3/backend/reload` | Reload backend components. |
| POST | `/api/v3/steamcmd/run` | Run the game-server SteamCMD update flow. |
| GET | `/api/v3/sscm/status` | Check whether SSCM is enabled. |
| POST | `/api/v3/sscm/commands` | Send one allowed server-console command. |

Server status contains `running`, `state`, `startedAt`, `uptimeSeconds`, `serverId` and `worldId`. Times use RFC 3339 and durations are numeric so clients do not need to parse UI strings. Player lists use `{ "players": [...] }`. An accepted console command returns its command and the state `queued`; that means SSCM received it, not that Stationeers has already executed it.

The connectivity result contains the check state, configured port, advertised address and the probe results for the tested UDP packet sizes. The POST route requires `server.control`, reserves the gameserver port temporarily and returns a conflict if another check is already running or the server is not stopped.

## Live streams

These authenticated endpoints return Server-Sent Events:

| Path | Stream |
|---|---|
| `/api/v3/streams/console` | Raw Stationeers console output. |
| `/api/v3/streams/events` | Parsed game and custom detection events. |
| `/api/v3/streams/logs/debug` | Debug log. |
| `/api/v3/streams/logs/info` | Informational log. |
| `/api/v3/streams/logs/warn` | Warning log. |
| `/api/v3/streams/logs/error` | Error log. |
| `/api/v3/streams/logs/backend` | Combined backend log. |

Keep the session cookie on browser streams. Browser `EventSource` cannot add a Bearer header; non-browser clients can use a normal streaming HTTP request with a PAT.

## Backups

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v3/backups?limit=N&include=summary&cursor=NAME` | List archives from the RAM inventory. |
| GET | `/api/v3/backups/analysis?name=NAME` | Read or request one archive analysis. |
| POST | `/api/v3/backups/restore` | Stop the game and restore the archive named by JSON `{"name":"NAME"}`. |
| GET | `/api/v3/backups/download?name=NAME` | Stream the selected archive. |

Backup lists return `{ "items": [...] }` and, when a limited page has more results, `nextCursor`. Pass that value back unchanged. List entries use `name`, `saveTime` and optional `summary`. The relative filename is the stable identity. URL-encode it when used as a query value. Absolute paths, traversal and old numeric index selectors are rejected. See [Backups and restores](Backup-System.md).

## Settings and maintenance

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v3/settings` | Read settings exposed by the web interface. |
| PATCH | `/api/v3/settings` | Change one or more exposed settings. |
| GET | `/api/v3/worldgen/catalog` | Read world-generation choices from the game installation. |
| POST | `/api/v3/advertiser/override` | Save the advertised address override. |
| POST | `/api/v3/tls/certificate` | Validate and install certificate material. |
| GET/POST | `/api/v3/update` | Read update state or trigger the update flow. |
| GET/POST/DELETE | `/api/v3/detections` | Read, save or remove custom detections. |
| POST | `/api/v3/setup/finalize` | Finish first-time configuration. |

During a brand-new setup only, `/api/v3/setup/settings` accepts wizard settings before the owner session exists. It disappears once owner bootstrap is complete.

The settings API deliberately does not expose the complete `config.json` file. Its typed contract contains the settings available in the web interface; omitted PATCH fields remain unchanged and unknown fields are rejected. Game-server passwords and secrets are returned to callers with `settings.view`. Internal SSUI credentials and the Discord bot token are not returned. `discordTokenConfigured` reports whether a bot token exists, while sending `discordToken` replaces or clears it.

Most changes apply immediately. SSUI reloads only the affected subsystem instead of restarting the complete backend. A response containing `restartRequired` names settings, currently `ssuiWebPort`, which were saved but need an SSUI restart before they take effect.

## StationeersLaunchPad and mods

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/v3/slp/install` | Install SLP. |
| POST | `/api/v3/slp/uninstall` | Remove SLP. |
| POST | `/api/v3/slp/reinstall` | Repair the SLP installation. |
| POST | `/api/v3/slp/packages` | Upload an exported mod package. |
| GET | `/api/v3/slp/mods` | List installed mods. |
| POST | `/api/v3/slp/mods/update` | Update installed Workshop mods. |
| POST | `/api/v3/slp/mods` | Install/update selected Workshop items. |

`GET /slp/mods` returns `{ "items": [...] }` with lower-camel-case fields. To install or update selected Workshop items, send `{"workshopIds":["123456789"]}`. IDs and full Steam Workshop URLs are accepted. Errors from SteamCMD use `502 Bad Gateway`; any captured tool output is available in `error.details.logs`.

## curl example

Use a certificate or CA that you verified independently:

```bash
curl --cacert ssui-cert.pem \
  --cookie-jar ssui.cookies \
  --header 'Content-Type: application/json' \
  --data @login.json \
  https://server.example:8443/api/v3/auth/login

curl --cacert ssui-cert.pem \
  --cookie ssui.cookies \
  https://server.example:8443/api/v3/backups?limit=20
```

For automation, create a scoped PAT in an authenticated browser session and use `Authorization: Bearer` instead of storing a password or browser cookie.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
