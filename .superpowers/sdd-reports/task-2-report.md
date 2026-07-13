# Task 2 Report: Database Layer

## What was implemented

Created the SQLite database layer for the Hermes Filebrowser in the `backend/internal/db/` package:

- **`models.go`** — `User` and `Session` structs with JSON tags (PasswordHash uses `json:"-"`)
- **`sqlite.go`** — `DB` struct wrapping `*sql.DB` with auto-migration and CRUD methods:
  - `New(path)` — opens SQLite, runs migration, returns DB
  - `Close()` — closes connection
  - `migrate()` — creates `users` and `sessions` tables (IF NOT EXISTS)
  - `EnsureAdmin(username, password)` — seeds admin if no users exist (bcrypt hashed)
  - `CreateUser(username, passwordHash, readOnly)` — inserts user, returns created User
  - `GetUserByUsername(username)` — queries by username
  - `CreateSession(userID, token, expiresAt)` — inserts session
  - `GetSessionByToken(token)` — queries by token
  - `DeleteSession(token)` — deletes session

Dependencies: `modernc.org/sqlite` (pure Go SQLite, no CGo), `golang.org/x/crypto/bcrypt` (password hashing).

## Test Results (TDD)

### RED phase (before implementation)
```
internal\db\sqlite_test.go:10:12: undefined: New
internal\db\sqlite_test.go:19:12: undefined: New
...
FAIL    github.com/youruser/hermes-filebrowser/internal/db [build failed]
```
Tests failed correctly — compilation errors because package didn't exist.

### GREEN phase (after implementation)
```
=== RUN   TestNewDB              --- PASS: 0.02s
=== RUN   TestCreateAndGetUser   --- PASS: 0.02s
=== RUN   TestCreateAndGetSession --- PASS: 0.03s
=== RUN   TestDeleteSession      --- PASS: 0.03s
=== RUN   TestEnsureAdminCreatesAdmin --- PASS: 0.06s
=== RUN   TestEnsureAdminIdempotent    --- PASS: 0.06s
=== RUN   TestDuplicateUser      --- PASS: 0.02s
=== RUN   TestGetNonExistentUser --- PASS: 0.02s
PASS
ok  	github.com/youruser/hermes-filebrowser/internal/db	2.457s
```

**8/8 passing, output pristine** (no warnings, no errors). Each test uses `t.TempDir()` for isolated temporary databases.

## Files changed

| File | Action | Lines |
|------|--------|-------|
| `backend/internal/db/models.go` | created | 15 |
| `backend/internal/db/sqlite.go` | created | 99 |
| `backend/internal/db/sqlite_test.go` | created | 170 |
| `backend/go.mod` | modified (deps added) | +4 |
| `backend/go.sum` | created | 204 |

## Self-review findings

### Matches spec
- Models match the exact field definitions and JSON tags from the brief
- All interface methods are implemented with correct signatures
- Auto-migration creates tables on `New()`
- `EnsureAdmin` uses bcrypt and is idempotent (skips if users exist)
- Build passes (`go build ./cmd/hermes` → success)

### Minor deviations
- `CreateUser` uses `INSERT ... RETURNING` (SQLite 3.35+) instead of separate INSERT + SELECT. This is cleaner and avoids a race window. The interface is identical.

### Issues
- `EnsureAdmin` silently ignores `QueryRow().Scan()` errors for the count query (matches brief's code exactly). If the query fails, `count` stays 0 and it attempts to insert. This is a minor concern — the migration runs first so the table always exists, making failure unlikely.

## Concerns

None significant. The database layer is clean and ready for Task 3 (Auth System) and Task 5 (API Handlers).
