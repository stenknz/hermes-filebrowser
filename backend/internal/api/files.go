package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/fs"
)

type fileHandler struct {
	svc *fs.Service
}

func NewFileHandler(svc *fs.Service) *fileHandler {
	return &fileHandler{svc: svc}
}

func (h *fileHandler) forUser(r *http.Request) *fs.Service {
	user := auth.GetUser(r)
	if user != nil && user.HomePath != "" {
		return h.svc.Scoped(user.HomePath)
	}
	return h.svc
}

func (h *fileHandler) List(w http.ResponseWriter, r *http.Request) {
	dirPath := r.URL.Query().Get("path")
	entries, err := h.forUser(r).List(dirPath)
	if err != nil {
		// If path with trailing slash fails, try without
		if len(dirPath) > 0 && dirPath[len(dirPath)-1] == '/' {
			dirPath = dirPath[:len(dirPath)-1]
			entries, err = h.forUser(r).List(dirPath)
		}
	}
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
	data, err := h.forUser(r).Read(filePath)
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
	data, err := h.forUser(r).Thumbnail(filePath)
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
	if err := h.forUser(r).Write(filepath.Join(dirPath, header.Filename), data); err != nil {
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
	if dirPath == "" {
		var body struct{ Path string `json:"path"` }
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			dirPath = body.Path
		}
	}
	if err := h.forUser(r).Mkdir(dirPath); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *fileHandler) MkdirPost(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.forUser(r).Mkdir(req.Path); err != nil {
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
		Path     string `json:"path"`
		Content  string `json:"content"`
		Type     string `json:"type"`
		Base64   bool   `json:"base64"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	// If type is "dir" or path ends with /, create a directory
	if req.Type == "dir" || strings.HasSuffix(req.Path, "/") {
		if err := h.forUser(r).Mkdir(req.Path); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
	// Decode base64 content if requested (accept both "base64": true and "encoding": "base64")
	data := []byte(req.Content)
	if req.Base64 || strings.ToLower(req.Encoding) == "base64" {
		var err error
		data, err = base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			jsonError(w, "invalid base64 content", http.StatusBadRequest)
			return
		}
	}
	if err := h.forUser(r).Write(req.Path, data); err != nil {
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
		OldPath     string `json:"oldPath"`
		NewPath     string `json:"newPath"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		From        string `json:"from"`
		To          string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	oldPath := req.OldPath
	if oldPath == "" {
		oldPath = req.Source
	}
	if oldPath == "" {
		oldPath = req.From
	}
	newPath := req.NewPath
	if newPath == "" {
		newPath = req.Destination
	}
	if newPath == "" {
		newPath = req.To
	}
	if oldPath == "" || newPath == "" {
		jsonError(w, "oldPath and newPath (or source/destination) required", http.StatusBadRequest)
		return
	}
	if err := h.forUser(r).Rename(oldPath, newPath); err != nil {
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
	if filePath == "" {
		var body struct{ Path string `json:"path"` }
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			filePath = body.Path
		}
	}
	if filePath == "" {
		jsonError(w, "path required", http.StatusBadRequest)
		return
	}
	if err := h.forUser(r).Delete(filePath); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *fileHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	var body struct {
		Path  string   `json:"path"`
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Path != "" {
		if err := h.forUser(r).Delete(body.Path); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	for _, name := range body.Names {
		if err := h.forUser(r).Delete(name); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
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
	if err := h.forUser(r).Copy(req.Source, req.Destination); err != nil {
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
	if err := h.forUser(r).Rename(req.Source, req.Destination); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *fileHandler) Stat(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		jsonError(w, "path required", http.StatusBadRequest)
		return
	}
	fullPath, err := h.forUser(r).SafePath(filePath)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		// For root dir that doesn't exist yet, return empty dir info
		if filePath == "." || filePath == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":    "",
				"path":    "",
				"size":    0,
				"isDir":   true,
				"modTime": time.Now().Format(time.RFC3339),
				"mode":    "drwxr-xr-x",
			})
			return
		}
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    info.Name(),
		"path":    filePath,
		"size":    info.Size(),
		"isDir":   info.IsDir(),
		"modTime": info.ModTime().Format(time.RFC3339),
		"mode":    info.Mode().String(),
	})
}

func (h *fileHandler) UploadRaw(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly() {
		jsonError(w, "read-only user", http.StatusForbidden)
		return
	}
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		jsonError(w, "path required", http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	if err := h.forUser(r).Write(filePath, data); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
