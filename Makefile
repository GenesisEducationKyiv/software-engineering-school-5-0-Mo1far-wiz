include .env

SHELL := /usr/bin/env bash

BINARY_NAME=weather-service
MAIN_PATH=./cmd/main.go

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: help build clean run migrate-up migrate-down db-up db-down up down lint

help:
	@echo "Usage: make [command]"
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
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@APP_PORT=$(APP_PORT) \
	DB_URL="$(DB_URL)" \
	./bin/$(BINARY_NAME)

migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database="$(DB_URL)" up

migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database="$(DB_URL)" down

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

lint:
	@golangci-lint run --config <(curl https://raw.githubusercontent.com/fabl3ss/genesis-se-school-linter/refs/heads/main/.golangci.yaml)