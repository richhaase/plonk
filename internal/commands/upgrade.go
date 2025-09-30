// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [manager:package|package|manager] ...",
	Short: "Upgrade packages across supported package managers",
	Long: `Upgrade packages managed by plonk to their latest versions.

This command only upgrades packages that are currently tracked in your lock file.
You can upgrade all packages, packages from specific managers, or individual packages.

Examples:
  plonk upgrade                      # Upgrade all packages managed by plonk
  plonk upgrade brew                 # Upgrade all Homebrew packages managed by plonk
  plonk upgrade ripgrep              # Upgrade ripgrep wherever it's managed by plonk
  plonk upgrade brew:ripgrep         # Upgrade only Homebrew's ripgrep
  plonk upgrade htop neovim         # Upgrade multiple packages
  plonk upgrade npm uv gem         # Upgrade all packages for these managers`,
	RunE:         runUpgrade,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

// upgradeSpec represents the parsed upgrade specification
type upgradeSpec struct {
	UpgradeAll     bool                // plonk upgrade (no args)
	ManagerTargets map[string][]string // manager -> packages to upgrade
}

// packageMatchInfo contains the information needed to match packages
type packageMatchInfo struct {
	Manager    string
	Name       string
	SourcePath string // for Go packages like github.com/rakyll/hey -> hey
	FullName   string // for npm scoped packages like @scope/name
	Resource   lock.ResourceEntry
}

// parseUpgradeArgs parses command arguments into upgrade specification
func parseUpgradeArgs(args []string, lockFile *lock.Lock) (upgradeSpec, error) {
	spec := upgradeSpec{
		ManagerTargets: make(map[string][]string),
	}

	if len(args) == 0 {
		spec.UpgradeAll = true
		return spec, nil
	}

	// Build enhanced package matching info from lock file
	var packageInfos []packageMatchInfo
	lockPackages := make(map[string][]string) // manager -> []packages (for backwards compatibility)

	for _, resource := range lockFile.Resources {
		if resource.Type == "package" {
			manager := extractManagerFromResource(resource)
			packageName := extractPackageNameFromResource(resource)
			if manager != "" && packageName != "" {
				lockPackages[manager] = append(lockPackages[manager], packageName)

				// Build enhanced matching info
				matchInfo := packageMatchInfo{
					Manager:  manager,
					Name:     packageName,
					Resource: resource,
				}

				// Extract source_path for Go packages
				if sourcePath, ok := resource.Metadata["source_path"]; ok {
					if sourcePathStr, ok := sourcePath.(string); ok {
						matchInfo.SourcePath = sourcePathStr
					}
				}

				// Extract full_name for npm scoped packages
				if fullName, ok := resource.Metadata["full_name"]; ok {
					if fullNameStr, ok := fullName.(string); ok {
						matchInfo.FullName = fullNameStr
					}
				}

				packageInfos = append(packageInfos, matchInfo)
			}
		}
	}

	for _, arg := range args {
		if strings.Contains(arg, ":") {
			// manager:package format
			parts := strings.SplitN(arg, ":", 2)
			manager := parts[0]
			packageName := parts[1]

			if packageName == "" {
				// "manager:" format is not supported - users should use just "manager"
				return spec, fmt.Errorf("invalid syntax '%s' - use '%s' to upgrade all packages for this manager", arg, manager)
			}

			// "manager:package" format - upgrade specific package
			matchedInfo := findPackageMatch(packageInfos, manager, packageName)
			if matchedInfo == nil {
				return spec, fmt.Errorf("package '%s' is not managed by plonk via '%s'", packageName, manager)
			}

			// Use the actual package name or source path as needed for the upgrade
			upgradeTarget := determineUpgradeTarget(*matchedInfo, packageName)
			spec.ManagerTargets[manager] = append(spec.ManagerTargets[manager], upgradeTarget)
		} else {
			// Could be package name or manager name

			// Check if it's a manager name
			if managerPackages, exists := lockPackages[arg]; exists {
				spec.ManagerTargets[arg] = managerPackages
				continue
			}

			// Check if it's a package name across all managers
			found := false
			for _, info := range packageInfos {
				if matchesPackage(info, arg) {
					upgradeTarget := determineUpgradeTarget(info, arg)
					spec.ManagerTargets[info.Manager] = append(spec.ManagerTargets[info.Manager], upgradeTarget)
					found = true
				}
			}

			if !found {
				return spec, fmt.Errorf("package '%s' is not managed by plonk", arg)
			}
		}
	}

	return spec, nil
}

// findPackageMatch finds a package in the given manager by name, source_path, or full_name
func findPackageMatch(packageInfos []packageMatchInfo, manager, packageName string) *packageMatchInfo {
	for _, info := range packageInfos {
		if info.Manager == manager && matchesPackage(info, packageName) {
			return &info
		}
	}
	return nil
}

// matchesPackage checks if a package matches the given name by any of its identifiers
func matchesPackage(info packageMatchInfo, packageName string) bool {
	// Match by binary/package name
	if info.Name == packageName {
		return true
	}

	// Match by source path (Go packages)
	if info.SourcePath != "" && info.SourcePath == packageName {
		return true
	}

	// Match by full name (npm scoped packages)
	if info.FullName != "" && info.FullName == packageName {
		return true
	}

	return false
}

// determineUpgradeTarget determines what target to pass to the package manager for upgrade
func determineUpgradeTarget(info packageMatchInfo, requestedName string) string {
	// For Go packages, if they requested the source path, we should upgrade with source path
	// If they requested the binary name, we still need to upgrade with source path for Go
	if info.Manager == "go" && info.SourcePath != "" {
		return info.SourcePath
	}

	// For npm scoped packages, if they have full_name, use it
	if info.Manager == "npm" && info.FullName != "" {
		return info.FullName
	}

	// For other cases, use the stored name
	return info.Name
}

// extractManagerFromResource extracts manager name from lock file resource
func extractManagerFromResource(resource lock.ResourceEntry) string {
	// Try to get manager from metadata first (v2 format)
	if managerVal, ok := resource.Metadata["manager"]; ok {
		if managerStr, ok := managerVal.(string); ok {
			return managerStr
		}
	}

	// If not in metadata, extract from ID prefix (fallback)
	if strings.Contains(resource.ID, ":") {
		parts := strings.SplitN(resource.ID, ":", 2)
		return parts[0]
	}

	return ""
}

// extractPackageNameFromResource extracts package name from lock file resource
func extractPackageNameFromResource(resource lock.ResourceEntry) string {
	// Try to get name from metadata first
	if nameVal, ok := resource.Metadata["name"]; ok {
		if nameStr, ok := nameVal.(string); ok {
			return nameStr
		}
	}

	// If not in metadata, extract from ID suffix (fallback)
	if strings.Contains(resource.ID, ":") {
		parts := strings.SplitN(resource.ID, ":", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}

	// If no colon, the entire ID is the package name
	return resource.ID
}

// runUpgrade executes the upgrade command
func runUpgrade(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories and config
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Load lock file (authoritative source)
	lockService := lock.NewYAMLLockService(configDir)
	lockFile, err := lockService.Read()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	// Parse upgrade specification
	spec, err := parseUpgradeArgs(args, lockFile)
	if err != nil {
		return err
	}

	// Populate spec with all packages if upgrading all
	if spec.UpgradeAll {
		for _, resource := range lockFile.Resources {
			if resource.Type == "package" {
				manager := extractManagerFromResource(resource)
				packageName := extractPackageNameFromResource(resource)
				if manager != "" && packageName != "" {
					spec.ManagerTargets[manager] = append(spec.ManagerTargets[manager], packageName)
				}
			}
		}
	}

	if len(spec.ManagerTargets) == 0 {
		output.Printf("No packages to upgrade\n")
		return nil
	}

	// Execute upgrades with injected registry
	registry := packages.NewManagerRegistry()
	results, err := Upgrade(cmd.Context(), spec, cfg, lockService, registry)
	if err != nil {
		return err
	}

	// Create output data
	outputData := upgradeResultsToOutput(results)

	// Create formatter and render
	formatter := output.NewUpgradeFormatter(outputData)
	if err := output.RenderOutput(formatter, format); err != nil {
		return err
	}

	// Print summary
	output.Printf("\nSummary: %d upgraded, %d failed, %d skipped\n",
		results.Summary.Upgraded, results.Summary.Failed, results.Summary.Skipped)

	// Return error if any packages failed to upgrade
	if results.Summary.Failed > 0 {
		return fmt.Errorf("failed to upgrade %d packages", results.Summary.Failed)
	}

	return nil
}

// Upgrade executes the upgrade operation using the provided dependencies.
// It is a thin wrapper around the internal execution function to enable
// dependency injection in tests and other callers.
func Upgrade(ctx context.Context, spec upgradeSpec, cfg *config.Config, lockService lock.LockService, registry *packages.ManagerRegistry) (upgradeResults, error) {
	return executeUpgrade(ctx, spec, cfg, lockService, registry)
}

// packageUpgradeResult represents the result of upgrading a single package
type packageUpgradeResult struct {
	Manager     string `json:"manager"`
	Package     string `json:"package"`
	FromVersion string `json:"from_version,omitempty"`
	ToVersion   string `json:"to_version,omitempty"`
	Status      string `json:"status"` // "upgraded", "failed", "skipped"
	Error       string `json:"error,omitempty"`
}

// upgradeResults represents the complete results of an upgrade operation
type upgradeResults struct {
	Results []packageUpgradeResult `json:"upgrades"`
	Summary upgradeSummary         `json:"summary"`
}

// upgradeSummary provides summary statistics
type upgradeSummary struct {
	Total    int `json:"total"`
	Upgraded int `json:"upgraded"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// executeUpgrade performs the actual upgrade operations
func executeUpgrade(ctx context.Context, spec upgradeSpec, cfg *config.Config, lockService lock.LockService, registry *packages.ManagerRegistry) (upgradeResults, error) {
	results := upgradeResults{
		Results: []packageUpgradeResult{},
		Summary: upgradeSummary{},
	}

	totalPackages := 0
	for _, pkgs := range spec.ManagerTargets {
		totalPackages += len(pkgs)
	}
	results.Summary.Total = totalPackages

	// Create spinner manager for all operations
	spinnerManager := output.NewSpinnerManager(totalPackages)

	// Iterate managers in deterministic order
	managerNames := make([]string, 0, len(spec.ManagerTargets))
	for name := range spec.ManagerTargets {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	for _, managerName := range managerNames {
		packageNames := append([]string(nil), spec.ManagerTargets[managerName]...)
		sort.Strings(packageNames)
		// Get the package manager
		mgr, err := registry.GetManager(managerName)
		if err != nil {
			// Add failures for all packages in this manager
			for _, pkg := range packageNames {
				spinner := spinnerManager.StartSpinner("Upgrading", fmt.Sprintf("%s (%s)", pkg, managerName))
				spinner.Error(fmt.Sprintf("Failed to upgrade %s: package manager '%s' not available", pkg, managerName))

				results.Results = append(results.Results, packageUpgradeResult{
					Manager: managerName,
					Package: pkg,
					Status:  "failed",
					Error:   fmt.Sprintf("package manager '%s' not available", managerName),
				})
				results.Summary.Failed++
			}
			continue
		}

		// Check if manager is available
		available, err := mgr.IsAvailable(ctx)
		if err != nil || !available {
			// Add failures for all packages in this manager
			for _, pkg := range packageNames {
				spinner := spinnerManager.StartSpinner("Upgrading", fmt.Sprintf("%s (%s)", pkg, managerName))
				spinner.Error(fmt.Sprintf("Failed to upgrade %s: package manager '%s' is not available", pkg, managerName))

				results.Results = append(results.Results, packageUpgradeResult{
					Manager: managerName,
					Package: pkg,
					Status:  "failed",
					Error:   fmt.Sprintf("package manager '%s' is not available", managerName),
				})
				results.Summary.Failed++
			}
			continue
		}

		// Check if manager supports upgrade
		upgrader, ok := mgr.(packages.PackageUpgrader)
		if !ok {
			// Add failures for all packages in this manager
			for _, pkg := range packageNames {
				spinner := spinnerManager.StartSpinner("Upgrading", fmt.Sprintf("%s (%s)", pkg, managerName))
				spinner.Error(fmt.Sprintf("Failed to upgrade %s: package manager '%s' does not support upgrade", pkg, managerName))

				results.Results = append(results.Results, packageUpgradeResult{
					Manager: managerName,
					Package: pkg,
					Status:  "failed",
					Error:   fmt.Sprintf("package manager '%s' does not support upgrade", managerName),
				})
				results.Summary.Failed++
			}
			continue
		}

		// Upgrade each package individually to provide per-package error handling
		for _, pkg := range packageNames {
			// Start spinner for this package
			spinner := spinnerManager.StartSpinner("Upgrading", fmt.Sprintf("%s (%s)", pkg, managerName))

			result := packageUpgradeResult{
				Manager: managerName,
				Package: pkg,
			}

			// Get current version before upgrade
			if version, err := mgr.InstalledVersion(ctx, pkg); err == nil {
				result.FromVersion = version
			}

			// Upgrade this package
			upgradeErr := upgrader.Upgrade(ctx, []string{pkg})

			if upgradeErr != nil {
				result.Status = "failed"
				result.Error = upgradeErr.Error()
				results.Summary.Failed++
				spinner.Error(fmt.Sprintf("Failed to upgrade %s: %s", pkg, upgradeErr.Error()))
			} else {
				// Get new version and update lock file
				if newVersion, err := mgr.InstalledVersion(ctx, pkg); err == nil {
					result.ToVersion = newVersion
					if result.FromVersion != newVersion {
						result.Status = "upgraded"
						results.Summary.Upgraded++

						// Update lock file immediately
						if err := updateLockFileForPackage(lockService, managerName, pkg, newVersion); err != nil {
							output.Printf("Warning: Failed to update lock file for %s: %v\n", pkg, err)
						}

						spinner.Success(fmt.Sprintf("upgraded %s %s → %s", pkg, result.FromVersion, result.ToVersion))
					} else {
						result.Status = "skipped"
						results.Summary.Skipped++
						spinner.Success(fmt.Sprintf("skipped %s (already up-to-date)", pkg))
					}
				} else {
					result.Status = "upgraded" // Assume success even if we can't get version
					results.Summary.Upgraded++
					spinner.Success(fmt.Sprintf("upgraded %s", pkg))
				}
			}

			results.Results = append(results.Results, result)
		}
	}

	return results, nil
}

// updateLockFileForPackage updates the lock file with new version information
func updateLockFileForPackage(lockService lock.LockService, manager, packageName, newVersion string) error {
	lockFile, err := lockService.Read()
	if err != nil {
		return err
	}

	// Find and update the resource
	for i, resource := range lockFile.Resources {
		if resource.Type == "package" {
			resourceManager := extractManagerFromResource(resource)
			resourcePackageName := extractPackageNameFromResource(resource)
			if resourceManager == manager && resourcePackageName == packageName {
				if lockFile.Resources[i].Metadata == nil {
					lockFile.Resources[i].Metadata = make(map[string]interface{})
				}
				lockFile.Resources[i].Metadata["version"] = newVersion
				lockFile.Resources[i].InstalledAt = time.Now().Format(time.RFC3339)
				break
			}
		}
	}

	return lockService.Write(lockFile)
}

// upgradeResultsToOutput converts upgrade results to output format
func upgradeResultsToOutput(results upgradeResults) output.UpgradeOutput {
	return output.UpgradeOutput{
		Command:    "upgrade",
		TotalItems: results.Summary.Total,
		Results:    convertUpgradeResults(results.Results),
		Summary: output.UpgradeSummary{
			Total:    results.Summary.Total,
			Upgraded: results.Summary.Upgraded,
			Failed:   results.Summary.Failed,
			Skipped:  results.Summary.Skipped,
		},
	}
}

// convertUpgradeResults converts internal results to output format
func convertUpgradeResults(results []packageUpgradeResult) []output.UpgradeResult {
	var outputResults []output.UpgradeResult
	for _, result := range results {
		outputResults = append(outputResults, output.UpgradeResult{
			Manager:     result.Manager,
			Package:     result.Package,
			FromVersion: result.FromVersion,
			ToVersion:   result.ToVersion,
			Status:      result.Status,
			Error:       result.Error,
		})
	}
	return outputResults
}
