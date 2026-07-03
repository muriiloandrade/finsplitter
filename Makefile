-include .env

.PHONY: *

# Migrations Config
MIGRATIONS_PATH ?= ./internal/gateways/postgres/migrations
DATABASE_URL ?= $(PG_URL)

UID=$(shell id -u)
GID=$(shell id -g)

# renovate: datasource=docker depName=migrate/migrate
MIGRATE_VERSION := v4.19.1@sha256:cc4ad8e19d66791e3689405d9a028ce6e9614f32032db14acda1469f7201d6e4
MIGRATE_CMD = docker run --rm -u $(UID):$(GID) \
	--add-host host.docker.internal:host-gateway \
	-v $(MIGRATIONS_PATH):/migrations \
	-w /migrations \
	--network finsplitter-net \
	migrate/migrate:$(MIGRATE_VERSION) \
	-path /migrations/ \
	-database "$(DATABASE_URL)"
# renovate: datasource=docker depName=sqlc/sqlc
SQLC_VERSION := 1.30.0@sha256:b8d1092c720438e093a231e75eba5d55b7696122f390292acabd5b6d3e986a12
SQLC_CMD = docker run --rm -u $(UID):$(GID) \
	--add-host host.docker.internal:host-gateway \
	-v .:/src \
	-w /src \
	sqlc/sqlc:$(SQLC_VERSION)
# renovate: datasource=docker depName=vektra/mockery
MOCKERY_VERSION := v3.7.1@sha256:bb74169bd86ecd32fa77b8b4646f266d907ac6cb9a21d5cb9de0f7b7ee91c20f
MOCKERY_CMD = docker run --rm -u $(UID):$(GID) \
	-v .:/src \
	-w /src \
	vektra/mockery:$(MOCKERY_VERSION)
# renovate: datasource=docker depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.11.3-alpine@sha256:b1c3de5862ad0a95b4e45a993b0f00415835d687e4f12c845c7493b86c13414e
GOLANGCI_LINT_CMD = docker run --rm -t -v $(shell pwd):/app -w /app \
	-v $(shell go env GOCACHE):/home/.cache/go-build \
	-e GOCACHE=/home/.cache/go-build \
	-e GOEXPERIMENT=jsonv2 \
	-v $(shell go env GOMODCACHE):/home/.cache/mod \
	-e GOMODCACHE=/home/.cache/mod \
	-v ~/.cache/golangci-lint:/home/.cache/golangci-lint \
	-e GOLANGCI_LINT_CACHE=/home/.cache/golangci-lint \
	golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint

# Default target
default: help

NAME=finsplitter
VERSION=prod
GIT_BUILD_TAG=localtag
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%S%Z)

start-infra:
	@echo "==> Running infra containers"
	@docker compose --profile infra --env-file .env up -d

stop-infra:
	@echo "==> Stopping infra containers"
	@docker compose --profile infra --env-file .env down -v --remove-orphans

start-monitoring: docker-network-setup
	@echo "==> Starting monitoring stack"
	@docker compose --profile monitoring --env-file .env up -d

stop-monitoring:
	@echo "==> Stopping monitoring stack"
	@docker compose --profile monitoring --env-file .env down -v --remove-orphans

start-dev start-debug: start-%:
start-%: docker-network-setup start-infra
	@echo "==> Running containers in $* mode"
	@if [ "$*" = "debug" ]; then \
		docker compose --profile infra --profile debug --env-file .env up --build --watch; \
	else \
		docker compose --profile infra --profile backend --env-file .env up --build --watch; \
	fi

stop-dev stop-debug: stop-%:
stop-%: stop-infra
	@echo "==> Stopping containers in $* mode"
	@if [ "$*" = "debug" ]; then \
		docker compose --profile infra --profile debug --env-file .env down --rmi local -v --remove-orphans; \
	else \
		docker compose --profile infra --profile backend --env-file .env down --rmi local -v --remove-orphans; \
	fi

build:
	@echo "==> Building Docker API image"
	@docker build --build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_BUILD_TAG=$(GIT_BUILD_TAG) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--target production --rm --compress -t ${NAME}:${VERSION} .

clean:
	@echo "==> Deleting Docker image"
	@docker rmi ${NAME}:${VERSION}

run-network-host: build
	@echo "==> Running Docker API image"
	@docker run --rm --network host --env-file .env --name ${NAME} -t ${NAME}:${VERSION}

run-network-compose: docker-network-setup build start-infra
	@echo "==> Running Docker API image"
	@docker run --rm --network finsplitter-net --env-file .env -p ${PORT}:${PORT} --name ${NAME} -t ${NAME}:${VERSION}

format:
	@echo "==> Formatting..."
	@$(GOLANGCI_LINT_CMD) fmt

lint:
	@echo "==> Linting..."
	@$(GOLANGCI_LINT_CMD) run

code-check: format lint
	@echo "==> Code check complete"

test:
	@echo "==> Running unit tests (short mode, skips integration)"
	@GOEXPERIMENT=jsonv2 go test -short ./...

test-int:
	@echo "==> Running all non-e2e tests (unit + integration)"
	@GOEXPERIMENT=jsonv2 go test ./...

test-e2e:
	@echo "==> Running e2e tests"
	@GOEXPERIMENT=jsonv2 go test -tags=e2e -v -count=1 -run "^TestE2E" \
	  ./internal/gateways/http/v1/auth/...

test-watch:
	@echo "==> Running unit tests in watch mode - IMPLEMENT ME"

test-cov:
	@echo "==> Running test coverage report"
	@GOEXPERIMENT=jsonv2 go test -coverprofile=coverage.out $$(GOEXPERIMENT=jsonv2 go list ./... | grep -v '/cmd/api' | grep -v '/api$$' | grep -v '/pkg/telemetry' | grep -v '/internal/config' | grep -v '/migrations' | grep -v '/sqlc')
	@echo ""
	@echo "==> Coverage by package:"
	@go tool cover -func=coverage.out | grep -v "^total:" | grep -v "mocks.gen.go" | grep -v "/testutils/"
	@echo ""
	@echo "==> Total Coverage:"
	@grep -v "mocks.gen.go" coverage.out | grep -v "testutils" > coverage_filtered.out
	@go tool cover -func=coverage_filtered.out | grep "^total:"

tools: install-lefthook
	@echo "==> Installing necessary development tools"
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "==> Dealt with tools used on project"

install-lefthook:
	@echo "==> Installing lefthook"
	@go tool lefthook install

docker-scout: build
	@echo "==> Search for vulnerabilities on prod image"
	@docker scout cves -e --only-fixed ${NAME}:${VERSION}

docker-network-setup:
	@echo "==> Setting up docker network if not exists"
	@docker network inspect finsplitter-net --format '{{.Id}}' 2>/dev/null || docker network create finsplitter-net

# === Generation Targets ===
generate: generate-sqlc generate-mocks

generate-sqlc:
	@echo "==> Generating SQLC code"
	@$(SQLC_CMD) generate

generate-mocks:
	@echo "==> Running mockery"
	@$(MOCKERY_CMD)

# === Migration Targets ===

migrate-check-vars:
ifndef DATABASE_URL
	$(error DATABASE_URL is not set. Please define it in your .env file or pass it as an argument)
endif

new-migration: migrate-check-vars
ifndef name
	$(error Usage: make new-migration name=<migration_name>)
endif
	@echo "==> Creating new migration: $(name)"
	@$(MIGRATE_CMD) create -ext sql $(name)

migrate-up: migrate-check-vars
	@echo "==> Applying migrations up $(if $(n),for $(n) steps...)"
	@$(MIGRATE_CMD) up $(n)

migrate-down: migrate-check-vars
	@echo "==> Reverting migrations down $(if $(n),for $(n) steps,all migrations)..."
	@$(MIGRATE_CMD) down $(if $(n),$(n),--all)

# === Application Targets ===

# Add a help target for better usability
help:
	@echo "Usage: make [target] [VARIABLE=value]"
	@echo ""
	@echo "Migration Targets:"
	@echo "  new-migration name=<name>  Create a new SQL migration file."
	@echo "  migrate-up [n=<steps>]     Apply migrations up (optionally specific number of steps)."
	@echo "  migrate-down [n=<steps>]   Revert migrations down (optionally specific number of steps, default: all)."
	@echo "Variables for migrations:"
	@echo "  MIGRATIONS_PATH            Path to migration files (default: $(MIGRATIONS_PATH))"
	@echo "  DATABASE_URL               Database connection string (must be set in .env or passed)"
	@echo ""
	@echo "Generation Targets:"
	@echo "  generate                   Generate all code (SQLC + mocks)."
	@echo "  generate-sqlc              Generate SQLC code only."
	@echo "  generate-mocks             Generate mock files only."
	@echo ""
	@echo "Monitoring Targets:"
	@echo "  start-monitoring           Start observability stack (OTel, Grafana, Tempo, Loki)"
	@echo "  stop-monitoring            Stop observability stack"
	@echo ""
	@echo "Other Targets:"
	@echo "  start-infra                Start infrastructure containers (e.g., database)."
	@echo "  stop-infra                 Stop infrastructure containers."
	@echo "  start-dev                  Start development environment (infra + app)."
	@echo "  start-debug                Start development environment in debug mode (infra + app)."
	@echo "  stop-dev                   Stop development environment."
	@echo "  build                      Build the production Docker image."
	@echo "  clean                      Remove the production Docker image."
	@echo "  run-network-host           Run the app using host network."
	@echo "  run-network-compose        Run the app using docker compose network."
	@echo "  code-check                 Run linters and formatters."
	@echo "  test                       Run unit tests (short mode, skips integration)."
	@echo "  test-int                   Run all non-e2e tests (unit + integration)."
	@echo "  test-e2e                   Run end-to-end auth tests (requires Docker)."
	@echo "  test-watch                 Run unit tests in watch mode (not implemented)."
	@echo "  test-cov                   Run test coverage report."
	@echo "  tools                      Install necessary development tools."
	@echo "  docker-scout               Scan the production image for vulnerabilities."
