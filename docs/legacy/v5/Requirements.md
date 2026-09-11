!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

Before installing StationeersServerUI, ensure your system meets the following requirements:

## System Requirements

- At **least** 4 processor cores (to run the Stationeers server)
  - SSUI runs on almost anything:
    - around `15mb RAM`
    - `Intel Core Potato Gen1`

- **Operating System**:
  - Windows
  - Ubuntu 22.04 or higher
  - Debian TRIXIE (13) or higher
 
- **Installation Location**:
  - An empty folder of your choice to install the software
  - At **least** 10gb Disk space

## Dependencies
SSUI **itself** should be able to run **without** installing dependencies on any OS. 

**The Stationeers Server on Windows** requires [The Visual C++ Redistributable](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist?view=msvc-170) on Windows wich is -depending on your Windows version- automatically installed by the Stationeers Server. 
`Just install it manually to be sure.` It can be Downloaded from the above link.

1. **If you run into really weird issues, ask on the SSUI Discord

## Resource Considerations

Depending on the size of your base, the complexity of systems, code, and player connection handling, the Stationeers server can require significant resources:

- Large setups may need 8+ CPU cores and 30GB+ of RAM for optimal performance
- Complex logic networks, high player counts, and sprawling bases with atmosphere calculations are particularly resource-intensive

## Network Requirements

If you plan to host the server for others to connect:

- Ability to configure port forwarding on your router (for the game server and optionally for the web UI)
- Stable internet connection with sufficient upload bandwidth


## Next Steps

- [Installation](Installation.md) - Learn how to install the server control
- [Docker Guide](Docker-Guide.md) - For containerized deployment options
