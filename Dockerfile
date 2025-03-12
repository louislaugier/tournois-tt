# Multi-stage build for tournois-tt project

# API build stage
FROM golang:1.23-alpine AS api-build
WORKDIR /go/src/tournois-tt/api

# Install build dependencies for Playwright
RUN apk add --no-cache \
    build-base \
    python3

COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./

# Install Playwright browsers during build
RUN go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps chromium

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/api ./cmd/main.go

# Frontend build stage
FROM node:20.18.1-alpine AS frontend-build
WORKDIR /app/frontend

# Install global dependencies
RUN npm install -g npm@latest

# Copy package files first for better caching
COPY frontend/package*.json ./

# Install dependencies with verbose output
RUN npm ci --verbose \
    || (echo "npm ci failed, trying npm install" && npm install --verbose)

# Copy the rest of the frontend files
COPY frontend ./

# Ensure node_modules is populated
RUN npm ls || true

# Build the frontend with production config
RUN npm run build \
    && echo "Build completed successfully" \
    && echo "Contents of build directory:" \
    && ls -la build \
    && echo "Contents of index.html:" \
    && cat build/index.html \
    || (echo "Build failed. Showing npm logs:" && cat /root/.npm/_logs/$(ls -t /root/.npm/_logs/ | head -n1) && exit 1)

# Production stage
FROM alpine:latest
WORKDIR /app

# Install necessary packages and Playwright dependencies
RUN apk add --no-cache \
    nginx \
    nodejs \
    npm \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    font-noto-emoji \
    # Playwright dependencies
    libx11 \
    libxcomposite \
    libxdamage \
    libxext \
    libxfixes \
    libxi \
    libxrandr \
    libxrender \
    libxss \
    libxtst \
    mesa-gbm \
    pango \
    cairo \
    alsa-lib \
    at-spi2-core \
    dbus-libs \
    eudev-libs \
    libxcb \
    libxkbcommon \
    wayland-libs-client

# Create /app/api directory
RUN mkdir -p /app/api

# Create necessary directories
RUN mkdir -p /app/api/cache

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /usr/share/nginx/html
# Verify the contents of the build directory
RUN echo "Contents of /usr/share/nginx/html:" \
    && ls -la /usr/share/nginx/html \
    && echo "Contents of index.html:" \
    && cat /usr/share/nginx/html/index.html

# Copy built API binary and Playwright browsers
COPY --from=api-build /go/bin/api /app/api
COPY --from=api-build /root/.cache/ms-playwright /root/.cache/ms-playwright

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy necessary configuration files and cache
COPY api/.env.example /app/api/.env
COPY api/air.toml /app/api/air.toml
COPY api/cache/geocoding_cache.json /app/api/cache/

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set environment variables for Playwright
ENV PLAYWRIGHT_BROWSERS_PATH=/root/.cache/ms-playwright
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

# Expose ports
EXPOSE 80

# Use the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]