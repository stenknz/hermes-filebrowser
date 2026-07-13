# Hermes Filebrowser — Design Spec

## Overview

A web-based file browser SPA ("Hermes Filebrowser") — single Go binary embedding a React frontend. Runs in Docker on a NAS behind a reverse proxy. Deployed via Portainer from Docker Hub images built by GitHub Actions.

## Tech Stack

- **Backend:** Go 1.23+, `chi` router, `modernc.org/sqlite` (pure Go SQLite)
- **Frontend:** React + Vite + Tailwind CSS + `react-icons` + `react-pdf`
- **Database:** SQLite (users, sessions)
- **CI/CD:** GitHub Actions → Docker Hub (multi-arch: linux/amd64 + linux/arm64)

## Project Layout

```
file-browser/
├── backend/
│   ├── cmd/hermes/
│   │   └── main.go              — entrypoint, server startup
│   ├── internal/
│   │   ├── api/
│   │   │   ├── routes.go        — route registration
│   │   │   ├── auth.go          — login/logout handlers
│   │   │   ├── files.go         — file CRUD handlers
│   │   │   ├── search.go        — search handler
│   │   │   └── middleware.go    — auth, CSRF, logging
│   │   ├── auth/
│   │   │   ├── session.go       — session token generation/validation
│   │   │   └── password.go      — bcrypt hashing
│   │   ├── db/
│   │   │   ├── sqlite.go        — DB init, migrations
│   │   │   └── models.go        — user, session structs
│   │   ├── fs/
│   │   │   └── service.go       — file ops with path validation
│   │   └── config/
│   │       └── config.go        — env var parsing
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Breadcrumb.tsx
│   │   │   ├── FileList.tsx
│   │   │   ├── FileRow.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── DropZone.tsx
│   │   │   ├── SearchBar.tsx
│   │   │   ├── PreviewPane.tsx
│   │   │   ├── UploadProgress.tsx
│   │   │   └── FileIcon.tsx
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   └── BrowserPage.tsx
│   │   ├── api/
│   │   │   └── client.ts        — fetch wrapper with auth
│   │   ├── context/
│   │   │   └── AuthContext.tsx
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── tsconfig.json
├── Dockerfile                    — multi-stage build, distroless final
├── docker-compose.yml
├── .github/workflows/
│   └── docker-publish.yml
├── AGENTS.md
└── README.md
```

## Backend API

| Method | Route | Purpose |
|--------|-------|---------|
| `POST` | `/api/login` | Auth → session token |
| `POST` | `/api/logout` | Invalidate session |
| `GET` | `/api/me` | Current user info |
| `GET` | `/api/files?path=` | List directory |
| `GET` | `/api/files/raw?path=` | Download/preview file |
| `GET` | `/api/files/thumbnail?path=` | Image thumbnail |
| `POST` | `/api/files/upload?path=` | Upload (multipart) |
| `POST` | `/api/files/dir?path=` | Create folder |
| `POST` | `/api/files/file?path=` | Create text file |
| `PUT` | `/api/files/rename` | Rename `{oldPath, newPath}` |
| `DELETE` | `/api/files?path=` | Delete |
| `POST` | `/api/files/copy` | Copy `{source, destination}` |
| `POST` | `/api/files/move` | Move `{source, destination}` |
| `GET` | `/api/search?q=&path=` | Search by name |

All responses: `{ "data": ... }` on success, `{ "error": "message" }` on failure.

## Auth

- Session tokens: UUID v4 stored in SQLite with expiry (24h default)
- Sent as `Authorization: Bearer` header + httpOnly cookie
- CSRF: double-submit cookie pattern (random token in cookie + header)
- Password: bcrypt
- Read-only mode per user enforced server-side

## File Operations

- All paths resolved relative to configured `FB_ROOT`
- Path traversal check: reject any path containing `..` after `filepath.Clean`
- Operations fail with 403 if user is read-only
- Upload: multipart form, streamed to disk (no temp file)
- Thumbnails: generated server-side for images (resize to 200px), cached in memory or temp dir

## Frontend

- Dark theme by default using Tailwind's `darkMode: 'class'`
- File icons: `react-icons` mapped by extension/MIME (`FiFolder`, `FiFileText`, `FiImage`, `FiFile`, etc.)
- PDF preview: `react-pdf` (PDF.js renderer)
- Image preview: inline `<img>` with thumbnail source
- Text preview: fetch raw content, render in `<pre>` with syntax highlighting
- Drag-and-drop: HTML5 DnD API → upload endpoint
- Sortable columns: client-side sort by name/size/date
- Responsive: sidebar collapses to top bar on <768px

## Error Handling

- HTTP codes: 400 bad request, 401 unauthorized, 403 read-only, 404 not found, 500 internal
- Error messages logged server-side, sanitized before returning to client (no stack traces)
- Frontend shows toast notifications on errors

## Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `FB_PORT` | `8080` | Listen port |
| `FB_ROOT` | `/data` | Root directory for file browsing |
| `FB_USERNAME` | `admin` | Default admin username |
| `FB_PASSWORD` | `admin` | Default admin password |
| `FB_DATABASE` | `/data/filebrowser.db` | SQLite DB path |

## Docker

- Multi-stage build: `golang:1.23-alpine` → `node:20-alpine` (frontend build) → `gcr.io/distroless/base`
- Final image ≈ 25MB
- Exposes port 8080, supports `X-Forwarded-For` for reverse proxy

## docker-compose.yml

```yaml
services:
  hermes:
    image: hermes-filebrowser:latest
    ports:
      - "8080:8080"
    environment:
      - FB_PORT=8080
      - FB_ROOT=/data
      - FB_USERNAME=admin
      - FB_PASSWORD=changeme
      - FB_DATABASE=/data/filebrowser.db
    volumes:
      - /volume2/HermesShared:/data
    restart: unless-stopped
```

## CI/CD (GitHub Actions)

On push to `main`:
1. Checkout code
2. Set up Docker Buildx
3. Build + push to Docker Hub (`docker.io/<user>/hermes-filebrowser:latest`, `:sha`)
4. Multi-arch: `linux/amd64`, `linux/arm64`

## Testing

- Backend: Go `httptest` for API handlers (temp SQLite + temp dir fixture)
- Frontend: Vitest + React Testing Library for component tests
- Security: test path traversal attempts, unauthenticated requests, read-only enforcement
