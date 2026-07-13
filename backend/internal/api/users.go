package api

import (
	"encoding/json"
	"net/http"

	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type usersHandler struct {
	db *db.DB
}

func NewUsersHandler(database *db.DB) *usersHandler {
	return &usersHandler{db: database}
}

func (h *usersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil || user.Role != db.RoleAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	users, err := h.db.ListUsers()
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": users})
}

func (h *usersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil || user.Role != db.RoleAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	var req struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Role     db.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
		return
	}
	if req.Role != db.RoleAdmin && req.Role != db.RoleEditor && req.Role != db.RoleViewer {
		req.Role = db.RoleViewer
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	u, err := h.db.CreateUser(req.Username, string(hash), req.Role)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func (h *usersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil || user.Role != db.RoleAdmin {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if req.ID == user.ID {
		http.Error(w, `{"error":"cannot delete yourself"}`, http.StatusBadRequest)
		return
	}
	if err := h.db.DeleteUser(req.ID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *usersHandler) ListApiTokens(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	tokens, err := h.db.ListApiTokens(user.ID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": tokens})
}

func (h *usersHandler) CreateApiToken(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	token, err := h.db.CreateApiToken(user.ID, auth.NewApiToken(), req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (h *usersHandler) DeleteApiToken(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := h.db.DeleteApiToken(req.ID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
