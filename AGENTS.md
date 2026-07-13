# AGENTS.md — FileBrowser

Web-based file browser SPA: Go backend (`chi` + `modernc.org/sqlite`) + React frontend (Vite + Tailwind). Single binary, dark theme, runs in Docker on NAS.

## Dev commands

| What | Command |
|------|---------|
| Run backend | `cd backend && go run ./cmd/hermes` |
| Run frontend dev | `cd frontend && npm run dev` (proxies `/api` to `:8080`) |
| Build frontend | `cd frontend && npm run build` |
| Build Go binary | `cd backend && go build ./cmd/hermes` |
| Run all Go tests | `cd backend && go test ./... -count=1` |
| Run single Go pkg | `cd backend && go test ./internal/<pkg>/ -v` |
| Run frontend typecheck | `cd frontend && npx tsc --noEmit` |
| Full Docker build | `docker build -t filebrowser .` (multi-stage, distroless) |
| GitHub Actions CI | Push to `master` → builds multi-arch image → pushes to Docker Hub |

## Architecture notes

- Module path: `github.com/stenknz/hermes-filebrowser`
- Frontend embeds into Go binary via `//go:embed` — Dockerfile copies `frontend/dist/` to `backend/internal/api/frontend/dist/` for the embed to find
- For local dev, Vite's proxy handles `/api` → backend — no need to rebuild frontend for backend changes
- `db.NewService` returns `(*Service, error)` — handle the error
- Session middleware checks `Authorization: Bearer` header first, then `session` cookie, then API tokens
- CSRF uses double-submit cookie pattern: `csrf_token` cookie + `X-CSRF-Token` header
- API tokens (prefix `fb_`) bypass CSRF — they authenticate via `Authorization: Bearer` header only
- Role system: `admin` (full), `editor` (read/write), `viewer` (read-only)

## Key conventions

- Go: `chi` router (not gin), `modernc.org/sqlite` (no CGo), UUID v4 sessions
- Frontend: Tailwind CSS variables via `:root`, dark theme only, `react-icons/fi` for file icons
- Error responses: `{"error":"message"}` JSON, appropriate HTTP codes
- All paths validated against `FB_ROOT` — `SafePath` prevents traversal
- DB files (`filebrowser.db*`) are hidden from file listing automatically

## API endpoints

| Method | Route | Purpose |
|--------|-------|---------|
| POST | `/api/login` | Auth |
| POST | `/api/logout` | End session |
| GET | `/api/me` | Current user |
| GET | `/api/files?path=` | List directory |
| GET | `/api/files/raw?path=` | Read file |
| POST | `/api/files/upload?path=` | Upload |
| POST | `/api/files/dir?path=` | Create folder |
| POST | `/api/files/file` | Create/edit file |
| PUT | `/api/files/rename` | Rename |
| DELETE | `/api/files?path=` | Delete |
| POST | `/api/files/copy` | Copy |
| POST | `/api/files/move` | Move |
| GET | `/api/search?q=&path=` | Search |
| GET | `/api/users` | List users (admin) |
| POST | `/api/users` | Create user (admin) |
| POST | `/api/users/delete` | Delete user (admin) |
| GET | `/api/tokens` | List API tokens |
| POST | `/api/tokens` | Create API token |
| POST | `/api/tokens/delete` | Revoke API token |

## Secrets for CI

- `DOCKERHUB_USERNAME` — Docker Hub username
- `DOCKERHUB_TOKEN` — Docker Hub access token
