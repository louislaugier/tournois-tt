# Multi-stage build for tournois-tt project

# Frontend build stage
FROM node:18-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend ./
RUN npm run build --verbose || (cat /root/.npm/_logs/$(ls -t /root/.npm/_logs/ | head -n1) && exit 1)

# API build stage
FROM golang:1.23-alpine AS api-build
WORKDIR /go/src/tournois-tt/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/api ./cmd/main.go

# Production stage
FROM alpine:latest
WORKDIR /app

# Install nginx
RUN apk add --no-cache nginx

# Create /app/api directory
RUN mkdir -p /app/api

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /usr/share/nginx/html

# Copy built API binary
COPY --from=api-build /go/bin/api /app/api

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy necessary configuration files
COPY api/.env.example /app/api/.env
COPY api/air.toml /app/api/air.toml

# Expose ports
EXPOSE 80

# Create startup script
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'nginx' >> /start.sh && \
    echo 'cd /app/api && ./api &' >> /start.sh && \
    echo 'wait' >> /start.sh && \
    chmod +x /start.sh && \
    chmod +x /app/api

# Command to run the application
CMD ["/start.sh"]