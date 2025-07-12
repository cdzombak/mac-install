package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
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
		fmt.Fprintf(os.Stderr, "Error: Multiple software items match '%s'. Please be more specific.\n\n", o.onlyTarget)
		fmt.Fprintf(os.Stderr, "Found matches:\n")
		for _, match := range matches {
			fmt.Fprintf(os.Stderr, "  - %s (artifact: %s)\n", match.software.GetDisplayName(), match.software.Artifact)
		}
		return fmt.Errorf("ambiguous target")
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