# Excalidraw Go File Store

A simple, efficient binary data storage server implemented in Go. This server provides endpoints for storing and retrieving binary data, designed to work with Excalidraw or similar applications.

## Features

- Pure Go implementation using standard library
- No external dependencies except for logging
- File-based persistent storage
- CORS support
- Simple API endpoints
- Structured logging with slog and prettyslog
- Configurable via environment variables

## API Endpoints

### Upload Binary Data
- **URL:** `/api/v2/post/`
- **Method:** `POST`
- **Max File Size:** 50MB
- **Response:** JSON containing `dataKey` and `url` for retrieval

### Retrieve Binary Data
- **URL:** `/api/v2/{dataKey}`
- **Method:** `GET`
- **Response:** Binary data stream

## Configuration

The server can be configured using the following environment variables:
- `PORT`: Server port (default: 8080)
- `DATA_DIR`: Directory for storing files (default: ./data)

## Running the Server

1. Ensure you have Go 1.21 or later installed
2. Clone the repository
3. Run the server:
   ```bash
   go run main.go
   ```

The server will create a data directory (specified by DATA_DIR) to store uploaded files.

## Docker Support

Build and run with Docker:

```bash
# Build the image
docker build -t excalidraw-go-file-store .

# Run the container
docker run -p 8080:8080 -v /path/to/host/data:/data excalidraw-go-file-store
```

## Development

The server uses:
- File-based storage system with unique identifiers
- Thread-safe operations
- Structured logging with slog and prettyslog for better debugging
- Docker multi-stage builds for minimal image size

## Security Considerations

- CORS is implemented but should be configured for your specific use case
- File size is limited to 50MB by default
- Comprehensive error handling and logging
- Production deployments should consider additional security measures
- Runs as non-root user in Docker 