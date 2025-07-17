package checklist

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddSoftwareSteps(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	manager := New(checklistFile)

	steps := []string{"Step 1", "Step 2"}
	caveats := "Important caveats here"

	err := manager.AddSoftwareSteps("Test Software", "Test", steps, caveats)
	if err != nil {
		t.Fatalf("Failed to add software steps: %v", err)
	}

	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## Test") {
		t.Error("Header not found in checklist")
	}

	if !strings.Contains(contentStr, "- [ ] Step 1") {
		t.Error("Step 1 not found in checklist")
	}

	if !strings.Contains(contentStr, "- [ ] Step 2") {
		t.Error("Step 2 not found in checklist")
	}

	if !strings.Contains(contentStr, "### Caveats") {
		t.Error("Caveats section not found in checklist")
	}

	if !strings.Contains(contentStr, "Important caveats here") {
		t.Error("Caveats content not found in checklist")
	}
}

func TestAddSoftwareStepsIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	manager := New(checklistFile)

	steps := []string{"Step 1"}

	err := manager.AddSoftwareSteps("Test Software", "Test", steps, "")
	if err != nil {
		t.Fatalf("Failed to add software steps: %v", err)
	}

	err = manager.AddSoftwareSteps("Test Software", "Test", steps, "")
	if err != nil {
		t.Fatalf("Failed to add software steps (second time): %v", err)
	}

	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist file: %v", err)
	}

	contentStr := string(content)
	headerCount := strings.Count(contentStr, "## Test")

	if headerCount != 1 {
		t.Errorf("Expected 1 header, found %d", headerCount)
	}
}

func TestAddInstallStep(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	manager := New(checklistFile)

	err := manager.AddInstallStep("Test Software", "Test")
	if err != nil {
		t.Fatalf("Failed to add install step: %v", err)
	}

	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## Test") {
		t.Error("Header not found in checklist")
	}

	if !strings.Contains(contentStr, "- [ ] Install Test Software") {
		t.Error("Install step not found in checklist")
	}
}

func TestHeaderExists(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	initialContent := `# System Setup

## Existing App

- [ ] Existing step

## Another App

- [ ] Another step
`

	if err := os.WriteFile(checklistFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	manager := New(checklistFile)

	exists, err := manager.headerExists("Existing App")
	if err != nil {
		t.Fatalf("Failed to check header exists: %v", err)
	}

	if !exists {
		t.Error("Should detect existing header")
	}

	exists, err = manager.headerExists("Non-existing App")
	if err != nil {
		t.Fatalf("Failed to check header exists: %v", err)
	}

	if exists {
		t.Error("Should not detect non-existing header")
	}
}

func TestHeaderExistsPublic(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	initialContent := `# System Setup

## Test App

- [ ] Test step
`

	if err := os.WriteFile(checklistFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	manager := New(checklistFile)

	exists, err := manager.HeaderExists("Test App")
	if err != nil {
		t.Fatalf("Failed to check header exists: %v", err)
	}

	if !exists {
		t.Error("Should detect existing header")
	}

	exists, err = manager.HeaderExists("Missing App")
	if err != nil {
		t.Fatalf("Failed to check header exists: %v", err)
	}

	if exists {
		t.Error("Should not detect non-existing header")
	}
}

func TestAddSoftwareStepsForExisting(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")

	initialContent := `# System Setup

## Existing App

- [ ] Existing step
`

	if err := os.WriteFile(checklistFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	manager := New(checklistFile)

	steps := []string{"New step 1", "New step 2"}
	caveats := "Test caveats"

	// Should not add because header already exists and force=false
	err := manager.AddSoftwareSteps("Existing App", "Existing App", steps, caveats)
	if err != nil {
		t.Fatalf("Failed to add software steps: %v", err)
	}

	// Should add because this is for a new app
	err = manager.AddSoftwareStepsForExisting("New App", "App note", steps, caveats)
	if err != nil {
		t.Fatalf("Failed to add software steps for existing: %v", err)
	}

	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist file: %v", err)
	}

	contentStr := string(content)

	// Should contain the new app header
	if !strings.Contains(contentStr, "## New App") {
		t.Error("New app header not found in checklist")
	}

	// Should contain the new steps
	if !strings.Contains(contentStr, "- [ ] New step 1") {
		t.Error("New step 1 not found in checklist")
	}

	if !strings.Contains(contentStr, "- [ ] New step 2") {
		t.Error("New step 2 not found in checklist")
	}

	// Should contain caveats
	if !strings.Contains(contentStr, "Test caveats") {
		t.Error("Caveats not found in checklist")
	}
}
