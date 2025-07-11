package orchestrator

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cdzombak/mac-install/internal/checklist"
	"github.com/cdzombak/mac-install/internal/colors"
	"github.com/cdzombak/mac-install/internal/config"
	"github.com/cdzombak/mac-install/internal/installer"
	"github.com/cdzombak/mac-install/internal/state"
)

type Orchestrator struct {
	config    *config.Config
	installer *installer.Installer
	checklist *checklist.Manager
	state     *state.Store
}

func New(cfg *config.Config, configDir string) *Orchestrator {
	return &Orchestrator{
		config:    cfg,
		installer: installer.New(configDir),
		checklist: checklist.New(cfg.Checklist),
	}
}

func (o *Orchestrator) Run() error {
	var err error
	o.state, err = state.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}

	if err := o.processInternalArtifacts(); err != nil {
		return fmt.Errorf("failed to process internal artifacts: %w", err)
	}

	for _, group := range o.config.InstallGroups {
		fmt.Printf("\n=== %s ===\n", colors.Group(group.Group))
		
		for _, software := range group.Software {
			if err := o.processSoftware(software, group.IsOptional()); err != nil {
				return fmt.Errorf("failed to process %s: %w", software.GetDisplayName(), err)
			}
		}
	}

	fmt.Printf("\n%s\n", colors.Success("Installation completed successfully!"))
	return nil
}

func (o *Orchestrator) processInternalArtifacts() error {
	if !o.config.RequiresHomebrew() {
		return nil
	}

	fmt.Printf("\n=== %s ===\n", colors.Group("Internal Requirements"))
	
	internalConfig, err := config.LoadInternal()
	if err != nil {
		return fmt.Errorf("failed to load internal configuration: %w", err)
	}

	for _, group := range internalConfig.InstallGroups {
		for _, software := range group.Software {
			if err := o.processSoftware(software, false); err != nil {
				return fmt.Errorf("failed to process internal artifact %s: %w", software.GetDisplayName(), err)
			}
		}
	}

	return nil
}

func (o *Orchestrator) processSoftware(software config.Software, isOptional bool) error {
	fmt.Printf("\n%s %s%s\n", colors.Info("•"), colors.Software(software.GetDisplayName()), colors.Dim("..."))

	if isOptional && software.ShouldPersist() && o.state.IsExcluded(software.GetDisplayName()) {
		exclusionFile := o.state.GetExclusionFilePath(software.GetDisplayName())
		fmt.Printf("  %s\n", colors.Dim(fmt.Sprintf("Skipped (previously excluded) - to unset: rm %s", exclusionFile)))
		return nil
	}

	artifactExists := o.installer.ArtifactExists(software.Artifact)
	softwareInstalled := false

	if artifactExists {
		fmt.Printf("  %s\n", colors.Success("Already installed"))
		
		// Check if checklist items exist for this already-installed software
		if len(software.Checklist) > 0 {
			headerName := software.GetDisplayName()
			headerExists, err := o.checklist.HeaderExists(headerName)
			if err != nil {
				return fmt.Errorf("failed to check checklist header: %w", err)
			}
			
			if !headerExists {
				fmt.Printf("  %s\n", colors.Info("Adding missing checklist items..."))
				
				var caveats string
				if o.wasInstalledViaHomebrew(software.Install) {
					packageName := o.getBrewPackageName(software.Install)
					if packageName != "" {
						caveats, _ = o.installer.GetBrewCaveats(packageName)
					}
				}
				
				if err := o.checklist.AddSoftwareStepsForExisting(software.GetDisplayName(), headerName, software.Checklist, caveats); err != nil {
					return fmt.Errorf("failed to add checklist items for existing software: %w", err)
				}
			}
		}
	} else {
		if len(software.Install) == 0 {
			fmt.Printf("  %s\n", colors.Warning("No installation steps defined, adding to checklist"))
			headerName := software.GetDisplayName()
			return o.checklist.AddInstallStep(software.GetDisplayName(), headerName)
		}

		if isOptional {
			shouldInstall, err := o.promptForInstallation(&software)
			if err != nil {
				return err
			}

			if !shouldInstall {
				if software.ShouldPersist() {
					if err := o.state.SetExcluded(software.GetDisplayName()); err != nil {
						return fmt.Errorf("failed to save exclusion state: %w", err)
					}
					fmt.Printf("  %s\n", colors.Dim("Skipped (excluded by user, choice saved)"))
				} else {
					fmt.Printf("  %s\n", colors.Dim("Skipped (excluded by user)"))
				}
				return nil
			}
		}

		fmt.Printf("  %s\n", colors.Info("Installing..."))
		if err := o.installer.Install(software.Install, software.Artifact); err != nil {
			return err
		}

		if !o.installer.ArtifactExists(software.Artifact) {
			return fmt.Errorf("installation completed but artifact %s not found", software.Artifact)
		}

		softwareInstalled = true
		fmt.Printf("  %s\n", colors.Success("Installed successfully"))
	}

	if o.installer.ArtifactExists(software.Artifact) && len(software.Configure) > 0 {
		fmt.Printf("  %s\n", colors.Info("Configuring..."))
		if err := o.installer.Configure(software.Configure); err != nil {
			return err
		}
		fmt.Printf("  %s\n", colors.Success("Configured successfully"))
	}

	if softwareInstalled && len(software.Checklist) > 0 {
		fmt.Printf("  %s\n", colors.Info("Adding checklist items..."))
		headerName := software.GetDisplayName()
		
		var caveats string
		if o.wasInstalledViaHomebrew(software.Install) {
			packageName := o.getBrewPackageName(software.Install)
			if packageName != "" {
				caveats, _ = o.installer.GetBrewCaveats(packageName)
			}
		}
		
		if err := o.checklist.AddSoftwareSteps(software.GetDisplayName(), headerName, software.Checklist, caveats); err != nil {
			return fmt.Errorf("failed to add checklist items: %w", err)
		}
	}

	return nil
}

func (o *Orchestrator) promptForInstallation(software *config.Software) (bool, error) {
	promptText := fmt.Sprintf("Install %s?", software.GetDisplayName())
	if software.Note != "" {
		promptText = fmt.Sprintf("Install %s (%s)?", software.GetDisplayName(), software.Note)
	}
	fmt.Printf("  %s (y/N): ", colors.Prompt(promptText))
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func (o *Orchestrator) wasInstalledViaHomebrew(installSteps []map[string]string) bool {
	for _, step := range installSteps {
		for method := range step {
			if method == "brew" || method == "cask" {
				return true
			}
		}
	}
	return false
}

func (o *Orchestrator) getBrewPackageName(installSteps []map[string]string) string {
	for _, step := range installSteps {
		for method, value := range step {
			if method == "brew" || method == "cask" {
				return value
			}
		}
	}
	return ""
}