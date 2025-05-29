include .env

.PHONY: *

# Migrations Config
MIGRATIONS_PATH ?= ./internal/gateways/postgres/migrations
DATABASE_URL ?= $(PG_URL)

UID=$(shell id -u)
GID=$(shell id -g)
MIGRATE_CMD = docker run --rm -u $(UID):$(GID) \
 	--add-host host.docker.internal:host-gateway \
 	-v $(MIGRATIONS_PATH):/migrations \
 	-w /migrations \
 	--network finsplitter-net \
 	migrate/migrate:v4.18.2 \
 	-path /migrations/ \
 	-database "$(DATABASE_URL)"
SQLC_CMD = docker run --rm -u $(UID):$(GID) \
 	--add-host host.docker.internal:host-gateway \
 	-v .:/src \
 	-w /src \
 	--network finsplitter-net \
 	sqlc/sqlc:latest

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

code-check:
	@echo "==> Linting..."
	@$(golanci-lint run)
	@echo "==> Formatting..."
	@$(golanci-lint fmt)

test:
	@echo "==> Running unit tests"
	@go test ./...

test-watch:
	@echo "==> Running unit tests in watch mode - IMPLEMENT ME"

test-cov:
	@echo "==> Running test coverage report - IMPLEMENT ME"

tools: install-golangci-lint install-lefthook
	@echo "==> Dealt with tools used on project"

install-golangci-lint:
	@echo "==> Intalling golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ${HOME}/go/bin v2.0.2

install-lefthook:
	@echo "==> Intalling lefthook"
	@go tool lefthook install

docker-scout: build
	@echo "==> Search for vulnerabilities on prod image"
	@docker scout cves -e --only-fixed ${NAME}:${VERSION}

# === Generation Targets ===
generate: generate-sqlc

generate-sqlc:
	@echo "==> Generating SQLC code"
	@$(SQLC_CMD) generate

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
