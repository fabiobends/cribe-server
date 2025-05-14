.PHONY: services-up run dev test migrate-create migrate-up migrate-down clean-volumes create-db

services-up:
	APP_ENV=$(APP_ENV) POSTGRES_PORT=$(POSTGRES_PORT) docker compose -p cribe-$(APP_ENV) -f compose.local.yml up -d

services-down:
	APP_ENV=$(APP_ENV) POSTGRES_PORT=$(POSTGRES_PORT) docker compose -p cribe-$(APP_ENV) -f compose.local.yml down --remove-orphans

services-down-dev:
	APP_ENV=development POSTGRES_PORT=5432 docker compose -p cribe-development -f compose.local.yml down --remove-orphans

services-up-dev:
	APP_ENV=development POSTGRES_PORT=5432 docker compose -p cribe-development -f compose.local.yml up -d

services-down-test:
	APP_ENV=test POSTGRES_PORT=5433 docker compose -p cribe-test -f compose.local.yml down --remove-orphans

services-up-test:
	APP_ENV=test POSTGRES_PORT=5433 docker compose -p cribe-test -f compose.local.yml up -d

clean-db-dev:
	docker volume rm cribe-development_development_db_data

clean-db-test:
	docker volume rm cribe-test_test_db_data

clean-docker:
	docker system prune -a --volumes

build:
	go build -o cribe-server ./cmd/app

run:
	APP_ENV=development POSTGRES_PORT=5432 make services-up
	@echo "Waiting for database to be ready..."
	@sleep 1
	godotenv -f .env.development go run cmd/app/main.go

dev:
	make services-up-dev
	@echo "Waiting for database to be ready..."
	@sleep 1
	godotenv -f .env.development air

test:
	@echo "Running tests..."
	make services-up-test
	@echo "Waiting for test database to be ready..."
	@sleep 1
	godotenv -f .env.test go test -v ./...
	@echo "Teardown test environment..."
	make services-down-test

# Migrations

migration_path=./infra/migrations

migrate-create:
	migrate create -ext sql -dir $(migration_path) -seq $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@source .env.development && migrate -path $(migration_path) -database $$DATABASE_URL up

migrate-down:
	@source .env.development && migrate -path $(migration_path) -database $$DATABASE_URL down

# Git hooks
setup-hooks:
	@echo "Setting up Git hooks..."
	@mkdir -p .git/hooks
	@cp scripts/pre-commit .git/hooks/pre-commit
	@cp scripts/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-commit .git/hooks/pre-push
	@echo "Git hooks installed successfully!"

## Sanity checks

lint:
	golangci-lint run

format:
	go fmt ./...

update-deps:
	go get -u ./...
	go mod tidy

typecheck:
	go vet ./...
