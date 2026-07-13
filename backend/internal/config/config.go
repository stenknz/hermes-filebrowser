package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Root         string
	Username     string
	Password     string
	DatabasePath string
}

func Load() *Config {
	return &Config{
		Port:         getEnvInt("FB_PORT", 8080),
		Root:         getEnv("FB_ROOT", "/data"),
		Username:     getEnv("FB_USERNAME", "admin"),
		Password:     getEnv("FB_PASSWORD", "admin"),
		DatabasePath: getEnv("FB_DATABASE", "/data/filebrowser.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
