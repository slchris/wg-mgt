# Build stage for frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copy frontend package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ ./

# Build frontend (outputs to ../cmd/wg-mgt/web/dist)
RUN mkdir -p ../cmd/wg-mgt/web && npm run build


# Build stage for backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set Go proxy for faster downloads (use Chinese mirror if needed)
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from frontend stage
COPY --from=frontend-builder /app/cmd/wg-mgt/web/dist ./cmd/wg-mgt/web/dist

# Build the binary with static linking for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o wg-mgt ./cmd/wg-mgt


# Final stage - minimal runtime image
FROM golang:1.24-alpine AS runtime

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binary from builder
COPY --from=backend-builder /app/wg-mgt .

# Copy default config if exists
COPY --from=backend-builder /app/configs ./configs

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Environment variables
ENV WG_MGT_DB_PATH=/app/data/wg-mgt.db
ENV WG_MGT_PORT=8080
ENV WG_MGT_HOST=0.0.0.0

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./wg-mgt"]
