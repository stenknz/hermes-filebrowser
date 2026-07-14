package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/config"
	"github.com/stenknz/hermes-filebrowser/internal/db"
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
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, expiresAt := auth.NewSessionToken()
	if err := h.db.CreateSession(user.ID, token, expiresAt.Format(time.RFC3339)); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
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
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user})
}
