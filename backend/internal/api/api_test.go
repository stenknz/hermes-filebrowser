package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/config"
	"github.com/stenknz/hermes-filebrowser/internal/db"
	"github.com/stenknz/hermes-filebrowser/internal/fs"
	"golang.org/x/crypto/bcrypt"
)

// --- Helpers ---

func setupTestEnv(t *testing.T) (*db.DB, *config.Config, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	d, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("db.New() error = %v", err)
	}
	root := t.TempDir()
	cfg := &config.Config{
		Port:         8080,
		Root:         root,
		Username:     "admin",
		Password:     "admin",
		DatabasePath: dbPath,
	}
	return d, cfg, root
}

func mustCreateUser(t *testing.T, d *db.DB, username, password string, role db.Role) *db.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	user, err := d.CreateUser(username, string(hash), role)
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	return user
}

func mustCreateSession(t *testing.T, d *db.DB, userID int64) (string, time.Time) {
	t.Helper()
	token, expiresAt := auth.NewSessionToken()
	if err := d.CreateSession(userID, token, expiresAt.Format(time.RFC3339)); err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}
	return token, expiresAt
}

func withSession(d *db.DB, h http.HandlerFunc) http.Handler {
	return auth.SessionMiddleware(d)(h)
}

func authenticatedRequest(t *testing.T, method, target string, body []byte, token string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withCSRF(req *http.Request, csrfToken string) *http.Request {
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	return req
}

// --- Middleware Tests ---

func TestCSRFMiddleware_SafeMethodsPass(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: got %d, want %d", method, rec.Code, http.StatusOK)
		}
	}
}

func TestCSRFMiddleware_MissingCookie(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCSRFMiddleware_Mismatch(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "abc"})
	req.Header.Set("X-CSRF-Token", "def")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCSRFMiddleware_Valid(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	req.Header.Set("X-CSRF-Token", "valid-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGenerateCSRFToken_NotEmpty(t *testing.T) {
	token := GenerateCSRFToken()
	if token == "" {
		t.Error("token should not be empty")
	}
	if len(token) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(token))
	}
}

// --- Auth Handler Tests ---

func TestLogin_Valid(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()
	mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)

	h := NewAuthHandler(d, cfg)
	body := `{"username":"admin","password":"admin123"}`
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}

	cookies := rec.Result().Cookies()
	var hasSession, hasCSRF bool
	for _, c := range cookies {
		if c.Name == "session" && c.Value != "" {
			hasSession = true
		}
		if c.Name == "csrf_token" && c.Value != "" {
			hasCSRF = true
		}
	}
	if !hasSession {
		t.Error("expected session cookie")
	}
	if !hasCSRF {
		t.Error("expected csrf_token cookie")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()
	mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)

	h := NewAuthHandler(d, cfg)
	body := `{"username":"admin","password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()

	h := NewAuthHandler(d, cfg)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(`not json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLogout(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	h := NewAuthHandler(d, cfg)
	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Session should be deleted
	_, err := d.GetSessionByToken(token)
	if err == nil {
		t.Error("session should have been deleted")
	}

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session" || c.Name == "csrf_token" {
			if c.MaxAge != -1 {
				t.Errorf("cookie %s should have MaxAge=-1", c.Name)
			}
		}
	}
}

func TestMe_Authenticated(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	h := NewAuthHandler(d, cfg)
	handler := auth.SessionMiddleware(d)(http.HandlerFunc(h.Me))
	req := authenticatedRequest(t, "GET", "/api/me", nil, token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	userMap, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user object in response")
	}
	if userMap["username"] != "admin" {
		t.Errorf("username = %v, want admin", userMap["username"])
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()

	h := NewAuthHandler(d, cfg)
	req := httptest.NewRequest("GET", "/api/me", nil)
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// --- File Handler Tests ---

func TestFileList(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(root, "subdir"), 0755)

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/files?path=.", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	fh.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("expected data array")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 entries, got %d", len(data))
	}
}

func TestFileRead(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "readme.txt"), []byte("file content"), 0644)

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/files/raw?path=readme.txt", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	fh.Read(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != "file content" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "file content")
	}
	if rec.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestFileRead_NotFound(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/files/raw?path=nonexistent.txt", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	fh.Read(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestFileThumbnail(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	imgPath := filepath.Join(root, "test.png")
	f, _ := os.Create(imgPath)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 400, 300)))
	f.Close()

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/files/thumbnail?path=test.png", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	fh.Thumbnail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Error("thumbnail body should not be empty")
	}
}

func TestFileUpload(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", "uploaded.txt")
	part.Write([]byte("upload content"))
	w.Close()

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/upload?path=.", buf.Bytes(), token), csrftoken)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	handler := auth.SessionMiddleware(d)(http.HandlerFunc(fh.Upload))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	data, err := os.ReadFile(filepath.Join(root, "uploaded.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "upload content" {
		t.Errorf("file content = %q, want %q", string(data), "upload content")
	}
}

func TestFileUpload_ReadOnly(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "readonly", "pass", db.RoleViewer)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/upload?path=.", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	handler := auth.SessionMiddleware(d)(http.HandlerFunc(fh.Upload))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCreateDir(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/dir?path=newdir", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.CreateDir).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusCreated)
	}
	if _, err := os.Stat(filepath.Join(root, "newdir")); os.IsNotExist(err) {
		t.Error("directory should exist")
	}
}

func TestCreateFile(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	body := `{"path":"newfile.txt","content":"hello world"}`
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/file", []byte(body), token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.CreateFile).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	data, _ := os.ReadFile(filepath.Join(root, "newfile.txt"))
	if string(data) != "hello world" {
		t.Errorf("content = %q, want %q", string(data), "hello world")
	}
}

func TestRename(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "old.txt"), []byte("data"), 0644)

	body := `{"oldPath":"old.txt","newPath":"new.txt"}`
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "PUT", "/api/files/rename", []byte(body), token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.Rename).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "old.txt")); !os.IsNotExist(err) {
		t.Error("old file should not exist")
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); os.IsNotExist(err) {
		t.Error("new file should exist")
	}
}

func TestDelete(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "del.txt"), []byte("data"), 0644)

	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "DELETE", "/api/files?path=del.txt", nil, token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.Delete).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if _, err := os.Stat(filepath.Join(root, "del.txt")); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}
}

func TestCopy(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "src.txt"), []byte("copy data"), 0644)

	body := `{"source":"src.txt","destination":"dst.txt"}`
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/copy", []byte(body), token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.Copy).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	srcData, _ := os.ReadFile(filepath.Join(root, "src.txt"))
	dstData, _ := os.ReadFile(filepath.Join(root, "dst.txt"))
	if string(srcData) != string(dstData) {
		t.Error("copy content mismatch")
	}
}

func TestMove(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "source.txt"), []byte("move data"), 0644)

	body := `{"source":"source.txt","destination":"dest.txt"}`
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "POST", "/api/files/move", []byte(body), token), csrftoken)
	rec := httptest.NewRecorder()

	fh := NewFileHandler(svc)
	withSession(d, fh.Move).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "source.txt")); !os.IsNotExist(err) {
		t.Error("source should not exist after move")
	}
	data, _ := os.ReadFile(filepath.Join(root, "dest.txt"))
	if string(data) != "move data" {
		t.Errorf("content = %q", string(data))
	}
}

func TestReadOnlyBlocksWrite(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "readonly", "pass", db.RoleViewer)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	fh := NewFileHandler(svc)
	csrftoken := GenerateCSRFToken()

	t.Run("CreateDir", func(t *testing.T) {
		req := withCSRF(authenticatedRequest(t, "POST", "/api/files/dir?path=x", nil, token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.CreateDir).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("CreateFile", func(t *testing.T) {
		body := `{"path":"x","content":"x"}`
		req := withCSRF(authenticatedRequest(t, "POST", "/api/files/file", []byte(body), token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.CreateFile).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("Rename", func(t *testing.T) {
		body := `{"oldPath":"a","newPath":"b"}`
		req := withCSRF(authenticatedRequest(t, "PUT", "/api/files/rename", []byte(body), token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.Rename).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		req := withCSRF(authenticatedRequest(t, "DELETE", "/api/files?path=x", nil, token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.Delete).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("Copy", func(t *testing.T) {
		body := `{"source":"a","destination":"b"}`
		req := withCSRF(authenticatedRequest(t, "POST", "/api/files/copy", []byte(body), token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.Copy).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("Move", func(t *testing.T) {
		body := `{"source":"a","destination":"b"}`
		req := withCSRF(authenticatedRequest(t, "POST", "/api/files/move", []byte(body), token), csrftoken)
		rec := httptest.NewRecorder()
		withSession(d, fh.Move).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

// --- Search Handler Tests ---

func TestSearch(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "alpha.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(root, "beta.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(root, "gamma.txt"), []byte("c"), 0644)

	sh := NewSearchHandler(svc)
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/search?q=alpha&path=.", nil, token), csrftoken)
	rec := httptest.NewRecorder()
	sh.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("expected data array")
	}
	if len(data) != 1 {
		t.Errorf("expected 1 result, got %d", len(data))
	}
}

func TestSearch_MissingQuery(t *testing.T) {
	d, _, root := setupTestEnv(t)
	defer d.Close()
	user := mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)
	token, _ := mustCreateSession(t, d, user.ID)

	svc, err := fs.NewService(root)
	if err != nil {
		t.Fatal(err)
	}

	sh := NewSearchHandler(svc)
	csrftoken := GenerateCSRFToken()
	req := withCSRF(authenticatedRequest(t, "GET", "/api/search?path=.", nil, token), csrftoken)
	rec := httptest.NewRecorder()
	sh.Search(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- Router Tests ---

func TestNewRouter(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()
	mustCreateUser(t, d, "admin", "admin123", db.RoleEditor)

	router := NewRouter(d, cfg)
	if router == nil {
		t.Fatal("router should not be nil")
	}

	// Test login endpoint (no auth required)
	body := `{"username":"admin","password":"admin123"}`
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login via router got %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRouter_ProtectedEndpointWithoutAuth(t *testing.T) {
	d, cfg, _ := setupTestEnv(t)
	defer d.Close()

	router := NewRouter(d, cfg)
	req := httptest.NewRequest("GET", "/api/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Should be 401 since no auth, but might be 403 due to CSRF
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Errorf("expected 401 or 403, got %d: %s", rec.Code, rec.Body.String())
	}
}
