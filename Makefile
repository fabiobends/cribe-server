.PHONY: services-up run dev test

services-up:
	docker compose up -d

services-down:
	docker compose down

build:
	go build -o cribe-server ./cmd/app

run: services-up
	godotenv -f .env.dev go run cmd/app/main.go

dev: services-up
	godotenv -f .env.dev air

test: services-up
	godotenv -f .env.dev go test -v ./...
