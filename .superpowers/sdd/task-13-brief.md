### Task 13: Docker Build + Docker Compose + GitHub Actions + README

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.github/workflows/docker-publish.yml`
- Create: `README.md`

- [ ] **Step 1: Dockerfile (multi-stage)**

File: `Dockerfile`
```dockerfile
# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/ ./
COPY --from=frontend /app/frontend/dist ./internal/api/frontend/dist
RUN CGO_ENABLED=0 go build -o /hermes ./cmd/hermes

# Stage 3: Distroless runtime
FROM gcr.io/distroless/base-debian12
COPY --from=backend /hermes /hermes
EXPOSE 8080
ENTRYPOINT ["/hermes"]
```

- [ ] **Step 2: docker-compose.yml**

File: `docker-compose.yml`
```yaml
services:
  hermes:
    build: .
    ports:
      - "8080:8080"
    environment:
      - FB_PORT=8080
      - FB_ROOT=/data
      - FB_USERNAME=admin
      - FB_PASSWORD=changeme
      - FB_DATABASE=/data/filebrowser.db
    volumes:
      - /volume2/HermesShared:/data
    restart: unless-stopped
```

- [ ] **Step 3: GitHub Actions workflow**

File: `.github/workflows/docker-publish.yml`
```yaml
name: Build and Push to Docker Hub

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ secrets.DOCKER_USERNAME }}/hermes-filebrowser:latest,${{ secrets.DOCKER_USERNAME }}/hermes-filebrowser:${{ github.sha }}
          platforms: linux/amd64,linux/arm64
```

- [ ] **Step 4: README.md**

Write a README.md with:
- Project description
- Quick start (docker-compose up)
- Environment variables table
- Building from source (Docker or manual)
- Reverse proxy notes (X-Forwarded-For support)
- Deploy with Portainer steps

- [ ] **Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml .github/workflows/docker-publish.yml README.md
git commit -m "feat: add Docker build, compose, CI/CD, and README"
```
