// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <package>",
	Short: "Search for packages across package managers",
	Long: `Search for packages across available package managers in parallel.

Searches all available package managers simultaneously with a 3-second timeout.
Use prefix syntax to search only a specific manager.

Examples:
  plonk search git              # Search all managers for git package
  plonk search typescript       # Search all managers for typescript
  plonk search brew:ripgrep     # Search only Homebrew for ripgrep
  plonk search npm:@types/node  # Search only npm for @types/node`,
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
		return fmt.Errorf("invalid output format: %w", err)
	}

	packageSpec := args[0]

	// Parse prefix syntax
	manager, packageName := ParsePackageSpec(packageSpec)

	// Validate manager if prefix specified
	if manager != "" && !IsValidManager(manager) {
		errorMsg := FormatNotFoundError("package manager", manager, GetValidManagers())
		return fmt.Errorf("%s", errorMsg)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Perform search
	var searchResult SearchOutput
	if manager != "" {
		// Search specific manager
		searchResult, err = searchSpecificManager(ctx, manager, packageName)
	} else {
		// Search all managers in parallel
		searchResult, err = searchAllManagersParallel(ctx, packageName)
	}

	if err != nil {
		return err
	}

	return RenderOutput(searchResult, format)
}

// searchSpecificManager searches only the specified manager
func searchSpecificManager(ctx context.Context, managerName, packageName string) (SearchOutput, error) {
	// Get the specific manager
	registry := packages.NewManagerRegistry()
	manager, err := registry.GetManager(managerName)
	if err != nil {
		return SearchOutput{}, fmt.Errorf("failed to get manager %s: %w", managerName, err)
	}

	// Check if manager is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		return SearchOutput{}, fmt.Errorf("failed to check %s availability: %w", managerName, err)
	}
	if !available {
		return SearchOutput{
			Package: packageName,
			Status:  "manager-unavailable",
			Message: fmt.Sprintf("Package manager '%s' is not available on this system", managerName),
		}, nil
	}

	// Search for the package
	results, err := manager.Search(ctx, packageName)
	if err != nil {
		return SearchOutput{}, fmt.Errorf("failed to search for %s in %s: %w", packageName, managerName, err)
	}

	if len(results) == 0 {
		return SearchOutput{
			Package: packageName,
			Status:  "not-found",
			Message: fmt.Sprintf("Package '%s' not found in %s", packageName, managerName),
		}, nil
	}

	return SearchOutput{
		Package: packageName,
		Status:  "found",
		Message: fmt.Sprintf("Found %d result(s) for '%s' in %s", len(results), packageName, managerName),
		Results: []SearchResultEntry{
			{
				Manager:  managerName,
				Packages: results,
			},
		},
	}, nil
}

// searchAllManagersParallel searches all managers in parallel with timeout
func searchAllManagersParallel(ctx context.Context, packageName string) (SearchOutput, error) {
	// Get available managers
	availableManagers, err := getAvailableManagersMap(ctx)
	if err != nil {
		return SearchOutput{}, err
	}

	if len(availableManagers) == 0 {
		return SearchOutput{
			Package: packageName,
			Status:  "no-managers",
			Message: "No package managers are available on this system",
		}, nil
	}

	// Channel for results and wait group for synchronization
	type managerResult struct {
		Manager  string
		Packages []string
		Error    error
	}

	resultsChan := make(chan managerResult, len(availableManagers))
	var wg sync.WaitGroup

	// Search each manager in parallel
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(managerName string, mgr packages.PackageManager) {
			defer wg.Done()

			// Create a child context for this manager
			managerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			results, err := mgr.Search(managerCtx, packageName)
			resultsChan <- managerResult{
				Manager:  managerName,
				Packages: results,
				Error:    err,
			}
		}(name, manager)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var searchResults []SearchResultEntry
	var errors []string

	for result := range resultsChan {
		if result.Error != nil {
			// Handle timeout or other errors gracefully
			if ctx.Err() == context.DeadlineExceeded {
				errors = append(errors, fmt.Sprintf("%s: timeout", result.Manager))
			} else {
				errors = append(errors, fmt.Sprintf("%s: %v", result.Manager, result.Error))
			}
			continue
		}

		if len(result.Packages) > 0 {
			searchResults = append(searchResults, SearchResultEntry{
				Manager:  result.Manager,
				Packages: result.Packages,
			})
		}
	}

	// Build response
	if len(searchResults) == 0 {
		message := fmt.Sprintf("Package '%s' not found in any available package manager", packageName)
		if len(errors) > 0 {
			message += fmt.Sprintf(" (errors: %s)", strings.Join(errors, ", "))
		}
		return SearchOutput{
			Package: packageName,
			Status:  "not-found",
			Message: message,
		}, nil
	}

	totalResults := 0
	managerNames := make([]string, len(searchResults))
	for i, result := range searchResults {
		totalResults += len(result.Packages)
		managerNames[i] = result.Manager
	}

	message := fmt.Sprintf("Found %d result(s) for '%s' across %d manager(s): %s",
		totalResults, packageName, len(searchResults), strings.Join(managerNames, ", "))

	if len(errors) > 0 {
		message += fmt.Sprintf(" (some managers had errors: %s)", strings.Join(errors, ", "))
	}

	return SearchOutput{
		Package: packageName,
		Status:  "found-multiple",
		Message: message,
		Results: searchResults,
	}, nil
}

// getAvailableManagersMap returns a map of available package managers
func getAvailableManagersMap(ctx context.Context) (map[string]packages.PackageManager, error) {
	registry := packages.NewManagerRegistry()
	availableManagers := make(map[string]packages.PackageManager)

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
			return nil, err
		} else if available {
			availableManagers[name] = manager
		}
	}

	return availableManagers, nil
}

// Output structures

type SearchResultEntry struct {
	Manager  string   `json:"manager" yaml:"manager"`
	Packages []string `json:"packages" yaml:"packages"`
}

type SearchOutput struct {
	Package string              `json:"package" yaml:"package"`
	Status  string              `json:"status" yaml:"status"`
	Message string              `json:"message" yaml:"message"`
	Results []SearchResultEntry `json:"results,omitempty" yaml:"results,omitempty"`
}

// TableOutput generates human-friendly table output for search command
func (s SearchOutput) TableOutput() string {
	var output strings.Builder

	switch s.Status {
	case "found":
		output.WriteString(fmt.Sprintf("📦 %s\n", s.Message))
		if len(s.Results) > 0 && len(s.Results[0].Packages) > 0 {
			output.WriteString("\nMatching packages:\n")
			for _, pkg := range s.Results[0].Packages {
				output.WriteString(fmt.Sprintf("  • %s\n", pkg))
			}
			output.WriteString(fmt.Sprintf("\nInstall with: plonk install %s:%s\n", s.Results[0].Manager, s.Package))
		}

	case "found-multiple":
		output.WriteString(fmt.Sprintf("📦 %s\n", s.Message))
		output.WriteString("\nResults by manager:\n")
		for _, result := range s.Results {
			output.WriteString(fmt.Sprintf("\n%s:\n", result.Manager))
			for _, pkg := range result.Packages {
				output.WriteString(fmt.Sprintf("  • %s\n", pkg))
			}
		}
		output.WriteString(fmt.Sprintf("\nInstall examples:\n"))
		for _, result := range s.Results {
			output.WriteString(fmt.Sprintf("  • plonk install %s:%s\n", result.Manager, s.Package))
		}

	case "not-found":
		output.WriteString(fmt.Sprintf("❌ %s\n", s.Message))

	case "no-managers":
		output.WriteString(fmt.Sprintf("⚠️  %s\n", s.Message))
		output.WriteString("\nPlease install a package manager (Homebrew or NPM) to search for packages.\n")

	case "manager-unavailable":
		output.WriteString(fmt.Sprintf("⚠️  %s\n", s.Message))

	default:
		output.WriteString(fmt.Sprintf("❓ %s\n", s.Message))
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (s SearchOutput) StructuredData() any {
	return s
}
