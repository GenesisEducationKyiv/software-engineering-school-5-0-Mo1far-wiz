include .env

BINARY\_NAME=weather-service
MAIN\_PATH=./cmd/main.go

DB\_URL=postgres\://\$(DB\_USER):\$(DB\_PASSWORD)@\$(DB\_HOST):\$(DB\_PORT)/\$(DB\_NAME)?sslmode=\$(DB\_SSL\_MODE)
MIGRATION\_PATH=\$(MIGRATION\_PATH)

.PHONY: help build clean run migrate-up migrate-down db-up db-down up down

help:
	@echo "Usage: make \[command]"
	@echo "Commands:"
	@echo "  build         Build the application"
	@echo "  run           Run the application"
	@echo "  clean         Clean build artifacts"
	@echo "  migrate-up    Apply database migrations"
	@echo "  migrate-down  Rollback database migrations"
	@echo "  db-up         Start the database container"
	@echo "  db-down       Stop the database container"
	@echo "  up            Start all services via docker-compose"
	@echo "  down          Stop all services via docker-compose"

clean:
	@rm -rf bin
	@go clean

build:
	@mkdir -p bin
	@go build -o bin/\$(BINARY\_NAME) \$(MAIN\_PATH)

run: build
	@APP\_PORT=\$(APP\_PORT)&#x20;
	DB\_URL="\$(DB\_URL)"&#x20;
	./bin/\$(BINARY\_NAME)

migrate-up:
	@migrate -path=\$(MIGRATION\_PATH) -database="\$(DB\_URL)" up

migrate-down:
	@migrate -path=\$(MIGRATION\_PATH) -database="\$(DB\_URL)" down

# Docker commands

db-up:
	@echo "Starting database container..."
	@docker-compose up -d postgres

db-down:
	@echo "Stopping database container..."
	@docker-compose stop postgres

up:
	@echo "Starting all services..."
	@docker-compose up -d

down:
	@echo "Stopping all services..."
	@docker-compose down
