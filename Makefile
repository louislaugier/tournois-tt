.PHONY: help build up down restart logs test test-instagram e2e-meta build-prd run-prd
.PHONY: ig-image ig-image-random ig-image-random-local

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
	@echo "  make test-instagram-follow - Run Instagram follow/unfollow bot test"
	@echo "  make test-instagram-follow-vet - Run go vet + Instagram follow/unfollow bot test"
	@echo "  make e2e-meta      	 - Run Meta E2E test in API container (posts REAL image post, story, thread & FB post!)"
	@echo "  make post-full ID=1234  - Post tournament to Instagram FEED + STORY + THREADS"
	@echo "  make post-story ID=1234 - Post tournament to Instagram STORY ONLY (no feed, no threads)"
	@echo "  make post-multiple IDS=\"3340 3336\" - Post multiple tournaments"
	@echo "  make cache-stats        - Show Instagram posted cache statistics"
	@echo "  make cache-remove IDS=\"3340,3336\" - Remove tournaments from posted cache"
	@echo "  make cache-sync         - Sync cache with Instagram API (detect deleted posts)"
	@echo "  make shell-api          - Open shell in API container"
	@echo "  make ig-image ID=1234   - Generate Instagram images (feed + story) for tournament ID"
	@echo "  make ig-image-feed ID=1234 - Generate only feed image (1080x1080)"
	@echo "  make ig-image-story ID=1234 - Generate only story image (1080x1920)"
	@echo "  make ig-image-random    - Generate feed + story for a random tournament (Docker)"
	@echo "  make ig-image-random-story - Generate ONLY story for a random tournament (Docker)"
	@echo "  make ig-image-random-local - Generate feed + story locally (NO Docker, NO posting)"
	@echo "  make ig-image-random-local-story - Generate ONLY story locally (NO Docker, NO posting)"
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

test-instagram-follow:
	@echo "Running Instagram follow/unfollow test..."
	docker-compose exec -e GIN_MODE=release -w /go/src/tournois-tt/api api go run cmd/test-instagram-follow/main.go

test-instagram-follow-vet:
	@echo "Running go vet on Instagram bot code..."
	docker-compose exec api go vet ./pkg/instagram/bot/
	@echo "✅ No linting issues found!"
	@echo ""
	@echo "Running Instagram follow/unfollow test..."
	docker-compose exec api go run cmd/test-instagram-follow/main.go


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
	@echo ""
	@echo "🔄 Rebuilding API container to pick up .env changes..."
	docker-compose up -d --force-recreate --build api
	@echo "⏳ Waiting 10 seconds for API to start..."
	sleep 10
	@echo "Running E2E test in API container..."
	docker-compose exec -e E2E_TEST_ENABLED=true -e TEST_TOURNAMENT_ID=$(TEST_TOURNAMENT_ID) api go run -a ./cmd/test-instagram-e2e/

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

# Instagram image generation
ig-image:
	@if [ -z "$(ID)" ]; then echo "Usage: make ig-image ID=<tournament_id>"; exit 1; fi
	@echo "Generating feed (1080x1080) and story (1080x1920) images..."
	docker-compose exec api go run cmd/generate-instagram-images/main.go --ids $(ID)

ig-image-feed:
	@if [ -z "$(ID)" ]; then echo "Usage: make ig-image-feed ID=<tournament_id>"; exit 1; fi
	@echo "Generating feed image only (1080x1080)..."
	docker-compose exec api go run cmd/generate-instagram-images/main.go --ids $(ID) --story=false

ig-image-story:
	@if [ -z "$(ID)" ]; then echo "Usage: make ig-image-story ID=<tournament_id>"; exit 1; fi
	@echo "Generating story image only (1080x1920)..."
	docker-compose exec api go run cmd/generate-instagram-images/main.go --ids $(ID) --feed=false

ig-image-random:
	@echo "Generating Instagram images (feed + story) for a random tournament..."
	docker-compose exec api go run cmd/test-instagram-image/main.go

ig-image-random-story:
	@echo "Generating ONLY story image (1080x1920) for a random tournament..."
	docker-compose exec api go run cmd/test-instagram-image/main.go --story-only

ig-image-random-local:
	@echo "Generating Instagram images locally (no Docker, no posting)..."
	cd api && go run cmd/test-instagram-image/main.go

ig-image-random-local-story:
	@echo "Generating ONLY story image (1080x1920) locally (no Docker, no posting)..."
	cd api && go run cmd/test-instagram-image/main.go --story-only

# Post full to Instagram (feed + story + threads)
post-full:
	@if [ -z "$(ID)" ]; then echo "Usage: make post-full ID=<tournament_id>"; exit 1; fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Post Tournament $(ID) to Instagram FULL"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  This will post to Instagram Feed + Story + Threads"
	@echo ""
	cd api && go run cmd/post-instagram-full/main.go --id $(ID) --yes

# Post multiple tournaments
post-multiple:
	@if [ -z "$(IDS)" ]; then echo "Usage: make post-multiple IDS=\"3340 3336\""; exit 1; fi
	@echo "Posting multiple tournaments to Instagram..."
	cd api && ./scripts/post_multiple.sh $(IDS)

# Post story only to Instagram
post-story:
	@if [ -z "$(ID)" ]; then echo "Usage: make post-story ID=<tournament_id>"; exit 1; fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Post Tournament $(ID) to Instagram STORY ONLY"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  This will post to Instagram Story (no feed, no threads)"
	@echo ""
	cd api && go run cmd/post-instagram-story/main.go --id $(ID)

# Post to Threads only
post-threads:
	@if [ -z "$(ID)" ]; then echo "Usage: make post-threads ID=<tournament_id>"; exit 1; fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Post Tournament $(ID) to Threads ONLY"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  This will post to Threads only"
	@echo ""
	cd api && go run cmd/post-threads-only/main.go --id $(ID) --yes


# Cache management
cache-stats:
	@echo "📊 Instagram Posted Cache Statistics"
	@echo ""
	cd api && go run cmd/sync-instagram-cache/main.go

cache-remove:
	@if [ -z "$(IDS)" ]; then echo "Usage: make cache-remove IDS=\"3340,3336\""; exit 1; fi
	@echo "🗑️  Removing tournaments from cache..."
	@echo ""
	cd api && go run cmd/sync-instagram-cache/main.go --remove "$(IDS)"

cache-sync:
	@echo "🔄 Syncing cache with Instagram/Threads APIs..."
	@echo "   (This will detect and remove deleted posts from cache)"
	@echo ""
	cd api && go run cmd/sync-cache/main.go

refresh-tournaments:
	@echo "🔄 Refreshing tournaments..."
	@echo ""
	cd api && go run cmd/refresh-tournaments/main.go