!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# API Documentation

This documentation provides a comprehensive overview of the Game Server API endpoints for server management, authentication, configuration, and monitoring. 

## Table of Contents
- [Authentication](#authentication)
- [Server Management](#server-management)
- [Configuration](#configuration)
- [Backups](#backups)
- [Custom Detections](#custom-detections)
- [Real-time Events (SSE)](#real-time-events-sse)
- [Postman Collections](#postman)

_Some endpoints use GET even though other types might be more adequate. This will be addressed._

## Authentication

Authentication is required for most API endpoints. The API uses JWT tokens stored in HTTP cookies for authentication.

### Authentication Methods

There are two types of authentication tokens:

1. **User Tokens**: Short-lived tokens (24 hours) obtained via login
2. **API Keys**: Long-lived tokens (1+ months) for programmatic access

### Token Lifetime

- **User login tokens**: 24 hours (1440 minutes)
- **API keys**: 1 month (default) or custom duration up to several months

### Cookie-Based Authentication

All authentication tokens are returned as HTTP-only cookies named `AuthToken`. The cookie must be included in subsequent requests. **There are no alternative authentication methods** - cookies are required.

Example cookie format:
```
AuthToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjcxNzYzMTQsImlkIjoiYWRtaW4iLCJpc3MiOiJTdGF0aW9uZWVyc1NlcnZlclVJIn0.kgA150QH__8lgoZgoUSLkfqG_JbahIDQbybtYRkYYTc
```

!!! important
    If auth is enabled, an access token cookie is required for most API endpoints. Only the login endpoint and API key setup endpoints are accessible without authentication. If auth is disabled via config.json, the middleware skips the cookie check and the entire API becomes public.

### Login Endpoint

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/login` | POST | Authenticate and obtain a 24-hour access token cookie |

#### Example Request
```json
{
  "username": "admin",
  "password": "password"
}
```

#### Response
- **Success (200)**: Sets `AuthToken` cookie, valid for 24 hours
- **Failure (401)**: Invalid credentials

### API Key Endpoints

API keys are designed for long-term programmatic access and use a special username prefix `apikey-`.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v2/auth/setup/apikey` | GET | Generate an API key with 1 month expiration |
| `/api/v2/auth/setup/apikey` | POST | Generate an API key with custom duration |

#### GET Request
Simply call the endpoint to receive an API key valid for 1 month:

```bash
GET /api/v2/auth/setup/apikey
```

#### POST Request
Specify a custom duration in months:

```json
{
  "durationMonths": 6
}
```

#### Response
Both endpoints return the JWT token as an `AuthToken` cookie that can be used for the specified duration.

### Initial Setup

To obtain your first API token when authentication is enabled:

1. Temporarily disable auth in `config.json`
2. Call `/api/v2/auth/setup/apikey` (GET or POST)
3. Save the returned cookie/token
4. Re-enable auth in `config.json` and restart SSUI or run `rlb` in the SSUICLI
5. Use the saved token for all subsequent requests

Alternatively, you can log in using the WebUI, and visit /api/v2/auth/setup/apikey to get a 1 month API token, then use that to POST to the apikey endpoint from curl or postman to obtain a longer lasting one.

!!! note
    ApiKeys are not fully fleshed-out, but exist and are usable. As you see, obtaining isn't easy per-se, sorry bout that.

### User Management

Users are stored with bcrypt-hashed passwords. User credentials are managed through the configuration system. Passwords are hashed using bcrypt with default cost factor (10).

## Server Management

Control the game server status with these endpoints.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v2/server/start` | GET | Start the game server |
| `/api/v2/server/stop` | GET | Stop the game server |
| `/api/v2/server/status` | GET | Gets the server status in a structured Json format |
| `/start` | GET | (Legacy) Start the game server |
| `/stop` | GET | (Legacy) Stop the game server |
| `/` | GET | Access the server's HTML interface |

## Configuration

Manage server configuration settings through this API.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/setup` | GET | Access the setup wizard HTML interface |
| `/saveconfigasjson` | POST | (Legacy) Update and save server configuration (form data) |
| `/api/v2/saveconfig` | POST | Update and save server configuration (JSON) |
| `none` | none | There is NO endpoint to GET the current config |

!!! important
    The `/saveconfigasjson` endpoint only allows config values present on the config page. Use the `/api/v2/saveconfig` endpoint for full configuration management.

### Configuration Parameters

See [Configuration](Configuration.md)

## Backups

Manage server backups with these endpoints.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v2/backups` | GET | Get a list of all available backups |
| `/api/v2/backups/restore` | GET | Restore a specific backup |
| `/backups` | GET | (Legacy) Get a list of all available backups |
| `/backups/restore` | GET | (Legacy) Restore a specific backup |

### Restore Backup URL Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `index` | Index of the backup to restore | `7` |

## Custom Detections

Create and manage custom event detections for server monitoring.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v2/custom-detections` | GET | Get all custom detections |
| `/api/v2/custom-detections` | POST | Create a new custom detection |
| `/api/v2/custom-detections/delete/` | DELETE | Delete a custom detection |

### Types of Custom Detections

#### Keyword Detection
Triggers when an exact string is found in logs or events.

```json
{
  "type": "keyword",
  "pattern": "UnloadTime",
  "eventType": "CUSTOM_DETECTION",
  "message": "UnloadTime found"
}
```

#### Regex Detection
Uses regular expressions to match patterns and capture groups.

```json
{
  "type": "regex",
  "pattern": "Player (.+) has reached level (\\d+)",
  "eventType": "CUSTOM_DETECTION",
  "message": "Player {1} reached level {2}!"
}
```

**Do NOT use another event type!**

### Delete Custom Detection Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `id` | UUID of the detection to delete | `bd923b8c-f773-42ca-baee-e35084e3b989` |

## Real-time Events (SSE)

Monitor server events in real-time using Server-Sent Events.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/console` | GET | Stream server console output |
| `/events` | GET | Stream server events |

## Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 303 | See Other (after successful configuration) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 405 | Method Not Allowed |
| 500 | Internal Server Error |

## Postman

Here you can download Postman collections for all API Endpoints.

### V5

_(Collections to be added)_

### V4

[1- Login](https://github.com/user-attachments/files/19542755/1-.Login.postman_collection.json)

[2- Authenticated UI](https://github.com/user-attachments/files/19542756/2-.Authenticated.UI.postman_collection.json)

[3- Start & Stop the Server](https://github.com/user-attachments/files/19542757/3-.Start.Stop.the.Server.postman_collection.json)

[4- Configuration API](https://github.com/user-attachments/files/19542758/4-.Configuration.API.postman_collection.json)

[5- Backups](https://github.com/user-attachments/files/19542759/5-.Backups.postman_collection.json)

[6- Custom Detections API](https://github.com/user-attachments/files/19542753/6-.Custom.Detections.API.postman_collection.json)

[7- SSE](https://github.com/user-attachments/files/19542754/7-.SSE.postman_collection.json)

## Additional Notes

- All authenticated endpoints require an `AuthToken` cookie obtained through the login or API key endpoints
- Cookies are the only supported authentication method - Bearer tokens and other methods are not supported
- The server can be managed both through the API and the web interface
- Custom detections support both simple keyword matching and complex regex patterns
- Real-time monitoring is available through SSE endpoints
- JWT tokens are signed using HS256 algorithm with a secret key configured in the server settings
- All passwords are hashed using bcrypt before storage
