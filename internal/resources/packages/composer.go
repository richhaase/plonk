// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ComposerManager manages Composer global packages.
type ComposerManager struct {
	binary string
}

// NewComposerManager creates a new Composer manager.
func NewComposerManager() *ComposerManager {
	return &ComposerManager{
		binary: "composer",
	}
}

// ListInstalled lists all globally installed Composer packages.
func (c *ComposerManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, c.binary, "global", "show", "--format=json")
	if err != nil {
		// Composer global show can return non-zero exit codes when no packages are installed
		if exitCode, ok := ExtractExitCode(err); ok {
			// Only treat it as an error if the exit code indicates a real failure
			if exitCode > 1 {
				return nil, fmt.Errorf("composer global show command failed with severe error: %w", err)
			}
			// Exit code 1 might just mean no packages installed - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, err
		}
	}

	return c.parseListOutput(output), nil
}

// parseListOutput parses composer global show JSON output to extract package names
func (c *ComposerManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || result == "{}" {
		return []string{}
	}

	// Parse JSON output
	var listResult struct {
		Installed []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"installed"`
	}

	if err := json.Unmarshal(output, &listResult); err != nil {
		// If JSON parsing fails, return empty list
		return []string{}
	}

	// Extract package names from installed array
	packages := make([]string, 0, len(listResult.Installed))
	for _, pkg := range listResult.Installed {
		if pkg.Name != "" {
			packages = append(packages, pkg.Name)
		}
	}

	// Sort packages for consistent output
	sort.Strings(packages)

	return packages
}

// Install installs a global Composer package.
func (c *ComposerManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, c.binary, "global", "require", name)
	if err != nil {
		return c.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a global Composer package.
func (c *ComposerManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, c.binary, "global", "remove", name)
	if err != nil {
		return c.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed globally.
func (c *ComposerManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	_, err := ExecuteCommand(ctx, c.binary, "global", "show", name)
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (composer not found, permission issues, etc.)
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	return true, nil
}

// Search searches for packages in Packagist registry.
func (c *ComposerManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := ExecuteCommand(ctx, c.binary, "search", query, "--format=json")
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitCode, ok := ExtractExitCode(err); ok {
			// For composer search, exit code 1 usually means no results found
			if exitCode == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, fmt.Errorf("composer search command failed for %s: %w", query, err)
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, err
	}

	return c.parseSearchOutput(output), nil
}

// parseSearchOutput parses composer search JSON output
func (c *ComposerManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || result == "{}" || result == "[]" {
		return []string{}
	}

	// Try to parse as JSON object (composer search format)
	var searchResults map[string]interface{}
	if err := json.Unmarshal(output, &searchResults); err == nil {
		packages := make([]string, 0, len(searchResults))
		for pkgName := range searchResults {
			if pkgName != "" {
				packages = append(packages, pkgName)
			}
		}
		sort.Strings(packages)
		return packages
	}

	// Fallback to line parsing if JSON fails
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for package names in format "vendor/package"
		if strings.Contains(line, "/") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			// Extract package name - typically first word on line
			parts := strings.Fields(line)
			if len(parts) > 0 && strings.Contains(parts[0], "/") {
				packages = append(packages, parts[0])
			}
		}
	}

	return packages
}

// Info retrieves detailed information about a package from Packagist.
func (c *ComposerManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	// Use composer show for info (works for both installed and available packages)
	output, err := ExecuteCommand(ctx, c.binary, "show", name, "--format=json")
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			return nil, fmt.Errorf("package '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
	}

	info := c.parseInfoOutput(output, name)
	info.Manager = "composer"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses composer show JSON output
func (c *ComposerManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{Name: name}

	// Try to parse as JSON
	var showResult struct {
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Description string   `json:"description"`
		Homepage    string   `json:"homepage"`
		Keywords    []string `json:"keywords"`
		License     []string `json:"license"`
		Authors     []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Require map[string]string `json:"require"`
	}

	if err := json.Unmarshal(output, &showResult); err == nil {
		info.Name = showResult.Name
		info.Version = showResult.Version
		info.Description = showResult.Description
		info.Homepage = showResult.Homepage

		// Convert require map to dependencies slice
		for depName := range showResult.Require {
			// Skip PHP version constraints and platform packages
			if depName != "php" && !strings.HasPrefix(depName, "ext-") && !strings.HasPrefix(depName, "lib-") {
				info.Dependencies = append(info.Dependencies, depName)
			}
		}
		return info
	}

	// Fallback to manual parsing if JSON fails
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") && strings.Contains(line, ":") {
			info.Name = cleanValue(extractValueAfterColon(line))
		} else if strings.HasPrefix(line, "version") && strings.Contains(line, ":") {
			info.Version = cleanValue(extractValueAfterColon(line))
		} else if strings.HasPrefix(line, "description") && strings.Contains(line, ":") {
			info.Description = cleanValue(extractValueAfterColon(line))
		} else if strings.HasPrefix(line, "homepage") && strings.Contains(line, ":") {
			info.Homepage = cleanValue(extractValueAfterColon(line))
		}
	}

	return info
}

// InstalledVersion retrieves the installed version of a global Composer package
func (c *ComposerManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed globally", name)
	}

	// Get version using composer global show with specific package
	output, err := ExecuteCommand(ctx, c.binary, "global", "show", name, "--format=json")
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	// Parse JSON output
	var showResult struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(output, &showResult); err != nil {
		// Fallback to text parsing
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "version") && strings.Contains(line, ":") {
				version := cleanValue(extractValueAfterColon(line))
				if version != "" {
					return version, nil
				}
			}
		}
		return "", fmt.Errorf("failed to parse composer JSON output for %s: %w", name, err)
	}

	if showResult.Version != "" {
		return showResult.Version, nil
	}

	return "", fmt.Errorf("package '%s' version not found in composer output", name)
}

// extractValueAfterColon extracts the value after a colon from a string
func extractValueAfterColon(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// cleanValue removes quotes and extra whitespace from a value
func cleanValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func init() {
	RegisterManager("composer", func() PackageManager {
		return NewComposerManager()
	})
}

// IsAvailable checks if composer is installed and accessible
func (c *ComposerManager) IsAvailable(ctx context.Context) (bool, error) {
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

// SupportsSearch returns true as Composer supports package search
func (c *ComposerManager) SupportsSearch() bool {
	return true
}

// handleInstallError processes install command errors
func (c *ComposerManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "could not find package") ||
			strings.Contains(outputStr, "Package not found") ||
			strings.Contains(outputStr, "No matching package found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}
		if strings.Contains(outputStr, "Your requirements could not be resolved") {
			return fmt.Errorf("dependency resolution failed for package '%s'", packageName)
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

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}

// handleUninstallError processes uninstall command errors
func (c *ComposerManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "is not installed") ||
			strings.Contains(outputStr, "Package not found") ||
			strings.Contains(outputStr, "does not exist") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
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

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}
