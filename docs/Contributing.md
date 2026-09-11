# Contributing

Report bugs, improve the docs or send a focused code change through the [SSUI repository](https://github.com/SteamServerUI/StationeersServerUI). For a large change, describe the problem in an issue or [Discord](https://discord.gg/8n3vN92MyJ) first so work does not overlap.

Read the project's [license](https://github.com/SteamServerUI/StationeersServerUI/blob/main/LICENSE) before publishing forks or redistributing builds.

## Send a useful report

Include SSUI version/branch, platform, game build, reproduction steps and the first relevant error. For documentation, link the page and say which step or claim failed. [Support packages](Support-Packages.md) help with runtime failures; review them before posting.

## Prepare a change

Use `nightly` as the development baseline unless the issue targets another branch. Keep the change focused and describe the resulting behavior in the pull request. Do not change version numbers or release machinery incidentally.

Run tests for the affected behavior and the relevant build from the [developer guide](Developer-Documentation.md#building). Use disposable directories for game, backup and update testing. Your actual world is not a test fixture with unusually good atmosphere simulation.

For documentation changes, verify claims against handlers and callers, not just comments. Check navigation and anchors in rendered GitHub Markdown. Keep operational warnings close to the action they affect.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
