package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")
	
	configContent := `checklist: /Users/test/SystemSetup.md

install_groups:
  - group: Test Group
    software:
      - name: Test App
        artifact: /Applications/Test.app
        note: This is a test note
        install:
          - brew: test-package
        configure:
          - run: echo "test"
        checklist:
          - Test checklist item
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Checklist != "/Users/test/SystemSetup.md" {
		t.Errorf("Expected checklist path '/Users/test/SystemSetup.md', got '%s'", config.Checklist)
	}

	if len(config.InstallGroups) != 1 {
		t.Fatalf("Expected 1 install group, got %d", len(config.InstallGroups))
	}

	group := config.InstallGroups[0]
	if group.Group != "Test Group" {
		t.Errorf("Expected group name 'Test Group', got '%s'", group.Group)
	}

	if len(group.Software) != 1 {
		t.Fatalf("Expected 1 software item, got %d", len(group.Software))
	}

	software := group.Software[0]
	if software.Name != "Test App" {
		t.Errorf("Expected software name 'Test App', got '%s'", software.Name)
	}

	if software.Artifact != "/Applications/Test.app" {
		t.Errorf("Expected artifact '/Applications/Test.app', got '%s'", software.Artifact)
	}

	if software.Note != "This is a test note" {
		t.Errorf("Expected note 'This is a test note', got '%s'", software.Note)
	}
}

func TestExpandVariables(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	
	config := &Config{
		Checklist: "$HOME/SystemSetup.md",
		InstallGroups: []InstallGroup{
			{
				Group: "Test",
				Software: []Software{
					{
						Name:     "Test",
						Artifact: "$HOME/Applications/Test.app",
					},
				},
			},
		},
	}

	config.expandVariables()

	expectedChecklist := strings.ReplaceAll("$HOME/SystemSetup.md", "$HOME", homeDir)
	if config.Checklist != expectedChecklist {
		t.Errorf("Expected checklist '%s', got '%s'", expectedChecklist, config.Checklist)
	}

	expectedArtifact := strings.ReplaceAll("$HOME/Applications/Test.app", "$HOME", homeDir)
	if config.InstallGroups[0].Software[0].Artifact != expectedArtifact {
		t.Errorf("Expected artifact '%s', got '%s'", expectedArtifact, config.InstallGroups[0].Software[0].Artifact)
	}
}

func TestGetArtifactDisplayName(t *testing.T) {
	tests := []struct {
		artifact string
		expected string
	}{
		{"/Applications/Test.app", "Test.app"},
		{"/Users/test/Applications/Test.app", "Test.app"},
		{"/opt/homebrew/bin/test", "test"},
		{"/usr/local/bin/test", "test"},
		{"custom-artifact", "custom-artifact"},
	}

	for _, test := range tests {
		software := Software{Artifact: test.artifact}
		result := software.GetArtifactDisplayName()
		if result != test.expected {
			t.Errorf("For artifact '%s', expected '%s', got '%s'", test.artifact, test.expected, result)
		}
	}
}

func TestIsOptional(t *testing.T) {
	tests := []struct {
		name     string
		optional *bool
		expected bool
	}{
		{"default (nil)", nil, true},
		{"explicitly true", boolPtr(true), true},
		{"explicitly false", boolPtr(false), false},
	}

	for _, test := range tests {
		group := InstallGroup{Optional: test.optional}
		result := group.IsOptional()
		if result != test.expected {
			t.Errorf("Test '%s': expected %v, got %v", test.name, test.expected, result)
		}
	}
}

func TestRequiresHomebrew(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name: "no homebrew",
			config: &Config{
				InstallGroups: []InstallGroup{
					{
						Software: []Software{
							{Install: []map[string]string{{"npm": "package"}}},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "with brew",
			config: &Config{
				InstallGroups: []InstallGroup{
					{
						Software: []Software{
							{Install: []map[string]string{{"brew": "package"}}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "with cask",
			config: &Config{
				InstallGroups: []InstallGroup{
					{
						Software: []Software{
							{Install: []map[string]string{{"cask": "package"}}},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		result := test.config.RequiresHomebrew()
		if result != test.expected {
			t.Errorf("Test '%s': expected %v, got %v", test.name, test.expected, result)
		}
	}
}

func TestLoadInternal(t *testing.T) {
	config, err := LoadInternal()
	if err != nil {
		t.Fatalf("Failed to load internal config: %v", err)
	}

	if len(config.InstallGroups) == 0 {
		t.Error("Internal config should have install groups")
	}

	if config.InstallGroups[0].Group != "Core Dependencies" {
		t.Errorf("Expected group name 'Core Dependencies', got '%s'", config.InstallGroups[0].Group)
	}

	if config.InstallGroups[0].IsOptional() {
		t.Error("Internal requirements should not be optional")
	}
}

func TestShouldPersist(t *testing.T) {
	tests := []struct {
		name     string
		persist  *bool
		expected bool
	}{
		{"default (nil)", nil, false},
		{"explicitly true", boolPtr(true), true},
		{"explicitly false", boolPtr(false), false},
	}

	for _, test := range tests {
		software := Software{Persist: test.persist}
		result := software.ShouldPersist()
		if result != test.expected {
			t.Errorf("Test '%s': expected %v, got %v", test.name, test.expected, result)
		}
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		software Software
		expected string
	}{
		{
			name: "with name provided",
			software: Software{
				Name:     "Custom Name",
				Artifact: "/Applications/Test.app",
			},
			expected: "Custom Name",
		},
		{
			name: "no name, Applications artifact",
			software: Software{
				Name:     "",
				Artifact: "/Applications/Test.app",
			},
			expected: "Test.app",
		},
		{
			name: "no name, bin artifact",
			software: Software{
				Name:     "",
				Artifact: "/usr/local/bin/tool",
			},
			expected: "tool",
		},
		{
			name: "no name, custom path",
			software: Software{
				Name:     "",
				Artifact: "/custom/path/artifact",
			},
			expected: "/custom/path/artifact",
		},
		{
			name: "empty name defaults to artifact display",
			software: Software{
				Name:     "",
				Artifact: "/Library/Fonts/font.ttf",
			},
			expected: "/Library/Fonts/font.ttf",
		},
		{
			name: "whitespace-only name defaults to artifact display",
			software: Software{
				Name:     "   ",
				Artifact: "/Applications/Test.app",
			},
			expected: "Test.app",
		},
		{
			name: "tab-only name defaults to artifact display",
			software: Software{
				Name:     "\t\t",
				Artifact: "/usr/local/bin/tool",
			},
			expected: "tool",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.software.GetDisplayName()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}