// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Display overall plonk status",
	Long: `Display a detailed list of all plonk-managed items and their status.

Shows:
- All managed packages and dotfiles
- Missing items that need to be installed
- Configuration and lock file status

Examples:
  plonk status    # Show all managed items
  plonk st        # Short alias`,
	RunE:         runStatus,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

	// Reconcile dotfiles with injected config
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile dotfiles
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Get package status from lock file
	packageResult := getPackageStatus(ctx, configDir)

	// Convert to output summary
	summary := convertToSummary(dotfileResult, packageResult)

	// Check file existence and validity
	configPath := filepath.Join(configDir, "plonk.yaml")
	lockPath := filepath.Join(configDir, "plonk.lock")

	configExists := false
	configValid := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		// Config is valid only if it loaded without error
		configValid = (configLoadErr == nil)
	}

	lockExists := false
	if _, err := os.Stat(lockPath); err == nil {
		lockExists = true
	}

	// Create formatter data directly
	formatterData := output.StatusOutput{
		ConfigPath:   configPath,
		LockPath:     lockPath,
		ConfigExists: configExists,
		ConfigValid:  configValid,
		LockExists:   lockExists,
		StateSummary: summary,
		ConfigDir:    configDir,
		HomeDir:      homeDir,
	}
	formatter := output.NewStatusFormatter(formatterData)
	output.RenderOutput(formatter)
	return nil
}

// packageStatus holds status information about tracked packages
type packageStatus struct {
	Managed []output.Item
	Missing []output.Item
}

// getPackageStatus reads the lock file and checks which packages are installed
func getPackageStatus(ctx context.Context, configDir string) packageStatus {
	result := packageStatus{}

	lockSvc := lock.NewLockV3Service(configDir)
	lockFile, err := lockSvc.Read()
	if err != nil {
		// No lock file or read error - return empty
		return result
	}

	// Check each package
	for manager, pkgs := range lockFile.Packages {
		mgr, err := packages.GetManager(manager)
		if err != nil {
			// Unknown manager - mark all as missing
			for _, pkg := range pkgs {
				result.Missing = append(result.Missing, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateMissing,
				})
			}
			continue
		}

		for _, pkg := range pkgs {
			installed, _ := mgr.IsInstalled(ctx, pkg)
			if installed {
				result.Managed = append(result.Managed, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateManaged,
				})
			} else {
				result.Missing = append(result.Missing, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateMissing,
				})
			}
		}
	}

	return result
}

// convertToSummary combines dotfile and package results into a unified summary
func convertToSummary(dotResult dotfiles.Result, pkgResult packageStatus) output.Summary {
	// Convert dotfiles result to output.Result
	dotfileOutput := output.Result{
		Domain:    "dotfile",
		Managed:   convertDotfileItemsToOutput(dotResult.Managed),
		Missing:   convertDotfileItemsToOutput(dotResult.Missing),
		Untracked: convertDotfileItemsToOutput(dotResult.Untracked),
	}

	// Create package result
	packageOutput := output.Result{
		Domain:  "package",
		Managed: pkgResult.Managed,
		Missing: pkgResult.Missing,
	}

	totalManaged := len(dotResult.Managed) + len(pkgResult.Managed)
	totalMissing := len(dotResult.Missing) + len(pkgResult.Missing)
	totalUntracked := len(dotResult.Untracked)

	return output.Summary{
		TotalManaged:   totalManaged,
		TotalMissing:   totalMissing,
		TotalUntracked: totalUntracked,
		Results:        []output.Result{packageOutput, dotfileOutput},
	}
}
