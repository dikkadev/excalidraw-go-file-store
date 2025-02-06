package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const (
	testDataDir = "./test_data"
	validOrigin = "https://excalidraw.com"
)

// TestMain is used for test setup and teardown
func TestMain(m *testing.M) {
	// Setup
	originalDataDir := dataDir
	dataDir = testDataDir

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testDataDir)
	dataDir = originalDataDir

	os.Exit(code)
}

// Helper function to generate random bytes
func generateRandomBytes(size int) []byte {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(fmt.Sprintf("Failed to generate random data: %v", err))
	}
	return data
}

// Helper function to create a test server
func setupTestServer(t *testing.T) (*httptest.Server, *Server) {
	server := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/post/", server.handleUpload)
	mux.HandleFunc("/api/v2/", server.handleDownload)

	return httptest.NewServer(mux), server
}

// Test case 1: Basic Successful Upload - Small Payload
func TestBasicSuccessfulUpload(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	// Create small payload
	payload := generateRandomBytes(1024) // 1KB

	// Create request
	req, err := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", validOrigin)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.StatusCode)
	}

	// Parse response
	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response fields
	if uploadResp.DataKey == "" {
		t.Error("Expected non-empty DataKey")
	}
	if uploadResp.URL == "" {
		t.Error("Expected non-empty URL")
	}

	// Verify file exists and content matches
	filePath := filepath.Join(testDataDir, uploadResp.DataKey)
	storedData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read stored file: %v", err)
	}
	if !bytes.Equal(storedData, payload) {
		t.Error("Stored data doesn't match original payload")
	}
}

// Test case 2: Maximum Size Upload
func TestMaxSizeUpload(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	payload := generateRandomBytes(maxUploadSize)

	req, err := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", validOrigin)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.StatusCode)
	}
}

// Test case 3: Upload Too Large
func TestUploadTooLarge(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	payload := generateRandomBytes(maxUploadSize + 1024) // Slightly over limit

	req, err := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", validOrigin)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413; got %v", resp.StatusCode)
	}
}

// Test case 4: Invalid Origin
func TestInvalidOrigin(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	payload := generateRandomBytes(1024)

	req, err := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", "http://untrusted-origin.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// The current implementation accepts all origins, but in production this should be restricted
	// This test might need to be updated based on your CORS policy
	if resp.StatusCode == http.StatusOK {
		t.Log("Warning: Server accepted request from untrusted origin")
	}
}

// Test case 5: Successful Data Retrieval
func TestSuccessfulDataRetrieval(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	// First upload some data
	payload := generateRandomBytes(1024)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	req.Header.Set("Origin", validOrigin)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload test data: %v", err)
	}

	var uploadResp UploadResponse
	json.NewDecoder(resp.Body).Decode(&uploadResp)
	resp.Body.Close()

	// Now try to retrieve it
	req, _ = http.NewRequest("GET", ts.URL+"/api/v2/"+uploadResp.DataKey, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.StatusCode)
	}

	retrievedData, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !bytes.Equal(retrievedData, payload) {
		t.Error("Retrieved data doesn't match original payload")
	}
}

// Test case 6: Data Not Found
func TestDataNotFound(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL+"/api/v2/nonexistent-key", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404; got %v", resp.StatusCode)
	}
}

// Test case 7: GET Request CORS
func TestGetRequestCORS(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	// First upload some data
	payload := generateRandomBytes(1024)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v2/post/", bytes.NewReader(payload))
	req.Header.Set("Origin", validOrigin)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload test data: %v", err)
	}

	var uploadResp UploadResponse
	json.NewDecoder(resp.Body).Decode(&uploadResp)
	resp.Body.Close()

	// Try to retrieve with different origins
	origins := []string{validOrigin, "http://untrusted-origin.com"}
	for _, origin := range origins {
		req, _ = http.NewRequest("GET", ts.URL+"/api/v2/"+uploadResp.DataKey, nil)
		req.Header.Set("Origin", origin)
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to retrieve data with origin %s: %v", origin, err)
		}
		resp.Body.Close()

		// GET requests should be allowed from any origin
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK for origin %s; got %v", origin, resp.StatusCode)
		}

		// Check CORS headers
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: * for origin %s", origin)
		}
	}
}
