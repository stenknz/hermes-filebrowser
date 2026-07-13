### Task 1: Go Module + Config + Main Skeleton

**Files:**
- Create: `backend/go.mod`
- Create: `backend/internal/config/config.go`
- Create: `backend/cmd/hermes/main.go`

**Interfaces:**
- Consumes: nothing
- Produces: `config.Config` struct with fields: `Port, Root, Username, Password, DatabasePath`

- [ ] **Step 1: Initialize Go module**

Run: `cd backend && go mod init github.com/<user>/file-browser`

```bash
cd backend
go mod init github.com/youruser/hermes-filebrowser
```

- [ ] **Step 2: Create config package**

File: `backend/internal/config/config.go`
```go
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
```

- [ ] **Step 3: Create main.go skeleton**

File: `backend/cmd/hermes/main.go`
```go
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
```

- [ ] **Step 4: Tidy and verify**

```bash
cd backend && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: add Go module, config, and main skeleton"
```
