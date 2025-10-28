#!/bin/bash

echo "🔑 Facebook Page Access Token Generator"
echo "======================================"
echo ""

# Get user token from environment or prompt
if [ -z "$FACEBOOK_USER_TOKEN" ]; then
    read -p "Enter your Facebook USER access token: " FACEBOOK_USER_TOKEN
fi

echo ""
echo "📡 Fetching your pages..."
echo ""

# Call the API to get all pages
response=$(curl -s "https://graph.facebook.com/me/accounts?access_token=$FACEBOOK_USER_TOKEN")

echo "Raw response:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# Extract the page token for page ID 61582857840582
page_token=$(echo "$response" | jq -r '.data[] | select(.id == "61582857840582") | .access_token' 2>/dev/null)

if [ -z "$page_token" ] || [ "$page_token" == "null" ]; then
    echo "❌ Could not find page token for ID: 61582857840582"
    echo ""
    echo "Your pages:"
    echo "$response" | jq -r '.data[] | "  - \(.name) (ID: \(.id))"' 2>/dev/null || echo "$response"
    exit 1
fi

echo "✅ Found your page!"
echo ""
echo "📋 Page Access Token:"
echo "===================="
echo "$page_token"
echo ""
echo "📝 Add this to your .env file:"
echo "   FACEBOOK_ACCESS_TOKEN=$page_token"
echo ""

