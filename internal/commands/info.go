// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <package>",
	Short: "Show detailed information about a package",
	Long: `Show detailed information about a package including version, description, homepage,
dependencies, and installation status.

The info command prioritizes information sources in this order:
1. If managed by plonk: Shows plonk-managed information
2. If installed but not managed: Shows installed package information
3. If available but not installed: Shows available package information

Use prefix syntax to get info from a specific manager.

Examples:
  plonk info git              # Show information about git package
  plonk info brew:ripgrep     # Show ripgrep info specifically from Homebrew
  plonk info npm:typescript   # Show TypeScript info from npm
  plonk info --output json go # Show information in JSON format`,
	Args:         cobra.ExactArgs(1),
	RunE:         runInfo,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Load configuration (not strictly required but consistent with other commands)
	configDir := config.GetDefaultConfigDirectory()
	_ = config.LoadWithDefaults(configDir)

	// Execute pure info logic
	infoResult, err := Info(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	// Convert to output package type and create formatter
	formatterData := output.InfoOutput{
		Package:     infoResult.Package,
		Status:      infoResult.Status,
		Message:     infoResult.Message,
		PackageInfo: infoResult.PackageInfo,
	}
	formatter := output.NewInfoFormatter(formatterData)
	return output.RenderOutput(formatter, format)
}

// Info performs the info lookup and returns typed results (pure logic)
func Info(ctx context.Context, packageSpec string) (InfoOutput, error) {
	// Parse prefix syntax
	manager, packageName := parsePackageSpec(packageSpec)

	// Validate manager if prefix specified
	if manager != "" && !isValidManager(manager) {
		errorMsg := formatNotFoundError("package manager", manager, getValidManagers())
		return InfoOutput{}, fmt.Errorf("%s", errorMsg)
	}

	if manager != "" {
		// Get info from specific manager
		return getInfoFromSpecificManager(ctx, manager, packageName)
	}
	// Use priority logic: managed → installed → available
	return getInfoWithPriorityLogic(ctx, packageName)
}

// getInfoFromSpecificManager gets info from a specific package manager only
func getInfoFromSpecificManager(ctx context.Context, managerName, packageName string) (InfoOutput, error) {
	// Get the specific manager
	registry := packages.NewManagerRegistry()
	manager, err := registry.GetManager(managerName)
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to get manager %s: %w", managerName, err)
	}

	// Check if manager is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to check %s availability: %w", managerName, err)
	}
	if !available {
		return InfoOutput{
			Package: packageName,
			Status:  "manager-unavailable",
			Message: fmt.Sprintf("Package manager '%s' is not available on this system", managerName),
		}, nil
	}

	// Check if package is managed by plonk
	configDir := config.GetDefaultConfigDirectory()
	lockService := lock.NewYAMLLockService(configDir)
	isManaged := lockService.HasPackage(managerName, packageName)

	// Check if package is installed
	installed, err := manager.IsInstalled(ctx, packageName)
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to check if package %s is installed in %s: %w", packageName, managerName, err)
	}

	// Get package info
	info, err := manager.Info(ctx, packageName)
	if err != nil {
		return InfoOutput{
			Package: packageName,
			Status:  "not-found",
			Message: fmt.Sprintf("Package '%s' not found in %s", packageName, managerName),
		}, nil
	}

	// Determine status based on installation and management
	status := "available"
	message := fmt.Sprintf("Package '%s' available in %s", packageName, managerName)
	if isManaged {
		status = "managed"
		message = fmt.Sprintf("Package '%s' is managed by plonk via %s", packageName, managerName)
	} else if installed {
		status = "installed"
		message = fmt.Sprintf("Package '%s' is installed via %s (not managed by plonk)", packageName, managerName)
	}

	return InfoOutput{
		Package:     packageName,
		Status:      status,
		Message:     message,
		PackageInfo: info,
	}, nil
}

// getInfoWithPriorityLogic implements the priority logic: managed → installed → available
func getInfoWithPriorityLogic(ctx context.Context, packageName string) (InfoOutput, error) {
	// Get config directory and initialize lock service
	configDir := config.GetDefaultConfigDirectory()
	lockService := lock.NewYAMLLockService(configDir)

	// 1. Check if package is managed by plonk
	packageLocations := lockService.FindPackage(packageName)
	if len(packageLocations) > 0 {
		// Use the first location found
		location := packageLocations[0]

		// Extract manager and version from metadata
		managerName, _ := location.Metadata["manager"].(string)
		version, _ := location.Metadata["version"].(string)

		// Try to get more detailed info from the manager
		registry := packages.NewManagerRegistry()
		if manager, err := registry.GetManager(managerName); err == nil {
			if available, err := manager.IsAvailable(ctx); err == nil && available {
				if info, err := manager.Info(ctx, packageName); err == nil {
					info.Installed = true // Override to reflect managed status
					message := fmt.Sprintf("Package '%s' is managed by plonk via %s", packageName, managerName)
					if len(packageLocations) > 1 {
						message += fmt.Sprintf(" (and %d other manager(s))", len(packageLocations)-1)
					}
					return InfoOutput{
						Package:     packageName,
						Status:      "managed",
						Message:     message,
						PackageInfo: info,
					}, nil
				}
			}
		}

		// Fallback to lock file info if manager call fails
		message := fmt.Sprintf("Package '%s' is managed by plonk via %s", packageName, managerName)
		if len(packageLocations) > 1 {
			message += fmt.Sprintf(" (and %d other manager(s))", len(packageLocations)-1)
		}
		return InfoOutput{
			Package: packageName,
			Status:  "managed",
			Message: message,
			PackageInfo: &packages.PackageInfo{
				Name:      packageName,
				Version:   version,
				Manager:   managerName,
				Installed: true,
			},
		}, nil
	}

	// 2. Check if package is installed (but not managed)
	availableManagers, err := getAvailableManagersMap(ctx)
	if err != nil {
		return InfoOutput{}, err
	}

	if len(availableManagers) == 0 {
		return InfoOutput{
			Package: packageName,
			Status:  "no-managers",
			Message: "No package managers are available on this system",
		}, nil
	}

	for name, manager := range availableManagers {
		installed, err := manager.IsInstalled(ctx, packageName)
		if err != nil {
			continue // Skip managers that fail
		}
		if installed {
			// Get detailed info
			info, err := manager.Info(ctx, packageName)
			if err != nil {
				// Fallback info if detailed lookup fails
				info = &packages.PackageInfo{
					Name:      packageName,
					Manager:   name,
					Installed: true,
				}
			}
			return InfoOutput{
				Package:     packageName,
				Status:      "installed",
				Message:     fmt.Sprintf("Package '%s' is installed via %s (not managed by plonk)", packageName, name),
				PackageInfo: info,
			}, nil
		}
	}

	// 3. Package not installed, search for available packages
	for name, manager := range availableManagers {
		info, err := manager.Info(ctx, packageName)
		if err != nil {
			continue // Package not found in this manager
		}

		return InfoOutput{
			Package:     packageName,
			Status:      "available",
			Message:     fmt.Sprintf("Package '%s' available in %s (not installed)", packageName, name),
			PackageInfo: info,
		}, nil
	}

	// Package not found anywhere
	return InfoOutput{
		Package: packageName,
		Status:  "not-found",
		Message: fmt.Sprintf("Package '%s' not found in any available package manager", packageName),
	}, nil
}

// Output structures

type InfoOutput struct {
	Package     string                `json:"package" yaml:"package"`
	Status      string                `json:"status" yaml:"status"`
	Message     string                `json:"message" yaml:"message"`
	PackageInfo *packages.PackageInfo `json:"package_info,omitempty" yaml:"package_info,omitempty"`
}
