.PHONY: build test lint migrate help

help:
	@echo "NektoKZ Makefile"
	@echo "Targets:"
	@echo "  build   - Build all services"
	@echo "  test    - Run tests for all services"
	@echo "  lint    - Run linter (placeholder)"
	@echo "  migrate - Show migration instructions"

build:
	go build ./...

test:
	go test ./...

lint:
	@echo "Linting... (golangci-lint run ./...)"

migrate:
	@echo "Migrations instructions:"
	@echo "Use migrate tool: migrate -path ./<service>/migrations -database \$$DB_URL up"
