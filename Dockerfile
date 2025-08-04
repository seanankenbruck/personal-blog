# Multi-stage build for Go application
FROM golang:1.24.5-alpine AS builder

# Install git and ca-certificates (needed for private repos and HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached unless go.mod/go.sum changes)
RUN go mod download && go mod verify

# Copy only the source code directories that are needed for building
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY static/ ./static/
COPY templates/ ./templates/

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main cmd/main.go

# Final stage - minimal image
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy passwd file for non-root user
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /app/main /main

# Copy static files and templates
COPY --from=builder /app/static /static
COPY --from=builder /app/templates /templates

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/main"]