package db

import (
	"database/sql"

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

func (d *DB) CreateUser(username, passwordHash string, readOnly bool) (*User, error) {
	u := &User{}
	err := d.conn.QueryRow("INSERT INTO users (username, password_hash, read_only) VALUES (?, ?, ?) RETURNING id, username, password_hash, read_only",
		username, passwordHash, readOnly).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.ReadOnly)
	if err != nil {
		return nil, err
	}
	return u, nil
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
