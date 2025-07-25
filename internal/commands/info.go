// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <package>",
	Short: "Show detailed information about a package",
	Long: `Show detailed information about a package including version, description, homepage,
dependencies, and installation status.

The info command shows different information based on whether the package is installed:

1. If the package is installed: Shows installed version and additional details
2. If not installed: Shows available version and details from package repositories

Examples:
  plonk info git              # Show information about git package
  plonk info typescript       # Show information about typescript package
  plonk info --output json go # Show information in JSON format`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	packageName := args[0]

	// Create context
	ctx := context.Background()

	// Perform info lookup
	infoResult, err := performPackageInfo(ctx, packageName)
	if err != nil {
		return fmt.Errorf("failed to get package information: %w", err)
	}

	return RenderOutput(infoResult, format)
}

// performPackageInfo performs the info lookup according to the specified behavior
func performPackageInfo(ctx context.Context, packageName string) (InfoOutput, error) {
	// Get config directory and initialize lock service
	configDir := config.GetDefaultConfigDirectory()
	lockService := lock.NewYAMLLockService(configDir)

	// Check lock file first to see if package is managed
	packageLocations := lockService.FindPackage(packageName)

	// Get available managers
	availableManagers, err := getAvailableManagers(ctx)
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to get available managers: %w", err)
	}

	if len(availableManagers) == 0 {
		return InfoOutput{
			Package: packageName,
			Status:  "no-managers",
			Message: "No package managers are available on this system",
		}, nil
	}

	// If package is in lock file, use info from lock file
	if len(packageLocations) > 0 {
		// Use the first location found (TODO: handle multiple installations)
		location := packageLocations[0]

		// Build message that mentions if there are multiple installations
		message := fmt.Sprintf("Package '%s' is installed via %s", packageName, location.Manager)
		if len(packageLocations) > 1 {
			message += fmt.Sprintf(" (and %d other manager(s))", len(packageLocations)-1)
		}

		return InfoOutput{
			Package: packageName,
			Status:  "installed",
			Message: message,
			PackageInfo: &packages.PackageInfo{
				Name:      packageName,
				Version:   location.Entry.Version,
				Manager:   location.Manager,
				Installed: true,
			},
		}, nil
	}

	// Otherwise, check if package is installed and get info from the installing manager
	installedManager, packageInfo, err := findInstalledPackageInfo(ctx, packageName, availableManagers)
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to check if package %s is installed: %w", packageName, err)
	}

	if installedManager != "" {
		return InfoOutput{
			Package:     packageName,
			Status:      "installed",
			Message:     fmt.Sprintf("Package '%s' is installed via %s", packageName, installedManager),
			PackageInfo: packageInfo,
		}, nil
	}

	// Package is not installed, determine search strategy
	defaultManager, err := getDefaultManager()
	if err != nil {
		return InfoOutput{}, fmt.Errorf("failed to get default manager: %w", err)
	}

	if defaultManager != "" {
		// We have a default manager, check it first
		return getInfoWithDefaultManager(ctx, packageName, defaultManager, availableManagers)
	} else {
		// No default manager, check all managers
		return getInfoFromAllManagers(ctx, packageName, availableManagers)
	}
}

// findInstalledPackageInfo checks if the package is installed by any manager and returns its info
func findInstalledPackageInfo(ctx context.Context, packageName string, managers map[string]packages.PackageManager) (string, *packages.PackageInfo, error) {
	for name, manager := range managers {
		installed, err := manager.IsInstalled(ctx, packageName)
		if err != nil {
			return "", nil, fmt.Errorf("failed to check if package %s is installed in %s: %w", packageName, name, err)
		}
		if installed {
			info, err := manager.Info(ctx, packageName)
			if err != nil {
				return "", nil, fmt.Errorf("failed to get package %s info from %s: %w", packageName, name, err)
			}
			return name, info, nil
		}
	}
	return "", nil, nil
}

// getInfoWithDefaultManager gets info from the default manager first, then others if needed
func getInfoWithDefaultManager(ctx context.Context, packageName string, defaultManager string, availableManagers map[string]packages.PackageManager) (InfoOutput, error) {
	// Check default manager first
	defaultMgr, exists := availableManagers[defaultManager]
	if !exists {
		return InfoOutput{}, fmt.Errorf("default manager '%s' is not available", defaultManager)
	}

	info, err := defaultMgr.Info(ctx, packageName)
	if err != nil {
		// If not found in default manager, try other managers
		otherManagers := make(map[string]packages.PackageManager)
		for name, manager := range availableManagers {
			if name != defaultManager {
				otherManagers[name] = manager
			}
		}
		return getInfoFromOtherManagers(ctx, packageName, defaultManager, otherManagers)
	}

	return InfoOutput{
		Package:     packageName,
		Status:      "available-default",
		Message:     fmt.Sprintf("Package '%s' available in %s (default)", packageName, defaultManager),
		PackageInfo: info,
	}, nil
}

// getInfoFromAllManagers gets info from all available managers
func getInfoFromAllManagers(ctx context.Context, packageName string, availableManagers map[string]packages.PackageManager) (InfoOutput, error) {
	var foundInfo *packages.PackageInfo
	var foundManager string

	for name, manager := range availableManagers {
		info, err := manager.Info(ctx, packageName)
		if err != nil {
			continue // Package not found in this manager
		}

		// Use the first manager where we find the package
		if foundInfo == nil {
			foundInfo = info
			foundManager = name
		}
	}

	if foundInfo == nil {
		return InfoOutput{
			Package: packageName,
			Status:  "not-found",
			Message: fmt.Sprintf("Package '%s' not found in any package manager", packageName),
		}, nil
	}

	return InfoOutput{
		Package:     packageName,
		Status:      "available-multiple",
		Message:     fmt.Sprintf("Package '%s' available from %s", packageName, foundManager),
		PackageInfo: foundInfo,
	}, nil
}

// getInfoFromOtherManagers gets info from managers other than the default
func getInfoFromOtherManagers(ctx context.Context, packageName string, defaultManager string, otherManagers map[string]packages.PackageManager) (InfoOutput, error) {
	var foundInfo *packages.PackageInfo
	var foundManager string

	for name, manager := range otherManagers {
		info, err := manager.Info(ctx, packageName)
		if err != nil {
			continue // Package not found in this manager
		}

		// Use the first manager where we find the package
		if foundInfo == nil {
			foundInfo = info
			foundManager = name
		}
	}

	if foundInfo == nil {
		return InfoOutput{
			Package: packageName,
			Status:  "not-found",
			Message: fmt.Sprintf("Package '%s' not found in %s (default) or any other package manager", packageName, defaultManager),
		}, nil
	}

	return InfoOutput{
		Package:     packageName,
		Status:      "available-other",
		Message:     fmt.Sprintf("Package '%s' not found in %s (default), but available from %s", packageName, defaultManager, foundManager),
		PackageInfo: foundInfo,
	}, nil
}

// Output structures

type InfoOutput struct {
	Package     string                `json:"package" yaml:"package"`
	Status      string                `json:"status" yaml:"status"`
	Message     string                `json:"message" yaml:"message"`
	PackageInfo *packages.PackageInfo `json:"package_info,omitempty" yaml:"package_info,omitempty"`
}

// TableOutput generates human-friendly table output for info command
func (i InfoOutput) TableOutput() string {
	var output strings.Builder

	switch i.Status {
	case "installed":
		output.WriteString(fmt.Sprintf("✅ %s\n", i.Message))
		if i.PackageInfo != nil {
			output.WriteString(i.formatPackageInfo())
		}

	case "available-default":
		output.WriteString(fmt.Sprintf("📦 %s\n", i.Message))
		if i.PackageInfo != nil {
			output.WriteString(i.formatPackageInfo())
		}

	case "available-multiple":
		output.WriteString(fmt.Sprintf("📦 %s\n", i.Message))
		if i.PackageInfo != nil {
			output.WriteString(i.formatPackageInfo())
		}

	case "available-other":
		output.WriteString(fmt.Sprintf("📦 %s\n", i.Message))
		if i.PackageInfo != nil {
			output.WriteString(i.formatPackageInfo())
		}

	case "not-found":
		output.WriteString(fmt.Sprintf("❌ %s\n", i.Message))

	case "no-managers":
		output.WriteString(fmt.Sprintf("⚠️  %s\n", i.Message))
		output.WriteString("\nPlease install a package manager (Homebrew or NPM) to get package information.\n")

	default:
		output.WriteString(fmt.Sprintf("❓ %s\n", i.Message))
	}

	return output.String()
}

// formatPackageInfo formats the package information for display
func (i InfoOutput) formatPackageInfo() string {
	if i.PackageInfo == nil {
		return ""
	}

	var output strings.Builder
	info := i.PackageInfo

	output.WriteString("\n")
	output.WriteString(fmt.Sprintf("Name: %s\n", info.Name))

	if info.Version != "" {
		output.WriteString(fmt.Sprintf("Version: %s\n", info.Version))
	}

	if info.Description != "" {
		output.WriteString(fmt.Sprintf("Description: %s\n", info.Description))
	}

	if info.Homepage != "" {
		output.WriteString(fmt.Sprintf("Homepage: %s\n", info.Homepage))
	}

	output.WriteString(fmt.Sprintf("Manager: %s\n", info.Manager))
	output.WriteString(fmt.Sprintf("Installed: %t\n", info.Installed))

	if info.InstalledSize != "" {
		output.WriteString(fmt.Sprintf("Size: %s\n", info.InstalledSize))
	}

	if len(info.Dependencies) > 0 {
		output.WriteString(fmt.Sprintf("Dependencies (%d):\n", len(info.Dependencies)))
		for _, dep := range info.Dependencies {
			output.WriteString(fmt.Sprintf("  • %s\n", dep))
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (i InfoOutput) StructuredData() any {
	return i
}
