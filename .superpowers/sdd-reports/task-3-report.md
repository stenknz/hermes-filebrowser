# Task 3 Report: Auth System

## What I implemented

- **`internal/auth/password.go`** — `CheckPassword(hash, password string) bool` using `bcrypt.CompareHashAndPassword`
- **`internal/auth/session.go`** — `NewSessionToken()` (UUID v4 + 24h expiry), `GetUser(r *http.Request) *db.User`, `SessionMiddleware(database *db.DB)` (extracts token from Authorization: Bearer header or session cookie, validates expiry, injects user into context)
- **`internal/db/sqlite.go`** — added `GetUserByID(id int64) (*User, error)` query method

## Test results (TDD evidence)

**RED phase** — tests written before implementation, all failed with `undefined` errors:
```
internal\auth\auth_test.go:20:6: undefined: CheckPassword
internal\auth\auth_test.go:36:22: undefined: NewSessionToken
... (11 failures)
```

**GREEN phase** — all 11 tests pass:
```
=== RUN   TestCheckPassword_Valid — PASS
=== RUN   TestCheckPassword_Invalid — PASS
=== RUN   TestNewSessionToken_NotEmpty — PASS
=== RUN   TestNewSessionToken_Unique — PASS
=== RUN   TestGetUser_NilWhenNoUser — PASS
=== RUN   TestSessionMiddleware_AuthorizationHeader — PASS
=== RUN   TestSessionMiddleware_Cookie — PASS
=== RUN   TestSessionMiddleware_AuthorizationTakesPriority — PASS
=== RUN   TestSessionMiddleware_ExpiredToken — PASS
=== RUN   TestSessionMiddleware_InvalidToken — PASS
=== RUN   TestSessionMiddleware_NoToken — PASS
PASS  ok  github.com/youruser/hermes-filebrowser/internal/auth  2.369s
```

**Build:** `go build ./cmd/hermes` → success
**DB tests:** `go test ./internal/db/... -v` → all 8 existing tests pass
**Vet:** `go vet ./internal/auth/...` and `./internal/db/...` → clean

## Files changed

| File | Action |
|------|--------|
| `backend/internal/auth/password.go` | Created |
| `backend/internal/auth/session.go` | Created |
| `backend/internal/auth/auth_test.go` | Created |
| `backend/internal/db/sqlite.go` | Modified — added `GetUserByID` |
| `backend/go.mod` | Modified — `github.com/google/uuid` promoted to direct dep |
| `backend/go.sum` | Modified — updated checksums |

## Self-review findings

- Code matches the task brief spec exactly
- `GetUserByID` uses `int64` parameter matching `User.ID` type from models
- SessionMiddleware correctly prioritizes Authorization header over cookie
- Expired tokens are silently ignored (user stays nil, no error returned)
- Token extraction does not attempt to parse malformed Authorization headers (just checks prefix)
- All edge cases covered: valid, invalid, expired, no token, header vs cookie priority

## Concerns

None.
