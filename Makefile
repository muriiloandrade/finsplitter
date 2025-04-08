include .env

.PHONY: *

NAME=finsplitter
VERSION=prod

start-infra:
	@echo "==> Running infra containers"
	@docker compose --profile infra --env-file .env up -d

stop-infra:
	@echo "==> Stopping infra containers"
	@docker compose --profile infra --env-file .env down -v --remove-orphans

start-dev: start-infra
	@echo "==> Running development containers"
	@docker compose --profile backend --env-file .env up --build

stop-dev: stop-infra
	@echo "==> Stopping development containers"
	@docker compose --profile backend --env-file .env down --rmi local -v --remove-orphans

build:
	@echo "==> Building Docker API image"
	@docker build --target production --rm --compress -t ${NAME}:${VERSION} .

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

tools: install-golangci-lint
	@echo "==> Dealt with tools used on project"

install-golangci-lint:
	@echo "==> Intalling golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ${HOME}/go/bin v2.0.2

docker-scout: build
	@echo "==> Search for vulnerabilities on prod image"
	@docker scout cves -e --only-fixed ${NAME}:${VERSION}