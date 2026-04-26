.PHONY: build test lint migrate help up down ps logs migrate-all migrate-user migrate-chat migrate-moderation

SERVICE_PATTERNS=./api-gateway/... ./user-service/... ./matchmaking-service/... ./chat-service/... ./moderation-service/... ./notification-service/...

help:
	@echo "NektoKZ Makefile"
	@echo "Targets:"
	@echo "  build   - Build all services"
	@echo "  test    - Run tests for all services"
	@echo "  up      - Start everything"
	@echo "  up-infra - Start infra only"
	@echo "  up-services - Start services only"
	@echo "  down    - Stop everything"
	@echo "  ps      - Show containers"
	@echo "  logs    - Tail logs"
	@echo "  lint    - Run linter (placeholder)"
	@echo "  migrate - Run all migrations"
	@echo "  migrate-user - Run user-service migrations"
	@echo "  migrate-chat - Run chat-service migrations"
	@echo "  migrate-moderation - Run moderation-service migrations"

build:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build.ps1

test:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/test.ps1

up: up-infra up-services
	@echo "Project NektoKZ is up and running"

up-infra:
	docker compose -f docker-compose.infra.yml up -d

up-services:
	docker compose -f docker-compose.services.yml up -d --build

down:
	docker compose -f docker-compose.services.yml down
	docker compose -f docker-compose.infra.yml down

ps:
	docker compose -f docker-compose.infra.yml ps
	docker compose -f docker-compose.services.yml ps

logs:
	docker compose -f docker-compose.infra.yml logs -f --tail=100 & docker compose -f docker-compose.services.yml logs -f --tail=100

lint:
	@echo "Linting... (golangci-lint run ./...)"

migrate:
migrate-all:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/migrate.ps1 -Service all

migrate-user:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/migrate.ps1 -Service user

migrate-chat:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/migrate.ps1 -Service chat

migrate-moderation:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/migrate.ps1 -Service moderation

migrate: migrate-all
 
swagger:
	swag init -g cmd/main.go -d ./api-gateway --output ./api-gateway/docs
