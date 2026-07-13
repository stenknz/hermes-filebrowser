package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/fs"
)

type resourcesHandler struct {
	svc *fs.Service
}

func NewResourcesHandler(svc *fs.Service) *resourcesHandler {
	return &resourcesHandler{svc: svc}
}

func (h *resourcesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"cannot read body"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Action      string `json:"action"`
		Path        string `json:"path"`
		Content     string `json:"content"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		OldPath     string `json:"oldPath"`
		NewPath     string `json:"newPath"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	switch req.Action {
	case "mkdir":
		if user.ReadOnly {
			http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
			return
		}
		if err := h.svc.Mkdir(req.Path); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	case "write":
		if user.ReadOnly {
			http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
			return
		}
		if err := h.svc.Write(req.Path, []byte(req.Content)); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	case "copy":
		if user.ReadOnly {
			http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
			return
		}
		if err := h.svc.Copy(req.Source, req.Destination); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "move":
		if user.ReadOnly {
			http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
			return
		}
		if err := h.svc.Rename(req.Source, req.Destination); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusBadRequest)
	}
}

func (h *resourcesHandler) HandlePatch(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := h.svc.Rename(req.OldPath, req.NewPath); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *resourcesHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
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

func (h *resourcesHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user.ReadOnly {
		http.Error(w, `{"error":"read-only user"}`, http.StatusForbidden)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, `{"error":"invalid multipart form"}`, http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"missing file field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, `{"error":"failed to read file"}`, http.StatusInternalServerError)
		return
	}
	dirPath := r.URL.Query().Get("path")
	targetPath := filepath.Join(dirPath, header.Filename)
	if err := h.svc.Write(targetPath, data); err != nil {
		http.Error(w, `{"error":"write failed"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *resourcesHandler) HandleRaw(w http.ResponseWriter, r *http.Request) {
	filePath := chi.URLParam(r, "*")
	if filePath == "" {
		filePath = chi.URLParam(r, "path")
	}
	if filePath == "" {
		filePath = r.URL.Query().Get("path")
	}
	filePath = strings.TrimPrefix(filePath, "/")
	data, err := h.svc.Read(filePath)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Write(data)
}
