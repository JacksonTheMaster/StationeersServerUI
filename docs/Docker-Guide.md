---
description: Run a Stationeers dedicated server with SSUI in Docker using one bind-mounted application directory.
---

# Docker

The official image runs SSUI and Stationeers in a Linux amd64 container, with the entire runtime installation persisted at `/app`. The image already includes SteamCMD's Linux runtime dependency.

Docker is a good choice if you already know how docker in general, published ports and image updates work. 
It is NOT a required compatibility layer and it does NOT automatically make networking or backups easier. 
If `docker compose` is new vocabulary for you, the native [Windows or Linux installation](Installation.md) will usually be the shorter first evening.

If you are, however, absolutely ready to slap SSUI on your 256GB RAM container cluster, go right ahead:

!!! important
    Do not actually deploy SSUI or the Stationeers server in general on a cluster / in a High Availability setup.
    SSUI _might in theory_ work given the config is atomic, but the gameserver certainly won't like it.

## 1. Create the deployment

Install Docker Engine with Compose on your host, if not already installed.
I personally recommend their convenience scripts, but you do you.

Create an empty deployment directory somewhere and save this as `compose.yml`:

```yaml
services:
  ssui:
    image: ghcr.io/steamserverui/stationeersserverui:latest
    container_name: stationeers-server-ui
    ports:
      - "8443:8443/tcp"
      - "27016:27016/udp"
    volumes:
      - ./server-data:/app:rw
    stop_grace_period: 30s
    restart: unless-stopped
```

Create `server-data` next to `compose.yaml` before starting the container:

```bash
mkdir -p server-data
```

Keep `/app` as one bind mount and do not add random mounts for `UIMod`, `saves`, `SSUI`, configuration or backups unless you absolutely need them. The directory contains the complete installation: game files, saves, SSUI configuration, TLS certs, backup metadata. Splitting child directories into additional mounts makes migrations and permissions harder to reason about.

!!! important
    TCP `8443` is for the SSUI web interface. UDP `27016` is for players.

`ghcr.io/steamserverui/stationeersserverui:latest` is the recommended tag and follows stable SSUI releases. Release candidates receive their own tags and never replace `latest`. If you require a deployment to stay on one exact build, use a version tag such as `6.0.0` or an image digest from the [container registry](https://github.com/SteamServerUI/StationeersServerUI/pkgs/container/stationeersserverui). Images are published for `linux/amd64`.

You can also append `nightly` like so: `ghcr.io/steamserverui/stationeersserverui-nightly:latest`. This gives you the most recent development build. Use it when you are deliberately testing current changes and are prepared to report regressions.

## 2. Start and configure

From that deployment directory:

```bash
docker compose up -d
docker compose logs --follow ssui
```

Wait for SteamCMD installation and web-server startup. The web server will only start once SteamCMD has installed the game files. This may take a while. Open `https://<docker-host-LAN-IP-or-whatever-you-have>:8443`, (or whatever port you set), then follow [first-time setup](First-Time-Setup.md) normally.

The default compose above has **no** interactive terminal. To run Stationeers automatically after container restarts, enable **Auto Start Server on Startup** on the config page after the initial installation works. The container restart policy starts **SSUI**; that setting starts the game.

For internet players, forward UDP `27016` to the Docker host. If you change `GamePort`, change the container mapping too.

## Keep the data

`server-data` is the complete `/app` installation on the host. It contains the game files, saves, configuration, mods, TLS keys and backup metadata, so recreating the container does not recreate the server.

Do **not** use `docker compose down -v` as a cleanup habit: with a bind mount, **Docker will not protect you from deleting or moving the host directory yourself either.**

Files created there are owned by root because SSUI currently runs as root inside the container. Avoid adding separate mounts below `/app`; a child mount hides data already present at that path and is a common source of missing saves, configs and backups.

### Existing Docker installations

The one-mount layout is the supported v6 layout. Existing installations that use separate mounts for `UIMod`, saves or the old SSUI data should be migrated manually:

1. Stop the old container.
2. Create `server-data`.
3. Copy the complete existing installation into the matching paths below `server-data`.
4. Keep the old folders until the new container has started and the saves, configuration and backups have been verified.
5. Replace the old volume entries with only `./server-data:/app:rw`.

Do not delete the old folders as part of the migration. SSUI cannot reliably discover host paths that are not mounted into the container, so it cannot safely perform this Docker layout migration by itself.

The current Dockerfile has no `USER` instruction and runs as root inside the container. This avoids the UID/GID mismatch common with Linux bind mounts. **It does not require `--privileged`, host networking or access to the Docker socket.**

At startup SSUI checks that `/app` and its managed data are accessible. In a container it also requires the saves tree to be writable because Stationeers runs in the same container. A failed check names the inaccessible path and exits instead of starting a partially working server. The image sets `SSUI_CONTAINER=1`; keep that variable when adapting the image or entrypoint.

## Stop the container

1. Save the world and wait for its saved event.
2. Stop Stationeers through SSUI and confirm Stopped.
3. Stop the container:

   ```bash
   docker compose stop ssui
   ```

SSUI handles Docker's `SIGTERM`, stops the managed game process and flushes the backup manager before exiting. The example Compose configuration allows 30 seconds for this. Saving and stopping through the UI before planned maintenance is still the safest option, especially for a busy server.

## Update the container

Save and stop the game, then back up the persistent data before an upgrade. Run:

```bash
docker compose pull ssui
docker compose up -d ssui
docker compose logs --follow ssui
```

On each container start the entrypoint copies the image's SSUI binary and license into `/app`. All data stays in this folder. The built-in updater continues checking for new SSUI releases and shows them in the UI, but it will **not** replace the executable in a container. Pulling and recreating the image is the only supported SSUI update path. Game-server updates through SteamCMD are separate and continue to work normally.

The image includes a health check against SSUI's internal `/healthz` endpoint. It allows a long initial start period because the first SteamCMD and game-server installation can take several minutes. If you override `SSUI_WEB_PORT`, publish the same port; the health check reads that environment variable automatically.

## Diagnose the deployment

```bash
docker compose ps
docker compose logs --tail 200 ssui
docker inspect stationeers-server-ui
```

Check installation errors, port conflicts, writable volumes and memory limits. Size resources for the actual world; the repository's four/eight CPU and eight/sixteen GB Compose values are examples, not measured minimums.

For network storage, test reconnects and archive writes as described in [Backups](Backup-System.md#storage-requirements). A container adds no durability to the underlying disk.

<details>
<summary>Build the image from SSUI source</summary>

The repository's `.docker/compose.yml` builds from `.docker/Dockerfile`, persists `/app` and includes example resource limits. From a source checkout:

```bash
docker compose -f .docker/compose.yml build
```

The multi-stage build compiles the frontend and Go application. See the [developer guide](Developer-Documentation.md#building) before running development builds against any valuable save.
 
</details>

---

[Documentation home](index.md) · [All guides](index.md#run-your-server)
