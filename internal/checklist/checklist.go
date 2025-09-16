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

func (m *Manager) AddSoftwareSteps(displayName, note string, steps []string, caveats string) error {
	return m.addSoftwareStepsForce(displayName, note, steps, caveats, false)
}

func (m *Manager) AddSoftwareStepsForExisting(displayName, note string, steps []string, caveats string) error {
	return m.addSoftwareStepsForce(displayName, note, steps, caveats, true)
}

func (m *Manager) addSoftwareStepsForce(displayName string, note string, steps []string, caveats string, force bool) error {
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
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", err)
		}
	}()

	if !headerExists {
		if _, err := fmt.Fprintf(file, "\n## %s\n\n", displayName); err != nil {
			return err
		}

		if note != "" {
			note = strings.ReplaceAll(note, "\n", "\n> ")
			note = strings.TrimSuffix(note, "\n")
			if _, err := fmt.Fprintf(file, "> %s\n\n", note); err != nil {
				return err
			}
		}

		for _, step := range steps {
			if _, err := fmt.Fprintf(file, "- [ ] %s\n", step); err != nil {
				return err
			}
		}

		if caveats != "" {
			if _, err := fmt.Fprintf(file, "\n### Caveats\n\n```\n%s\n```\n", caveats); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) AddInstallStep(displayName, note string) error {
	return m.AddSoftwareSteps(displayName, note, []string{fmt.Sprintf("Install %s", displayName)}, "")
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
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	headerLine := fmt.Sprintf("## %s", displayName)

	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == headerLine {
			return true, nil
		}
	}

	return false, scanner.Err()
}
