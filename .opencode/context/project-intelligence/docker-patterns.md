<!-- Context: project-intelligence/docker | Priority: medium | Version: 1.0 | Updated: 2026-02-16 -->

# Docker Patterns

**Purpose**: Docker containerization patterns for finsplitter.
**Last Updated**: 2026-02-16

## Multi-Stage Build
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /binary ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /binary .
CMD ["./binary"]
```

## Key Practices
- Use specific base image tags (not `latest`)
- Use `.dockerignore` to exclude unnecessary files
- Run as non-root user
- Use exec form for CMD/ENTRYPOINT
- Implement HEALTHCHECK

## Compose Profiles
```yaml
services:
  app:
    build: .
    profiles: [backend, debug]
  db:
    image: postgres:16
    profiles: [infra]
```

## Security
- Scan images with Trivy/Clair
- Don't store secrets in images
- Use read-only root filesystem when possible
- Limit container capabilities

## 📂 Codebase References
- **Dockerfile**: `Dockerfile`
- **Compose**: `compose.yml`, `compose.infra.yml`
- **Config**: `.dockerignore`
