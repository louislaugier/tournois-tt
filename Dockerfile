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

# Install necessary packages, Playwright dependencies, and poppler-utils for pdftotext
RUN apt-get update && apt-get install -y \
    nginx \
    nodejs \
    npm \
    wget \
    git \
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libdbus-1-3 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libcairo2 \
    libasound2 \
    libatspi2.0-0 \
    libwayland-client0 \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

# Install Go to install playwright browsers easier
RUN wget https://dl.google.com/go/go1.23.3.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz \
    && rm go1.23.3.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# Install Playwright browsers for Go
RUN git clone https://github.com/playwright-community/playwright-go.git \
    && cd playwright-go && go mod tidy && go run cmd/playwright/main.go install --with-deps chromium && cd ..

# Create /app/api directory
RUN mkdir -p /app/api

# Create necessary directories
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