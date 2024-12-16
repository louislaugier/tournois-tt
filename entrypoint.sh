#!/bin/sh

# Print debug information
echo "Starting entrypoint script..."
echo "Contents of /usr/share/nginx/html:"
ls -la /usr/share/nginx/html

# Start Nginx in the background with debug logging
echo "Starting Nginx..."
nginx -g "daemon off;" &

# Start the API in the background
echo "Starting API..."
cd /app/api
./api &

# Wait for any background process to exit
wait -n

# Exit with the status of the process that exited first
exit $?