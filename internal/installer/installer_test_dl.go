package installer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDL(t *testing.T) {
	tempDir := t.TempDir()
	installer := New(tempDir)

	// Create a test HTTP server
	testContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// Test downloading to a file
	targetFile := filepath.Join(tempDir, "downloaded-file.txt")
	installSteps := []map[string]string{
		{"dl": server.URL + "/testfile.txt"},
	}

	err := installer.Install(installSteps, targetFile)
	if err != nil {
		t.Fatalf("Install with dl should not error: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("Downloaded file should exist")
	}

	// Verify the content
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}
}

func TestInstallDLWithDirectory(t *testing.T) {
	tempDir := t.TempDir()
	installer := New(tempDir)

	// Create a test HTTP server
	testContent := "nested file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// Test downloading to a file in a nested directory
	targetFile := filepath.Join(tempDir, "nested", "dir", "downloaded-file.txt")
	installSteps := []map[string]string{
		{"dl": server.URL + "/testfile.txt"},
	}

	err := installer.Install(installSteps, targetFile)
	if err != nil {
		t.Fatalf("Install with dl to nested directory should not error: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("Downloaded file should exist in nested directory")
	}

	// Verify the content
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}
}

func TestInstallDLServerError(t *testing.T) {
	tempDir := t.TempDir()
	installer := New(tempDir)

	// Create a test HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	targetFile := filepath.Join(tempDir, "failed-download.txt")
	installSteps := []map[string]string{
		{"dl": server.URL + "/nonexistent.txt"},
	}

	err := installer.Install(installSteps, targetFile)
	if err == nil {
		t.Error("Install with dl should error when server returns 404")
	}

	// Verify the file was not created
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Error("File should not exist when download fails")
	}
}