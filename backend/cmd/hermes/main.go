package main

import (
	"log"

	"github.com/youruser/hermes-filebrowser/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("Hermes Filebrowser starting on port %d", cfg.Port)
	_ = cfg
}
