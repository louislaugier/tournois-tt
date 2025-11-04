#!/bin/bash

# Script to post multiple tournaments to Instagram
# Usage: ./scripts/post_multiple.sh 3340 3336

if [ $# -eq 0 ]; then
    echo "Usage: $0 <tournament_id1> <tournament_id2> ..."
    exit 1
fi

for id in "$@"; do
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo "  Posting Tournament $id"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    
    # Run the posting command
    go run cmd/post-instagram-full/main.go --id "$id"
    
    result=$?
    if [ $result -ne 0 ]; then
        echo "❌ Failed to post tournament $id (exit code: $result)"
        echo ""
        read -p "Continue with next tournament? (yes/no): " continue_response
        if [ "$continue_response" != "yes" ]; then
            echo "Stopped by user"
            exit 1
        fi
    else
        echo "✅ Successfully posted tournament $id"
    fi
    
    # Wait between posts to avoid rate limiting
    if [ "$id" != "${@: -1}" ]; then
        echo ""
        echo "⏳ Waiting 10 seconds before next tournament..."
        sleep 10
    fi
done

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ All tournaments processed!"
echo "════════════════════════════════════════════════════════════════"

