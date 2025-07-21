// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/executor"
	"github.com/richhaase/plonk/internal/managers/parsers"
)

// NpmManager manages NPM packages using BaseManager for common functionality.
type NpmManager struct {
	*BaseManager
}

// NewNpmManager creates a new NPM manager with the default executor.
func NewNpmManager() *NpmManager {
	return newNpmManager(nil)
}

// NewNpmManagerWithExecutor creates a new NPM manager with a custom executor for testing.
func NewNpmManagerWithExecutor(exec executor.CommandExecutor) *NpmManager {
	return newNpmManager(exec)
}

// newNpmManager creates an NPM manager with the given executor.
func newNpmManager(exec executor.CommandExecutor) *NpmManager {
	config := ManagerConfig{
		BinaryName:  "npm",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"list", "-g", "--depth=0", "--json"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", "-g", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", "-g", pkg}
		},
	}

	// Add npm-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "404", "E404", "Not found")
	errorMatcher.AddPattern(ErrorTypePermission, "EACCES")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "ENOENT", "cannot remove")

	var base *BaseManager
	if exec == nil {
		base = NewBaseManager(config)
	} else {
		base = NewBaseManagerWithExecutor(config, exec)
	}
	base.ErrorMatcher = errorMatcher

	return &NpmManager{
		BaseManager: base,
	}
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	// Call the binary directly to handle npm's unique exit code behavior
	output, err := n.Executor.Execute(ctx, n.GetBinary(), n.Config.ListArgs()...)
	if err != nil {
		// npm list can return non-zero exit codes even when working correctly
		// (e.g., when there are peer dependency warnings)
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			// Only treat it as an error if the exit code indicates a real failure
			if execErr.ExitCode() > 1 {
				return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
					"npm list command failed with severe error")
			}
			// Exit code 1 might just be warnings - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
				"failed to execute npm list command")
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
	return n.ExecuteInstall(ctx, name)
}

// Uninstall removes a global NPM package.
func (n *NpmManager) Uninstall(ctx context.Context, name string) error {
	return n.ExecuteUninstall(ctx, name)
}

// IsInstalled checks if a specific package is installed globally.
func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	_, err := n.Executor.Execute(ctx, n.GetBinary(), "list", "-g", name)
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (npm not found, permission issues, etc.)
		return false, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "check", name,
			"failed to check package installation status")
	}
	return true, nil
}

// Search searches for packages in NPM registry.
func (n *NpmManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := n.Executor.Execute(ctx, n.GetBinary(), "search", query, "--json")
	if err != nil {
		// Check if this is a real error vs expected conditions
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			// For npm search, exit code 1 usually means no results found
			if execErr.ExitCode() == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
				"npm search command failed")
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "search",
			"failed to execute npm search command")
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
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	// Always use npm view for info (works for both installed and available packages)
	output, err := n.Executor.Execute(ctx, n.GetBinary(), "view", name, "--json")
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
			return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
				fmt.Sprintf("package '%s' not found", name)).
				WithSuggestionMessage(fmt.Sprintf("Search available packages: npm search %s", name))
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get package info")
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
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check package installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("package '%s' is not installed globally", name))
	}

	// Get version using npm list with specific package
	output, err := n.Executor.Execute(ctx, n.GetBinary(), "list", "-g", name, "--depth=0", "--json")
	if err != nil {
		// Try alternative approach if JSON fails
		return n.getVersionFromLS(ctx, name)
	}

	// Try to parse JSON output
	var listResult struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &listResult); err == nil {
		if dep, ok := listResult.Dependencies[name]; ok && dep.Version != "" {
			return dep.Version, nil
		}
	}

	// Fallback to manual parsing
	version := parsers.ExtractVersion(output, `"version":`)
	if version != "" {
		return cleanJSONValue(version), nil
	}

	// Final fallback to ls approach
	return n.getVersionFromLS(ctx, name)
}

// getVersionFromLS gets version using npm ls command as fallback
func (n *NpmManager) getVersionFromLS(ctx context.Context, name string) (string, error) {
	output, err := n.Executor.Execute(ctx, n.GetBinary(), "ls", "-g", name, "--depth=0")
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	result := strings.TrimSpace(string(output))
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines like "├── packagename@version" or "└── packagename@version"
		if strings.Contains(line, name+"@") {
			parts := strings.Split(line, "@")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[len(parts)-1])
				// Clean up any extra characters that might be in the version
				if idx := strings.Index(version, " "); idx > 0 {
					version = version[:idx]
				}
				return version, nil
			}
		}
	}

	return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
		fmt.Sprintf("could not extract version for package '%s' from npm output", name))
}

// cleanJSONValue removes quotes and commas from a JSON value
func cleanJSONValue(value string) string {
	value = strings.Trim(value, `"`)
	value = strings.TrimSuffix(value, ",")
	return value
}
