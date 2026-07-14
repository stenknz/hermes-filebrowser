package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"

	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/fs"
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
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": entries})
}

func (h *fileHandler) Read(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		jsonError(w, "path required", http.StatusBadRequest)
		return
	}
	data, err := h.svc.Read(filePath)
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
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
		jsonError(w, "cannot generate thumbnail", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(data)
}

func (h *fileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	dirPath := r.URL.Query().Get("path")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "invalid multipart form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	if err := h.svc.Write(filepath.Join(dirPath, header.Filename), data); err != nil {
		jsonError(w, "write failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) CreateDir(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	dirPath := r.URL.Query().Get("path")
	if err := h.svc.Mkdir(dirPath); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Write(req.Path, []byte(req.Content)); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) Rename(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Rename(req.OldPath, req.NewPath); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *fileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	filePath := r.URL.Query().Get("path")
	if err := h.svc.Delete(filePath); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *fileHandler) Copy(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Copy(req.Source, req.Destination); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *fileHandler) Move(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Rename(req.Source, req.Destination); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
