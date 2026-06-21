.PHONY: build build-check test lint help up down ps logs up-build up-infra up-services up-services-build rebuild-services
.PHONY: logs-infra logs-services coturn-status migrate migrate-all migrate-user migrate-chat migrate-moderation swagger

SERVICES = api-gateway user-service chat-service matchmaking-service moderation-service notification-service

help:
	@echo "Mathalama Talk Makefile"
	@echo "Targets:"
	@echo "  build          - Build all services"
	@echo "  build-check    - Quick go build for all services"
	@echo "  test           - Run tests for all services"
	@echo "  lint           - Run golangci-lint"
	@echo "  up             - Start everything"
	@echo "  up-build       - Start everything (with build)"
	@echo "  up-infra       - Start infra only"
	@echo "  up-services    - Start services only (no build)"
	@echo "  up-services-build - Start services only (with build)"
	@echo "  rebuild-services - Rebuild + recreate services"
	@echo "  down           - Stop everything"
	@echo "  ps             - Show containers"
	@echo "  logs           - Tail logs"
	@echo "  logs-infra     - Logs for infra (postgres, redis, coturn)"
	@echo "  logs-services  - Logs for microservices"
	@echo "  coturn-status  - Check coturn sessions"
	@echo "  migrate        - Run all migrations"
	@echo "  migrate-user   - Run user-service migrations"
	@echo "  migrate-chat   - Run chat-service migrations"
	@echo "  migrate-moderation - Run moderation-service migrations"
	@echo "  swagger        - Regenerate swagger docs"

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && CGO_ENABLED=0 go build -o ../../bin/$$svc ./cmd/main.go && cd ../..; \
	done
	@echo "All services built successfully"

build-check:
	@for svc in $(SERVICES); do \
		echo "Checking $$svc..."; \
		go build ./services/$$svc/cmd/main.go || exit 1; \
	done
	@echo "All services compile OK"

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... -v -count=1 && cd ../..; \
	done

up: up-infra up-services
	@echo "Project Mathalama Talk is up and running"

up-build: up-infra up-services-build
	@echo "Project Mathalama Talk is up and running (built)"

up-infra:
	docker compose up -d postgres redis nats coturn jaeger

up-services:
	docker compose up -d api-gateway user-service matchmaking-service chat-service moderation-service notification-service web

up-services-build:
	docker compose up -d --build api-gateway user-service matchmaking-service chat-service moderation-service notification-service web

rebuild-services:
	docker compose up -d --build --force-recreate api-gateway user-service matchmaking-service chat-service moderation-service notification-service web

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f --tail=50

logs-infra:
	docker compose logs -f postgres redis nats coturn jaeger

logs-services:
	docker compose logs -f api-gateway user-service matchmaking-service chat-service moderation-service notification-service web

coturn-status:
	docker exec coturn turnadmin -l

lint:
	golangci-lint run ./services/...

migrate: migrate-all

migrate-all: migrate-user migrate-chat migrate-moderation
	@echo "All migrations completed"

migrate-user:
	docker compose -f docker-compose.yml run --rm migrate-user

migrate-chat:
	docker compose -f docker-compose.yml run --rm migrate-chat

migrate-moderation:
	docker compose -f docker-compose.yml run --rm migrate-moderation

swagger:
	swag init -g cmd/main.go -d ./services/api-gateway --output ./services/api-gateway/docs
