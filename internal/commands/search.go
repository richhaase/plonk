// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <package>",
	Short: "Search for packages across package managers",
	Long: `Search for packages across available package managers.

The search behavior depends on installation status and configuration:

1. If the package is installed: Shows which manager installed it
2. If not installed and available in default manager: Shows results from default manager
3. If not installed and not in default manager: Shows all managers that have the package
4. If no default manager is configured: Shows all managers that have the package

Examples:
  plonk search git              # Search for git package
  plonk search typescript       # Search for typescript package
  plonk search nonexistent      # Search for package that doesn't exist`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "search", "output-format", "invalid output format")
	}

	packageName := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform search
	searchResult, err := performPackageSearch(ctx, packageName)
	if err != nil {
		return errors.Wrap(err, errors.ErrCommandExecution, errors.DomainCommands, "search", "failed to search for package")
	}

	return RenderOutput(searchResult, format)
}

// performPackageSearch performs the search logic according to the specified behavior
func performPackageSearch(ctx context.Context, packageName string) (SearchOutput, error) {
	// Get available managers
	availableManagers, err := getAvailableManagers(ctx)
	if err != nil {
		return SearchOutput{}, errors.Wrap(err, errors.ErrInternal, errors.DomainPackages, "get-managers", "failed to get available managers")
	}

	if len(availableManagers) == 0 {
		return SearchOutput{
			Package: packageName,
			Status:  "no-managers",
			Message: "No package managers are available on this system",
		}, nil
	}

	// Check if package is installed
	installedManager, err := findInstalledPackage(ctx, packageName, availableManagers)
	if err != nil {
		return SearchOutput{}, errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "check", packageName, "failed to check if package is installed")
	}

	if installedManager != "" {
		return SearchOutput{
			Package:          packageName,
			Status:           "installed",
			Message:          fmt.Sprintf("Package '%s' is installed via %s", packageName, installedManager),
			InstalledManager: installedManager,
		}, nil
	}

	// Package is not installed, determine search strategy
	defaultManager, err := getDefaultManager()
	if err != nil {
		return SearchOutput{}, errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "get-default", "failed to get default manager")
	}

	if defaultManager != "" {
		// We have a default manager, search it first
		return searchWithDefaultManager(ctx, packageName, defaultManager, availableManagers)
	} else {
		// No default manager, search all managers
		return searchAllManagers(ctx, packageName, availableManagers)
	}
}

// getAvailableManagers returns a map of available package managers
func getAvailableManagers(ctx context.Context) (map[string]managers.PackageManager, error) {
	sharedCtx := runtime.GetSharedContext()
	registry := sharedCtx.ManagerRegistry()
	availableManagers := make(map[string]managers.PackageManager)

	for _, name := range registry.GetAllManagerNames() {
		// Skip cargo for search since it doesn't support search well
		if name == "cargo" {
			continue
		}

		manager, err := registry.GetManager(name)
		if err != nil {
			continue // Skip unsupported managers
		}

		if available, err := manager.IsAvailable(ctx); err != nil {
			return nil, errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", name, "failed to check manager availability")
		} else if available {
			availableManagers[name] = manager
		}
	}

	return availableManagers, nil
}

// findInstalledPackage checks if the package is installed by any manager
func findInstalledPackage(ctx context.Context, packageName string, managers map[string]managers.PackageManager) (string, error) {
	for name, manager := range managers {
		installed, err := manager.IsInstalled(ctx, packageName)
		if err != nil {
			return "", errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "check", packageName, fmt.Sprintf("failed to check if package is installed in %s", name))
		}
		if installed {
			return name, nil
		}
	}
	return "", nil
}

// getDefaultManager gets the default manager from configuration
func getDefaultManager() (string, error) {
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadConfigWithDefaults(configDir)

	return cfg.Resolve().GetDefaultManager(), nil
}

// searchWithDefaultManager searches the default manager first, then others if needed
func searchWithDefaultManager(ctx context.Context, packageName string, defaultManager string, availableManagers map[string]managers.PackageManager) (SearchOutput, error) {
	// Search default manager first
	defaultMgr, exists := availableManagers[defaultManager]
	if !exists {
		return SearchOutput{}, errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "check", fmt.Sprintf("default manager '%s' is not available", defaultManager)).WithSuggestionMessage(getManagerInstallSuggestion(defaultManager))
	}

	results, err := defaultMgr.Search(ctx, packageName)
	if err != nil {
		return SearchOutput{}, errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "search", packageName, fmt.Sprintf("failed to search in %s", defaultManager))
	}

	// Check if package is found in default manager
	found := false
	for _, result := range results {
		if result == packageName {
			found = true
			break
		}
	}

	if found {
		return SearchOutput{
			Package:        packageName,
			Status:         "found-default",
			Message:        fmt.Sprintf("Package '%s' available in %s (default)", packageName, defaultManager),
			DefaultManager: defaultManager,
			Results:        results,
		}, nil
	}

	// Not found in default manager, search other managers
	otherManagers := make(map[string]managers.PackageManager)
	for name, manager := range availableManagers {
		if name != defaultManager {
			otherManagers[name] = manager
		}
	}

	return searchOtherManagers(ctx, packageName, defaultManager, otherManagers)
}

// searchAllManagers searches all available managers
func searchAllManagers(ctx context.Context, packageName string, availableManagers map[string]managers.PackageManager) (SearchOutput, error) {
	var foundManagers []string
	allResults := make(map[string][]string)

	for name, manager := range availableManagers {
		results, err := manager.Search(ctx, packageName)
		if err != nil {
			return SearchOutput{}, errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "search", packageName, fmt.Sprintf("failed to search in %s", name))
		}

		// Check if exact package name is found
		found := false
		for _, result := range results {
			if result == packageName {
				found = true
				break
			}
		}

		if found {
			foundManagers = append(foundManagers, name)
			allResults[name] = results
		}
	}

	if len(foundManagers) == 0 {
		return SearchOutput{
			Package: packageName,
			Status:  "not-found",
			Message: fmt.Sprintf("Package '%s' not found in any package manager", packageName),
		}, nil
	}

	return SearchOutput{
		Package:        packageName,
		Status:         "found-multiple",
		Message:        fmt.Sprintf("Package '%s' available from: %s", packageName, strings.Join(foundManagers, ", ")),
		FoundManagers:  foundManagers,
		ManagerResults: allResults,
	}, nil
}

// searchOtherManagers searches managers other than the default
func searchOtherManagers(ctx context.Context, packageName string, defaultManager string, otherManagers map[string]managers.PackageManager) (SearchOutput, error) {
	var foundManagers []string
	allResults := make(map[string][]string)

	for name, manager := range otherManagers {
		results, err := manager.Search(ctx, packageName)
		if err != nil {
			return SearchOutput{}, errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "search", packageName, fmt.Sprintf("failed to search in %s", name))
		}

		// Check if exact package name is found
		found := false
		for _, result := range results {
			if result == packageName {
				found = true
				break
			}
		}

		if found {
			foundManagers = append(foundManagers, name)
			allResults[name] = results
		}
	}

	if len(foundManagers) == 0 {
		return SearchOutput{
			Package:        packageName,
			Status:         "not-found",
			Message:        fmt.Sprintf("Package '%s' not found in %s (default) or any other package manager", packageName, defaultManager),
			DefaultManager: defaultManager,
		}, nil
	}

	return SearchOutput{
		Package:        packageName,
		Status:         "found-other",
		Message:        fmt.Sprintf("Package '%s' not found in %s (default), but available from: %s", packageName, defaultManager, strings.Join(foundManagers, ", ")),
		DefaultManager: defaultManager,
		FoundManagers:  foundManagers,
		ManagerResults: allResults,
	}, nil
}

// Output structures

type SearchOutput struct {
	Package          string              `json:"package" yaml:"package"`
	Status           string              `json:"status" yaml:"status"`
	Message          string              `json:"message" yaml:"message"`
	InstalledManager string              `json:"installed_manager,omitempty" yaml:"installed_manager,omitempty"`
	DefaultManager   string              `json:"default_manager,omitempty" yaml:"default_manager,omitempty"`
	FoundManagers    []string            `json:"found_managers,omitempty" yaml:"found_managers,omitempty"`
	Results          []string            `json:"results,omitempty" yaml:"results,omitempty"`
	ManagerResults   map[string][]string `json:"manager_results,omitempty" yaml:"manager_results,omitempty"`
}

// TableOutput generates human-friendly table output for search command
func (s SearchOutput) TableOutput() string {
	var output strings.Builder

	switch s.Status {
	case "installed":
		output.WriteString(fmt.Sprintf("✅ %s\n", s.Message))

	case "found-default":
		output.WriteString(fmt.Sprintf("📦 %s\n", s.Message))
		if len(s.Results) > 1 {
			output.WriteString("\nRelated packages:\n")
			for _, result := range s.Results {
				if result != s.Package {
					output.WriteString(fmt.Sprintf("  • %s\n", result))
				}
			}
		}

	case "found-multiple":
		output.WriteString(fmt.Sprintf("📦 %s\n", s.Message))
		output.WriteString("\nInstall with:\n")
		for _, manager := range s.FoundManagers {
			output.WriteString(fmt.Sprintf("  • %s: plonk pkg add %s\n", manager, s.Package))
		}

	case "found-other":
		output.WriteString(fmt.Sprintf("📦 %s\n", s.Message))
		output.WriteString("\nInstall with:\n")
		for _, manager := range s.FoundManagers {
			output.WriteString(fmt.Sprintf("  • %s: plonk pkg add %s\n", manager, s.Package))
		}

	case "not-found":
		output.WriteString(fmt.Sprintf("❌ %s\n", s.Message))

	case "no-managers":
		output.WriteString(fmt.Sprintf("⚠️  %s\n", s.Message))
		output.WriteString("\nPlease install a package manager (Homebrew or NPM) to search for packages.\n")

	default:
		output.WriteString(fmt.Sprintf("❓ %s\n", s.Message))
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (s SearchOutput) StructuredData() any {
	return s
}
