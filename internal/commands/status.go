// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

	// Reconcile dotfiles with injected config
	cfg := config.LoadWithDefaults(configDir)

	// Create DotfileManager and reconcile directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return err
	}

	// Get package status from lock file
	ctx := context.Background()
	packageResult, err := getPackageStatus(ctx, configDir)
	if err != nil {
		return err
	}

	// Convert to output summary
	summary := convertStatusToSummary(statuses, packageResult)

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
	Errors  []output.Item
}

// getPackageStatus reads the lock file and checks which packages are installed
func getPackageStatus(ctx context.Context, configDir string) (packageStatus, error) {
	result := packageStatus{}

	// Check if lock file exists first
	lockPath := filepath.Join(configDir, "plonk.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		// No lock file yet - this is fine, just no packages tracked
		return result, nil
	}

	lockSvc := lock.NewLockV3Service(configDir)
	lockFile, err := lockSvc.Read()
	if err != nil {
		return result, fmt.Errorf("failed to read lock file: %w", err)
	}

	// Check each package (sorted for deterministic output)
	managers := make([]string, 0, len(lockFile.Packages))
	for manager := range lockFile.Packages {
		managers = append(managers, manager)
	}
	sort.Strings(managers)
	for _, manager := range managers {
		pkgs := lockFile.Packages[manager]
		mgr, err := packages.GetManager(manager)
		if err != nil {
			// Unknown/unsupported manager - mark all as errors (not missing)
			for _, pkg := range pkgs {
				result.Errors = append(result.Errors, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateError,
					Error:   fmt.Sprintf("unsupported manager: %s", manager),
				})
			}
			continue
		}

		var managerBroken bool
		var managerErr string
		for _, pkg := range pkgs {
			// Short-circuit remaining packages if the manager itself is broken
			// (e.g., binary not on PATH) to avoid repeated failing subprocesses.
			if managerBroken {
				result.Errors = append(result.Errors, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateError,
					Error:   managerErr,
				})
				continue
			}

			installed, err := mgr.IsInstalled(ctx, pkg)
			if err != nil {
				managerBroken = true
				managerErr = err.Error()
				result.Errors = append(result.Errors, output.Item{
					Name:    pkg,
					Manager: manager,
					State:   output.StateError,
					Error:   err.Error(),
				})
				continue
			}
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

	return result, nil
}

// convertStatusToSummary combines dotfile statuses and package results into a unified summary
func convertStatusToSummary(statuses []dotfiles.DotfileStatus, pkgResult packageStatus) output.Summary {
	// Convert dotfiles to output format
	managedItems, missingItems, errorItems := convertDotfileStatusToOutput(statuses)

	dotfileOutput := output.Result{
		Domain:  "dotfile",
		Managed: managedItems,
		Missing: missingItems,
		Errors:  errorItems,
	}

	// Create package result
	packageOutput := output.Result{
		Domain:  "package",
		Managed: pkgResult.Managed,
		Missing: pkgResult.Missing,
		Errors:  pkgResult.Errors,
	}

	totalManaged := len(managedItems) + len(pkgResult.Managed)
	totalMissing := len(missingItems) + len(pkgResult.Missing)
	totalErrors := len(errorItems) + len(pkgResult.Errors)

	return output.Summary{
		TotalManaged:   totalManaged,
		TotalMissing:   totalMissing,
		TotalUntracked: 0,
		TotalErrors:    totalErrors,
		Results:        []output.Result{packageOutput, dotfileOutput},
	}
}
