.PHONY: help build up down restart logs test test-instagram e2e-instagram meme-list meme-random meme-gen meme-all build-prd run-prd

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
	@echo "  make e2e-instagram      - Run Instagram E2E test in API container (posts REAL image!)"
	@echo "  make shell-api          - Open shell in API container"
	@echo ""
	@echo "Meme Generator commands:"
	@echo "  make meme-list          - List all meme templates"
	@echo "  make meme-list-cat CAT=fftt - List memes by category (fftt, match, club, etc.)"
	@echo "  make meme-random        - Generate a random meme"
	@echo "  make meme-gen ID=gratte_10_9 - Generate specific meme by ID"
	@echo "  make meme-all           - Generate ALL memes (75+ videos!)"

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

# E2E Instagram test
e2e-instagram:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Instagram & Threads E2E Test (Docker)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  WARNING: This will post REAL content to Instagram & Threads!"
	@echo ""
	@echo "Prerequisites:"
	@echo "  1. INSTAGRAM_ENABLED=true in .env"
	@echo "  2. INSTAGRAM_ACCESS_TOKEN set in .env"
	@echo "  3. INSTAGRAM_PAGE_ID set in .env"
	@echo "  4. THREADS_ENABLED=true (optional) in .env"
	@echo "  5. THREADS_USER_ID (optional) in .env"
	@echo "Running E2E test in API container..."
	docker-compose exec -e E2E_TEST_ENABLED=true api go run -a ./cmd/test-instagram-e2e/

# Meme Generator commands
meme-list:
	@echo "🏓 Listing all meme templates..."
	@cd api && go run cmd/meme-generator/main.go -list

meme-list-cat:
ifndef CAT
	@echo "❌ Error: Please specify category with CAT=<category>"
	@echo "Examples: make meme-list-cat CAT=fftt"
	@echo "Categories: fftt, match, club, competitive, championship, money, equipment, travel, tournament, classement, community, season"
	@exit 1
endif
	@echo "🏓 Listing memes for category: $(CAT)"
	@cd api && go run cmd/meme-generator/main.go -list -category $(CAT)

meme-random:
	@echo "🎲 Generating random meme..."
	@cd api && go run cmd/meme-generator/main.go -random
	@echo ""
	@echo "✅ Check ./api/meme-output/ for the video"

meme-gen:
ifndef ID
	@echo "❌ Error: Please specify meme ID with ID=<id>"
	@echo "Example: make meme-gen ID=gratte_10_9"
	@echo "Use 'make meme-list' to see all available IDs"
	@exit 1
endif
	@echo "🎬 Generating meme: $(ID)"
	@cd api && go run cmd/meme-generator/main.go -id $(ID)
	@echo ""
	@echo "✅ Check ./api/meme-output/ for the video"

meme-all:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  🏓 Generate ALL Memes (75+ videos)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  WARNING: This will generate 75+ video files!"
	@echo "    This may take several minutes..."
	@echo ""
	@read -p "Press Enter to continue or Ctrl+C to cancel: " confirm
	@cd api && go run cmd/meme-generator/main.go -all
	@echo ""
	@echo "✅ All memes generated in ./api/meme-output/"

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