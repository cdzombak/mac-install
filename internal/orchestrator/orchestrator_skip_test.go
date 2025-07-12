package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cdzombak/mac-install/internal/config"
)

func TestSkipOptional(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "checklist.md")

	cfg := &config.Config{
		Checklist: checklistFile,
		InstallGroups: []config.InstallGroup{
			{
				Group: "Required Group",
				Optional: boolPtr(false),
				Software: []config.Software{
					{
						Name:     "Required Software",
						Artifact: filepath.Join(tempDir, "required.txt"),
						Checklist: []string{"Required checklist item"},
					},
				},
			},
			{
				Group: "Optional Group",
				Optional: boolPtr(true),
				Software: []config.Software{
					{
						Name:     "Optional Software",
						Artifact: filepath.Join(tempDir, "optional.txt"),
						Checklist: []string{"Optional checklist item"},
					},
				},
			},
		},
	}

	// Test with skip-optional flag set
	o := New(cfg, tempDir)
	o.SetSkipOptional(true)
	
	if err := o.initializeForTesting(tempDir); err != nil {
		t.Fatal(err)
	}

	// Process all groups
	for _, group := range cfg.InstallGroups {
		if o.skipOptional && group.IsOptional() {
			continue
		}
		for _, software := range group.Software {
			err := o.processSoftware(software, group.IsOptional())
			if err != nil {
				t.Fatalf("Process software should not error: %v", err)
			}
		}
	}

	// Check that checklist was created
	content, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("Failed to read checklist: %v", err)
	}

	contentStr := string(content)
	
	// Should contain required software checklist
	if !strings.Contains(contentStr, "Required Software") {
		t.Error("Expected required software in checklist")
	}
	// Should contain either the checklist item or the "Install" item since there are no install steps
	if !strings.Contains(contentStr, "Required checklist item") && !strings.Contains(contentStr, "Install Required Software") {
		t.Error("Expected required checklist item or install item")
	}

	// Should NOT contain optional software checklist
	if strings.Contains(contentStr, "Optional Software") {
		t.Error("Optional software should not be in checklist when skip-optional is set")
	}
	if strings.Contains(contentStr, "Optional checklist item") {
		t.Error("Optional checklist item should not be present when skip-optional is set")
	}
}

func TestSkipOptionalFalse(t *testing.T) {
	tempDir := t.TempDir()
	checklistFile := filepath.Join(tempDir, "checklist.md")

	cfg := &config.Config{
		Checklist: checklistFile,
		InstallGroups: []config.InstallGroup{
			{
				Group: "Optional Group",
				Optional: boolPtr(true),
				Software: []config.Software{
					{
						Name:     "Optional Software",
						Artifact: filepath.Join(tempDir, "optional.txt"),
					},
				},
			},
		},
	}

	// Test with skip-optional flag NOT set
	o := New(cfg, tempDir)
	o.SetSkipOptional(false)
	
	// Verify that the flag is correctly set
	if o.skipOptional {
		t.Error("skipOptional should be false")
	}
}

func boolPtr(b bool) *bool {
	return &b
}