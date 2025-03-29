#!/bin/bash
set -e

echo "Starting environment setup with enhanced diagnostics"
export DEBUG_LEVEL=verbose

# Create runtime directories
mkdir -p /tmp/runtime-dir
chmod 0700 /tmp/runtime-dir

# Configure shared memory for browser with increased size
if [ ! -d /dev/shm ]; then
  echo "WARNING: /dev/shm is not mounted, creating tmpfs for /dev/shm"
  mkdir -p /dev/shm
  mount -t tmpfs -o size=4g tmpfs /dev/shm
  echo "Created tmpfs with 4GB size for /dev/shm"
else
  echo "Using existing /dev/shm"
  df -h /dev/shm
fi

# Ensure proper permissions for shared memory
chmod 1777 /dev/shm
chmod 1777 /tmp
echo "Shared memory setup completed"

# Start Xvfb with better error handling and monitoring
echo "Starting Xvfb virtual display on :99 with improved configuration"
rm -f /tmp/.X99-lock || true
Xvfb :99 -screen 0 1920x1080x24 -ac +extension GLX +render -noreset &
XVFB_PID=$!
sleep 3

# Enhanced verification that Xvfb is running
if ! ps -p $XVFB_PID > /dev/null; then
  echo "ERROR: Xvfb failed to start - trying again with fallback configuration"
  Xvfb :99 -screen 0 1280x1024x24 -nolisten tcp -ac +extension RANDR &
  XVFB_PID=$!
  sleep 3
  
  if ! ps -p $XVFB_PID > /dev/null; then
    echo "CRITICAL ERROR: Xvfb failed to start even with fallback configuration"
    exit 1
  fi
fi

echo "Xvfb started successfully with PID $XVFB_PID"

# Basic validation of X server
if ! DISPLAY=:99 xdpyinfo >/dev/null 2>&1; then
  echo "WARNING: X server validation failed, but continuing anyway"
else
  echo "X server validation passed"
fi

# Enhanced environment configuration
export PLAYWRIGHT_BROWSERS_PATH=/usr/local/ms-playwright
export CONTAINER_RUNTIME=1
export DISPLAY=:99
export CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox
export XDG_RUNTIME_DIR=/tmp/runtime-dir

# Additional browser configuration for Docker environments
export PLAYWRIGHT_SKIP_BROWSER_VALIDATION=1
export DEBUG=pw:api
export PLAYWRIGHT_DOCKER=1
export PW_EXPERIMENTAL_SERVICE_WORKER_NETWORK_EVENTS=1
export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
export PWDEBUG=console

# Enhanced browser stability settings
export PLAYWRIGHT_BROWSER_WS_ENDPOINT_TIMEOUT=120000
export PLAYWRIGHT_CONNECTION_TIMEOUT=120000
export PLAYWRIGHT_BROWSER_LAUNCH_TIMEOUT=120000

# Run verification checks for the browser environment
echo "Development environment setup - Verifying browser environment..."
echo "Browser location: $PLAYWRIGHT_BROWSERS_PATH"
echo "Display: $DISPLAY"
echo "Running with container optimizations"

# Enhanced browser location checking
CHROMIUM_DIRS=$(find $PLAYWRIGHT_BROWSERS_PATH -name "chromium*" -type d 2>/dev/null || echo "None found")
echo "Chromium directories found: $CHROMIUM_DIRS"

CHROMIUM_DIR=$(find $PLAYWRIGHT_BROWSERS_PATH -name "chromium-*" -type d 2>/dev/null | head -n 1)
if [ -n "$CHROMIUM_DIR" ]; then
  echo "Found Chromium installation at: $CHROMIUM_DIR"
  CHROME_BINARY=$(find $CHROMIUM_DIR -name "*chrome*" -type f -executable 2>/dev/null | head -n 1)
  if [ -n "$CHROME_BINARY" ]; then
    echo "Chrome binary found at: $CHROME_BINARY"
    echo "Ensuring browser executables have proper permissions"
    find $CHROMIUM_DIR -name "*chrome*" -type f -exec chmod +x {} \; 2>/dev/null || true
    find $CHROMIUM_DIR -name "*chromium*" -type f -exec chmod +x {} \; 2>/dev/null || true
  else
    echo "WARNING: No chrome binary found in $CHROMIUM_DIR"
    ls -la $CHROMIUM_DIR
  fi
else
  echo "WARNING: No Chromium installation found at $PLAYWRIGHT_BROWSERS_PATH"
  echo "Available directories in PLAYWRIGHT_BROWSERS_PATH:"
  ls -la $PLAYWRIGHT_BROWSERS_PATH || echo "Cannot list directory"
fi

# Create browser sandbox if it doesn't exist
if [ ! -e "$CHROME_DEVEL_SANDBOX" ]; then
  echo "Creating chrome-devel-sandbox at $CHROME_DEVEL_SANDBOX"
  echo '#!/bin/bash
exec "$@"' > $CHROME_DEVEL_SANDBOX
  chmod 4755 $CHROME_DEVEL_SANDBOX
fi

# Verify system resources
echo "System resources:"
free -h
df -h

# Trap to ensure clean shutdown of background processes
trap 'echo "Shutting down Xvfb with PID $XVFB_PID"; kill $XVFB_PID 2>/dev/null || true' EXIT

# Execute the command passed to docker run
echo "Starting development server with enhanced environment..."
exec "$@" 