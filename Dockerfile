# Multi-stage build for tournois-tt project

# API build stage
FROM --platform=linux/amd64 golang:1.23 AS api-build
WORKDIR /go/src/tournois-tt/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/api ./cmd/main.go

# Frontend build stage
FROM --platform=linux/amd64 node:20.18.1-alpine AS frontend-build
WORKDIR /app/frontend

# Install global dependencies
RUN npm install -g npm@latest

# Copy package files first for better caching
COPY frontend/package*.json ./

# Install dependencies
RUN npm i

# Copy the rest of the frontend files
COPY frontend ./

# Build the frontend with production config
RUN npm run build

# Production stage
FROM --platform=linux/amd64 debian:bullseye-slim
WORKDIR /app

# Install necessary packages for stable Chromium operation
RUN apt-get update && apt-get install -y \
    nginx \
    nodejs \
    npm \
    wget \
    git \
    ca-certificates \
    fonts-liberation \
    fonts-noto-color-emoji \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libc6 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libexpat1 \
    libfontconfig1 \
    libgbm1 \
    libgcc1 \
    libglib2.0-0 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libpulse0 \
    libsoup2.4-1 \
    libwayland-client0 \
    libwayland-cursor0 \
    libwayland-egl1 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxkbcommon0 \
    libxrandr2 \
    libxrender1 \
    libxshmfence1 \
    libxss1 \
    libxtst6 \
    lsb-release \
    procps \
    xdg-utils \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

# Configure environment for stable browser operation
ENV PLAYWRIGHT_BROWSERS_PATH=/usr/local/ms-playwright \
    PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0 \
    DEBIAN_FRONTEND=noninteractive \
    DISPLAY=:99 \
    XDG_RUNTIME_DIR=/tmp/runtime-dir \
    CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox \
    CONTAINER_RUNTIME=1

# Setup directories for Playwright
RUN mkdir -p /usr/local/ms-playwright \
    && mkdir -p ${XDG_RUNTIME_DIR} \
    && chmod 0700 ${XDG_RUNTIME_DIR}

# Install Go to install playwright browsers easier
RUN wget https://dl.google.com/go/go1.23.3.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz \
    && rm go1.23.3.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# Install Playwright browsers for Go with more robust setup
RUN git clone https://github.com/playwright-community/playwright-go.git \
    && cd playwright-go && go mod tidy && go run cmd/playwright/main.go install --with-deps chromium \
    && cd .. \
    && rm -rf playwright-go

# Create /app/api directory and other necessary directories
RUN mkdir -p /app/api/cache

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /usr/share/nginx/html

# Copy built API binary
COPY --from=api-build /go/bin/api /app/api

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy necessary configuration files and cache
COPY api/cache/ /app/api/cache/

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose ports
EXPOSE 80

# Use the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]