package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		artifact string
		expected string
	}{
		{"/Applications/Test.app", "Test"},
		{"/Users/test/Applications/Test.app", "Test"},
		{"/opt/homebrew/bin/test", "test"},
		{"/usr/local/bin/test", "test"},
		{"custom-artifact", "custom-artifact"},
		{homeDir + "/.asdf/plugins/python", "~/.asdf/plugins/python"},
		{homeDir + "/Applications/Test.app", "Test"},
		{homeDir + "/custom/path", "~/custom/path"},
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
			expected: "Test",
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
			expected: "Test",
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

func TestExpandEnvVariables(t *testing.T) {
	config := &Config{}

	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
		hasError bool
	}{
		{
			name:     "simple environment variable",
			input:    "$HOME/.asdf/python/$ENV_ASDF_PY",
			envVars:  map[string]string{"ASDF_PY": "3.12"},
			expected: "$HOME/.asdf/python/3.12",
			hasError: false,
		},
		{
			name:     "multiple environment variables",
			input:    "$ENV_PREFIX/bin/$ENV_VERSION/tool",
			envVars:  map[string]string{"PREFIX": "/opt", "VERSION": "v1.2.3"},
			expected: "/opt/bin/v1.2.3/tool",
			hasError: false,
		},
		{
			name:     "no environment variables",
			input:    "$HOME/.config/app/config.json",
			envVars:  map[string]string{},
			expected: "$HOME/.config/app/config.json",
			hasError: false,
		},
		{
			name:     "missing environment variable",
			input:    "$ENV_MISSING_VAR/path",
			envVars:  map[string]string{},
			expected: "",
			hasError: true,
		},
		{
			name:     "mixed variables",
			input:    "$HOME/$ENV_SUBDIR/file.txt",
			envVars:  map[string]string{"SUBDIR": "Documents"},
			expected: "$HOME/Documents/file.txt",
			hasError: false,
		},
		{
			name:     "environment variable with underscores and numbers",
			input:    "/path/$ENV_APP_V2_CONFIG/file",
			envVars:  map[string]string{"APP_V2_CONFIG": "config-v2"},
			expected: "/path/config-v2/file",
			hasError: false,
		},
		{
			name:     "same environment variable used twice",
			input:    "$ENV_DIR/$ENV_DIR/nested",
			envVars:  map[string]string{"DIR": "shared"},
			expected: "shared/shared/nested",
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range test.envVars {
				t.Setenv(key, value)
			}

			// Clear any environment variables that shouldn't be set
			if test.hasError {
				for key := range test.envVars {
					if key == "MISSING_VAR" {
						os.Unsetenv(key)
					}
				}
			}

			result, err := config.expandEnvVariables(test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != test.expected {
					t.Errorf("Expected '%s', got '%s'", test.expected, result)
				}
			}
		})
	}
}

func TestExpandVariablesWithEnv(t *testing.T) {
	// Set up test environment variables
	t.Setenv("TEST_VAR", "test-value")
	t.Setenv("CONFIG_DIR", "configs")

	yamlData := `
checklist: $HOME/checklist-$ENV_CONFIG_DIR.md
install_groups:
  - group: Test Group
    software:
      - name: Test App
        artifact: $HOME/.local/$ENV_TEST_VAR/app
        install:
          - run: echo test
`

	config := &Config{}
	err := yaml.Unmarshal([]byte(yamlData), config)
	if err != nil {
		t.Fatal(err)
	}

	err = config.expandVariables()
	if err != nil {
		t.Fatalf("expandVariables should not error: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedChecklist := homeDir + "/checklist-configs.md"
	if config.Checklist != expectedChecklist {
		t.Errorf("Expected checklist '%s', got '%s'", expectedChecklist, config.Checklist)
	}

	expectedArtifact := homeDir + "/.local/test-value/app"
	if config.InstallGroups[0].Software[0].Artifact != expectedArtifact {
		t.Errorf("Expected artifact '%s', got '%s'", expectedArtifact, config.InstallGroups[0].Software[0].Artifact)
	}
}

func TestExpandVariablesErrorHandling(t *testing.T) {
	yamlData := `
checklist: $HOME/checklist.md
install_groups:
  - group: Test Group
    software:
      - name: Test App
        artifact: $HOME/.local/$ENV_MISSING_VAR/app
        install:
          - run: echo test
`

	config := &Config{}
	err := yaml.Unmarshal([]byte(yamlData), config)
	if err != nil {
		t.Fatal(err)
	}

	err = config.expandVariables()
	if err == nil {
		t.Error("expandVariables should error when environment variable is missing")
	}

	if !strings.Contains(err.Error(), "MISSING_VAR") {
		t.Errorf("Error should mention the missing variable name, got: %v", err)
	}

	if !strings.Contains(err.Error(), "Test App") {
		t.Errorf("Error should mention the software name, got: %v", err)
	}
}

func TestExpandTildePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standalone tilde",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "tilde with slash",
			input:    "~/Documents",
			expected: homeDir + "/Documents",
		},
		{
			name:     "tilde with nested path",
			input:    "~/Documents/Projects/test.txt",
			expected: homeDir + "/Documents/Projects/test.txt",
		},
		{
			name:     "path without tilde",
			input:    "/absolute/path/to/file",
			expected: "/absolute/path/to/file",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "relative path without tilde",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "tilde with special characters",
			input:    "~/Documents & Files",
			expected: homeDir + "/Documents & Files",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := expandTildePath(test.input, homeDir)
			if result != test.expected {
				t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
