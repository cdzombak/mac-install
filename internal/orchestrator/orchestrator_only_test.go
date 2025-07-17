package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cdzombak/mac-install/internal/config"
)

func TestFindMatchingSoftware(t *testing.T) {
	cfg := &config.Config{
		InstallGroups: []config.InstallGroup{
			{
				Group: "Development Tools",
				Software: []config.Software{
					{
						Name:     "Visual Studio Code",
						Artifact: "/Applications/Visual Studio Code.app",
					},
					{
						Name:     "Autodesk Fusion",
						Artifact: "/Applications/Autodesk Fusion.app",
					},
					{
						// No name - should match on artifact basename
						Artifact: "/Applications/Docker.app",
					},
				},
			},
		},
	}

	o := &Orchestrator{config: cfg}

	tests := []struct {
		target   string
		expected int
		names    []string
	}{
		{"Code", 1, []string{"Visual Studio Code"}},
		{"code", 1, []string{"Visual Studio Code"}}, // case insensitive
		{"Autodesk", 1, []string{"Autodesk Fusion"}},
		{"Docker", 1, []string{"Docker"}}, // matches artifact basename
		{"docker", 1, []string{"Docker"}}, // case insensitive
		{"Visual", 1, []string{"Visual Studio Code"}},
		{"Studio", 1, []string{"Visual Studio Code"}},
		{"nonexistent", 0, []string{}},
		{"", 3, []string{"Visual Studio Code", "Autodesk Fusion", "Docker"}}, // empty matches everything
	}

	for _, test := range tests {
		matches := o.findMatchingSoftware(test.target)
		if len(matches) != test.expected {
			t.Errorf("Target '%s': expected %d matches, got %d", test.target, test.expected, len(matches))
			continue
		}

		for i, expectedName := range test.names {
			if i >= len(matches) {
				t.Errorf("Target '%s': expected match %d to be '%s', but not enough matches", test.target, i, expectedName)
				continue
			}
			actualName := matches[i].software.GetDisplayName()
			if actualName != expectedName {
				t.Errorf("Target '%s': expected match %d to be '%s', got '%s'", test.target, i, expectedName, actualName)
			}
		}
	}
}

func TestRunOnlyTargetSingleMatch(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "checklist.md")

	cfg := &config.Config{
		Checklist: checklistFile,
		InstallGroups: []config.InstallGroup{
			{
				Group: "Test Group",
				Software: []config.Software{
					{
						Name:      "Target Software",
						Artifact:  filepath.Join(tempDir, "target.txt"),
						Checklist: []string{"Target checklist item"},
					},
					{
						Name:      "Other Software",
						Artifact:  filepath.Join(tempDir, "other.txt"),
						Checklist: []string{"Other checklist item"},
					},
				},
			},
		},
	}

	o := New(cfg, tempDir)
	o.SetOnlyTarget("Target")

	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	err := o.runOnlyTarget()
	if err != nil {
		t.Fatalf("runOnlyTarget should not error: %v", err)
	}

	// Check that checklist was created for target software only
	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist: %v", err)
	}

	contentStr := string(content)

	// Should contain target software
	if !strings.Contains(contentStr, "Target Software") {
		t.Error("Expected target software in checklist")
	}

	// Should NOT contain other software
	if strings.Contains(contentStr, "Other Software") {
		t.Error("Other software should not be in checklist when using -only")
	}
}

func TestRunOnlyTargetNoMatch(t *testing.T) {
	cfg := &config.Config{
		InstallGroups: []config.InstallGroup{
			{
				Group: "Test Group",
				Software: []config.Software{
					{
						Name:     "Some Software",
						Artifact: "/Applications/Some.app",
					},
				},
			},
		},
	}

	o := New(cfg, t.TempDir())
	o.SetOnlyTarget("Nonexistent")

	if err := o.initializeForTesting(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	err := o.runOnlyTarget()
	if err == nil {
		t.Error("Expected error when no software matches target")
	}
	if !strings.Contains(err.Error(), "no software found matching") {
		t.Errorf("Expected 'no software found matching' error, got: %v", err)
	}
}

func TestRunOnlyTargetMultipleMatches(t *testing.T) {
	cfg := &config.Config{
		InstallGroups: []config.InstallGroup{
			{
				Group: "Test Group",
				Software: []config.Software{
					{
						Name:     "Test Software One",
						Artifact: "/Applications/TestOne.app",
					},
					{
						Name:     "Test Software Two",
						Artifact: "/Applications/TestTwo.app",
					},
				},
			},
		},
	}

	o := New(cfg, t.TempDir())
	o.SetOnlyTarget("Test")

	if err := o.initializeForTesting(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	err := o.runOnlyTarget()
	if err == nil {
		t.Error("Expected error when multiple software items match target")
	}
	if !strings.Contains(err.Error(), "ambiguous target") {
		t.Errorf("Expected 'ambiguous target' error, got: %v", err)
	}
}
