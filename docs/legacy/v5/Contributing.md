!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

!!! caution
    Section 2 of the License clearly states that forks are allowed only for non-commercial use and primarily for contributing back to the main repository via Pull Requests or Issues (Section 2.3). If your fork shows no activity or is not indicating an intent to contribute, the Licensor reserves the right to file a DCMA removal request and further action if applicable.

Thanks for your interest in contributing! StationeersServerUI is a passion project that’s grown from a niche tool into something with big potential. Whether you’re a coder, designer, or just have ideas, there’s a place for you.

## Project Background & Current Status

StationeersServerUI began as my seccond go project and has grown into something with significant potential. However, the reality is that maintaining and expanding this project alone has become increasingly challenging while I write new code and features. Just recently in V4, I ripped out probably 60% of the Code and rewrote it.

If you're interested in joining a project with real-world utility, visible impact, and exciting technical challenges, this is an excellent opportunity to make a difference. Even a few hours of your time could significantly move this project forward.

## Official Repository

This project's home is at https://github.com/SteamServerUI/StationeersServerUI.

While forks are welcome for testing, please contribute back to the main repository.


## Branching Strategy

We use the following branch structure to maintain code quality:

- **`main`**: Stable, production-ready code. Merges here are polished and tested.
- **`nightly`**: The bleeding-edge development branch - mostly stable, but still a work in progress.
- **Feature branches**: Spin off from `nightly` for specific feature work. Name them like `add-cool-thing`, e.g., `add-linux-vine`. Keep them short-lived.
- **Fix branches**: Spin off from the branch with the issue for your work. Name it like `fix-important-thing`, e.g., `fix-steamcmd-installer`. Keep them short-lived and possibly tied to an issue or PR.

## Versioning

We follow [semantic versioning](https://semver.org/) in spirit - `MAJOR.MINOR.PATCH` - but the process is automated via `build.go`:

- **`MAJOR`**: Bumped _manually_ for big, incompatible changes or major milestones.
- **`MINOR`**: Incremented _manually_ for backwards-compatible features.
- **`PATCH`**: _Auto-incremented on each build_ by `build.go` for bug fixes and small tweaks.

The script updates `src/config/config.go`, builds the executable, and cleans up old builds. No need to mess with versions yourself - focus on the code, and the build handles the rest!

## How You Can Help

1. **Report Issues**
   - Open an issue on GitHub for bugs or feature requests
   - Provide detailed information to help reproduce the issue

2. **Submit Pull Requests**
   - Fork the repository only if it is **really** necesary.
   - Create a new branch from `nightly` for your changes
   - Make your changes and test them thoroughly
   - Submit a pull request with a clear description of your changes

3. **Improve Documentation**
   - Help improve this wiki
   - Create tutorials or guides for specific use cases
   - Report unclear or outdated documentation

4. **Testing**
   - Test new features and report any issues
   - Verify bug fixes work as expected
   - Test on different platforms and configurations

## Code Style Guidelines

- Follow standard Go coding conventions
- Write clear, concise comments
- NEEDED: write tests!

## Next Steps

- [Acknowledgments](Acknowledgments.md) - View project contributors and acknowledgments
