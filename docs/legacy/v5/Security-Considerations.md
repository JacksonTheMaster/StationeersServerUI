!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# Security Considerations

Securing your StationeersServerUI installation is crucial, especially if exposing it beyond your local network.

## Authentication Overview

Authentication is fully handled at the application level by the backend (Go server). It uses:

- **Password Storage**: Passwords are hashed using bcrypt (strong, salted hashing).
- **Session Management**: JSON Web Tokens (JWT) signed with HMAC-SHA256 using a server-side secret key.
- **Token Delivery**: JWT is stored in an HttpOnly, Secure, SameSite=Strict cookie named `AuthToken`.
- **API Keys**: Support for long-lived API keys with configurable expiration (default 1 month).

## Key security features:
- No plaintext passwords are ever stored or transmitted.
- A random JWT secret key is automatically generated at first startup. Could be explicitly changed to something else in the config.json.
 - The JWT key is stored in plaintext in the config.json
- All routes besides the login page are guarded by middleware that validates the JWT cookie.
- First-time setup encourages enabling authentication and creating users.

## Multiple users:
- Multiple users can be added throuh the /changeuser page
 - Each user has the same permission level - User A can change User B's password. Be aware of that. For community access, consider using the Discord integration instead.

Authentication can be completely disabled, but this is **strongly discouraged** for any public facing deployment.

## Setting Up Credentials

During first-time setup (or manually):

1. **Via Web UI Setup** (recommended for initial deployment):
   - Access `/setup` on fresh install.
   - Register one or more users (passwords automatically hashed).
   - Finalize setup to enable authentication.

2. **Manually via config.json**:
   - In the `UIMod` folder, edit `config.json`.
   - Add users to the `"users"` map with bcrypt-hashed passwords.
   - Example:
     ```json
     "users": {
       "admin": "$2a$10$examplehashedpasswordhere..."
     }


### Network Security

1. **Firewall Configuration**
   - Only open the necessary ports (27015, 27016 for the game server)
   - Keep the web UI port (8443) restricted to best practices
   - Consider using Windows Firewall or iptables (Linux) to restrict access

2. **Reverse Proxy Setup**
   - If you need remote access to the web UI, set up a reverse proxy with:
     - ([Traefik](https://traefik.io/) - might be worth checking out!)
     - Rate limiting to prevent brute force attacks

### Application Security

2. **Discord Integration**
   - Keep your Discord bot token secure
   - Use Discord's role-based permissions to restrict command access
   - Only give administrative command access to trusted users


### Docker Security

If using Docker:

1. **Container Isolation**
   - Don't run containers with `--privileged` flag
   - Use volume mounts instead of bind mounts where possible

3. **Network Configuration**
   - Use Docker's network controls to limit container access

## Next Steps

- [Configuration](Configuration.md) - Review proper Configuration procedures
- [Docker Guide](Docker-Guide.md) - Secure containerized deployment
