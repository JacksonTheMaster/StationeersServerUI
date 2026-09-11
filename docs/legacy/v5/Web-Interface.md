!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

StationeersServerUI provides a user-friendly web interface for managing your server. Here's how to use it effectively.

## Accessing the Interface

- Open your web browser
- Navigate to `https://<IP-OF-YOUR-SERVER>:8443`
  - Replace `<IP-OF-YOUR-SERVER>` with the local IP address of your server

## Main Controls

The main page provides these core functions:

- **Start Server**: Launches the Stationeers server
- **Stop Server**: Gracefully shuts down the server
- **Console Output**: View real-time server output directly in the browser

## Managing Configurations

See [Configuration](Configuration.md)

## Backup Management

The **Backups** section provides tools for managing your server backups:

- **List Backups**: View all available backups for the specified save
- **Restore Backup**: Select and restore a specific backup
- **Auto-Cleanup**: System automatically deletes backups older than 2 days _(more functionality here was planned, but halted. This should be optional or atleast properly documented)_


## Next Steps

- [First-Time Setup](First-Time-Setup.md) - Configure your server for first use
- [Discord Integration](Discord-Integration.md) - Set up Discord integration for remote management
- [Security Considerations](Security-Considerations.md) - Important security notes
