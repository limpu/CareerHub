.PHONY: all bootstrap dev test lint migrate compose-up compose-down verify-docs

all: build

bootstrap:
	@echo "Bootstrapping development environment..."
	@pnpm install
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env from .env.example"; fi

dev:
	@echo "Starting local development services..."
	@pnpm dev:web

build:
	@echo "Building monorepo packages..."
	@pnpm build
	@cd services/core && go build ./cmd/...

test:
	@echo "Running tests..."
	@pnpm test
	@cd services/core && go test ./...

lint:
	@echo "Linting code..."
	@pnpm lint
	@cd services/core && go vet ./...

migrate:
	@echo "Running database migrations validation..."
	@cd services/core && go run ./cmd/migrate -dir ../../migrations -cmd validate

compose-up:
	@echo "Starting local infrastructure containers..."
	@docker compose -f deploy/compose/docker-compose.yml up -d

compose-down:
	@echo "Stopping local infrastructure containers..."
	@docker compose -f deploy/compose/docker-compose.yml down

verify-docs:
	@echo "Verifying documentation suite..."
	@powershell -ExecutionPolicy Bypass -File ./scripts/verify-docs.ps1