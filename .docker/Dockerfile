###############
# Frontend build
###############
FROM node:24-bookworm-slim AS frontend-builder

WORKDIR /src/frontend

# Only copy frontend to leverage Docker layer caching for npm install
COPY ./frontend/package.json ./frontend/package-lock.json* ./
RUN npm install

COPY ./frontend/ ./
# Build Svelte UI into ../SSUI/onboard_bundled/v2 (as configured in vite.config.js)
RUN npm run build

############
# Go build
############
FROM golang:1.26-bookworm AS go-builder

WORKDIR /src

# Pre-copy go.mod and go.sum for better caching
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . ./

COPY --from=frontend-builder /src/SSUI /src/SSUI

# Build the server (embed will include SSUI/* at build time)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/StationeersServerUI ./server.go

############
# Runtime
############
FROM debian:13-slim AS runner

WORKDIR /app

ENV HOME=/app
ENV SSUI_CONTAINER=1

# Minimal runtime dependencies. SteamCMD needs 32-bit runtime.
RUN dpkg --add-architecture i386 \
 && apt-get update -y \
 && apt-get install -y --no-install-recommends \
    ca-certificates \
    locales \
    lib32gcc-s1 \
    file \
    wget \
 && rm -rf /var/lib/apt/lists/*

# Copy compiled binary and license
RUN mkdir -p /opt/SSUIBuildFiles
COPY --from=go-builder /out/StationeersServerUI /opt/SSUIBuildFiles/StationeersServerUI
COPY ./LICENSE /opt/SSUIBuildFiles/LICENSE
COPY ./.docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /opt/SSUIBuildFiles/StationeersServerUI
RUN chmod +x /entrypoint.sh


EXPOSE 8443/tcp 27016/udp 27015/udp

HEALTHCHECK --interval=30s --timeout=5s --start-period=15m --retries=3 \
  CMD wget --no-check-certificate --spider "https://127.0.0.1:${SSUI_WEB_PORT:-8443}/healthz" || exit 1

ENTRYPOINT ["/entrypoint.sh"]
