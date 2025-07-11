package config

import (
	_ "embed"
	"os"
	"path/filepath"
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

	config.expandVariables()
	return &config, nil
}

func (c *Config) expandVariables() {
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
		}
	}

	c.Checklist = strings.ReplaceAll(c.Checklist, "$HOME", homeDir)
	c.Checklist = strings.ReplaceAll(c.Checklist, "$BREW", brewPrefix)
}

func (s *Software) GetArtifactDisplayName() string {
	artifact := s.Artifact
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

	config.expandVariables()
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