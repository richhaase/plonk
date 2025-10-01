// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct {
	binary string
	exec   CommandExecutor
}

// brewInfoV2JSON represents the JSON structure from brew info --json=v2
type brewInfoV2JSON struct {
	Formulae []brewFormulaJSON `json:"formulae"`
	Casks    []brewCaskJSON    `json:"casks"`
}

// brewFormulaJSON represents a formula in the v2 JSON format
type brewFormulaJSON struct {
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases"`
	Installed []struct {
		Version string `json:"version"`
	} `json:"installed"`
	Versions struct {
		Stable string `json:"stable"`
	} `json:"versions"`
}

// brewCaskJSON represents a cask in the v2 JSON format
type brewCaskJSON struct {
	Token string   `json:"token"`
	Name  []string `json:"name"`
	// Note: Casks don't have the same installed/versions structure as formulae
}

// brewPackageInfo represents unified package info for both formulae and casks
type brewPackageInfo struct {
	Name      string
	Aliases   []string
	Installed []struct {
		Version string `json:"version"`
	}
	Versions struct {
		Stable string `json:"stable"`
	}
}

// NewHomebrewManager creates a new homebrew manager with default executor.
func NewHomebrewManager() *HomebrewManager {
	return NewHomebrewManagerWithExecutor(nil)
}

// NewHomebrewManagerWithExecutor creates a new homebrew manager with the provided executor.
// If executor is nil, uses the default executor.
func NewHomebrewManagerWithExecutor(executor CommandExecutor) *HomebrewManager {
	if executor == nil {
		executor = defaultExecutor
	}
	return &HomebrewManager{
		binary: "brew",
		exec:   executor,
	}
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	// Get basic list first
	output, err := ExecuteWith(ctx, h.exec, h.binary, "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	packages := SplitLines(output)

	// Try to enrich with aliases from JSON
	packagesInfo, err := h.getInstalledPackagesInfo(ctx)
	if err == nil {
		// Build a set of all names and aliases
		packageSet := make(map[string]bool)
		for _, pkg := range packagesInfo {
			packageSet[pkg.Name] = true
			for _, alias := range pkg.Aliases {
				packageSet[alias] = true
			}
		}

		// Convert to slice
		var allPackages []string
		for name := range packageSet {
			allPackages = append(allPackages, name)
		}
		return allPackages, nil
	}

	// Fallback to simple list if JSON fails
	return packages, nil
}

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
	output, err := CombinedOutputWith(ctx, h.exec, h.binary, "install", name)
	if err != nil {
		return h.handleInstallError(err, []byte(output), name)
	}
	return nil
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error {
	output, err := CombinedOutputWith(ctx, h.exec, h.binary, "uninstall", name)
	if err != nil {
		return h.handleUninstallError(err, []byte(output), name)
	}
	return nil
}

// getInstalledPackagesInfo returns detailed information about all installed packages (formulae and casks)
func (h *HomebrewManager) getInstalledPackagesInfo(ctx context.Context) ([]brewPackageInfo, error) {
	output, err := ExecuteWith(ctx, h.exec, h.binary, "info", "--installed", "--json=v2")
	if err != nil {
		return nil, fmt.Errorf("failed to get installed packages info: %w", err)
	}

	var v2Data brewInfoV2JSON
	if err := json.Unmarshal(output, &v2Data); err != nil {
		return nil, fmt.Errorf("failed to parse brew info JSON: %w", err)
	}

	var allPackages []brewPackageInfo

	// Convert formulae to unified format
	for _, formula := range v2Data.Formulae {
		allPackages = append(allPackages, brewPackageInfo{
			Name:      formula.Name,
			Aliases:   formula.Aliases,
			Installed: formula.Installed,
			Versions:  formula.Versions,
		})
	}

	// Convert casks to unified format
	for _, cask := range v2Data.Casks {
		allPackages = append(allPackages, brewPackageInfo{
			Name:    cask.Token, // Use token as name to match brew list output
			Aliases: []string{}, // Casks don't have aliases
		})
	}

	return allPackages, nil
}

// IsInstalled checks if a specific package is installed.
func (h *HomebrewManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	packages, err := h.getInstalledPackagesInfo(ctx)
	if err != nil {
		// Fallback to simple brew list check
		_, err := ExecuteWith(ctx, h.exec, h.binary, "list", name)
		if err == nil {
			return true, nil
		}
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			return false, nil
		}
		return false, err
	}

	// Check both package names and aliases
	for _, pkg := range packages {
		if pkg.Name == name {
			return true, nil
		}
		for _, alias := range pkg.Aliases {
			if alias == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// Search searches for packages in Homebrew repositories.
func (h *HomebrewManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := ExecuteWith(ctx, h.exec, h.binary, "search", query)
	if err != nil {
		return nil, fmt.Errorf("failed to search homebrew packages for %s: %w", query, err)
	}

	return h.parseSearchOutput(output), nil
}

// parseSearchOutput parses brew search output
func (h *HomebrewManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || strings.Contains(result, "No formula found") {
		return []string{}
	}

	// Parse output into package list
	packages := strings.Split(result, "\n")
	var filteredPackages []string
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			// Skip header lines and empty lines
			if !strings.HasPrefix(trimmed, "=") && !strings.HasPrefix(trimmed, "If you meant") {
				filteredPackages = append(filteredPackages, trimmed)
			}
		}
	}

	return filteredPackages
}

// Info retrieves detailed information about a package.
func (h *HomebrewManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	output, err := ExecuteWith(ctx, h.exec, h.binary, "info", name)
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			return nil, fmt.Errorf("package '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
	}

	// Check if installed
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	info := h.parseInfoOutput(output, name)
	info.Manager = "brew"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses brew info output
func (h *HomebrewManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{
		Name: name,
	}

	lines := strings.Split(string(output), "\n")
	versionFound := false

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// First line usually contains name and version
		if strings.HasPrefix(line, name) && strings.Contains(line, ":") {
			// Extract version from lines like "package: stable 1.2.3"
			parts := strings.Fields(line)
			for j, part := range parts {
				if part == "stable" && j+1 < len(parts) {
					info.Version = parts[j+1]
					versionFound = true
					break
				}
			}

			// Check for description on the same line after version
			if colonIndex := strings.Index(line, ":"); colonIndex > 0 && colonIndex < len(line)-1 {
				desc := strings.TrimSpace(line[colonIndex+1:])
				// Remove version info if present
				if versionFound && strings.Contains(desc, "stable") {
					// Look for description after stable version
					stableIndex := strings.Index(desc, "stable")
					if stableIndex >= 0 {
						remainingDesc := desc[stableIndex:]
						parts := strings.Fields(remainingDesc)
						if len(parts) > 2 { // "stable", "version", "description..."
							info.Description = strings.TrimSpace(strings.Join(parts[2:], " "))
						}
					}
				} else if desc != "" && !strings.Contains(desc, "stable") {
					info.Description = desc
				}
			}
		}

		// Look for description on the next line if not found yet
		if info.Description == "" && i > 0 && versionFound {
			if line != "" && !strings.HasPrefix(line, "From:") && !strings.HasPrefix(line, "URL:") &&
				!strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "=") {
				info.Description = line
			}
		}

		// Look for homepage
		if strings.HasPrefix(line, "From:") || strings.HasPrefix(line, "URL:") {
			// Extract URL from lines like "From: https://..."
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "http") {
					info.Homepage = part
					break
				}
			}
		}
	}

	return info
}

// InstalledVersion retrieves the installed version of a package
func (h *HomebrewManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	packages, err := h.getInstalledPackagesInfo(ctx)
	if err != nil {
		// Fallback to brew list --versions
		output, err := ExecuteWith(ctx, h.exec, h.binary, "list", "--versions", name)
		if err != nil {
			return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
		}
		version := h.extractVersion(output, name)
		if version == "" {
			return "", fmt.Errorf("could not extract version for package '%s'", name)
		}
		return version, nil
	}

	// Find package by name or alias
	for _, pkg := range packages {
		if pkg.Name == name || contains(pkg.Aliases, name) {
			if len(pkg.Installed) > 0 && pkg.Installed[0].Version != "" {
				return pkg.Installed[0].Version, nil
			}
			// No installed version info, return stable version
			if pkg.Versions.Stable != "" {
				return pkg.Versions.Stable, nil
			}
			return "", fmt.Errorf("no version information available for package '%s'", name)
		}
	}

	return "", fmt.Errorf("package '%s' is not installed", name)
}

// extractVersion extracts version from brew list --versions output
func (h *HomebrewManager) extractVersion(output []byte, name string) string {
	result := strings.TrimSpace(string(output))
	// Output format: "package 1.2.3 1.2.2" (latest version first)
	if strings.HasPrefix(result, name+" ") {
		parts := strings.Fields(result)
		if len(parts) >= 2 {
			return parts[1] // Return the first version (latest)
		}
	}
	return ""
}

// IsAvailable checks if homebrew is installed and accessible
func (h *HomebrewManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailableWith(h.exec, h.binary) {
		return false, nil
	}

	err := VerifyBinaryWith(ctx, h.exec, h.binary, []string{"--version"})
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

// CheckHealth performs a comprehensive health check of the Homebrew installation
func (h *HomebrewManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "Homebrew Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Homebrew is available and properly configured",
	}

	// Check basic availability first
	available, err := h.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = "Homebrew availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking homebrew: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "Homebrew is recommended but not available"
		check.Issues = []string{"Homebrew is recommended for best compatibility"}
		check.Suggestions = []string{
			`Install Homebrew: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
			"After installation, ensure brew is in your PATH",
		}
		return check, nil
	}

	// Discover homebrew bin directory dynamically
	binDir, err := h.getBinDirectory(ctx)
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not determine Homebrew bin directory"
		check.Issues = []string{fmt.Sprintf("Error discovering bin directory: %v", err)}
		return check, nil
	}

	// Check if bin directory is in PATH
	pathCheck := checkDirectoryInPath(binDir)
	check.Details = append(check.Details, fmt.Sprintf("Homebrew bin directory: %s", binDir))

	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = "Homebrew bin directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Status = "warn"
		check.Message = "Homebrew bin directory does not exist"
		check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
	} else {
		check.Details = append(check.Details, "Homebrew bin directory is in PATH")
	}

	return check, nil
}

// getBinDirectory discovers the actual homebrew bin directory
func (h *HomebrewManager) getBinDirectory(ctx context.Context) (string, error) {
	output, err := ExecuteWith(ctx, h.exec, h.binary, "--prefix")
	if err != nil {
		return "", fmt.Errorf("failed to get homebrew prefix: %w", err)
	}

	prefix := strings.TrimSpace(string(output))
	return filepath.Join(prefix, "bin"), nil
}

// SelfInstall installs Homebrew using the official installer script
func (h *HomebrewManager) SelfInstall(ctx context.Context) error {
	// Check if already available (idempotent)
	if available, _ := h.IsAvailable(ctx); available {
		return nil
	}

	// Execute official installer script
	script := `curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash`
	return executeInstallScript(ctx, script, "Homebrew")
}

// Upgrade upgrades one or more packages to their latest versions
func (h *HomebrewManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// Upgrade all packages
		output, err := CombinedOutputWith(ctx, h.exec, h.binary, "upgrade")
		if err != nil {
			return h.handleUpgradeError(err, []byte(output), "all packages")
		}
		return nil
	}

	// Upgrade specific packages
	args := append([]string{"upgrade"}, packages...)
	output, err := CombinedOutputWith(ctx, h.exec, h.binary, args...)
	if err != nil {
		return h.handleUpgradeError(err, []byte(output), strings.Join(packages, ", "))
	}
	return nil
}

func init() {
	RegisterManagerV2("brew", func(exec CommandExecutor) PackageManager {
		return NewHomebrewManagerWithExecutor(exec)
	})
}

// handleUpgradeError processes upgrade command errors
func (h *HomebrewManager) handleUpgradeError(err error, output []byte, packages string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no available formula") || strings.Contains(outputStr, "no formulae found") {
			return fmt.Errorf("one or more packages not found: %s", packages)
		}
		if strings.Contains(outputStr, "nothing to upgrade") || strings.Contains(outputStr, "already up-to-date") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") {
			return fmt.Errorf("permission denied upgrading %s", packages)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("package upgrade failed: %s", errorOutput)
			}
			return fmt.Errorf("package upgrade failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute upgrade command: %w", err)
}

// Dependencies returns package managers this manager depends on for self-installation
func (h *HomebrewManager) Dependencies() []string {
	return []string{} // Homebrew is independent - uses official installer script
}

// handleInstallError processes install command errors
func (h *HomebrewManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no available formula") || strings.Contains(outputStr, "no formulae found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already installed") {
			return nil // Already installed is success
		}
		if strings.Contains(outputStr, "permission denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}
		if strings.Contains(outputStr, "because it is required by") || strings.Contains(outputStr, "still has dependents") {
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("package installation failed: %s", errorOutput)
			}
			return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors
func (h *HomebrewManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no such keg") || strings.Contains(outputStr, "not installed") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}
		if strings.Contains(outputStr, "because it is required by") || strings.Contains(outputStr, "still has dependents") {
			return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("package uninstallation failed: %s", errorOutput)
			}
			return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute uninstall command: %w", err)
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
