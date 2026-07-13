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

**Important notes:**
- `fs.NewService` now returns `(*fs.Service, error)` — handle the error in routes.go
- `auth` package is `github.com/youruser/hermes-filebrowser/internal/auth` — add the import to routes.go
- The `//go:embed frontend/dist/*` directive is relative to the file location. The Dockerfile copies frontend build output to `backend/internal/api/frontend/dist/` so the embed finds it. For now, the embed is fine as-is — just create an empty `frontend/dist/.gitkeep` so the build works.

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
	"github.com/youruser/hermes-filebrowser/internal/config"
	"github.com/youruser/hermes-filebrowser/internal/db"
)

type authHandler struct {
	db  *db.DB
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
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/youruser/hermes-filebrowser/internal/auth"
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

		fileSvc, err := fs.NewService(cfg.Root)
		if err != nil {
			log.Fatalf("failed to create file service: %v", err)
		}
		fh := NewFileHandler(fileSvc)
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

		searchSvc, err := fs.NewService(cfg.Root)
		if err != nil {
			log.Fatalf("failed to create search service: %v", err)
		}
		sh := NewSearchHandler(searchSvc)
		r.Get("/api/search", sh.Search)
	})

	// Serve embedded frontend
	r.Group(func(r chi.Router) {
		subFS, err := fs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			log.Fatalf("failed to get frontend sub FS: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		r.Handle("/*", fileServer)
	})

	return r
}
```

- [ ] **Step 6: Verify build**

```bash
cd backend && go get github.com/go-chi/chi/v5 && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/api/
git commit -m "feat: add API handlers and chi router"
```
