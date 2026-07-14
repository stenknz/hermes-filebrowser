package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/stenknz/hermes-filebrowser/internal/fs"
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
		jsonError(w, "missing query", http.StatusBadRequest)
		return
	}
	entries, err := h.svc.List(basePath)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
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
