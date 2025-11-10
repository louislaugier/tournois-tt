#!/bin/bash

# Quick Commands for Instagram Bot

# 1. RESTART CONTAINERS
echo "🔄 Restarting containers..."
docker-compose down
docker-compose up -d
echo "✅ Containers restarted"

# 2. TEST BOT
echo ""
echo "🤖 Testing bot..."
cd api && ./test-bot.sh

# 3. CHECK LOGS
echo ""
echo "📋 To watch logs, run:"
echo "   docker logs -f tournois-tt-api-1"

# 4. CHECK BOT STATE
echo ""
echo "📊 To check bot state, run:"
echo "   docker exec tournois-tt-api-1 cat /go/src/tournois-tt/api/tmp/bot_data/bot_state.json"
