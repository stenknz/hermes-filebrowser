package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stenknz/hermes-filebrowser/internal/api"
	"github.com/stenknz/hermes-filebrowser/internal/config"
	"github.com/stenknz/hermes-filebrowser/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := database.EnsureAdmin(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	router := api.NewRouter(database, cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("FileBrowser listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
