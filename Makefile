.PHONY: build test lint migrate help up down ps logs migrate-all migrate-user migrate-chat migrate-moderation
.PHONY: build test lint migrate help up down ps logs migrate-all migrate-user migrate-chat migrate-moderation build-check logs-infra logs-services coturn-status
.PHONY: up-build up-services-build rebuild-services

SERVICE_PATTERNS=./api-gateway/... ./user-service/... ./matchmaking-service/... ./chat-service/... ./moderation-service/... ./notification-service/...

help:
	@echo "NektoKZ Makefile"
	@echo "Targets:"
	@echo "  build   - Build all services"
	@echo "  build-check - Quick go build for all services"
	@echo "  coturn-status - Check coturn sessions"
	@echo "  logs-infra - Logs for infra (postgres, redis, coturn)"
	@echo "  logs-services - Logs for microservices"
	@echo "  test    - Run tests for all services"
	@echo "  up      - Start everything"
	@echo "  up-build - Start everything (with build)"
	@echo "  up-infra - Start infra only"
	@echo "  up-services - Start services only (no build)"
	@echo "  up-services-build - Start services only (with build)"
	@echo "  rebuild-services - Rebuild + recreate services"
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

build-check:
	go build ./api-gateway/cmd/main.go
	go build ./user-service/cmd/main.go
	go build ./chat-service/cmd/main.go
	go build ./matchmaking-service/cmd/main.go
	go build ./moderation-service/cmd/main.go
	go build ./notification-service/cmd/main.go

test:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/test.ps1

up: up-infra up-services
	@echo "Project NektoKZ is up and running"

up-build: up-infra up-services-build
	@echo "Project NektoKZ is up and running (built)"

up-infra:
	docker compose -f docker-compose.infra.yml up -d

up-services:
	docker compose -f docker-compose.services.yml up -d

up-services-build:
	docker compose -f docker-compose.services.yml up -d --build

rebuild-services:
	docker compose -f docker-compose.services.yml up -d --build --force-recreate

down:
	docker compose -f docker-compose.services.yml down
	docker compose -f docker-compose.infra.yml down

ps:
	docker compose -f docker-compose.infra.yml ps
	docker compose -f docker-compose.services.yml ps

logs:
	docker compose -f docker-compose.infra.yml logs -f --tail=50

logs-infra:
	docker compose -f docker-compose.infra.yml logs -f

logs-services:
	docker compose -f docker-compose.services.yml logs -f

coturn-status:
	docker exec nektokz-coturn turnadmin -l

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
