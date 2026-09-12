---
title: Early SSUI releases: v1 to v4
description: A short historical guide to the v1, v2, v3 and v4 Stationeers Server UI branches.
---

# Early releases: v1 to v4

!!! danger "Historical code only"
    The branches on this page are preserved for history and code archaeology. They are not supported installation targets. For a current server, use the [v6 documentation](../index.md) and the [current downloads](../Downloads.md).

    Need a human? The [SSUI Discord](https://discord.gg/8n3vN92MyJ) is the right place to ask. Please include the version or branch you are actually running.

This is the short version of SSUI's early history. The old branches do not contain a proper versioned documentation set, so these notes are reconstructed from their source, branch README files and commit history. They are meant to explain what a branch was, not to turn it into a supported product again.

## At a glance

| Branch | Version visible in the source/history | What it represents |
|---|---|---|
| [v1](https://github.com/SteamServerUI/StationeersServerUI/tree/v1) | v1.1-era | The first Windows-focused web UI and backup controls. |
| [v2](https://github.com/SteamServerUI/StationeersServerUI/tree/v2) | v2.4.3 | A more complete setup flow, JSON configuration and early Discord integration. |
| [v3](https://github.com/SteamServerUI/StationeersServerUI/tree/v3) | v3.0.1 | The move toward Linux, Docker, generalized detections and SSE. |
| [v4](https://github.com/SteamServerUI/StationeersServerUI/tree/v4) | v4.6.10 | The larger pre-v5 configuration, setup and updater foundation. |

The branch names are the useful part here. They are not a promise that the branch tip is a clean release or that an old build can still download everything it once expected.

## v1: the first usable web UI

The v1 branch is the original small Stationeers management tool. Its README calls it **Stationeers Dedicated Server Control v1.1**, and its implementation is focused on Windows: the server process is started through PowerShell and the UI is served by a small Go application.

The early UI already covered the useful basics:

- start and stop the gameserver;
- view server output;
- edit the game configuration;
- list and restore backups; and
- call the same operations through a small REST API.

The backup API used numeric indexes, and the source looked for backups below the game's own `saves/<world>/backup` path. That is important historical context only. Current SSUI identifies archives by filename and stores its managed data below `SSUI/`.

Linux and Docker were not part of this branch. Authentication and the current permission model did not exist yet either.

## v2: setup and Discord arrive

The v2 line keeps the original shape but grows around it. The source contains a first-time setup path, SteamCMD installation helpers, a JSON configuration file and the first complete Discord integration work.

The old v2 documentation describes Discord status notifications, server control, backup actions, player notifications and role-based channel access. It also still describes the earlier numeric backup API and the separate `furtherconfig` flow. Those details belong to the v2 code, not to the current API.

The branch tip represents the v2.4.3 line. It is useful when following the history of the installer and Discord integration, but it is not a sensible base for a new server.

## v3: the cross-platform step

The v3 line is where the project starts looking like the multi-platform tool people recognize today. The branch contains Linux and Docker support, cross-platform SteamCMD helpers, a more developed UI and the generalized detection work.

Its history also introduces server events over SSE and a detection-events view. The old release notes describe Linux support as experimental at that point, and the Stationeers Linux server itself had platform-specific rough edges. That was a snapshot of the time, not a statement about current SSUI support.

The branch's old API and backup handling still follow the earlier conventions. Do not combine its `UIMod`, config or Docker files with a current installation.

## v4: the pre-v5 foundation

The v4 branch reaches the `4.6.10` line and is the last of these early branch snapshots before the later v5 work. It contains the larger JSON configuration model, environment-variable fallbacks, integrated setup and updater helpers, and several rounds of server lifecycle and Windows error-handling fixes.

It still uses the old `UIMod` layout, old API conventions and the older safe-backup paths. The configuration and updater code are useful for understanding how SSUI evolved, but they are not migration instructions for v6.

## v5: the long stable line

v5 was not just another short-lived development branch. It was the stable SSUI line for a long time and the version most existing installations came from before v6. The [v5 documentation](v5/Home.md) is therefore kept as a real reference section, not just a footnote in the history.

The v5 line carried the established Windows, Linux and Docker deployments, the browser UI, Discord integration, SteamCMD setup, log detections and the older backup system. It also accumulated the configuration and operational behavior that many existing servers still rely on.

The [v5 installation guide](v5/Installation.md), [configuration reference](v5/Configuration.md), [backup documentation](v5/Backup-System.md), [Discord guide](v5/Discord-Integration.md) and [old API reference](v5/API.md) are available from the v5 section in the sidebar. They describe v5 as it was; they are not v6 instructions with a different badge.

If you are moving from v5, use the current [upgrade guide](../Upgrade-Guide.md) first. Authentication, API routes, backup identity and storage paths, runtime data and parts of the Discord integration changed in v6. Do not copy a v5 configuration or API example into a new v6 installation by accident.

## What to use today

If you arrived here because you want to host a Stationeers server, you probably want the [current installation guide](../Installation.md), not an old branch. SSUI is intentionally simple on the normal path: download one executable, put it in a writable folder, run it and follow the browser setup.

Use the [v5 legacy documentation](v5/Home.md) only when you are maintaining an existing v5 installation or checking how an old setting worked. For everything current, start with [What is SSUI?](../Features.md), [Getting started](../Getting-Started.md) or [Downloads](../Downloads.md).

If you are studying the old code itself, open the branch links above and treat each branch as an archived snapshot. Keep it isolated from a current installation and expect old release services, paths and API assumptions to have changed.
