# Hermes Filebrowser

A self-hosted file browser with a React frontend and Go backend. Browse, upload, search, preview, and manage files on your server through a clean web UI.

## Quick Start

```bash
docker compose up -d
```

Open http://localhost:8080 and log in with `admin` / `changeme`.

## Environment Variables

| Variable       | Default                | Description                  |
|----------------|------------------------|------------------------------|
| `FB_PORT`      | `8080`                 | HTTP listen port             |
| `FB_ROOT`      | `/data`                | Root directory to serve      |
| `FB_USERNAME`  | `admin`                | Admin login username         |
| `FB_PASSWORD`  | `admin`                | Admin login password         |
| `FB_DATABASE`  | `/data/filebrowser.db` | SQLite database file path    |

## Building from Source

### Docker

```bash
docker build -t hermes-filebrowser .
docker run -p 8080:8080 -v /path/to/files:/data hermes-filebrowser
```

### Manual

**Frontend:**
```bash
cd frontend
npm ci
npm run build
```

**Backend:**
```bash
cd backend
go build -o hermes ./cmd/hermes
cp -r ../frontend/dist internal/api/frontend/dist
FB_ROOT=/path/to/files ./hermes
```

## Reverse Proxy

The server respects `X-Forwarded-For` headers via `chi.middleware.RealIP`. Ensure your proxy (nginx, Caddy, Traefik) forwards this header.

## Deploy with Portainer

1. Go to **Stacks** > **Add stack**
2. Name: `hermes-filebrowser`
3. Paste the `docker-compose.yml` content into the editor
4. Adjust the volume path under `volumes:` to point to your host directory
5. Change `FB_USERNAME` / `FB_PASSWORD` as desired
6. Click **Deploy the stack**
