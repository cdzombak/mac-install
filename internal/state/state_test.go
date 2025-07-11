package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	store, err := NewStore()
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	expectedStateDir := filepath.Join(homeDir, ".config", "dotfiles", "software")
	if store.stateDir != expectedStateDir {
		t.Errorf("Expected state dir '%s', got '%s'", expectedStateDir, store.stateDir)
	}

	if _, err := os.Stat(expectedStateDir); os.IsNotExist(err) {
		t.Error("State directory was not created")
	}
}

func TestIsExcluded(t *testing.T) {
	tempDir := t.TempDir()
	store := &Store{stateDir: tempDir}

	if store.IsExcluded("test-software") {
		t.Error("Software should not be excluded initially")
	}

	flagFile := filepath.Join(tempDir, "no-test-software")
	if err := os.WriteFile(flagFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	if !store.IsExcluded("test-software") {
		t.Error("Software should be excluded after creating flag file")
	}
}

func TestSetExcluded(t *testing.T) {
	tempDir := t.TempDir()
	store := &Store{stateDir: tempDir}

	if err := store.SetExcluded("test-software"); err != nil {
		t.Fatalf("Failed to set excluded: %v", err)
	}

	if !store.IsExcluded("test-software") {
		t.Error("Software should be excluded after calling SetExcluded")
	}

	flagFile := filepath.Join(tempDir, "no-test-software")
	if _, err := os.Stat(flagFile); os.IsNotExist(err) {
		t.Error("Flag file was not created")
	}
}

func TestNormalizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test Software", "test-software"},
		{"Test/Path", "test-path"},
		{"UPPERCASE", "uppercase"},
		{"Mixed Case With Spaces", "mixed-case-with-spaces"},
		{"Boop.app", "boop"},
		{"Test.App", "test"},
		{"NoExtension", "noextension"},
	}

	for _, test := range tests {
		result := normalizeFilename(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func TestGetExclusionFilePath(t *testing.T) {
	tempDir := t.TempDir()
	store := &Store{stateDir: tempDir}

	tests := []struct {
		softwareName string
		expected     string
	}{
		{"Boop.app", filepath.Join(tempDir, "no-boop")},
		{"Test Software", filepath.Join(tempDir, "no-test-software")},
		{"Simple", filepath.Join(tempDir, "no-simple")},
	}

	for _, test := range tests {
		result := store.GetExclusionFilePath(test.softwareName)
		if result != test.expected {
			t.Errorf("For software '%s', expected '%s', got '%s'", test.softwareName, test.expected, result)
		}
	}
}