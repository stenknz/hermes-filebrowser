package api

import (
	"embed"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/stenknz/hermes-filebrowser/internal/auth"
	"github.com/stenknz/hermes-filebrowser/internal/config"
	"github.com/stenknz/hermes-filebrowser/internal/db"
	"github.com/stenknz/hermes-filebrowser/internal/fs"

	iofs "io/fs"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func NewRouter(database *db.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware)
	r.Use(SecurityHeadersMiddleware)

	authMw := auth.SessionMiddleware(database)

	// Auth routes (no auth required)
	ah := NewAuthHandler(database, cfg)
	r.Post("/api/login", ah.Login)

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(requireAuth)
		r.Use(JSONContentType)
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

		rh := NewResourcesHandler(fileSvc)
		r.Post("/api/resources", rh.HandlePost)
		r.Patch("/api/resources", rh.HandlePatch)
		r.Delete("/api/resources", rh.HandleDelete)
		r.Post("/api/upload", rh.HandleUpload)
		r.Get("/api/raw/*", rh.HandleRaw)
		r.Get("/api/raw/{path}", rh.HandleRaw)

		sh := NewSearchHandler(fileSvc)
		r.Get("/api/search", sh.Search)

		uh := NewUsersHandler(database)
		r.Get("/api/users", uh.ListUsers)
		r.Post("/api/users", uh.CreateUser)
		r.Post("/api/users/delete", uh.DeleteUser)
		r.Get("/api/tokens", uh.ListApiTokens)
		r.Post("/api/tokens", uh.CreateApiToken)
		r.Post("/api/tokens/delete", uh.DeleteApiToken)
	})

	// Serve embedded frontend (SPA fallback)
	r.Group(func(r chi.Router) {
		subFS, err := iofs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			log.Fatalf("failed to get frontend sub FS: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		// Read and cache index.html for SPA fallback
		indexData := func() []byte {
			f, err := subFS.Open("index.html")
			if err != nil {
				return nil
			}
			defer f.Close()
			info, _ := f.Stat()
			data := make([]byte, info.Size())
			f.Read(data)
			return data
		}()
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path != "/" {
				_, err := iofs.Stat(subFS, path[1:])
				if err == nil {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexData)
		}))
	})

	return r
}
