// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/resources/packages/parsers"
)

// NpmManager manages NPM packages.
type NpmManager struct {
	*StandardManager
}

// NewNpmManager creates a new NPM manager.
func NewNpmManager() *NpmManager {
	return newNpmManager()
}

// newNpmManager creates an NPM manager.
func newNpmManager() *NpmManager {
	config := GetNpmConfig()
	standardManager := NewStandardManager(config)

	return &NpmManager{
		StandardManager: standardManager,
	}
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	// Call the binary directly to handle npm's unique exit code behavior
	cmd := exec.CommandContext(ctx, n.Binary, "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		// npm list can return non-zero exit codes even when working correctly
		// (e.g., when there are peer dependency warnings)
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			// Only treat it as an error if the exit code indicates a real failure
			if execErr.ExitCode() > 1 {
				return nil, fmt.Errorf("npm list command failed with severe error: %w", err)
			}
			// Exit code 1 might just be warnings - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, fmt.Errorf("failed to execute npm list command: %w", err)
		}
	}

	return n.parseListOutput(output), nil
}

// parseListOutput parses npm list JSON output to extract package names
func (n *NpmManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	// Parse JSON output
	var listResult struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &listResult); err != nil {
		// If JSON parsing fails, return empty list
		return []string{}
	}

	// Extract package names from dependencies
	packages := make([]string, 0, len(listResult.Dependencies))
	for pkgName := range listResult.Dependencies {
		packages = append(packages, pkgName)
	}

	// Sort packages for consistent output
	sort.Strings(packages)

	return packages
}

// Install installs a global NPM package.
func (n *NpmManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, n.Binary, "install", "-g", name)
	if err != nil {
		return n.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a global NPM package.
func (n *NpmManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, n.Binary, "uninstall", "-g", name)
	if err != nil {
		return n.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed globally.
func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	checkCmd := exec.CommandContext(ctx, n.Binary, "list", "-g", name)
	_, err := checkCmd.Output()
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (npm not found, permission issues, etc.)
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	return true, nil
}

// Search searches for packages in NPM registry.
func (n *NpmManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, n.Binary, "search", query, "--json")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			// For npm search, exit code 1 usually means no results found
			if execErr.ExitCode() == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, fmt.Errorf("npm search command failed for %s: %w", query, err)
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, fmt.Errorf("failed to execute npm search command: %w", err)
	}

	return n.parseSearchOutput(output), nil
}

// parseSearchOutput parses npm search JSON output
func (n *NpmManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		return []string{}
	}

	// Try to parse as JSON array
	var searchResults []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &searchResults); err == nil {
		packages := make([]string, 0, len(searchResults))
		for _, pkg := range searchResults {
			if pkg.Name != "" {
				packages = append(packages, pkg.Name)
			}
		}
		return packages
	}

	// Fallback to line parsing if JSON fails
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, `"name":`) {
			// Extract package name from JSON line like: "name": "package-name",
			parts := strings.Split(line, `"name":`)
			if len(parts) > 1 {
				namepart := strings.TrimSpace(parts[1])
				// Clean up quotes and commas
				namepart = strings.Trim(namepart, ` "',`)
				if namepart != "" {
					packages = append(packages, namepart)
				}
			}
		}
	}

	return packages
}

// Info retrieves detailed information about a package from NPM.
func (n *NpmManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := n.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	// Always use npm view for info (works for both installed and available packages)
	cmd := exec.CommandContext(ctx, n.Binary, "view", name, "--json")
	output, err := cmd.Output()
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
			return nil, fmt.Errorf("package '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
	}

	info := n.parseInfoOutput(output, name)
	info.Manager = "npm"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses npm view JSON output
func (n *NpmManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{Name: name}

	// Try to parse as JSON
	var viewResult struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		Description  string            `json:"description"`
		Homepage     string            `json:"homepage"`
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &viewResult); err == nil {
		info.Name = viewResult.Name
		info.Version = viewResult.Version
		info.Description = viewResult.Description
		info.Homepage = viewResult.Homepage

		// Convert dependencies map to slice
		for depName := range viewResult.Dependencies {
			info.Dependencies = append(info.Dependencies, depName)
		}
		return info
	}

	// Fallback to manual parsing if JSON fails
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"name":`) {
			info.Name = cleanJSONValue(parsers.ExtractVersion([]byte(line), `"name":`))
		} else if strings.Contains(line, `"version":`) {
			info.Version = cleanJSONValue(parsers.ExtractVersion([]byte(line), `"version":`))
		} else if strings.Contains(line, `"description":`) {
			info.Description = cleanJSONValue(parsers.ExtractVersion([]byte(line), `"description":`))
		} else if strings.Contains(line, `"homepage":`) {
			info.Homepage = cleanJSONValue(parsers.ExtractVersion([]byte(line), `"homepage":`))
		}
	}

	// Extract dependencies separately as they're nested
	info.Dependencies = n.extractDependencies(string(output))

	return info
}

// extractDependencies extracts dependencies from npm view JSON output
func (n *NpmManager) extractDependencies(jsonOutput string) []string {
	var dependencies []string
	lines := strings.Split(jsonOutput, "\n")
	inDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"dependencies":`) {
			inDependencies = true
			continue
		}
		if inDependencies {
			if strings.Contains(line, `}`) && !strings.Contains(line, `"`) {
				break
			}
			if strings.Contains(line, `"`) && strings.Contains(line, `:`) {
				// Extract package name from dependency line
				parts := strings.Split(line, `"`)
				if len(parts) > 1 {
					depName := parts[1]
					if depName != "" {
						dependencies = append(dependencies, depName)
					}
				}
			}
		}
	}

	return dependencies
}

// GetInstalledVersion retrieves the installed version of a global NPM package
func (n *NpmManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := n.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed globally", name)
	}

	// Get version using npm list with specific package
	cmd := exec.CommandContext(ctx, n.Binary, "list", "-g", name, "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	// Parse JSON output
	var listResult struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &listResult); err != nil {
		return "", fmt.Errorf("failed to parse npm JSON output for %s: %w", name, err)
	}

	if dep, ok := listResult.Dependencies[name]; ok && dep.Version != "" {
		return dep.Version, nil
	}

	return "", fmt.Errorf("package '%s' not found in npm list output", name)
}

// cleanJSONValue removes quotes and commas from a JSON value
func cleanJSONValue(value string) string {
	return CleanJSONValue(value)
}

// IsAvailable checks if npm is installed and accessible
func (n *NpmManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(n.Binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, n.Binary, []string{"--version"})
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

// SupportsSearch returns true as NPM supports package search
func (n *NpmManager) SupportsSearch() bool {
	return true
}

// handleInstallError processes install command errors using ErrorMatcher
func (n *NpmManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if exitCode, ok := ExtractExitCode(err); ok {
		errorType := n.ErrorMatcher.MatchError(outputStr)

		switch errorType {
		case ErrorTypeNotFound:
			return fmt.Errorf("package '%s' not found", packageName)

		case ErrorTypeAlreadyInstalled:
			// Package is already installed - this is typically fine
			return nil

		case ErrorTypePermission:
			return fmt.Errorf("permission denied installing %s", packageName)

		case ErrorTypeLocked:
			return fmt.Errorf("package manager database is locked")

		case ErrorTypeNetwork:
			return fmt.Errorf("network error during installation")

		case ErrorTypeBuild:
			return fmt.Errorf("failed to build package '%s'", packageName)

		case ErrorTypeDependency:
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)

		default:
			// Only treat non-zero exit codes as errors
			if exitCode != 0 {
				return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors using ErrorMatcher
func (n *NpmManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if exitCode, ok := ExtractExitCode(err); ok {
		errorType := n.ErrorMatcher.MatchError(outputStr)

		switch errorType {
		case ErrorTypeNotInstalled:
			// Package is not installed - this is typically fine for uninstall
			return nil

		case ErrorTypePermission:
			return fmt.Errorf("permission denied uninstalling %s", packageName)

		case ErrorTypeLocked:
			return fmt.Errorf("package manager database is locked")

		case ErrorTypeDependency:
			return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)

		default:
			// Only treat non-zero exit codes as errors
			if exitCode != 0 {
				return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute uninstall command: %w", err)
}
