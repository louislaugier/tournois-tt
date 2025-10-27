.PHONY: help build up down restart logs test test-instagram e2e-meta build-prd run-prd

# Default target
help:
	@echo "Available commands:"
	@echo "  make build              - Build Docker containers"
	@echo "  make up                 - Start all services"
	@echo "  make down               - Stop all services"
	@echo "  make restart            - Restart all services"
	@echo "  make logs               - Show logs from all services"
	@echo "  make logs-api           - Show logs from API service"
	@echo "  make test               - Run all tests in API container"
	@echo "  make test-instagram     - Run Instagram unit tests in API container"
	@echo "  make e2e-meta      	 - Run Meta E2E test in API container (posts REAL image post, story, thread & FB post!)"
	@echo "  make shell-api          - Open shell in API container"
	@echo ""

# Docker commands
build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

logs:
	docker-compose logs -f

logs-api:
	docker-compose logs -f api

# Test commands
test:
	@echo "Running all tests in API container..."
	docker-compose exec api go test ./...

test-instagram:
	@echo "Running Instagram unit tests..."
	docker-compose exec api go test -v ./pkg/instagram/

# E2E Meta test
e2e-meta:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Meta E2E Test (Docker)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  WARNING: This will post REAL content to Meta Platforms!"
	@echo ""
	@echo "Prerequisites:"
	@echo "  1. INSTAGRAM_ENABLED=true in .env"
	@echo "  2. INSTAGRAM_ACCESS_TOKEN set in .env"
	@echo "  3. INSTAGRAM_PAGE_ID set in .env"
	@echo "  4. THREADS_ENABLED=true (optional) in .env"
	@echo "  5. THREADS_USER_ID (optional) in .env"
	@echo "Running E2E test in API container..."
	docker-compose exec -e E2E_TEST_ENABLED=true api go run -a ./cmd/test-instagram-e2e/

# Shell access
shell-api:
	docker-compose exec api /bin/sh

# Clean up
clean:
	docker-compose down -v
	docker system prune -f

# Development helpers
dev-build:
	docker-compose build --no-cache

dev-restart-api:
	docker-compose restart api
	docker-compose logs -f api


build-prd:
	docker build -t tournois-tt . --no-cache

run-prd:
	docker run -p 80:80 tournois-tt