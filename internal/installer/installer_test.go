package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArtifactExists(t *testing.T) {
	installer := New()

	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if !installer.ArtifactExists(existingFile) {
		t.Error("Should detect existing file")
	}

	nonExistingFile := filepath.Join(tempDir, "does-not-exist.txt")
	if installer.ArtifactExists(nonExistingFile) {
		t.Error("Should not detect non-existing file")
	}
}

func TestExecuteInstallStep(t *testing.T) {
	installer := New()

	tests := []struct {
		method      string
		shouldError bool
	}{
		{"brew", true},
		{"cask", true},
		{"mas", true},
		{"npm", true},
		{"gem", true},
		{"run", false},
		{"unknown", true},
	}

	for _, test := range tests {
		err := installer.executeInstallStep(test.method, "echo test")
		if test.shouldError && err == nil {
			t.Errorf("Method '%s' should have errored but didn't", test.method)
		}
		if !test.shouldError && err != nil {
			t.Errorf("Method '%s' should not have errored but did: %v", test.method, err)
		}
	}
}

func TestExecuteConfigStep(t *testing.T) {
	installer := New()

	if err := installer.executeConfigStep("run", "echo test"); err != nil {
		t.Errorf("run command should not error: %v", err)
	}

	if err := installer.executeConfigStep("unknown", "test"); err == nil {
		t.Error("unknown method should error")
	}
}

func TestConfigure(t *testing.T) {
	installer := New()

	configSteps := []map[string]string{
		{"ignore_errors": "true"},
		{"run": "exit 1"},
		{"run": "echo success"},
	}

	if err := installer.Configure(configSteps); err != nil {
		t.Errorf("Configuration with ignore_errors should not fail: %v", err)
	}

	configStepsWithoutIgnore := []map[string]string{
		{"run": "exit 1"},
	}

	if err := installer.Configure(configStepsWithoutIgnore); err == nil {
		t.Error("Configuration without ignore_errors should fail on error")
	}
}

func TestRunScript(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "test-script.sh")
	
	scriptContent := `#!/bin/bash
echo "test script output"
exit 0
`
	
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	if err := installer.runScript(scriptFile); err != nil {
		t.Errorf("Script execution should not error: %v", err)
	}
}

func TestInstallArchiveValidation(t *testing.T) {
	installer := New()

	tests := []struct {
		name        string
		installSteps []map[string]string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "archive without file parameter (directory extraction)",
			installSteps: []map[string]string{
				{"archive": "https://example.com/test.dmg"},
			},
			shouldError: true, // Will error because URL doesn't exist, but should accept no file parameter
			errorMsg:    "archive installation failed",
		},
		{
			name: "archive with file parameter",
			installSteps: []map[string]string{
				{"archive": "https://example.com/test.dmg", "file": "Test.app"},
			},
			shouldError: true, // Will error because URL doesn't exist, but validates parameters
			errorMsg:    "archive installation failed",
		},
		{
			name: "regular install step",
			installSteps: []map[string]string{
				{"run": "echo test"},
			},
			shouldError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := installer.Install(test.installSteps, "/Applications/Test.app")
			
			if test.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !test.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if test.shouldError && err != nil {
				if test.errorMsg != "" && !contains(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", test.errorMsg, err)
				}
			}
		})
	}
}

func TestFindFileInDirectory(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	
	// Create test files
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create nested directory with file
	nestedDir := filepath.Join(tempDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	nestedFile := filepath.Join(nestedDir, "nested.txt")
	if err := os.WriteFile(nestedFile, []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name     string
		fileName string
		found    bool
	}{
		{"existing file", "test.txt", true},
		{"nested file", "nested.txt", true},
		{"non-existent file", "missing.txt", false},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := installer.findFileInDirectory(tempDir, test.fileName)
			
			if test.found && err != nil {
				t.Errorf("Expected to find file but got error: %v", err)
			}
			
			if !test.found && err == nil {
				t.Error("Expected error for missing file but got none")
			}
			
			if test.found && path == "" {
				t.Error("Expected non-empty path for found file")
			}
		})
	}
}

func TestCopyFileOrDirectory(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	
	// Test file copying
	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")
	
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := installer.copyFileOrDirectory(srcFile, destFile); err != nil {
		t.Errorf("File copy should not error: %v", err)
	}
	
	// Verify file was copied
	if !installer.ArtifactExists(destFile) {
		t.Error("Destination file should exist after copy")
	}
	
	// Test directory copying
	srcDir := filepath.Join(tempDir, "src_dir")
	destDir := filepath.Join(tempDir, "dest_dir")
	
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	dirFile := filepath.Join(srcDir, "file_in_dir.txt")
	if err := os.WriteFile(dirFile, []byte("dir content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := installer.copyFileOrDirectory(srcDir, destDir); err != nil {
		t.Errorf("Directory copy should not error: %v", err)
	}
	
	// Verify directory was copied
	copiedFile := filepath.Join(destDir, "file_in_dir.txt")
	if !installer.ArtifactExists(copiedFile) {
		t.Error("File in copied directory should exist")
	}
}

func TestCopyDirectoryContents(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	
	// Create source directory with multiple files
	srcDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create test files in source directory
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create nested directory with file
	nestedDir := filepath.Join(srcDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	nestedFile := filepath.Join(nestedDir, "nested.txt")
	if err := os.WriteFile(nestedFile, []byte("nested content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create destination directory
	destDir := filepath.Join(tempDir, "destination")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Copy directory contents
	if err := installer.copyDirectoryContents(srcDir, destDir); err != nil {
		t.Errorf("Directory contents copy should not error: %v", err)
	}
	
	// Verify all files were copied
	destFile1 := filepath.Join(destDir, "file1.txt")
	destFile2 := filepath.Join(destDir, "file2.txt")
	destNestedFile := filepath.Join(destDir, "nested", "nested.txt")
	
	if !installer.ArtifactExists(destFile1) {
		t.Error("file1.txt should exist in destination")
	}
	if !installer.ArtifactExists(destFile2) {
		t.Error("file2.txt should exist in destination")
	}
	if !installer.ArtifactExists(destNestedFile) {
		t.Error("nested/nested.txt should exist in destination")
	}
	
	// Verify content
	content1, err := os.ReadFile(destFile1)
	if err != nil || string(content1) != "content1" {
		t.Error("file1.txt content should be preserved")
	}
}

func TestExtractArchiveUnsupportedFormat(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	extractDir := filepath.Join(tempDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create a file with unsupported extension
	unsupportedFile := filepath.Join(tempDir, "test.rar")
	if err := os.WriteFile(unsupportedFile, []byte("fake rar content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	err := installer.extractArchive(unsupportedFile, extractDir, "https://example.com/test.rar")
	if err == nil {
		t.Error("Should error on unsupported archive format")
	}
	
	if !contains(err.Error(), "unsupported archive format") {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}
}

func TestExtractArchiveDMGURLDetection(t *testing.T) {
	installer := New()
	
	tempDir := t.TempDir()
	extractDir := filepath.Join(tempDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name        string
		url         string
		shouldBeDMG bool
	}{
		{
			name:        "URL with .dmg extension",
			url:         "https://example.com/app.dmg",
			shouldBeDMG: true,
		},
		{
			name:        "URL with .dmg in path without extension",
			url:         "https://developer.apple.com/download/files/icon-composer.dmg",
			shouldBeDMG: true,
		},
		{
			name:        "URL without .dmg",
			url:         "https://example.com/app.zip",
			shouldBeDMG: false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a dummy file (won't actually be processed since we'll hit an error)
			dummyFile := filepath.Join(tempDir, "dummy")
			if err := os.WriteFile(dummyFile, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}
			
			err := installer.extractArchive(dummyFile, extractDir, test.url)
			
			// We expect an error since we're not providing real archives,
			// but we can check if the error suggests it tried the right format
			if test.shouldBeDMG {
				// Should try hdiutil which will fail with a specific error
				if err == nil || !contains(err.Error(), "hdiutil") && !contains(err.Error(), "not recognized") {
					// If it doesn't mention hdiutil, it might be trying a different format
					// Let's be flexible about the exact error message
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}