#!/bin/bash
set -e

# Create runtime directories
mkdir -p /tmp/runtime-dir
chmod 0700 /tmp/runtime-dir

# Configure shared memory for browser
if [ ! -d /dev/shm ]; then
  echo "WARNING: /dev/shm is not mounted, creating tmpfs for /dev/shm"
  mkdir -p /dev/shm
  mount -t tmpfs -o size=512m tmpfs /dev/shm
fi

# Ensure proper permissions
chmod 1777 /dev/shm
chmod 1777 /tmp

# Start Nginx in background
echo "Starting nginx..."
nginx -g "daemon off;" &

# Wait for Nginx to start
sleep 2

# Set environment variables for browser
export PLAYWRIGHT_BROWSERS_PATH=/usr/local/ms-playwright
export CONTAINER_RUNTIME=1
export CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox
export XDG_RUNTIME_DIR=/tmp/runtime-dir

# Give feedback about the environment
echo "Container environment ready."
echo "Browser location: $PLAYWRIGHT_BROWSERS_PATH"
echo "Running with container optimizations"

# Start API service
echo "Starting API service..."
cd /app/api && ./api

# Keep container running
wait