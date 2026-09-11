!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

This guide covers how to run StationeersServerUI using Docker and Docker Compose.

!!! warning
    For reliable backup operations, use local storage or **block-level** network storage (iSCSI, Fibre Channel). Network shares like SMB, NFS1/2, CIFS are NOT supported and cause issues. 
    See [Backup-System-v2/Important Notes on Filesystem Support](https://github.com/SteamServerUI/StationeersServerUI/wiki/Backup-System-v2#important-notes-on-filesystem-support) for further details if you want to use Network Storage.

## Running through docker cli

1a. **Create a named volume** *(Optional)*

   `docker volume create app-data`

   *(Optional)* This ensures the volume exists and can be managed separately if needed


1b. **Run the Image**

   ```sh
   docker run -d --name stationeers-server-ui -p 8443:8443 -p 27016:27016/udp -p 27016:27016/tcp -p 27015:27015/udp -p 27015:27015/tcp -v app-data:/app -v "./saves:/app/saves:rw" -v "./UIMod/config:/app/UIMod/config:rw" -v "./UIMod/tls:/app/UIMod/tls:rw" ghcr.io/steamserverui/stationeersserverui:latest
   ```

2. **Check if the container is running:**

   ```sh
   docker ps
   ```

3. **View logs (if debugging is needed):**

   ```sh
   docker logs stationeers-server-ui
   ```

4. **Stop the container:**

   ```sh
   docker stop stationeers-server-ui
   ```

## Running with Docker Compose

1. **Create (or use provided) compose.yml File**

```yaml
services:
  stationeers-server-ui:
    container_name: stationeers-server-ui
    image: ghcr.io/steamserverui/stationeersserverui:latest
    deploy:
      resources:
        limits:
          cpus: '8'
          memory: 16G
        reservations:
          cpus: '4'
          memory: 8G
    ports:
      - "8443:8443"
      - "27016:27016/udp"
      - "27016:27016/tcp"
      - "27015:27015/udp"
      - "27015:27015/tcp"
    volumes:
      - app-data:/app
      - ./saves:/app/saves:rw
      - ./UIMod/config:/app/UIMod/config:rw
      - ./UIMod/tls:/app/UIMod/tls:rw
    restart: unless-stopped

volumes:
  app-data:
```

2. **Run Docker Compose**

   ```sh
   docker compose up -d
   ```

3. **Check Logs (Optional)**

   ```sh
   docker compose logs -f
   ```

   Press **CTRL+C** to exit the log view.


## Next Steps

- [First-Time Setup](First-Time-Setup.md) - Configure your server for first use
- [Security Considerations](Security-Considerations.md) - Important security notes
