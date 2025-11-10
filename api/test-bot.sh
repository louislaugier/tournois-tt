#!/bin/bash

# Instagram Bot Test Script
# This script tests the Instagram follow/unfollow bot in the Docker environment

set -e

echo "🚀 Instagram Bot Test Script"
echo "=============================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found!"
    echo "Please create a .env file with the required environment variables:"
    echo "  INSTAGRAM_BOT_USERNAME=your_username"
    echo "  INSTAGRAM_BOT_PASSWORD=your_password"
    echo "  INSTAGRAM_BOT_TOTP_SECRET=your_totp_secret"
    echo "  INSTAGRAM_BOT_HEADLESS=true"
    exit 1
fi

# Source environment variables
set -a
source .env
set +a

# Check required variables
if [ -z "$INSTAGRAM_BOT_USERNAME" ] || [ -z "$INSTAGRAM_BOT_PASSWORD" ]; then
    echo "❌ Error: INSTAGRAM_BOT_USERNAME and INSTAGRAM_BOT_PASSWORD must be set in .env file"
    exit 1
fi

echo "✅ Environment variables loaded"
echo "   Username: $INSTAGRAM_BOT_USERNAME"
echo "   Headless: ${INSTAGRAM_BOT_HEADLESS:-true}"
echo ""

# Check if containers are running
if ! docker ps | grep -q tournois-tt-api; then
    echo "❌ Error: API container is not running"
    echo "Please start the containers with: docker-compose up -d"
    exit 1
fi

echo "✅ Docker containers are running"
echo ""

# Run the test inside the container
echo "🤖 Starting bot test..."
echo "=============================="

docker exec -it tournois-tt-api-1 /bin/sh -c "cd /go/src/tournois-tt/api && go run cmd/test-instagram-follow/main.go"

echo ""
echo "=============================="
echo "✅ Test completed!"
