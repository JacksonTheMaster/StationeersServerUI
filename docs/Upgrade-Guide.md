# Upgrade from SSUI v5

SSUI v6 keeps the familiar WebUI and the Stationeers focus, but it changes some of the backend contracts underneath. Most ordinary installations can upgrade normally. The important part is knowing whether you are an ordinary installation before you click through.

!!! note "The short answer"
    If you use the WebUI, a native installation and normal local storage, v6 is intended to be a straightforward upgrade. Take a copy of the complete installation first, then follow the normal update path. Pay extra attention if you use the API directly, old numeric backup indexes, split Docker mounts or custom storage.

## What changes in v6

| Area | v5 behavior | v6 behavior |
|---|---|---|
| Managed runtime data | `UIMod` | `SSUI` |
| HTTP API | `/api/v2` routes and route-specific responses | `/api/v3` with a shared response contract |
| Backup selection | Numeric indexes may be used by old clients | Backup filenames are the stable identity |
| Authentication | Legacy sessions and API keys | Users, permission groups, sessions and scoped personal tokens |
| Docker layout | Several child mounts were common in older examples | One writable `./server-data:/app:rw` bind mount is the supported v6 layout |

The old v5 pages remain available under [Legacy](legacy/v5/Home.md). They are useful for identifying an old setting, but they are not a v6 operating manual.

## Before you upgrade

1. Stop Stationeers cleanly and confirm it is stopped.
2. Make a copy of the complete SSUI installation and the active save.
3. If you have API clients, list every route and backup selector they use.
4. If you run Docker, read the [one-mount layout](Docker-Guide.md#1-create-the-deployment) before replacing the Compose file.
5. Keep the old `UIMod` folder and old safe-backup directories until the new installation has been verified.

Do not delete `UIMod` as part of the upgrade. v6 copies managed data into `SSUI` and deliberately leaves the original available for rollback or inspection. It also leaves existing legacy safe-backup folders where they are.

For a cautious native upgrade, you can smoke-test the old path before deleting anything: after v6 has started successfully and you have checked the WebUI, stop SSUI and rename the old folder to something like `UIMod.v5-backup`. Start v6 again and exercise the normal server, backup and update paths. If everything is fine, remove the renamed folder later. Renaming is optional for ordinary installations, and it must happen after the first v6 migration, not before it. Do not use this as a workaround for split Docker mounts.

## Native Windows or Linux

Run the v6 executable from the existing installation directory. On the first start, SSUI migrates managed data when required and uses the new `SSUI` path afterwards. Existing saves are not moved into a different game-managed tree by this migration.

Then verify the boring things first:

- the browser opens and you can sign in;
- the server name, world and network settings are present;
- Stationeers starts and players can connect;
- a new autosave is handled by Backup Manager; and
- the new archive can be inspected and downloaded.

The [installation guide](Installation.md) covers the platform-specific executable and permissions. The [first-time setup guide](First-Time-Setup.md) is for a new owner or a new installation, not a reason to wipe a working v5 data directory.

## API and automation clients

This is the deliberate breaking part of the upgrade:

- move clients from `/api/v2` to `/api/v3`;
- replace numeric backup indexes with the archive `name` value;
- handle the shared `data` and `error` envelopes;
- use the v6 authentication flow and scoped personal tokens; and
- use non-GET methods for state-changing operations.

The complete route list and examples are in the [API reference](API.md). Do not keep a compatibility shim in a client and assume it is selecting the same backup. v6 rejects old numeric selectors instead of guessing.

## Docker installations

Existing split mounts cannot be migrated by SSUI because the container cannot inspect host paths that are not mounted into it. Copy the complete old installation into one host directory and mount that directory as `/app`:

```yaml
volumes:
  - ./server-data:/app:rw
```

Keep the old container and folders until the new layout has started and the saves, configuration, TLS files and backups have been checked. Docker updates come from pulling a new image; the in-container updater only reports that an image update exists.

See the [Docker guide](Docker-Guide.md) for the complete Compose example and migration steps.

## After the upgrade

Once the server has run normally for a while, remove only the old files you have independently decided to retire. Keep an off-host copy of important saves and backups. A successful application update is not a backup strategy; it is merely an update that happened to finish.

If something does not match the guide, stop before deleting data and use [Troubleshooting](Troubleshooting.md) or the [SSUI Discord](https://discord.gg/8n3vN92MyJ). The old folder is much more useful as evidence than it is as an improvised target for cleanup.

---

[Documentation home](index.md) · [Installation](Installation.md) · [API reference](API.md) · [Docker](Docker-Guide.md)
