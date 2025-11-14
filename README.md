# 💰 Finsplitter

> Financial expense splitting application built with Go

[![CI](https://github.com/muriiloandrade/finsplitter/actions/workflows/ci.yml/badge.svg)](https://github.com/muriiloandrade/finsplitter/actions/workflows/ci.yml)
[![Release](https://github.com/muriiloandrade/finsplitter/actions/workflows/release.yml/badge.svg)](https://github.com/muriiloandrade/finsplitter/actions/workflows/release.yml)
[![Security](https://github.com/muriiloandrade/finsplitter/actions/workflows/security.yml/badge.svg)](https://github.com/muriiloandrade/finsplitter/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/muriiloandrade/finsplitter)](https://goreportcard.com/report/github.com/muriiloandrade/finsplitter)

## 🚀 Quick Start

### Using Docker (Recommended)

Pull the image from GitHub Container Registry:

```bash
# Specific version
docker pull ghcr.io/muriiloandrade/finsplitter:v0.1.0
```

Run with docker-compose:

```bash
# Start infrastructure and application
make start-dev

# Start in debug mode (with Delve debugger)
make start-debug

# Stop everything
make stop-dev
```

### 🐳 Available Images

| Tag Pattern | Description | Platforms | Build Trigger |
|-------------|-------------|-----------|---------------|
| `v*.*.*` | Specific version releases | `linux/amd64`, `linux/arm64` | Release publication |
| `pr-*` | Pull request builds | `linux/amd64` | Pull requests |

## 🛠️ Development

### Prerequisites

- Go 1.25+ 
- Docker & Docker Compose
- PostgreSQL 18+ (via Docker)

### Local Setup

```bash
# Clone repository
git clone https://github.com/muriiloandrade/finsplitter.git
cd finsplitter

# Copy and configure environment variables
cp .env.example .env
# Edit .env with your settings

# Install development tools (lefthook and dlv)
make tools

# Start development environment
make start-dev
```

### Available Make Commands

#### 🔨 Development
```bash
make start-infra          # Start database only
make start-dev           # Start full development environment  
make start-debug         # Start with debugging enabled (Delve)
make stop-dev            # Stop development environment
make stop-infra          # Stop infrastructure only
```

#### 🧪 Code Quality & Testing
```bash
make code-check          # Run linters and formatters
make format              # Format code only
make lint                # Lint code only
make test                # Run unit tests
make docker-scout        # Security scan production image
```

#### 🏗️ Code Generation
```bash
make generate            # Generate code (SQLC + mocks)
make generate-sqlc       # Generate SQLC code only
make generate-mocks      # Generate mock files only
```

#### 🗄️ Database
```bash
make new-migration name=<name>  # Create new migration
make migrate-up [n=<steps>]                # Apply migrations (default: all)
make migrate-down [n=<steps>]              # Rollback migrations (default: all)
```

#### 🐳 Build & Deploy
```bash
make build               # Build production Docker image
make clean               # Remove production Docker image
make run-network-host    # Run with host network
make run-network-compose # Run with compose network
```

## 🏗️ Architecture

This project follows **Clean Architecture** principles with **Hexagonal/Ports-and-Adapters** pattern:

- **`internal/domain/entity`**: Core business entities (`CardBrand`, `Person`, `Card`)
- **`internal/app/ports`**: Repository interfaces defining data contracts
- **`internal/app/usecases`**: Business logic orchestration (e.g., `card-brand/create_card_brand.go`)
- **`internal/gateways/postgres`**: Database implementation using pgx/v5 and SQLC
- **`internal/gateways/http/v1`**: HTTP handlers using Huma v2 framework

### Key Features
- **🔄 Transaction Management**: Context-aware with `domain.Transactioner`
- **🔧 SQLC Integration**: Type-safe database access from SQL queries  
- **🧪 Comprehensive Testing**: Testify/mock with table-driven tests
- **📦 Dependency Injection**: Manual DI with clear boundaries
- **�️ OpenAPI**: Auto-generated docs with Huma v2

## 🔧 Configuration

Environment variables example provided as `.env.example`.

## 📋 API Documentation

Once the application is running, access:

- **OpenAPI Spec**: `http://localhost:3033/openapi`
- **Interactive Docs**: `http://localhost:3033/docs`
- **Health Checks**: 
  - Liveness: `http://localhost:3033/health/liveness`
  - Readiness: `http://localhost:3033/health/readiness`

### Current Endpoints
- `GET /card-brands` - List card brands
- `POST /card-brands` - Create card brand  
- `GET /card-brands/{id}` - Get card brand by ID
- `PATCH /card-brands/{id}` - Update card brand
- `DELETE /card-brands/{id}` - Delete card brand

## 🔄 CI/CD Pipeline

### Automated Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **CI** | Push to main, PRs | Quality checks, testing, Docker build |
| **Release** | Release published | Multi-arch production builds, security scans |
| **Security** | Push, PRs | Vulnerability scanning, SARIF reports |
| **Quality** | Via CI | Code formatting, linting (golangci-lint) |
| **Test** | Via CI | Unit tests with coverage |

### Release Process
1. **Development**: Work in feature branches, create PR
2. **Auto-labeling**: PRs automatically labeled using conventional commits
3. **CI Pipeline**: Quality checks + Docker build on PR merge
4. **Release Draft**: Auto-generated using Release Drafter
5. **Production Release**: Publish release → triggers multi-arch builds

### Container Security
All images include:
- ✅ **Multi-arch support** (amd64, arm64)
- ✅ **Distroless base** (minimal attack surface)
- ✅ **SBOM generation** (Software Bill of Materials)
- ✅ **Provenance attestation** (Build provenance)
- ✅ **Trivy security scanning** (Vulnerability reports)
- ✅ **Non-root execution** (Security hardening)

## 🛠️ Development Workflow

### Adding New Entities
To add a new entity (e.g., "Transaction"):

1. **Migration**: `make new-migration name=create_table_transaction`
2. **SQL Queries**: Write in `internal/gateways/postgres/sqlc/queries/transaction.sql`
3. **Generate Code**: `make generate-sqlc` 
4. **Domain Entity**: `internal/domain/entity/transaction.go`
5. **Repository Interface**: `internal/app/ports/transaction_repo.go`  
6. **Repository Implementation**: `internal/gateways/postgres/transaction.go`
7. **Use Cases**: `internal/app/usecases/transaction/`
8. **HTTP Handlers**: `internal/gateways/http/v1/transaction/`
9. **Wire Dependencies**: `cmd/api/main.go`
10. **Generate Mocks & Tests**: `make generate-mocks`

### Pre-commit Hooks
Lefthook automatically runs quality checks:
- Code formatting and linting via `make code-check`
- Configured in `lefthook.yml`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Commit using [Conventional Commits](https://conventionalcommits.org/):
   - `feat:` for new features
   - `fix:` for bug fixes  
   - `chore:` for maintenance
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Open a Pull Request (auto-labeled by conventional commit type)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
