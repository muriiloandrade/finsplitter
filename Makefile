-include .env

.PHONY: *

# Migrations Config
MIGRATIONS_PATH ?= ./internal/gateways/postgres/migrations
DATABASE_URL ?= $(PG_URL)

UID=$(shell id -u)
GID=$(shell id -g)
# renovate: datasource=docker depName=migrate/migrate
MIGRATE_VERSION := v4.19.0@sha256:d5c978181e3bfa55cc50e3bd8d7da3d87418a87693453250a8804b81ee6494db
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
MOCKERY_VERSION := v3.5.5@sha256:b5bb5f45647d3d7646496617113bc4a2bec4349df20d23b33afdbc73fa514ee1
MOCKERY_CMD = docker run --rm -u $(UID):$(GID) \
	-v .:/src \
	-w /src \
	vektra/mockery:$(MOCKERY_VERSION)
# renovate: datasource=docker depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.6.1-alpine@sha256:a7da5151e0bd61bd7f99e1ebd8e5e144b535b73b2762c498443ff4f6a4a538c4
GOLANGCI_LINT_CMD = docker run --rm -t -v $(shell pwd):/app -w /app \
	-v $(shell go env GOCACHE):/home/.cache/go-build \
	-e GOCACHE=/home/.cache/go-build \
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

start-dev start-debug: start-%:
start-%: start-infra
	@echo "==> Running containers in $* mode"
	@if [ "$*" = "debug" ]; then \
		docker compose --profile debug --env-file .env up --build; \
	else \
		docker compose --profile backend --env-file .env up --build; \
	fi

stop-dev stop-debug: stop-%:
stop-%: stop-infra
	@echo "==> Stopping containers in $* mode"
	@if [ "$*" = "debug" ]; then \
		docker compose --profile debug --env-file .env down --rmi local -v --remove-orphans; \
	else \
		docker compose --profile backend --env-file .env down --rmi local -v --remove-orphans; \
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

run-network-compose: build start-infra
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
	@echo "==> Running unit tests"
	@go test ./...

test-watch:
	@echo "==> Running unit tests in watch mode - IMPLEMENT ME"

test-cov:
	@echo "==> Running test coverage report - IMPLEMENT ME"

tools: install-lefthook
	@echo "==> Dealt with tools used on project"


install-lefthook:
	@echo "==> Installing lefthook"
	@go tool lefthook install

docker-scout: build
	@echo "==> Search for vulnerabilities on prod image"
	@docker scout cves -e --only-fixed ${NAME}:${VERSION}

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
	@echo "==> Reverting all migrations down"
	@$(MIGRATE_CMD) down --all

migrate-down-%: n=$*
migrate-down-%: migrate-check-vars
	@echo "==> Reverting migrations down $(if $(n),for $(n) steps...)"
	@$(MIGRATE_CMD) down $(n)

# === Application Targets ===

# Add a help target for better usability
help:
	@echo "Usage: make [target] [VARIABLE=value]"
	@echo ""
	@echo "Migration Targets:"
	@echo "  new-migration name=<name>  Create a new SQL migration file."
	@echo "  migrate-up [n=<steps>]     Apply migrations up (optionally specific number of steps)."
	@echo "  migrate-down-<steps>       Revert migrations down a specific amount of steps."
	@echo "  migrate-down-all           Revert all migrations down."
	@echo "Variables for migrations:"
	@echo "  MIGRATIONS_PATH            Path to migration files (default: $(MIGRATIONS_PATH))"
	@echo "  DATABASE_URL               Database connection string (must be set in .env or passed)"
	@echo ""
	@echo "Generation Targets:"
	@echo "  generate                   Generate code (e.g., SQLC)."
	@echo "  generate-sqlc              Generate SQLC code only."
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
	@echo "  test                       Run unit tests."
	@echo "  test-watch                 Run unit tests in watch mode (not implemented)."
	@echo "  test-cov                   Run test coverage report (not implemented)."
	@echo "  tools                      Install necessary development tools."
	@echo "  docker-scout               Scan the production image for vulnerabilities."
