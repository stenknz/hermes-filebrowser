package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stenknz/hermes-filebrowser/internal/db"
)

type contextKey string

const userKey contextKey = "user"
const apiTokenKey contextKey = "apiToken"

func IsApiTokenAuth(r *http.Request) bool {
	v, _ := r.Context().Value(apiTokenKey).(bool)
	return v
}

func NewSessionToken() (string, time.Time) {
	return uuid.New().String(), time.Now().Add(24 * time.Hour)
}

func NewApiToken() string {
	return "fb_" + uuid.New().String()
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
				// Try session token first
				session, err := database.GetSessionByToken(token)
				if err == nil {
					expiresAt, parseErr := time.Parse(time.RFC3339, session.ExpiresAt)
					if parseErr == nil && time.Now().Before(expiresAt) {
						user, _ := database.GetUserByID(session.UserID)
						if user != nil {
							ctx := context.WithValue(r.Context(), userKey, user)
							r = r.WithContext(ctx)
							next.ServeHTTP(w, r)
							return
						}
					}
				}
				// Try API token
				apiToken, err := database.GetApiTokenByToken(token)
				if err == nil {
					user, _ := database.GetUserByID(apiToken.UserID)
					if user != nil {
						ctx := context.WithValue(r.Context(), userKey, user)
						ctx = context.WithValue(ctx, apiTokenKey, true)
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
