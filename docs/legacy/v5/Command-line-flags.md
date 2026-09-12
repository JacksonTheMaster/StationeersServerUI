!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

The SSUI executable can be started with several command line flags:

| Flag                  | Alias    | Type    | Description                                                                                                                             |
|-----------------------|----------|---------|-----------------------------------------------------------------------------------------------------------------------------------------|
| AdvertiserOverride    |          | string  | Override the advertised server IP. For this, the ServerVisible setting must be set to false. Use "auto" for automatic public IP detection, an IPv4 address, or a DNS hostname (to allow server advertisement if you are behind a reverse proxy) |
| BackendEndpointPort   | -p       | string  | Override the backend endpoint port (e.g., 8080)                                                                                          |
| GameBranch            | -b       | string  | Override the game branch (e.g., beta)                                                                                                    |
| LogLevel              | -ll      | int     | Override the log level (e.g., 10)                                                                                                        |
| LogToFiles            | -lf      |         | Create log files for SSUI                                                                                                               |
| NoSanityCheck         |          |         | Skips the sanity check. Not recommended.                                                                                                |
| NoSteamCMD            |          |         | Skips SteamCMD installation                                                                                                             |
| RecoveryPassword      | -r       | string  | Adds a 'recovery' user (expects password as argument)                                                                                   |                                                           |
