### Task 3: Auth System

**Files:**
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/session.go`
- Modify: `backend/internal/db/sqlite.go` (add `GetUserByID` method)

**Interfaces:**
- Consumes: `db.DB` (with `GetSessionByToken`, `GetUserByID`), `config.Config`
- Produces: `CheckPassword(hash, password string) bool`, `NewSessionToken() (string, time.Time)`, `GetUser(r *http.Request) *db.User`, `SessionMiddleware(database *db.DB) func(http.Handler) http.Handler`

- [ ] **Step 1: Password hashing**

File: `backend/internal/auth/password.go`
```go
package auth

import "golang.org/x/crypto/bcrypt"

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

- [ ] **Step 2: Session token generation + middleware**

File: `backend/internal/auth/session.go`
```go
package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/youruser/hermes-filebrowser/internal/db"
)

type contextKey string

const userKey contextKey = "user"

func NewSessionToken() (string, time.Time) {
	return uuid.New().String(), time.Now().Add(24 * time.Hour)
}

func GetUser(r *http.Request) *db.User {
	u, _ := r.Context().Value(userKey).(*db.User)
	return u
}

func SessionMiddleware(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimPrefix(h, "Bearer ")
			}
			if token == "" {
				if c, err := r.Cookie("session"); err == nil {
					token = c.Value
				}
			}
			if token != "" {
				session, err := database.GetSessionByToken(token)
				if err == nil {
					expiresAt, _ := time.Parse(time.RFC3339, session.ExpiresAt)
					if time.Now().Before(expiresAt) {
						user, _ := database.GetUserByID(session.UserID)
						ctx := context.WithValue(r.Context(), userKey, user)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 3: GetUserByID helper — add method to db.DB** (add to `backend/internal/db/sqlite.go`)

```go
func (d *DB) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := d.conn.QueryRow("SELECT id, username, password_hash, read_only FROM users WHERE id = ?", id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.ReadOnly)
	if err != nil {
		return nil, err
	}
	return u, nil
}
```

- [ ] **Step 4: Add dependency**

```bash
cd backend && go get github.com/google/uuid && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth/ backend/internal/db/sqlite.go
git commit -m "feat: add auth system with password hashing and session middleware"
```
