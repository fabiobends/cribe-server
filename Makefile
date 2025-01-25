.PHONY: services-up run dev test migrate-create migrate-up migrate-down

services-up:
	APP_ENV=$(APP_ENV) docker compose -f compose.local.yml up -d

services-down:
	APP_ENV=$(APP_ENV) docker compose -f compose.local.yml down

build:
	go build -o cribe-server ./cmd/app

run:
	APP_ENV=test make services-down
	APP_ENV=development make services-up
	godotenv -f .env.development go run cmd/app/main.go

dev:
	APP_ENV=test make services-down
	APP_ENV=development make services-up
	godotenv -f .env.development air

test:
	APP_ENV=development make services-down
	APP_ENV=test make services-up
	godotenv -f .env.test go test -v ./...
	APP_ENV=test make services-down

clean-docker:
	docker system prune -a --volumes

# Migrations
migration_path=./infra/migrations

migrate-create:
	migrate create -ext sql -dir $(migration_path) -seq $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	migrate -path $(migration_path) -database ${DATABASE_URL} up

migrate-down:
	migrate -path $(migration_path) -database ${DATABASE_URL} down
