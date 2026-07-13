package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/youruser/hermes-filebrowser/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckPassword_Valid(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	if !CheckPassword(string(hash), "correct-password") {
		t.Error("CheckPassword returned false for correct password")
	}
}

func TestCheckPassword_Invalid(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	if CheckPassword(string(hash), "wrong-password") {
		t.Error("CheckPassword returned true for wrong password")
	}
}

func TestNewSessionToken_NotEmpty(t *testing.T) {
	token, expiresAt := NewSessionToken()
	if token == "" {
		t.Error("token should not be empty")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expiresAt should be in the future")
	}
}

func TestNewSessionToken_Unique(t *testing.T) {
	t1, _ := NewSessionToken()
	t2, _ := NewSessionToken()
	if t1 == t2 {
		t.Error("tokens should be unique")
	}
}

func TestGetUser_NilWhenNoUser(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	u := GetUser(r)
	if u != nil {
		t.Error("GetUser should return nil when no user in context")
	}
}

func TestSessionMiddleware_AuthorizationHeader(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	user, err := d.CreateUser("testuser", string(mustHash("pass")), false)
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	err = d.CreateSession(user.ID, "test-token-123", time.Now().Add(24*time.Hour).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser == nil {
		t.Fatal("expected user to be set from Authorization header")
	}
	if capturedUser.Username != "testuser" {
		t.Errorf("username = %q, want %q", capturedUser.Username, "testuser")
	}
}

func TestSessionMiddleware_Cookie(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	user, err := d.CreateUser("cookieuser", string(mustHash("pass")), false)
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	err = d.CreateSession(user.ID, "cookie-token-456", time.Now().Add(24*time.Hour).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "cookie-token-456"})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser == nil {
		t.Fatal("expected user to be set from cookie")
	}
	if capturedUser.Username != "cookieuser" {
		t.Errorf("username = %q, want %q", capturedUser.Username, "cookieuser")
	}
}

func TestSessionMiddleware_AuthorizationTakesPriority(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	user1, _ := d.CreateUser("beareruser", string(mustHash("pass")), false)
	user2, _ := d.CreateUser("cookieuser", string(mustHash("pass")), false)
	d.CreateSession(user1.ID, "bearer-token", time.Now().Add(24*time.Hour).Format(time.RFC3339))
	d.CreateSession(user2.ID, "cookie-token", time.Now().Add(24*time.Hour).Format(time.RFC3339))

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{Name: "session", Value: "cookie-token"})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser == nil || capturedUser.Username != "beareruser" {
		t.Errorf("expected bearer user, got %v", capturedUser)
	}
}

func TestSessionMiddleware_ExpiredToken(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	user, _ := d.CreateUser("expireduser", string(mustHash("pass")), false)
	d.CreateSession(user.ID, "expired-token", time.Now().Add(-1*time.Hour).Format(time.RFC3339))

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser != nil {
		t.Error("expected no user for expired token")
	}
}

func TestSessionMiddleware_InvalidToken(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer nonexistent-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser != nil {
		t.Error("expected no user for invalid token")
	}
}

func TestSessionMiddleware_NoToken(t *testing.T) {
	d := setupDB(t)
	defer d.Close()

	var capturedUser *db.User
	handler := SessionMiddleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedUser != nil {
		t.Error("expected no user when no token provided")
	}
}

func setupDB(t *testing.T) *db.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth_test.db")
	d, err := db.New(path)
	if err != nil {
		t.Fatalf("db.New() error = %v", err)
	}
	return d
}

func mustHash(password string) []byte {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return h
}


