### Task 6: Wire Everything Together in main.go

**Files:**
- Modify: `backend/cmd/hermes/main.go`

- [ ] **Step 1: Replace main.go with full server wiring**

File: `backend/cmd/hermes/main.go`
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/youruser/hermes-filebrowser/internal/api"
	"github.com/youruser/hermes-filebrowser/internal/config"
	"github.com/youruser/hermes-filebrowser/internal/db"
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
	log.Printf("Hermes Filebrowser listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
```

- [ ] **Step 2: Build and verify**

```bash
cd backend && go build ./cmd/hermes
```

- [ ] **Step 3: Commit**

```bash
git add backend/cmd/hermes/main.go
git commit -m "feat: wire up server with config, DB, and routes"
```
