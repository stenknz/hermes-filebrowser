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
