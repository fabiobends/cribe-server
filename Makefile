# ---------------------------
# VARIABLES
# ---------------------------
APP_ENV ?= development
POSTGRES_PORT ?= 5432
MIGRATION_PATH = ./infra/migrations

# ---------------------------
# COMMON DEVELOPER COMMANDS
# ---------------------------
dev:  ## Run with Air (hot reload, no debugger)
	@echo "Starting Air (hot reload)..."
	make services-up-dev
	@echo "Waiting for database to be ready..."
	@sleep 1
	godotenv -f .env.development air

dev-debug:  ## Run with Air + Delve (hot reload + debugger on port 2345)
	@echo "Starting Air with Delve (debugger)..."
	@echo "Killing port 2345 for debug server to attach to"
	make kill-port port=2345
	make services-up-dev
	@echo "Waiting for database to be ready..."
	@sleep 1
	godotenv -f .env.development air -c .air.debug.toml

run:  ## Run the app without hot reload
	@echo "Starting app..."
	make services-up-dev
	@echo "Waiting for database to be ready..."
	@sleep 1
	godotenv -f .env.development go run cmd/app/main.go

build:  ## Build the Go binary
	@echo "Building Go binary..."
	godotenv -f .env.development go build -o cribe-server ./cmd/app

test:  ## Run tests with test database
	@echo "Running tests..."
	make services-up-test
	@echo "Waiting for test database to be ready..."
	@sleep 1
	godotenv -f .env.test go test -p 1 -tags=test -coverprofile=coverage.out ./...
	@echo "Teardown test environment..."
	make services-down-test
	@echo "Generating coverage summary..."
	./scripts/coverage-summary.sh
	@echo "Checking coverage requirements..."
	./scripts/coverage-check.sh

view-coverage: ## Open coverage report in browser
	open coverage.html


# ---------------------------
# UTILITIES
# ---------------------------
kill-port: ## Kill a port
	@echo "Killing port $(port)..."
	lsof -ti:$(port) | xargs -r kill

clean-temp: ## Clean temp files
	rm -f cribe-server
	rm -f __debug_bin*
	rm -f coverage.out
	rm -f coverage.html

# ---------------------------
# DOCKER & DATABASE COMMANDS
# ---------------------------
services-up:  ## Start all services (current env)
	APP_ENV=$(APP_ENV) POSTGRES_PORT=$(POSTGRES_PORT) docker compose -p cribe-$(APP_ENV) -f compose.local.yml up -d

services-down:  ## Stop all services (current env)
	APP_ENV=$(APP_ENV) POSTGRES_PORT=$(POSTGRES_PORT) docker compose -p cribe-$(APP_ENV) -f compose.local.yml down --remove-orphans

services-up-dev:  ## Start dev services
	make services-up

services-down-dev:  ## Stop dev services
	make services-down

services-up-test:  ## Start test services
	APP_ENV=test POSTGRES_PORT=5433 make services-up

services-down-test:  ## Stop test services
	APP_ENV=test POSTGRES_PORT=5433 make services-down

clean-db-dev:  ## Remove dev DB volume
	docker volume rm cribe-development_development_db_data

clean-db-test:  ## Remove test DB volume
	docker volume rm cribe-test_test_db_data

clean-docker:  ## Remove all unused Docker data
	docker system prune -a --volumes

# ---------------------------
# MIGRATIONS
# ---------------------------
migrate-create:  ## Create a new migration
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(filter-out $@,$(MAKECMDGOALS))

migrate-up:  ## Apply all up migrations
	@source .env.development && migrate -path $(MIGRATION_PATH) -database $$DATABASE_URL up

migrate-down:  ## Rollback last migration
	@source .env.development && migrate -path $(MIGRATION_PATH) -database $$DATABASE_URL down

# ---------------------------
# GIT HOOKS
# ---------------------------
setup-hooks:  ## Install Git hooks
	@echo "Setting up Git hooks..."
	@git config core.hooksPath scripts
	@chmod +x scripts/pre-commit scripts/pre-push
	@echo "Git hooks installed successfully!"

# ---------------------------
# SANITY CHECKS
# ---------------------------
lint: ## Run linting
	golangci-lint run

format: ## Format code
	go fmt ./...

update-deps: ## Update dependencies
	go get -u ./...
	go mod tidy

typecheck: ## Run type checking
	go vet ./...

# ---------------------------
# HELP
# ---------------------------
help:
	@egrep -h '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | sed 's/:.*## /: /' | awk -F: '{printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: dev dev-debug run build test services-up services-down services-up-dev services-down-dev services-up-test services-down-test clean-db-dev clean-db-test clean-docker migrate-create migrate-up migrate-down setup-hooks help
