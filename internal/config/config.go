package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed internal.yaml
var internalConfigData []byte

type Config struct {
	Checklist     string         `yaml:"checklist"`
	InstallGroups []InstallGroup `yaml:"install_groups"`
}

type InstallGroup struct {
	Group    string     `yaml:"group"`
	Optional *bool      `yaml:"optional,omitempty"`
	Software []Software `yaml:"software"`
}

type Software struct {
	Name      string              `yaml:"name"`
	Artifact  string              `yaml:"artifact"`
	Note      string              `yaml:"note,omitempty"`
	Install   []map[string]string `yaml:"install,omitempty"`
	Configure []map[string]string `yaml:"configure,omitempty"`
	Checklist []string            `yaml:"checklist,omitempty"`
	Persist   *bool               `yaml:"persist,omitempty"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := config.expandVariables(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) expandVariables() error {
	homeDir, _ := os.UserHomeDir()
	brewPrefix := "/opt/homebrew"
	if _, err := os.Stat("/usr/local/bin/brew"); err == nil {
		brewPrefix = "/usr/local"
	}

	for i := range c.InstallGroups {
		for j := range c.InstallGroups[i].Software {
			software := &c.InstallGroups[i].Software[j]
			software.Artifact = strings.ReplaceAll(software.Artifact, "$HOME", homeDir)
			software.Artifact = strings.ReplaceAll(software.Artifact, "$BREW", brewPrefix)
			
			// Handle $ENV_ variables
			var err error
			software.Artifact, err = c.expandEnvVariables(software.Artifact)
			if err != nil {
				return fmt.Errorf("failed to expand environment variables in artifact path for %s: %w", software.Name, err)
			}
		}
	}

	c.Checklist = strings.ReplaceAll(c.Checklist, "$HOME", homeDir)
	c.Checklist = strings.ReplaceAll(c.Checklist, "$BREW", brewPrefix)
	
	// Handle $ENV_ variables in checklist
	var err error
	c.Checklist, err = c.expandEnvVariables(c.Checklist)
	if err != nil {
		return fmt.Errorf("failed to expand environment variables in checklist path: %w", err)
	}
	
	return nil
}

// expandEnvVariables expands environment variables using $ENV_ prefix
func (c *Config) expandEnvVariables(input string) (string, error) {
	// Regular expression to match $ENV_VARIABLE_NAME patterns
	envVarRegex := regexp.MustCompile(`\$ENV_([A-Z_][A-Z0-9_]*)`)
	
	// Find all matches
	matches := envVarRegex.FindAllStringSubmatch(input, -1)
	
	result := input
	for _, match := range matches {
		fullMatch := match[0]  // e.g., "$ENV_ASDF_PY"
		varName := match[1]    // e.g., "ASDF_PY"
		
		// Get the environment variable value
		envValue := os.Getenv(varName)
		if envValue == "" {
			return "", fmt.Errorf("environment variable %s is not set", varName)
		}
		
		// Replace the $ENV_VARIABLE_NAME with the actual value
		result = strings.ReplaceAll(result, fullMatch, envValue)
	}
	
	return result, nil
}

func (s *Software) GetArtifactDisplayName() string {
	artifact := s.Artifact
	
	// Replace home directory with tilde
	if homeDir, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(artifact, homeDir) {
			artifact = "~" + strings.TrimPrefix(artifact, homeDir)
		}
	}
	
	if strings.HasPrefix(artifact, "/Applications/") ||
		strings.Contains(artifact, "/Applications/") ||
		strings.Contains(artifact, "/bin/") {
		return filepath.Base(artifact)
	}
	return artifact
}

func (s *Software) GetDisplayName() string {
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	return s.GetArtifactDisplayName()
}

func (s *Software) ShouldPersist() bool {
	if s.Persist == nil {
		return false
	}
	return *s.Persist
}

func (g *InstallGroup) IsOptional() bool {
	if g.Optional == nil {
		return true
	}
	return *g.Optional
}

func LoadInternal() (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(internalConfigData, &config); err != nil {
		return nil, err
	}

	if err := config.expandVariables(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) RequiresHomebrew() bool {
	for _, group := range c.InstallGroups {
		for _, software := range group.Software {
			for _, installStep := range software.Install {
				for method := range installStep {
					if method == "brew" || method == "cask" {
						return true
					}
				}
			}
		}
	}
	return false
}