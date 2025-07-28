package orchestrator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cdzombak/mac-install/internal/colors"
	"github.com/cdzombak/mac-install/internal/config"
)

type softwareMatch struct {
	software config.Software
	group    config.InstallGroup
}

func (o *Orchestrator) runOnlyTarget() error {
	// Find all matching software
	matches := o.findMatchingSoftware(o.onlyTarget)

	if len(matches) == 0 {
		return fmt.Errorf("no software found matching '%s'", o.onlyTarget)
	}

	if len(matches) > 1 {
		fmt.Printf("Multiple software items match '%s'. Please select which one to install:\n\n", o.onlyTarget)
		for i, match := range matches {
			fmt.Printf("%d. %s (artifact: %s)\n", i+1, match.software.GetDisplayName(), match.software.Artifact)
		}
		fmt.Printf("\nEnter selection (1-%d): ", len(matches))
		
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read selection: %w", err)
		}
		
		selection, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || selection < 1 || selection > len(matches) {
			return fmt.Errorf("invalid selection: please enter a number between 1 and %d", len(matches))
		}
		
		// Use the selected match
		matches = []softwareMatch{matches[selection-1]}
	}

	// Process the single match
	match := matches[0]
	fmt.Printf("\n=== %s ===\n", colors.Group("Installing Single Target"))
	
	if err := o.processSoftware(match.software, match.group.IsOptional()); err != nil {
		return fmt.Errorf("failed to process %s: %w", match.software.GetDisplayName(), err)
	}

	fmt.Printf("\n%s\n", colors.Success("Installation completed successfully!"))
	return nil
}

func (o *Orchestrator) findMatchingSoftware(target string) []softwareMatch {
	var matches []softwareMatch
	targetLower := strings.ToLower(target)

	for _, group := range o.config.InstallGroups {
		for _, software := range group.Software {
			// Check if the user-chosen name contains the target
			if software.Name != "" && strings.Contains(strings.ToLower(software.Name), targetLower) {
				matches = append(matches, softwareMatch{software: software, group: group})
				continue
			}

			// Check if the artifact basename contains the target
			artifactBase := filepath.Base(software.Artifact)
			if strings.Contains(strings.ToLower(artifactBase), targetLower) {
				matches = append(matches, softwareMatch{software: software, group: group})
			}
		}
	}

	return matches
}