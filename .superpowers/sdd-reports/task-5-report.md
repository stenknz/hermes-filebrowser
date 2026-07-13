# Task 5 Report: API Handlers + Routes

## Status: DONE

## Commits
- `64e1602` feat: add API handlers and chi router

## Test Summary
- `go build ./cmd/hermes` — builds clean
- `go vet ./internal/api/...` — no issues
- `go test ./... -count=1` — all packages pass
- API package: **25 tests + 6 subtests = 31/31 passing**

## Files Created
| File | Purpose |
|------|---------|
| `backend/internal/api/middleware.go` | CSRF middleware, logging middleware, token generator |
| `backend/internal/api/auth.go` | Login/logout/me handlers |
| `backend/internal/api/files.go` | File CRUD handlers (list, read, upload, mkdir, create, rename, delete, copy, move, thumbnail) |
| `backend/internal/api/search.go` | Search handler (client-side filter over list) |
| `backend/internal/api/routes.go` | Chi router with auth + protected + frontend route groups |
| `backend/internal/api/api_test.go` | Comprehensive tests for all handlers and middleware |

## Notes
- Fixed import name conflict in `routes.go`: aliased `io/fs` as `iofs` (both `internal/fs` and `io/fs` would otherwise collide under the name `fs`)
- `fs.NewService` now returns `(*fs.Service, error)` — handled in `routes.go` with `log.Fatalf`
- Reused the existing `frontend/dist/.gitkeep` for the `//go:embed` directive
- All write operations check `user.ReadOnly` before mutating
- CSRF uses double-submit cookie pattern (safe methods pass through)
