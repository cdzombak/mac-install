package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cdzombak/mac-install/internal/config"
	"github.com/cdzombak/mac-install/internal/state"
)

func TestWasInstalledViaHomebrew(t *testing.T) {
	o := &Orchestrator{}

	tests := []struct {
		installSteps []map[string]string
		expected     bool
	}{
		{
			[]map[string]string{{"brew": "package"}},
			true,
		},
		{
			[]map[string]string{{"cask": "package"}},
			true,
		},
		{
			[]map[string]string{{"npm": "package"}},
			false,
		},
		{
			[]map[string]string{{"run": "echo test"}},
			false,
		},
		{
			[]map[string]string{{"brew": "package1"}, {"npm": "package2"}},
			true,
		},
	}

	for i, test := range tests {
		result := o.wasInstalledViaHomebrew(test.installSteps)
		if result != test.expected {
			t.Errorf("Test %d: expected %v, got %v", i, test.expected, result)
		}
	}
}

func TestGetBrewPackageName(t *testing.T) {
	o := &Orchestrator{}

	tests := []struct {
		installSteps []map[string]string
		expected     string
	}{
		{
			[]map[string]string{{"brew": "test-package"}},
			"test-package",
		},
		{
			[]map[string]string{{"cask": "test-cask"}},
			"test-cask",
		},
		{
			[]map[string]string{{"npm": "test-npm"}},
			"",
		},
		{
			[]map[string]string{{"brew": "first-package"}, {"cask": "second-package"}},
			"first-package",
		},
	}

	for i, test := range tests {
		result := o.getBrewPackageName(test.installSteps)
		if result != test.expected {
			t.Errorf("Test %d: expected '%s', got '%s'", i, test.expected, result)
		}
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Checklist: "/tmp/test-checklist.md",
		InstallGroups: []config.InstallGroup{
			{
				Group: "Test Group",
				Software: []config.Software{
					{
						Name:     "Test Software",
						Artifact: "/Applications/Test.app",
					},
				},
			},
		},
	}

	orchestrator := New(cfg, t.TempDir())

	if orchestrator.config != cfg {
		t.Error("Config not set correctly")
	}

	if orchestrator.installer == nil {
		t.Error("Installer not initialized")
	}

	if orchestrator.checklist == nil {
		t.Error("Checklist manager not initialized")
	}
}

func TestProcessSoftwareWithoutInstallSteps(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")
	
	cfg := &config.Config{
		Checklist: checklistFile,
	}

	o := New(cfg, t.TempDir())
	
	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	software := config.Software{
		Name:     "Test Software",
		Artifact: "/nonexistent/Test.app",
	}

	err := o.processSoftware(software, true)
	if err != nil {
		t.Fatalf("Process software should not error: %v", err)
	}

	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist: %v", err)
	}

	if len(content) == 0 {
		t.Error("Checklist should not be empty")
	}
}

func TestProcessSoftwareWithPersist(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")
	
	cfg := &config.Config{
		Checklist: checklistFile,
	}

	o := New(cfg, t.TempDir())
	
	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	persistTrue := true
	persistFalse := false

	tests := []struct {
		name     string
		software config.Software
		shouldPersistChoice bool
	}{
		{
			name: "persist true",
			software: config.Software{
				Name:     "Test Software Persist",
				Artifact: "/nonexistent/Test.app",
				Persist:  &persistTrue,
			},
			shouldPersistChoice: true,
		},
		{
			name: "persist false",
			software: config.Software{
				Name:     "Test Software No Persist",
				Artifact: "/nonexistent/Test.app", 
				Persist:  &persistFalse,
			},
			shouldPersistChoice: false,
		},
		{
			name: "persist nil (default)",
			software: config.Software{
				Name:     "Test Software Default",
				Artifact: "/nonexistent/Test.app",
			},
			shouldPersistChoice: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.software.ShouldPersist()
			if result != test.shouldPersistChoice {
				t.Errorf("Expected ShouldPersist() to return %v, got %v", test.shouldPersistChoice, result)
			}
		})
	}
}

func TestProcessSoftwareExistingWithMissingChecklist(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")
	
	// Create an existing artifact
	existingArtifact := filepath.Join(tempDir, "existing.txt")
	if err := os.WriteFile(existingArtifact, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	cfg := &config.Config{
		Checklist: checklistFile,
	}

	o := New(cfg, t.TempDir())
	
	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	software := config.Software{
		Name:      "Test Software",
		Artifact:  existingArtifact,
		Checklist: []string{"Manual step 1", "Manual step 2"},
	}

	err := o.processSoftware(software, true)
	if err != nil {
		t.Fatalf("Process software should not error: %v", err)
	}

	// Check that checklist was created
	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist: %v", err)
	}

	contentStr := string(content)
	
	// Should contain the header (uses software name when provided)
	if !strings.Contains(contentStr, "Test Software") {
		t.Error("Expected header not found in checklist")
	}
	
	// Should contain the manual steps
	if !strings.Contains(contentStr, "- [ ] Manual step 1") {
		t.Error("Manual step 1 not found in checklist")
	}
	
	if !strings.Contains(contentStr, "- [ ] Manual step 2") {
		t.Error("Manual step 2 not found in checklist")
	}
}

func TestProcessSoftwareWithoutNameUsesArtifactDisplayName(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "SystemSetup.md")
	
	// Create an existing artifact
	existingArtifact := filepath.Join(tempDir, "TestApp.app")
	if err := os.WriteFile(existingArtifact, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	cfg := &config.Config{
		Checklist: checklistFile,
	}

	o := New(cfg, t.TempDir())
	
	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	software := config.Software{
		// No Name field - should use artifact display name
		Artifact:  existingArtifact,
		Checklist: []string{"Manual step 1", "Manual step 2"},
	}

	err := o.processSoftware(software, true)
	if err != nil {
		t.Fatalf("Process software should not error: %v", err)
	}

	// Check that checklist was created
	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist: %v", err)
	}

	contentStr := string(content)
	
	// Should contain the header using artifact display name (full path since it's not in /Applications/ or /bin/)
	if !strings.Contains(contentStr, existingArtifact) {
		t.Error("Expected header with artifact path not found in checklist")
	}
	
	// Should contain the manual steps
	if !strings.Contains(contentStr, "- [ ] Manual step 1") {
		t.Error("Manual step 1 not found in checklist")
	}
}

func (o *Orchestrator) initializeForTesting(tempDir string) error {
	var err error
	o.state, err = state.NewStore()
	return err
}

func TestHasRunOrScriptSteps(t *testing.T) {
	o := &Orchestrator{}
	
	tests := []struct {
		name         string
		configSteps  []map[string]string
		expected     bool
	}{
		{
			name: "has run step",
			configSteps: []map[string]string{
				{"run": "echo test"},
			},
			expected: true,
		},
		{
			name: "has script step",
			configSteps: []map[string]string{
				{"script": "/path/to/script.sh"},
			},
			expected: true,
		},
		{
			name: "has both run and script",
			configSteps: []map[string]string{
				{"run": "echo test"},
				{"script": "/path/to/script.sh"},
			},
			expected: true,
		},
		{
			name: "has ignore_errors before run",
			configSteps: []map[string]string{
				{"ignore_errors": "true"},
				{"run": "echo test"},
			},
			expected: true,
		},
		{
			name: "no run or script steps",
			configSteps: []map[string]string{
				{"ignore_errors": "true"},
			},
			expected: false,
		},
		{
			name: "empty config steps",
			configSteps: []map[string]string{},
			expected: false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := o.hasRunOrScriptSteps(test.configSteps)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}