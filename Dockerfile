# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go mod tidy && go build -o /excalidraw-go-file-store

# Final stage
FROM alpine:latest

# Add non root user
RUN adduser -D -g '' appuser

# Create data directory and set permissions
RUN mkdir /data && chown appuser:appuser /data

# Copy binary from builder
COPY --from=builder /excalidraw-go-file-store /excalidraw-go-file-store

# Use non root user
USER appuser

# Set data directory environment variable
ENV DATA_DIR=/data

# Expose the default port
EXPOSE 8080

# Run the application
CMD ["/excalidraw-go-file-store"] 