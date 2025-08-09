// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// CondaManager manages conda packages using intelligent binary detection (mamba → conda).
type CondaManager struct {
	binary   string // Detected binary: "mamba" or "conda"
	useMamba bool   // Performance optimization indicator
}

// NewCondaManager creates a new conda manager with intelligent binary detection.
func NewCondaManager() *CondaManager {
	binary, useMamba := detectCondaVariant()
	return &CondaManager{
		binary:   binary,
		useMamba: useMamba,
	}
}

// detectCondaVariant performs two-way conda variant detection (mamba → conda).
func detectCondaVariant() (string, bool) {
	// Priority order: mamba → conda (mamba includes micromamba's mamba command)

	// Try mamba first (10-100x faster than conda, includes all mamba variants)
	if CheckCommandAvailable("mamba") && isCondaVariantFunctional("mamba") {
		return "mamba", true
	}

	// Fall back to conda (reliable baseline)
	if CheckCommandAvailable("conda") && isCondaVariantFunctional("conda") {
		return "conda", false
	}

	// Return conda as default (will fail later in IsAvailable if not found)
	return "conda", false
}

// isCondaVariantFunctional validates binary using existing plonk patterns.
func isCondaVariantFunctional(binary string) bool {
	// Follow existing plonk validation pattern exactly
	if !CheckCommandAvailable(binary) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := VerifyBinary(ctx, binary, []string{"--version"})
	return err == nil
}

// CondaListOutput represents the structure of conda/mamba list -n base --json output.
type CondaListOutput []CondaListItem

// CondaListItem represents a single package in conda list output.
type CondaListItem struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Build       string `json:"build"`
	Channel     string `json:"channel"`
	Platform    string `json:"platform"`
	BuildString string `json:"build_string"`
}

// CondaSearchOutput represents the structure of conda/mamba search --json output.
type CondaSearchOutput map[string][]CondaSearchItem

// CondaSearchItem represents search result information for a package.
type CondaSearchItem struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Build        string   `json:"build"`
	Channel      string   `json:"channel"`
	Dependencies []string `json:"depends"`
	License      string   `json:"license"`
	Summary      string   `json:"summary"`
	Description  string   `json:"description"`
	Homepage     string   `json:"home"`
	Size         int64    `json:"size"`
}

// CondaInfo represents the structure of conda/mamba info --json output.
type CondaInfo struct {
	Platform        string   `json:"platform"`
	CondaVersion    string   `json:"conda_version"`
	PythonVersion   string   `json:"python_version"`
	BaseEnvironment string   `json:"base_environment"`
	CondaPrefix     string   `json:"conda_prefix"`
	Channels        []string `json:"channels"`
	PackageCacheDir []string `json:"pkgs_dirs"`
	EnvironmentDirs []string `json:"envs_dirs"`
	VirtualPackages []string `json:"virtual_packages"`
}

// IsAvailable checks if conda/mamba is installed and accessible.
func (c *CondaManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(c.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, c.binary, []string{"--version"})
	if err != nil {
		// Check for context cancellation
		if IsContextError(err) {
			return false, err
		}
		// Binary exists but not functional - not an error condition
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all packages installed in the base environment.
func (c *CondaManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, c.binary, "list", "-n", "base", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed conda packages: %w", err)
	}

	var listOutput CondaListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse conda list output: %w", err)
	}

	var packages []string
	for _, item := range listOutput {
		packages = append(packages, item.Name)
	}

	sort.Strings(packages)
	return packages, nil
}

// Install installs a conda package globally in the base environment.
func (c *CondaManager) Install(ctx context.Context, name string) error {
	// Both mamba and conda use identical syntax
	output, err := ExecuteCommandCombined(ctx, c.binary, "install", "-n", "base", "-y", name)
	if err != nil {
		return c.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a conda package from the base environment.
func (c *CondaManager) Uninstall(ctx context.Context, name string) error {
	// Both mamba and conda use identical syntax
	output, err := ExecuteCommandCombined(ctx, c.binary, "remove", "-n", "base", "-y", name)
	if err != nil {
		return c.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed in the base environment.
func (c *CondaManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installed, err := c.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	for _, pkg := range installed {
		if pkg == name {
			return true, nil
		}
	}
	return false, nil
}

// InstalledVersion retrieves the installed version of a package.
func (c *CondaManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Get detailed package list
	output, err := ExecuteCommand(ctx, c.binary, "list", "-n", "base", "--json")
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	var listOutput CondaListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return "", fmt.Errorf("failed to parse conda list output: %w", err)
	}

	for _, item := range listOutput {
		if item.Name == name {
			return item.Version, nil
		}
	}

	return "", fmt.Errorf("version information not found for package '%s'", name)
}

// Search searches for conda packages in available repositories.
func (c *CondaManager) Search(ctx context.Context, query string) ([]string, error) {
	// Both mamba and conda use identical search syntax
	output, err := ExecuteCommand(ctx, c.binary, "search", query, "--json")
	if err != nil {
		// Check for no results vs real errors
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			outputStr := string(output)
			if strings.Contains(outputStr, "no packages found") ||
				strings.Contains(outputStr, "PackagesNotFoundError") {
				return []string{}, nil // No results is not an error
			}
		}
		return nil, fmt.Errorf("conda search failed for '%s': %w", query, err)
	}

	return c.parseSearchOutput(output), nil
}

// parseSearchOutput parses conda search JSON output.
func (c *CondaManager) parseSearchOutput(output []byte) []string {
	var searchResults CondaSearchOutput
	if err := json.Unmarshal(output, &searchResults); err != nil {
		return []string{} // Parsing error returns empty results
	}

	// Extract unique package names
	packages := make(map[string]bool)
	for packageName := range searchResults {
		packages[packageName] = true
	}

	// Convert to sorted slice
	result := make([]string, 0, len(packages))
	for pkg := range packages {
		result = append(result, pkg)
	}
	sort.Strings(result)

	return result
}

// Info retrieves detailed information about a package.
func (c *CondaManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "conda",
		Installed: installed,
	}

	if installed {
		// Get version from installed packages
		version, err := c.InstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}
	}

	// Get package information from conda search
	output, err := ExecuteCommand(ctx, c.binary, "search", name, "--info", "--json")
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			return nil, fmt.Errorf("package '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
	}

	packageInfo := c.parseInfoOutput(output, name)
	if packageInfo == nil {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	// Merge search info with installation info
	info.Description = packageInfo.Description
	info.Homepage = packageInfo.Homepage
	info.Dependencies = packageInfo.Dependencies
	if !installed && packageInfo.Version != "" {
		info.Version = packageInfo.Version
	}

	return info, nil
}

// parseInfoOutput parses conda search --info JSON output.
func (c *CondaManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	var searchResults CondaSearchOutput
	if err := json.Unmarshal(output, &searchResults); err != nil {
		return nil
	}

	// Find the requested package
	packages, exists := searchResults[name]
	if !exists || len(packages) == 0 {
		return nil
	}

	// Use the latest version (first in list)
	pkg := packages[0]

	return &PackageInfo{
		Name:         pkg.Name,
		Version:      pkg.Version,
		Description:  pkg.Summary,
		Homepage:     pkg.Homepage,
		Manager:      "conda",
		Dependencies: pkg.Dependencies,
	}
}

// CheckHealth performs a comprehensive health check of the conda installation.
func (c *CondaManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "Conda Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Conda is available and properly configured",
	}

	// Check availability
	available, err := c.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "warn"
		check.Message = "Conda availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking conda: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "Conda is not available"
		check.Issues = []string{"conda command not found"}
		check.Suggestions = []string{
			"Install via Homebrew: brew install micromamba (recommended)",
			"Install Miniconda: https://docs.conda.io/en/latest/miniconda.html",
			"Install Anaconda: https://www.anaconda.com/products/distribution",
			"Install Mamba: https://mamba.readthedocs.io/",
		}
		return check, nil
	}

	// Report detected variant for transparency
	variant := "conda"
	if c.useMamba {
		variant = "mamba (10-100x faster)"
	}
	check.Details = append(check.Details, fmt.Sprintf("Detected variant: %s", variant))

	// Get conda info for detailed diagnostics
	info, err := c.getCondaInfo(ctx)
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not retrieve conda information"
		check.Issues = []string{fmt.Sprintf("Error getting conda info: %v", err)}
		return check, nil
	}

	// Add conda version and environment details
	if info.CondaVersion != "" {
		check.Details = append(check.Details, fmt.Sprintf("Conda version: %s", info.CondaVersion))
	}
	if info.PythonVersion != "" {
		check.Details = append(check.Details, fmt.Sprintf("Python version: %s", info.PythonVersion))
	}
	if info.BaseEnvironment != "" {
		check.Details = append(check.Details, fmt.Sprintf("Base environment: %s", info.BaseEnvironment))
	}
	if info.Platform != "" {
		check.Details = append(check.Details, fmt.Sprintf("Platform: %s", info.Platform))
	}

	// Check channels configuration
	if len(info.Channels) > 0 {
		check.Details = append(check.Details, fmt.Sprintf("Configured channels: %d", len(info.Channels)))
		// Show top 3 channels
		channelCount := len(info.Channels)
		if channelCount > 3 {
			channelCount = 3
		}
		channelStr := strings.Join(info.Channels[:channelCount], ", ")
		check.Details = append(check.Details, fmt.Sprintf("Primary channels: %s", channelStr))
	}

	// Validate base environment access
	if info.BaseEnvironment == "" {
		check.Status = "warn"
		check.Message = "Conda base environment not properly configured"
		check.Issues = []string{"Base environment path not detected"}
		check.Suggestions = []string{"Reinitialize conda: conda init"}
	}

	return check, nil
}

// getCondaInfo retrieves comprehensive conda system information.
func (c *CondaManager) getCondaInfo(ctx context.Context) (*CondaInfo, error) {
	output, err := ExecuteCommand(ctx, c.binary, "info", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get conda info: %w", err)
	}

	var info CondaInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse conda info: %w", err)
	}

	return &info, nil
}

// SelfInstall installs mamba (preferred) for optimal performance.
func (c *CondaManager) SelfInstall(ctx context.Context) error {
	// Check if already available (any variant)
	if available, _ := c.IsAvailable(ctx); available {
		return nil
	}

	// Install micromamba via Homebrew (recommended method)
	return c.installMicromambaViaHomebrew(ctx)
}

// installMicromambaViaHomebrew installs micromamba via Homebrew (recommended method).
func (c *CondaManager) installMicromambaViaHomebrew(ctx context.Context) error {
	if available, _ := checkPackageManagerAvailable(ctx, "brew"); !available {
		return fmt.Errorf("homebrew not available")
	}
	// micromamba is the recommended conda variant via Homebrew as of 2025
	return executeInstallCommand(ctx, "brew", []string{"install", "micromamba"}, "micromamba")
}

// Upgrade upgrades one or more packages to their latest versions.
func (c *CondaManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// Get all installed packages first
		installed, err := c.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to list installed packages: %w", err)
		}

		// Update all packages
		var upgradeErrors []string
		for _, pkg := range installed {
			output, err := ExecuteCommandCombined(ctx, c.binary, "update", "-n", "base", "-y", pkg)
			if err != nil {
				upgradeErr := c.handleUpgradeError(err, output, pkg)
				if upgradeErr != nil {
					upgradeErrors = append(upgradeErrors, upgradeErr.Error())
				}
				continue
			}
		}

		if len(upgradeErrors) > 0 {
			return fmt.Errorf("some packages failed to upgrade: %s", strings.Join(upgradeErrors, "; "))
		}
		return nil
	}

	// Upgrade specific packages
	var upgradeErrors []string
	for _, pkg := range packages {
		output, err := ExecuteCommandCombined(ctx, c.binary, "update", "-n", "base", "-y", pkg)
		if err != nil {
			upgradeErr := c.handleUpgradeError(err, output, pkg)
			if upgradeErr != nil {
				upgradeErrors = append(upgradeErrors, upgradeErr.Error())
			}
			continue
		}
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("failed to upgrade packages: %s", strings.Join(upgradeErrors, "; "))
	}
	return nil
}

// handleInstallError processes install command errors.
func (c *CondaManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known conda error patterns
		if strings.Contains(outputStr, "packagenotfounderror") ||
			strings.Contains(outputStr, "no packages found matching") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "unsatisfiableerror") ||
			strings.Contains(outputStr, "conflicting dependencies") {
			return fmt.Errorf("dependency conflicts installing '%s'", packageName)
		}
		if strings.Contains(outputStr, "channelnotavailableerror") {
			return fmt.Errorf("conda channels unavailable for package '%s'", packageName)
		}
		if strings.Contains(outputStr, "environmentlockederror") {
			return fmt.Errorf("conda environment is locked")
		}
		if strings.Contains(outputStr, "condahttperror") ||
			strings.Contains(outputStr, "connection failed") {
			return fmt.Errorf("network error during conda operation")
		}
		if strings.Contains(outputStr, "already installed") ||
			strings.Contains(outputStr, "nothing to do") {
			// Package is already installed - this is typically fine
			return nil
		}

		// Standard exit code handling
		if exitCode != 0 {
			if len(output) > 0 {
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("conda installation failed: %s", errorOutput)
			}
			return fmt.Errorf("conda installation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute conda install command: %w", err)
}

// handleUninstallError processes uninstall command errors.
func (c *CondaManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "no packages found") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "environmentlockederror") {
			return fmt.Errorf("conda environment is locked")
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("conda uninstallation failed: %s", errorOutput)
			}
			return fmt.Errorf("conda uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute conda uninstall command: %w", err)
}

// handleUpgradeError processes upgrade command errors.
func (c *CondaManager) handleUpgradeError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "packagenotfounderror") ||
			strings.Contains(outputStr, "no packages found matching") ||
			strings.Contains(outputStr, "not found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "nothing to do") ||
			strings.Contains(outputStr, "all requested packages already installed") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "unsatisfiableerror") ||
			strings.Contains(outputStr, "conflicting dependencies") {
			return fmt.Errorf("dependency conflicts upgrading '%s'", packageName)
		}
		if strings.Contains(outputStr, "environmentlockederror") {
			return fmt.Errorf("conda environment is locked")
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("conda upgrade failed: %s", errorOutput)
			}
			return fmt.Errorf("conda upgrade failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute conda upgrade command: %w", err)
}

// Dependencies returns package managers this manager depends on for self-installation.
func (c *CondaManager) Dependencies() []string {
	return []string{"brew"} // micromamba installation requires Homebrew
}

func init() {
	RegisterManager("conda", func() PackageManager {
		return NewCondaManager()
	})
}
