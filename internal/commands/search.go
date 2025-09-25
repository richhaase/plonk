// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
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
	Args:         cobra.ExactArgs(1),
	RunE:         runSearch,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Load configuration
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Execute pure search logic
	res, err := Search(cmd.Context(), cfg, args[0])
	if err != nil {
		return err
	}

	// Convert to output package type and create formatter
	formatterData := output.SearchOutput{
		Package: res.Package,
		Status:  res.Status,
		Message: res.Message,
		Results: convertSearchResults(res.Results),
	}
	formatter := output.NewSearchFormatter(formatterData)
	return output.RenderOutput(formatter, format)
}

// Search performs package search based on config and returns typed results (pure logic)
func Search(ctx context.Context, cfg *config.Config, packageSpec string) (SearchOutput, error) {
	// Parse and validate search specification
	spec, err := packages.ValidateSpec(packageSpec, packages.ValidationModeSearch, "")
	if err != nil {
		return SearchOutput{}, err
	}

	manager := spec.Manager
	packageName := spec.Name

	// Create context with configurable timeout
	t := config.GetTimeouts(cfg)
	ctx, cancel := context.WithTimeout(ctx, t.Operation)
	defer cancel()

	// Perform search
	if manager != "" {
		return searchSpecificManager(ctx, manager, packageName)
	}
	return searchAllManagersParallel(ctx, cfg, packageName)
}

// convertSearchResults converts from command types to output types
func convertSearchResults(results []SearchResultEntry) []output.SearchResultEntry {
	converted := make([]output.SearchResultEntry, len(results))
	for i, result := range results {
		converted[i] = output.SearchResultEntry{
			Manager:  result.Manager,
			Packages: result.Packages,
		}
	}
	return converted
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
func searchAllManagersParallel(ctx context.Context, cfg *config.Config, packageName string) (SearchOutput, error) {
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

			// Create a child context for this manager with configurable timeout
			t := config.GetTimeouts(cfg)
			managerCtx, cancel := context.WithTimeout(ctx, t.Operation)
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
	var errors []error

	for result := range resultsChan {
		if result.Error != nil {
			// Handle timeout or other errors gracefully
			if ctx.Err() == context.DeadlineExceeded {
				errors = append(errors, fmt.Errorf("%s: timeout", result.Manager))
			} else {
				errors = append(errors, fmt.Errorf("%s: %w", result.Manager, result.Error))
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

	// Sort results for determinism
	sort.Slice(searchResults, func(i, j int) bool { return searchResults[i].Manager < searchResults[j].Manager })
	for i := range searchResults {
		sort.Strings(searchResults[i].Packages)
	}

	// Build response
	if len(searchResults) == 0 {
		message := fmt.Sprintf("Package '%s' not found in any available package manager", packageName)
		if len(errors) > 0 {
			var errorStrings []string
			for _, err := range errors {
				errorStrings = append(errorStrings, err.Error())
			}
			message += fmt.Sprintf(" (errors: %s)", strings.Join(errorStrings, ", "))
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
		var errorStrings []string
		for _, err := range errors {
			errorStrings = append(errorStrings, err.Error())
		}
		message += fmt.Sprintf(" (some managers had errors: %s)", strings.Join(errorStrings, ", "))
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
