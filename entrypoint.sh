#!/bin/sh

# Ensure Playwright directories exist
mkdir -p /root/.cache/ms-playwright

# Check permissions on Playwright directories
echo "Checking permissions on Playwright directories:"
ls -la $PLAYWRIGHT_BROWSERS_PATH
CHROMIUM_DIR=$(find $PLAYWRIGHT_BROWSERS_PATH -name "chromium-*" -type d | head -n 1)
if [ -n "$CHROMIUM_DIR" ]; then
    ls -la $CHROMIUM_DIR
    if [ -d "$CHROMIUM_DIR/chrome-linux" ]; then
        echo "Ensuring chrome executable has proper permissions:"
        ls -la $CHROMIUM_DIR/chrome-linux/chrome*
        chmod +x $CHROMIUM_DIR/chrome-linux/chrome
        chmod +x $CHROMIUM_DIR/chrome-linux/chrome_sandbox 2>/dev/null || true
    fi
else
    echo "No chromium directory found"
fi

# Create symbolic links to ensure compatibility with expected versions
echo "Creating compatibility symlinks for Playwright drivers..."
mkdir -p /root/.cache/ms-playwright-go/1.50.1

# Check if there's a browser to link
CHROME_BINARY=$(find /root/.cache/ms-playwright -name chrome -type f -executable | head -n 1)
if [ -n "$CHROME_BINARY" ]; then
    echo "Linking browser binary: $CHROME_BINARY"
    PARENT_DIR=$(dirname "$CHROME_BINARY")
    ln -s "$PARENT_DIR" /root/.cache/ms-playwright-go/1.50.1/
    echo "Created symlink: /root/.cache/ms-playwright-go/1.50.1/ -> $PARENT_DIR"
else
    echo "Warning: No Chrome binary found to link"
fi

# Set additional Playwright environment variables
export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1
export DEBUG=pw:api

# Start Nginx in the background with debug logging
echo "Starting Nginx..."
nginx -g "daemon off;" &

# Start the API in the background
echo "Starting API..."
cd /app/api
./api &

# Wait for any background process to exit
wait %1

# Exit with the status of the process that exited first
exit $?