# Multi-stage build for tournois-tt project

# API build stage
FROM golang:1.23-alpine AS api-build
WORKDIR /go/src/tournois-tt/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/api ./cmd/main.go

# Frontend build stage
FROM node:18-alpine AS frontend-build
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

# Install necessary packages
RUN apk add --no-cache nginx nodejs npm

# Create /app/api directory
RUN mkdir -p /app/api

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /usr/share/nginx/html
# Verify the contents of the build directory
RUN echo "Contents of /usr/share/nginx/html:" \
    && ls -la /usr/share/nginx/html \
    && echo "Contents of index.html:" \
    && cat /usr/share/nginx/html/index.html

# Copy built API binary
COPY --from=api-build /go/bin/api /app/api

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy necessary configuration files
COPY api/.env.example /app/api/.env
COPY api/air.toml /app/api/air.toml

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose ports
EXPOSE 80

# Use the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]