### Task 2: Database Layer

**Files:**
- Create: `backend/internal/db/sqlite.go`
- Create: `backend/internal/db/models.go`

**Interfaces:**
- Consumes: `config.Config.DatabasePath`
- Produces: `db.DB` struct with methods: `Init() error`, `Close() error`, `CreateUser(username, passwordHash string, readOnly bool) (*User, error)`, `GetUserByUsername(username string) (*User, error)`, `CreateSession(userID int64, token, expiresAt string) error`, `GetSessionByToken(token string) (*Session, error)`, `DeleteSession(token string) error`

- [ ] **Step 1: Create models**

File: `backend/internal/db/models.go`
```go
package db

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	ReadOnly     bool   `json:"readOnly"`
}

type Session struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}
```

- [ ] **Step 2: Create SQLite layer with auto-migration and default admin user creation**

File: `backend/internal/db/sqlite.go`
```go
package db

import (
	"database/sql"
	_ "modernc.org/sqlite"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			read_only INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			token TEXT UNIQUE NOT NULL,
			expires_at TEXT NOT NULL
		);
	`)
	return err
}

func (d *DB) EnsureAdmin(username, password string) error {
	var count int
	d.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec("INSERT INTO users (username, password_hash, read_only) VALUES (?, ?, 0)", username, string(hash))
	return err
}

func (d *DB) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := d.conn.QueryRow("SELECT id, username, password_hash, read_only FROM users WHERE username = ?", username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.ReadOnly)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *DB) CreateSession(userID int64, token, expiresAt string) error {
	_, err := d.conn.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)", userID, token, expiresAt)
	return err
}

func (d *DB) GetSessionByToken(token string) (*Session, error) {
	s := &Session{}
	err := d.conn.QueryRow("SELECT id, user_id, token, expires_at FROM sessions WHERE token = ?", token).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (d *DB) DeleteSession(token string) error {
	_, err := d.conn.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
```

- [ ] **Step 3: Add dependencies**

```bash
cd backend && go get modernc.org/sqlite golang.org/x/crypto/bcrypt && go mod tidy
```

- [ ] **Step 4: Build to verify**

```bash
cd backend && go build ./cmd/hermes
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/db/
git commit -m "feat: add SQLite database layer with users and sessions"
```
