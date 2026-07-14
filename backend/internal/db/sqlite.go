package db

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"
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
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA foreign_keys=ON")
	conn.Exec("PRAGMA busy_timeout=5000")
	if err := conn.Ping(); err != nil {
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
			role TEXT NOT NULL DEFAULT 'viewer',
			home_path TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			token TEXT UNIQUE NOT NULL,
			expires_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			token TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}
	// Handle migration from old read_only column to role
	d.conn.Exec("UPDATE users SET role = 'viewer' WHERE role = '' OR role IS NULL")
	return nil
}

func (d *DB) EnsureAdmin(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var count int
	if err := d.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		// Update password if admin user already exists with this username
		_, err = d.conn.Exec("UPDATE users SET password_hash = ? WHERE username = ? AND role = ?", string(hash), username, RoleAdmin)
		return err
	}
	_, err = d.conn.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", username, string(hash), RoleAdmin)
	return err
}

func (d *DB) CreateUser(username, passwordHash string, role Role, homePath ...string) (*User, error) {
	hp := ""
	if len(homePath) > 0 {
		hp = homePath[0]
	}
	u := &User{}
	err := d.conn.QueryRow("INSERT INTO users (username, password_hash, role, home_path) VALUES (?, ?, ?, ?) RETURNING id, username, password_hash, role, home_path",
		username, passwordHash, role, hp).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.HomePath)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *DB) ListUsers() ([]*User, error) {
	rows, err := d.conn.Query("SELECT id, username, password_hash, role, home_path FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.HomePath); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (d *DB) DeleteUser(id int64) error {
	if _, err := d.conn.Exec("DELETE FROM sessions WHERE user_id = ?", id); err != nil {
		return err
	}
	if _, err := d.conn.Exec("DELETE FROM api_tokens WHERE user_id = ?", id); err != nil {
		return err
	}
	_, err := d.conn.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (d *DB) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := d.conn.QueryRow("SELECT id, username, password_hash, role, home_path FROM users WHERE username = ?", username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.HomePath)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *DB) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := d.conn.QueryRow("SELECT id, username, password_hash, role, home_path FROM users WHERE id = ?", id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.HomePath)
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

func (d *DB) CreateApiToken(userID int64, token, name string) (*ApiToken, error) {
	t := &ApiToken{}
	err := d.conn.QueryRow("INSERT INTO api_tokens (user_id, token, name, created_at) VALUES (?, ?, ?, ?) RETURNING id, user_id, token, name, created_at",
		userID, token, name, time.Now().UTC().Format(time.RFC3339)).Scan(&t.ID, &t.UserID, &t.Token, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (d *DB) GetApiTokenByToken(token string) (*ApiToken, error) {
	t := &ApiToken{}
	err := d.conn.QueryRow("SELECT id, user_id, token, name, created_at FROM api_tokens WHERE token = ?", token).Scan(&t.ID, &t.UserID, &t.Token, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (d *DB) ListApiTokens(userID int64) ([]*ApiToken, error) {
	rows, err := d.conn.Query("SELECT id, user_id, token, name, created_at FROM api_tokens WHERE user_id = ? ORDER BY id", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []*ApiToken
	for rows.Next() {
		t := &ApiToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (d *DB) DeleteApiToken(id, userID int64) error {
	_, err := d.conn.Exec("DELETE FROM api_tokens WHERE id = ? AND user_id = ?", id, userID)
	return err
}
