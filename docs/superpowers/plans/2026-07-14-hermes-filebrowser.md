# Hermes Filebrowser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Hermes Filebrowser — a web-based file browser SPA (Go backend + React frontend) that runs in Docker on a NAS, deployable via Portainer.

**Architecture:** Single Go binary embedding a Vite-built React frontend via `embed.FS`. SQLite for auth/sessions. All file ops constrained to `FB_ROOT` with path traversal prevention. Chi router for HTTP.

**Tech Stack:** Go 1.23+, chi, modernc.org/sqlite, React 18, Vite, Tailwind CSS, react-icons, react-pdf.

## Global Constraints

- Go 1.23+, use `modernc.org/sqlite` (pure Go, no CGo, enables easy cross-compile for arm64)
- Use `chi` router (not gin, not stdlib mux)
- Frontend embedded via `//go:embed` in final binary
- All paths must be validated against `FB_ROOT` — reject any `..` traversal
- Session tokens: UUID v4, 24h expiry, stored in SQLite
- CSRF: double-submit cookie pattern
- Docker: multi-stage build with distroless final image
- Docker Hub multi-arch: linux/amd64 + linux/arm64
- Dark theme only, no light mode

---
### Task 1: Go Module + Config + Main Skeleton

**Files:**
- Create: `backend/go.mod`
- Create: `backend/internal/config/config.go`
- Create: `backend/cmd/hermes/main.go`

**Interfaces:**
- Consumes: nothing
- Produces: `config.Config` struct with fields: `Port, Root, Username, Password, DatabasePath`

- [ ] **Step 1: Initialize Go module**

Run: `cd backend && go mod init github.com/<user>/file-browser`

```bash
cd backend
go mod init github.com/youruser/hermes-filebrowser
```

- [ ] **Step 2: Create config package**

File: `backend/internal/config/config.go`
```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Root         string
	Username     string
	Password     string
	DatabasePath string
}

func Load() *Config {
	return &Config{
		Port:         getEnvInt("FB_PORT", 8080),
		Root:         getEnv("FB_ROOT", "/data"),
		Username:     getEnv("FB_USERNAME", "admin"),
		Password:     getEnv("FB_PASSWORD", "admin"),
		DatabasePath: getEnv("FB_DATABASE", "/data/filebrowser.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
```

- [ ] **Step 3: Create main.go skeleton**

File: `backend/cmd/hermes/main.go`
```go
package main

import (
	"log"

	"github.com/youruser/hermes-filebrowser/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("Hermes Filebrowser starting on port %d", cfg.Port)
	_ = cfg
}
```

- [ ] **Step 4: Tidy and verify**

```bash
cd backend && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: add Go module, config, and main skeleton"
```

---
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

---
### Task 3: Auth System

**Files:**
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/session.go`

**Interfaces:**
- Consumes: `db.DB`, `config.Config`
- Produces: `CheckPassword(hash, password string) bool`, `NewSessionToken() string`, `SessionMiddleware(next http.Handler) http.Handler` — extracts Bearer token, validates session, sets `context` with user info

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
git add backend/internal/auth/
git commit -m "feat: add auth system with password hashing and session middleware"
```

---
### Task 4: File Service

**Files:**
- Create: `backend/internal/fs/service.go`

**Interfaces:**
- Consumes: `config.Config.Root`
- Produces: `NewService(root string) *Service` with methods: `List(path string) ([]FileInfo, error)`, `Read(path string) ([]byte, error)`, `Write(path string, data []byte) error`, `Delete(path string) error`, `Rename(oldPath, newPath string) error`, `Copy(src, dst string) error`, `Mkdir(path string) error`, `Thumbnail(path string) ([]byte, error)`, `SafePath(path string) (string, error)`

- [ ] **Step 1: Create file service with path traversal prevention**

File: `backend/internal/fs/service.go`
```go
package fs

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

type FileInfo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	IsDir   bool        `json:"isDir"`
	ModTime time.Time   `json:"modTime"`
	Mode    os.FileMode `json:"mode"`
}

type Service struct {
	root string
}

func NewService(root string) *Service {
	absRoot, _ := filepath.Abs(root)
	return &Service{root: absRoot}
}

func (s *Service) SafePath(path string) (string, error) {
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "..") || strings.Contains(clean, "..") {
		return "", os.ErrPermission
	}
	fullPath := filepath.Join(s.root, clean)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, s.root) {
		return "", os.ErrPermission
	}
	return absPath, nil
}

func (s *Service) List(dirPath string) ([]FileInfo, error) {
	fullPath, err := s.SafePath(dirPath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var infos []FileInfo
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, FileInfo{
			Name:    e.Name(),
			Path:    filepath.Join(dirPath, e.Name()),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}
	return infos, nil
}

func (s *Service) Read(filePath string) ([]byte, error) {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fullPath)
}

func (s *Service) Write(filePath string, data []byte) error {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *Service) Delete(filePath string) error {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

func (s *Service) Rename(oldPath, newPath string) error {
	fullOld, err := s.SafePath(oldPath)
	if err != nil {
		return err
	}
	fullNew, err := s.SafePath(newPath)
	if err != nil {
		return err
	}
	return os.Rename(fullOld, fullNew)
}

func (s *Service) Copy(src, dst string) error {
	fullSrc, err := s.SafePath(src)
	if err != nil {
		return err
	}
	fullDst, err := s.SafePath(dst)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(fullSrc)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(fullDst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *Service) Mkdir(dirPath string) error {
	fullPath, err := s.SafePath(dirPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, 0755)
}

func (s *Service) Thumbnail(filePath string) ([]byte, error) {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return nil, err
	}
	src, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	const maxSize = 200
	newW, newH := bounds.Dx(), bounds.Dy()
	if newW > maxSize || newH > maxSize {
		ratio := float64(maxSize) / float64(max(newW, newH))
		newW = int(float64(newW) * ratio)
		newH = int(float64(newH) * ratio)
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 2: Add thumbnail dependency**

```bash
cd backend && go get golang.org/x/image && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/fs/
git commit -m "feat: add file service with path traversal prevention"
```

---
### Task 5: API Handlers + Routes

**Files:**
- Create: `backend/internal/api/middleware.go` — CSRF middleware, logging, CORS
- Create: `backend/internal/api/auth.go` — login/logout/me handlers
- Create: `backend/internal/api/files.go` — list/read/upload/rename/delete/copy/move/mkdir/thumbnail
- Create: `backend/internal/api/search.go` — search by name
- Create: `backend/internal/api/routes.go` — wire up chi router

**Interfaces:**
- Consumes: `*db.DB`, `*fs.Service`, `config.Config`
- Produces: `http.Handler` (chi router with all routes mounted)

- [ ] **Step 1: Middleware**

File: `backend/internal/api/middleware.go`
```go
package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie("csrf_token")
		if err != nil {
			http.Error(w, `{"error":"missing CSRF cookie"}`, http.StatusForbidden)
			return
		}
		header := r.Header.Get("X-CSRF-Token")
		if header == "" || header != cookie.Value {
			http.Error(w, `{"error":"CSRF token mismatch"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func GenerateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

- [ ] **Step 2: Auth handlers**

File: `backend/internal/api/auth.go`
```go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/youruser/hermes-filebrowser/internal/auth"
	"github.com/youruser/hermes-filebrowser/internal/db"
)

type authHandler struct {
	db *db.DB
	cfg *config.Config
}

func NewAuthHandler(database *db.DB, cfg *config.Config) *authHandler {
	return &authHandler{db: database, cfg: cfg}
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	token, expiresAt := auth.NewSessionToken()
	if err := h.db.CreateSession(user.ID, token, expiresAt.Format(time.RFC3339)); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: token, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteStrictMode,
		Expires: expiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name: "csrf_token", Value: GenerateCSRFToken(), Path: "/",
		HttpOnly: false, SameSite: http.SameSiteStrictMode,
		Expires: expiresAt,
	})
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token, "user": user,
	})
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("session")
	if c != nil {
		h.db.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "csrf_token", MaxAge: -1, Path: "/"})
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user})
}
```

- [ ] **Step 3: File handlers**

File: `backend/internal/api/files.go`
```go
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/youruser/hermes-filebrowser/internal/auth"
	"github.com/youruser/hermes-filebrowser/internal/fs"
)

type fileHandler struct {
	svc *fs.Service
}

func NewFileHandler(svc *fs.Service) *fileHandler {
	return &fileHandler{svc: svc}
}

func (h *fileHandler) List(w http.ResponseWriter, r *http.Request) {
	dirPath := r.URL.Query().Get("path")
	entries, err := h.svc.List(dirPath)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": entries})
}

func (h *fileHandler) Read(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	data, err := h.svc.Read(filePath)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	ext := filepath.Ext(filePath)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		w.Header().Set("Content-Type", "image/"+ext[1:])
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Write(data)
}

func (h *fileHandler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	data, err := h.svc.Thumbnail(filePath)
	if err != nil {
		http.Error(w, `{"error":"cannot generate thumbnail"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(data)
}

func (h *fileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	dirPath := r.URL.Query().Get("path")
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"missing file"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, _ := io.ReadAll(file)
	if err := h.svc.Write(filepath.Join(dirPath, header.Filename), data); err != nil {
		http.Error(w, `{"error":"write failed"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) CreateDir(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	dirPath := r.URL.Query().Get("path")
	if err := h.svc.Mkdir(dirPath); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.Write(req.Path, []byte(req.Content)); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) Rename(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.Rename(req.OldPath, req.NewPath); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *fileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	filePath := r.URL.Query().Get("path")
	if err := h.svc.Delete(filePath); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *fileHandler) Copy(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.Copy(req.Source, req.Destination); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *fileHandler) Move(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.Rename(req.Source, req.Destination); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 4: Search handler**

File: `backend/internal/api/search.go`
```go
package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/youruser/hermes-filebrowser/internal/fs"
)

type searchHandler struct {
	svc *fs.Service
}

func NewSearchHandler(svc *fs.Service) *searchHandler {
	return &searchHandler{svc: svc}
}

func (h *searchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	basePath := r.URL.Query().Get("path")
	if q == "" {
		http.Error(w, `{"error":"missing query"}`, http.StatusBadRequest)
		return
	}
	entries, err := h.svc.List(basePath)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	var results []fs.FileInfo
	q = strings.ToLower(q)
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), q) {
			results = append(results, e)
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": results})
}
```

- [ ] **Step 5: Routes**

File: `backend/internal/api/routes.go`
```go
package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/youruser/hermes-filebrowser/internal/config"
	"github.com/youruser/hermes-filebrowser/internal/db"
	"github.com/youruser/hermes-filebrowser/internal/fs"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func NewRouter(database *db.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(LoggingMiddleware)

	authMw := auth.SessionMiddleware(database)

	// Auth routes (no auth required)
	ah := NewAuthHandler(database, cfg)
	r.Post("/api/login", ah.Login)

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(CSRFMiddleware)
		r.Post("/api/logout", ah.Logout)
		r.Get("/api/me", ah.Me)

		fh := NewFileHandler(fs.NewService(cfg.Root))
		r.Get("/api/files", fh.List)
		r.Get("/api/files/raw", fh.Read)
		r.Get("/api/files/thumbnail", fh.Thumbnail)
		r.Post("/api/files/upload", fh.Upload)
		r.Post("/api/files/dir", fh.CreateDir)
		r.Post("/api/files/file", fh.CreateFile)
		r.Put("/api/files/rename", fh.Rename)
		r.Delete("/api/files", fh.Delete)
		r.Post("/api/files/copy", fh.Copy)
		r.Post("/api/files/move", fh.Move)

		sh := NewSearchHandler(fs.NewService(cfg.Root))
		r.Get("/api/search", sh.Search)
	})

	// Serve embedded frontend
	r.Group(func(r chi.Router) {
		subFS, _ := fs.Sub(frontendFS, "frontend/dist")
		fileServer := http.FileServer(http.FS(subFS))
		r.Handle("/*", fileServer)
	})

	return r
}
```

Note: The `frontend/dist` embed path requires the frontend to be built before `go build`. The Dockerfile handles this in multi-stage. For local dev, the frontend is served by Vite dev server.

- [ ] **Step 6: Verify build**

```bash
cd backend && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/api/
git commit -m "feat: add API handlers and chi router"
```

---
### Task 6: Wire Everything Together in main.go

**Files:**
- Modify: `backend/cmd/hermes/main.go`

- [ ] **Step 1: Replace main.go with full server wiring**

File: `backend/cmd/hermes/main.go`
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/youruser/hermes-filebrowser/internal/api"
	"github.com/youruser/hermes-filebrowser/internal/config"
	"github.com/youruser/hermes-filebrowser/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := database.EnsureAdmin(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	router := api.NewRouter(database, cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Hermes Filebrowser listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
```

- [ ] **Step 2: Build and verify**

```bash
cd backend && go build ./cmd/hermes
```

- [ ] **Step 3: Commit**

```bash
git add backend/cmd/hermes/main.go
git commit -m "feat: wire up server with config, DB, and routes"
```

---
### Task 7: Frontend Scaffolding

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tailwind.config.ts`
- Create: `frontend/postcss.config.js`
- Create: `frontend/tsconfig.json`
- Create: `frontend/tsconfig.node.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/index.css`
- Create: `frontend/src/vite-env.d.ts`

- [ ] **Step 1: Scaffold project with Vite**

```bash
mkdir frontend && cd frontend
npm create vite@latest . -- --template react-ts
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install react-router-dom react-icons react-pdf
```

- [ ] **Step 2: Configure Vite with Tailwind and API proxy**

File: `frontend/vite.config.ts`
```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false
  }
})
```

- [ ] **Step 3: Set up CSS**

File: `frontend/src/index.css`
```css
@import "tailwindcss";

:root {
  --color-bg: #0f0f0f;
  --color-surface: #1a1a1a;
  --color-border: #2a2a2a;
  --color-text: #e0e0e0;
  --color-text-muted: #888;
  --color-accent: #6c5ce7;
  --color-accent-hover: #5a4bd1;
  --color-danger: #e74c3c;
  --color-success: #2ecc71;
}

body {
  background-color: var(--color-bg);
  color: var(--color-text);
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}
```

- [ ] **Step 4: Create App.tsx with router**

File: `frontend/src/App.tsx`
```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import LoginPage from './pages/LoginPage'
import BrowserPage from './pages/BrowserPage'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<BrowserPage />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
```

- [ ] **Step 5: Verify frontend builds**

```bash
cd frontend && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/
git commit -m "feat: scaffold React frontend with Vite + Tailwind"
```

---
### Task 8: API Client + Auth Context

**Files:**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/context/AuthContext.tsx`

**Interfaces:**
- Consumes: nothing yet
- Produces: `apiClient` with methods: `get/put/post/delete(url, body?)` (auto-attaches auth headers, CSRF), `AuthContext` with: `user, token, login(username, password), logout(), isAuthenticated`

- [ ] **Step 1: API client**

File: `frontend/src/api/client.ts`
```ts
const BASE = ''

function getCSRFToken(): string {
  const m = document.cookie.match(/(?:^| )csrf_token=([^;]+)/)
  return m ? m[1] : ''
}

async function request(method: string, url: string, body?: unknown) {
  const headers: Record<string, string> = {
    'X-CSRF-Token': getCSRFToken(),
  }
  const token = localStorage.getItem('token')
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(`${BASE}${url}`, {
    method,
    headers,
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  get: (url: string) => request('GET', url),
  post: (url: string, body?: unknown) => request('POST', url, body),
  put: (url: string, body?: unknown) => request('PUT', url, body),
  delete: (url: string) => request('DELETE', url),
}
```

- [ ] **Step 2: Auth context**

File: `frontend/src/context/AuthContext.tsx`
```tsx
import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { api } from '../api/client'

interface User {
  id: number
  username: string
  readOnly: boolean
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType>(null!)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))

  useEffect(() => {
    if (token) {
      api.get('/api/me').then(res => setUser(res.user)).catch(() => logout())
    }
  }, [token])

  async function login(username: string, password: string) {
    const res = await api.post('/api/login', { username, password })
    localStorage.setItem('token', res.token)
    setToken(res.token)
    setUser(res.user)
  }

  function logout() {
    api.post('/api/logout').catch(() => {})
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, token, login, logout, isAuthenticated: !!user }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
```

- [ ] **Step 3: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/ frontend/src/context/
git commit -m "feat: add API client and auth context"
```

---
### Task 9: Login Page

**Files:**
- Create: `frontend/src/pages/LoginPage.tsx`

- [ ] **Step 1: Login page component**

File: `frontend/src/pages/LoginPage.tsx`
```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { login } = useAuth()
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await login(username, password)
      navigate('/')
    } catch (err: any) {
      setError(err.message)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg)]">
      <form onSubmit={handleSubmit} className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-8 w-full max-w-sm space-y-4">
        <h1 className="text-xl font-semibold text-center">Hermes Filebrowser</h1>
        {error && <p className="text-[var(--color-danger)] text-sm text-center">{error}</p>}
        <input className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]" placeholder="Username" value={username} onChange={e => setUsername(e.target.value)} />
        <input className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-[var(--color-accent)]" type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} />
        <button className="w-full bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white rounded-lg px-3 py-2 text-sm font-medium transition-colors" type="submit">Sign in</button>
      </form>
    </div>
  )
}
```

- [ ] **Step 2: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/LoginPage.tsx
git commit -m "feat: add login page"
```

---
### Task 10: Browser Page Layout

**Files:**
- Create: `frontend/src/pages/BrowserPage.tsx`
- Create: `frontend/src/components/Sidebar.tsx`
- Create: `frontend/src/components/Breadcrumb.tsx`
- Create: `frontend/src/components/FileList.tsx`
- Create: `frontend/src/components/FileRow.tsx`
- Create: `frontend/src/components/FileIcon.tsx`

**Interfaces:**
- Consumes: `useAuth()`, `api.get('/api/files?path=...')`
- Produces: folder tree, breadcrumb navigation, sortable file list with icons

- [ ] **Step 1: FileIcon component**

File: `frontend/src/components/FileIcon.tsx`
```tsx
import { FiFolder, FiFileText, FiImage, FiCode, FiArchive, FiFile, FiVideo, FiMusic } from 'react-icons/fi'

const iconMap: Record<string, any> = {
  dir: FiFolder,
  txt: FiFileText, md: FiFileText, json: FiCode, yml: FiCode, yaml: FiCode, xml: FiCode,
  js: FiCode, ts: FiCode, tsx: FiCode, jsx: FiCode, css: FiCode, html: FiCode, go: FiCode, py: FiCode,
  jpg: FiImage, jpeg: FiImage, png: FiImage, gif: FiImage, webp: FiImage, svg: FiImage,
  zip: FiArchive, tar: FiArchive, gz: FiArchive, rar: FiArchive, '7z': FiArchive,
  mp4: FiVideo, avi: FiVideo, mov: FiVideo, mkv: FiVideo,
  mp3: FiMusic, wav: FiMusic, flac: FiMusic, ogg: FiMusic,
}

export function FileIcon({ name, isDir, className = '' }: { name: string; isDir: boolean; className?: string }) {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const Icon = isDir ? FiFolder : iconMap[ext] || FiFile
  return <Icon className={`${className} ${isDir ? 'text-amber-400' : 'text-[var(--color-text-muted)]'}`} />
}
```

- [ ] **Step 2: Breadcrumb**

File: `frontend/src/components/Breadcrumb.tsx`
```tsx
interface Props {
  path: string
  onNavigate: (path: string) => void
}

export default function Breadcrumb({ path, onNavigate }: Props) {
  const parts = path.split('/').filter(Boolean)
  return (
    <nav className="flex items-center gap-1 text-sm text-[var(--color-text-muted)] px-4 py-2">
      <button onClick={() => onNavigate('')} className="hover:text-[var(--color-text)] transition-colors">Root</button>
      {parts.map((part, i) => {
        const p = '/' + parts.slice(0, i + 1).join('/')
        return (
          <span key={p} className="flex items-center gap-1">
            <span className="text-[var(--color-border)]">/</span>
            <button onClick={() => onNavigate(p)} className="hover:text-[var(--color-text)] transition-colors">{part}</button>
          </span>
        )
      })}
    </nav>
  )
}
```

- [ ] **Step 3: FileRow + FileList**

File: `frontend/src/components/FileRow.tsx`
```tsx
import { FileIcon } from './FileIcon'

interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: string
}

interface Props {
  file: FileInfo
  onNavigate: (path: string) => void
  onSelect?: (path: string) => void
  selected?: boolean
}

export default function FileRow({ file, onNavigate, onSelect, selected }: Props) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-2 rounded-lg cursor-pointer transition-colors ${
        selected ? 'bg-[var(--color-accent)]/20' : 'hover:bg-[var(--color-surface)]'
      }`}
      onClick={() => file.isDir ? onNavigate(file.path) : onSelect?.(file.path)}
    >
      <FileIcon name={file.name} isDir={file.isDir} className="w-5 h-5 shrink-0" />
      <span className="flex-1 truncate text-sm">{file.name}</span>
      {!file.isDir && (
        <span className="text-xs text-[var(--color-text-muted)] w-20 text-right">
          {file.size > 1024 * 1024
            ? (file.size / 1024 / 1024).toFixed(1) + ' MB'
            : file.size > 1024
            ? (file.size / 1024).toFixed(1) + ' KB'
            : file.size + ' B'}
        </span>
      )}
      <span className="text-xs text-[var(--color-text-muted)] w-24 text-right">
        {new Date(file.modTime).toLocaleDateString()}
      </span>
    </div>
  )
}
```

File: `frontend/src/components/FileList.tsx`
```tsx
import { useState } from 'react'
import FileRow from './FileRow'

interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: string
}

interface Props {
  files: FileInfo[]
  onNavigate: (path: string) => void
  sort: { key: string; dir: 'asc' | 'desc' }
  onSort: (key: string) => void
}

export default function FileList({ files, onNavigate, sort, onSort }: Props) {
  const [selected, setSelected] = useState<string | null>(null)
  const sorted = [...files].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    const dir = sort.dir === 'asc' ? 1 : -1
    switch (sort.key) {
      case 'size': return (a.size - b.size) * dir
      case 'modTime': return (new Date(a.modTime).getTime() - new Date(b.modTime).getTime()) * dir
      default: return a.name.localeCompare(b.name) * dir
    }
  })

  const SortHeader = ({ label, field }: { label: string; field: string }) => (
    <button onClick={() => onSort(field)} className="text-xs text-[var(--color-text-muted)] uppercase tracking-wider hover:text-[var(--color-text)] transition-colors">
      {label} {sort.key === field ? (sort.dir === 'asc' ? '↑' : '↓') : ''}
    </button>
  )

  return (
    <div className="flex-1 overflow-auto">
      <div className="flex items-center gap-3 px-4 py-2 border-b border-[var(--color-border)]">
        <span className="w-5 shrink-0" />
        <SortHeader label="Name" field="name" />
        <span className="flex-1" />
        <SortHeader label="Size" field="size" />
        <div className="w-24" />
      </div>
      {sorted.map(f => (
        <FileRow key={f.path} file={f} onNavigate={onNavigate} onSelect={setSelected} selected={selected === f.path} />
      ))}
    </div>
  )
}
```

- [ ] **Step 4: Sidebar**

File: `frontend/src/components/Sidebar.tsx`
```tsx
import { FiFolder, FiPlus, FiFilePlus } from 'react-icons/fi'

interface Props {
  currentPath: string
  onNavigate: (path: string) => void
  onNewFolder: () => void
  onNewFile: () => void
}

export default function Sidebar({ currentPath, onNavigate, onNewFolder, onNewFile }: Props) {
  return (
    <aside className="w-60 border-r border-[var(--color-border)] flex flex-col bg-[var(--color-surface)]">
      <div className="p-3 border-b border-[var(--color-border)] flex gap-2">
        <button onClick={onNewFolder} className="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors">
          <FiPlus className="w-3.5 h-3.5" /> Folder
        </button>
        <button onClick={onNewFile} className="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors">
          <FiFilePlus className="w-3.5 h-3.5" /> File
        </button>
      </div>
      <div className="flex-1 overflow-auto p-2">
        {/* Folder tree would recursively render here */}
        <div className="text-xs text-[var(--color-text-muted)] p-2">Folders</div>
      </div>
    </aside>
  )
}
```

- [ ] **Step 5: BrowserPage (main layout)**

File: `frontend/src/pages/BrowserPage.tsx`
```tsx
import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'
import Sidebar from '../components/Sidebar'
import Breadcrumb from '../components/Breadcrumb'
import FileList from '../components/FileList'

export default function BrowserPage() {
  const { user, logout, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })

  useEffect(() => {
    if (!isAuthenticated) navigate('/login')
  }, [isAuthenticated])

  const fetchFiles = useCallback(async (p: string) => {
    try {
      const res = await api.get(`/api/files?path=${encodeURIComponent(p)}`)
      setFiles(res.data)
    } catch { setFiles([]) }
  }, [])

  useEffect(() => { fetchFiles(path) }, [path, fetchFiles])

  return (
    <div className="h-screen flex flex-col bg-[var(--color-bg)]">
      <header className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
        <h1 className="text-sm font-medium flex items-center gap-2">
          <span className="text-[var(--color-accent)]">●</span> Hermes
        </h1>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-[var(--color-text-muted)]">{user?.username}</span>
          {user?.readOnly && <span className="text-xs text-amber-400">read-only</span>}
          <button onClick={logout} className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors">Sign out</button>
        </div>
      </header>
      <div className="flex flex-1 overflow-hidden">
        <Sidebar currentPath={path} onNavigate={setPath} onNewFolder={() => {}} onNewFile={() => {}} />
        <div className="flex-1 flex flex-col">
          <Breadcrumb path={path} onNavigate={setPath} />
          <FileList files={files} onNavigate={setPath} sort={sort} onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))} />
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 6: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/BrowserPage.tsx frontend/src/components/FileIcon.tsx frontend/src/components/FileRow.tsx frontend/src/components/FileList.tsx frontend/src/components/Breadcrumb.tsx frontend/src/components/Sidebar.tsx
git commit -m "feat: add browser page layout with sidebar, breadcrumb, file list"
```

---
### Task 11: File Operations UI (Upload, Rename, Delete, Copy, Move, New Folder/File)

**Files:**
- Create: `frontend/src/components/UploadProgress.tsx`
- Modify: `frontend/src/pages/BrowserPage.tsx` — add toolbar with action buttons, upload area, rename/move/copy dialogs
- Modify: `frontend/src/components/Sidebar.tsx` — wire up New Folder / New File

- [ ] **Step 1: Add file operation toolbar to BrowserPage**

Add action buttons between breadcrumb and file list: Upload, New Folder, New File, Rename, Delete, Copy, Move. Show them conditionally based on file selection. Include modals/prompts for rename, copy, move, new folder, new file.

File: `frontend/src/components/Toolbar.tsx` (if extracted) but for simplicity add directly in BrowserPage.

Add state to BrowserPage:
```tsx
const [selectedFile, setSelectedFile] = useState<string | null>(null)
const [showNewFolder, setShowNewFolder] = useState(false)
const [showNewFile, setShowNewFile] = useState(false)
const [showRename, setShowRename] = useState(false)
const [showCopy, setShowCopy] = useState(false)
const [showMove, setShowMove] = useState(false)
```

- [ ] **Step 2: Upload handling with DropZone**

Create `frontend/src/components/DropZone.tsx` that wraps the file list area. On drag over/files dropped, uploads via `api.post('/api/files/upload?path=...', formData)`.

```tsx
import { useState, DragEvent } from 'react'
import { api } from '../api/client'

interface Props {
  path: string
  onUploadComplete: () => void
  children: React.ReactNode
}

export default function DropZone({ path, onUploadComplete, children }: Props) {
  const [dragging, setDragging] = useState(false)

  async function handleDrop(e: DragEvent) {
    e.preventDefault()
    setDragging(false)
    const files = Array.from(e.dataTransfer.files)
    for (const file of files) {
      const fd = new FormData()
      fd.append('file', file)
      await api.post(`/api/files/upload?path=${encodeURIComponent(path)}`, fd)
    }
    onUploadComplete()
  }

  return (
    <div
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
      className={`flex-1 flex flex-col ${dragging ? 'ring-2 ring-[var(--color-accent)] bg-[var(--color-accent)]/5' : ''}`}
    >
      {children}
    </div>
  )
}
```

- [ ] **Step 3: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/DropZone.tsx
git commit -m "feat: add file operations UI and drag-drop upload"
```

---
### Task 12: Preview Pane + Search

**Files:**
- Create: `frontend/src/components/PreviewPane.tsx`
- Create: `frontend/src/components/SearchBar.tsx`
- Modify: `frontend/src/pages/BrowserPage.tsx` — add search bar, bottom preview pane

- [ ] **Step 1: PreviewPane**

File: `frontend/src/components/PreviewPane.tsx`
```tsx
import { useState, useEffect } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import { api } from '../api/client'
import { FileIcon } from './FileIcon'

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

interface Props {
  filePath: string | null
}

export default function PreviewPane({ filePath }: Props) {
  const [data, setData] = useState<string | null>(null)
  const ext = filePath?.split('.').pop()?.toLowerCase()

  useEffect(() => {
    if (!filePath) { setData(null); return }
    if (['jpg','jpeg','png','gif','webp','svg'].includes(ext || '')) {
      setData(`/api/files/raw?path=${encodeURIComponent(filePath)}`)
    } else {
      api.get(`/api/files/raw?path=${encodeURIComponent(filePath)}`).then(d => setData(d)).catch(() => setData(null))
    }
  }, [filePath])

  if (!filePath) return null

  const fileName = filePath.split('/').pop() || ''

  return (
    <div className="border-t border-[var(--color-border)] bg-[var(--color-surface)] p-4 max-h-64 overflow-auto">
      <div className="flex items-center gap-2 mb-3">
        <FileIcon name={fileName} isDir={false} className="w-4 h-4" />
        <span className="text-sm font-medium">{fileName}</span>
      </div>
      {['jpg','jpeg','png','gif','webp','svg'].includes(ext || '') && data && (
        <img src={data} alt={fileName} className="max-h-48 rounded" />
      )}
      {ext === 'pdf' && (
        <Document file={`/api/files/raw?path=${encodeURIComponent(filePath)}`}>
          <Page pageNumber={1} width={400} />
        </Document>
      )}
      {['txt','md','json','xml','yml','yaml','js','ts','jsx','tsx','css','html','go','py','sh','env','cfg','ini','log'].includes(ext || '') && data && (
        <pre className="text-xs leading-relaxed overflow-x-auto whitespace-pre-wrap">{data}</pre>
      )}
    </div>
  )
}
```

- [ ] **Step 2: SearchBar**

File: `frontend/src/components/SearchBar.tsx`
```tsx
import { useState, useEffect, useRef } from 'react'
import { FiSearch } from 'react-icons/fi'
import { api } from '../api/client'

interface Props {
  path: string
  onResults: (files: any[]) => void
  onClear: () => void
}

export default function SearchBar({ path, onResults, onClear }: Props) {
  const [query, setQuery] = useState('')
  const timer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    if (!query.trim()) { onClear(); return }
    clearTimeout(timer.current)
    timer.current = setTimeout(async () => {
      const res = await api.get(`/api/search?q=${encodeURIComponent(query)}&path=${encodeURIComponent(path)}`)
      onResults(res.data)
    }, 300)
    return () => clearTimeout(timer.current)
  }, [query, path])

  return (
    <div className="relative px-4 py-2">
      <FiSearch className="absolute left-6 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] w-4 h-4" />
      <input
        className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg pl-9 pr-3 py-1.5 text-sm focus:outline-none focus:border-[var(--color-accent)]"
        placeholder="Search files..."
        value={query}
        onChange={e => setQuery(e.target.value)}
      />
    </div>
  )
}
```

- [ ] **Step 3: Wire into BrowserPage — add search bar (above file list), preview pane (below file list)**

- [ ] **Step 4: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/PreviewPane.tsx frontend/src/components/SearchBar.tsx
git commit -m "feat: add preview pane and search bar"
```

---
### Task 13: Docker Build + Docker Compose + GitHub Actions + README

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.github/workflows/docker-publish.yml`
- Create: `README.md`

- [ ] **Step 1: Dockerfile (multi-stage)**

File: `Dockerfile`
```dockerfile
# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/ ./
COPY --from=frontend /app/frontend/dist ./internal/api/frontend/dist
RUN CGO_ENABLED=0 go build -o /hermes ./cmd/hermes

# Stage 3: Distroless runtime
FROM gcr.io/distroless/base-debian12
COPY --from=backend /hermes /hermes
EXPOSE 8080
ENTRYPOINT ["/hermes"]
```

- [ ] **Step 2: docker-compose.yml**

File: `docker-compose.yml`
```yaml
services:
  hermes:
    build: .
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

- [ ] **Step 3: GitHub Actions workflow**

File: `.github/workflows/docker-publish.yml`
```yaml
name: Build and Push to Docker Hub

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ secrets.DOCKER_USERNAME }}/hermes-filebrowser:latest,${{ secrets.DOCKER_USERNAME }}/hermes-filebrowser:${{ github.sha }}
          platforms: linux/amd64,linux/arm64
```

- [ ] **Step 4: README.md**

Basic build/run instructions, env vars table, reverse proxy notes (X-Forwarded-For), Portainer deploy steps.

- [ ] **Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml .github/workflows/docker-publish.yml README.md
git commit -m "feat: add Docker build, compose, CI/CD, and README"
```

---
### Task 14: Final Polish — Full Stack Integration Test

- [ ] **Step 1: Build the full Docker image locally**

```bash
docker build -t hermes-filebrowser:latest .
```

- [ ] **Step 2: Run and test**

```bash
docker run --rm -p 8080:8080 -e FB_USERNAME=test -e FB_PASSWORD=test hermes-filebrowser:latest
```

Verify: `curl -X POST localhost:8080/api/login -d '{"username":"test","password":"test"}'` returns token.

- [ ] **Step 3: Verify frontend is served**

Open `http://localhost:8080` — should see login page.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: final integration fixes"
```
