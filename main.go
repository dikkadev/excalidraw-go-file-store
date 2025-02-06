package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dikkadev/prettyslog"
)

const (
	maxUploadSize = 50 * 1024 * 1024 // 50MB max file size
)

var (
	dataDir = getDataDir() // Directory to store files
)

func init() {
	// Initialize the logger
	handler := prettyslog.NewPrettyslogHandler("excalidraw-go-file-store",
		prettyslog.WithLevel(slog.LevelDebug),
	)
	slog.SetDefault(slog.New(handler))
}

// getDataDir returns the data directory from environment variable or default
func getDataDir() string {
	if dir := os.Getenv("DATA_DIR"); dir != "" {
		return dir
	}
	return "./data"
}

type Server struct {
	mu sync.RWMutex
}

type UploadResponse struct {
	DataKey string `json:"dataKey"`
	URL     string `json:"url"`
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// CORS headers for upload
	origin := r.Header.Get("Origin")
	// In production, you should validate the origin against a whitelist
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("failed to create data directory", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Generate a unique filename
	dataKey := generateUniqueID()
	filePath := filepath.Join(dataDir, dataKey)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("failed to create file", "error", err, "path", filePath)
		http.Error(w, "Could not create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy the data
	if _, err := io.Copy(file, r.Body); err != nil {
		slog.Error("failed to copy data", "error", err)
		os.Remove(filePath) // Clean up on error

		// Check if the error is due to request body being too large
		if err.Error() == "http: request body too large" {
			http.Error(w, fmt.Sprintf("File too large. Maximum size is %d bytes", maxUploadSize), http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	// Construct the response
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/api/v2/%s", scheme, r.Host, dataKey)

	response := UploadResponse{
		DataKey: dataKey,
		URL:     url,
	}

	slog.Info("file uploaded successfully",
		"dataKey", dataKey,
		"size", r.ContentLength,
		"origin", origin,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// CORS headers for download - more permissive
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	// Extract the dataKey from the URL path
	dataKey := filepath.Base(r.URL.Path)
	filePath := filepath.Join(dataDir, dataKey)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Warn("file not found", "dataKey", dataKey)
		http.Error(w, "Could not find the file", http.StatusNotFound)
		return
	}

	// Open and serve the file
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("failed to open file", "error", err, "path", filePath)
		http.Error(w, "Could not read the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, file); err != nil {
		slog.Error("failed to send file", "error", err, "dataKey", dataKey)
	} else {
		slog.Info("file downloaded successfully", "dataKey", dataKey)
	}
}

func generateUniqueID() string {
	// Simple implementation - in production you might want something more sophisticated
	return fmt.Sprintf("%d", os.Getpid()) + fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	slog.Info("starting excalidraw-go-file-store server",
		"dataDir", dataDir,
		"maxUploadSize", maxUploadSize,
	)

	server := &Server{}

	// Set up routes
	http.HandleFunc("/api/v2/post/", server.handleUpload)
	http.HandleFunc("/api/v2/", server.handleDownload)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server listening", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
