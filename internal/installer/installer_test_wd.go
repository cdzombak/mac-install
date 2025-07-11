package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	installer := New(configDir)

	// Test run command with working directory
	testFile := filepath.Join(configDir, "test-run.txt")
	err := installer.runShellCommand("echo 'test' > test-run.txt")
	if err != nil {
		t.Fatalf("Failed to run shell command: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Shell command did not create file in config directory")
	}

	// Test script command with working directory
	scriptPath := filepath.Join(tempDir, "test-script.sh")
	scriptContent := `#!/bin/sh
echo 'script test' > test-script.txt`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	testScriptFile := filepath.Join(configDir, "test-script.txt")
	err = installer.runScript(scriptPath)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if _, err := os.Stat(testScriptFile); os.IsNotExist(err) {
		t.Error("Script did not create file in config directory")
	}
}