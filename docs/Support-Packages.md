# Support packages

A support package collects SSUI logs, system information and a configuration copy with known secret fields removed.

## Collect a package

Use **SSCLI**, the terminal running SSUI. These are not Stationeers game-console commands.

1. Enter `supportmode` (alias `sm`).
2. Reproduce the issue and note the time and first error.
3. Enter `supportpackage` (alias `sp`). Support mode must still be enabled.
4. Enter `supportmode` again to turn it off.
5. Open `support_package_<date>_<time>.zip` in the installation directory and verify its contents before sharing it.

Include the SSUI version/branch, operating system or image, game build, exact action and first relevant error in your report. Use the [issue tracker](https://github.com/SteamServerUI/StationeersServerUI/issues) or the project's [Discord](https://discord.gg/8n3vN92MyJ); share sensitive diagnostics through a private location agreed with support.

!!! warning
    Configuration sanitization is not log sanitization. Player names, Steam IDs, paths, IP addresses and even secrets previously printed by a command can remain in logs. Review the ZIP. “It says sanitized” is not a substitute for opening it.

<details>
<summary>Included files and removed fields</summary>

The generator includes files under `SSUI/logs`, sanitized `SSUI/config/config.json`, OS/version and architecture information, SSUI version/branch and creation time.

It removes `discordToken`, `users`, `JwtKey`, `AdminPassword`, `ServerAuthSecret` and `ServerPassword` from the configuration copy. It does not include an automatic scrub of arbitrary log content.

</details>

## Restore normal logging

Support mode enables debug, log level 10 and SSUI file logging, then reloads backend components. Turning it off sets debug false, log level 20 and file logging false; it does **not** restore your earlier custom values. Reapply those if needed.

If SSUI started with debug enabled, its profiling listener on port `6060` remains a separate concern. Restart without debug to close that listener. [Access and HTTPS](Security-Considerations.md#limit-network-access).

Check that the ZIP actually contains the expected files. The generator has limited error handling; a filename alone is not proof of a complete package.

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
