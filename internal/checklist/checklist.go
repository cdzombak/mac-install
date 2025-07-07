package checklist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Manager struct {
	checklistPath string
}

func New(checklistPath string) *Manager {
	return &Manager{checklistPath: checklistPath}
}

func (m *Manager) AddSoftwareSteps(softwareName, displayName string, steps []string, caveats string) error {
	return m.addSoftwareStepsForce(softwareName, displayName, steps, caveats, false)
}

func (m *Manager) AddSoftwareStepsForExisting(softwareName, displayName string, steps []string, caveats string) error {
	return m.addSoftwareStepsForce(softwareName, displayName, steps, caveats, true)
}

func (m *Manager) addSoftwareStepsForce(softwareName, displayName string, steps []string, caveats string, force bool) error {
	headerExists, err := m.headerExists(displayName)
	if err != nil {
		return err
	}

	if headerExists && !force {
		return nil
	}

	file, err := os.OpenFile(m.checklistPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if !headerExists {
		if _, err := file.WriteString(fmt.Sprintf("\n## %s\n\n", displayName)); err != nil {
			return err
		}

		for _, step := range steps {
			if _, err := file.WriteString(fmt.Sprintf("- [ ] %s\n", step)); err != nil {
				return err
			}
		}

		if caveats != "" {
			if _, err := file.WriteString(fmt.Sprintf("\n### Brew Caveats\n\n```\n%s\n```\n", caveats)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) AddInstallStep(softwareName, displayName string) error {
	return m.AddSoftwareSteps(softwareName, displayName, []string{fmt.Sprintf("Install %s", softwareName)}, "")
}

func (m *Manager) HeaderExists(displayName string) (bool, error) {
	return m.headerExists(displayName)
}

func (m *Manager) headerExists(displayName string) (bool, error) {
	file, err := os.Open(m.checklistPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	headerLine := fmt.Sprintf("## %s", displayName)
	
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == headerLine {
			return true, nil
		}
	}

	return false, scanner.Err()
}