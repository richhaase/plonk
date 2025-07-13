// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// NpmManager manages NPM packages.
type NpmManager struct{}

// NewNpmManager creates a new NPM manager.
func NewNpmManager() *NpmManager {
	return &NpmManager{}
}

// IsAvailable checks if NPM is installed and accessible.
func (n *NpmManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("npm")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify npm is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "npm", "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// npm exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "npm binary found but not functional")
	}

	return true, nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--parseable")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// npm list can return non-zero exit codes even when working correctly
			// (e.g., when there are peer dependency warnings)
			// Only treat it as an error if the exit code indicates a real failure
			if exitError.ExitCode() > 1 {
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

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output to extract package names
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "/") {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if pkg != "" && pkg != "lib" {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages, nil
}

// Install installs a global NPM package.
func (n *NpmManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// NPM often returns exit code 1 for warnings, check actual content
			if exitError.ExitCode() == 1 {
				// Check for "already installed" warnings
				if strings.Contains(outputStr, "already installed") || strings.Contains(outputStr, "up to date") {
					// Package is already installed - this is typically fine
					return nil
				}

				// Check for package not found
				if strings.Contains(outputStr, "404") || strings.Contains(outputStr, "Not found") || strings.Contains(outputStr, "E404") {
					return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
						fmt.Sprintf("package '%s' not found in npm registry", name)).
						WithSuggestionMessage(fmt.Sprintf("Search available packages: npm search %s", name))
				}

				// Check for permission errors
				if strings.Contains(outputStr, "EACCES") || strings.Contains(outputStr, "permission denied") {
					return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
						fmt.Sprintf("permission denied installing %s", name)).
						WithSuggestionMessage("Try running with sudo or fix npm permissions")
				}
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("package installation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute npm install command")
	}

	return nil
}

// Uninstall removes a global NPM package.
func (n *NpmManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "uninstall", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// NPM often returns exit code 1 for warnings, check actual content
			if exitError.ExitCode() == 1 {
				// Check for "not installed" warnings
				if strings.Contains(outputStr, "not installed") || strings.Contains(outputStr, "up to date") {
					// Package is not installed - this is typically fine for uninstall
					return nil
				}

				// Check for permission errors
				if strings.Contains(outputStr, "EACCES") || strings.Contains(outputStr, "permission denied") {
					return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
						fmt.Sprintf("permission denied uninstalling %s", name)).
						WithSuggestionMessage("Try running with sudo or fix npm permissions")
				}

				// Check for dependency issues (less common in npm global packages)
				if strings.Contains(outputStr, "ENOENT") || strings.Contains(outputStr, "cannot remove") {
					return errors.NewError(errors.ErrPackageUninstall, errors.DomainPackages, "uninstall",
						fmt.Sprintf("cannot uninstall %s: package files may be corrupted or missing", name)).
						WithSuggestionMessage(fmt.Sprintf("Try reinstalling first: npm install -g %s", name))
				}
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
				fmt.Sprintf("package uninstallation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute npm uninstall command")
	}

	return nil
}

// IsInstalled checks if a specific package is installed globally.
func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
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
	cmd := exec.CommandContext(ctx, "npm", "search", query, "--json")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// For npm search, exit code 1 usually means no results found
			if exitError.ExitCode() == 1 {
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

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		// No packages found - this is normal, not an error
		return []string{}, nil
	}

	// Parse JSON output to extract package names
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, `"name":`) {
			// Extract package name from JSON line like: "name": "package-name",
			parts := strings.Split(line, `"name":`)
			if len(parts) > 1 {
				namepart := strings.TrimSpace(parts[1])
				if strings.HasPrefix(namepart, `"`) && strings.Contains(namepart, `"`) {
					// Extract the name between quotes
					namepart = namepart[1:] // Remove leading quote
					if idx := strings.Index(namepart, `"`); idx > 0 {
						packageName := namepart[:idx]
						if packageName != "" {
							packages = append(packages, packageName)
						}
					}
				}
			}
		}
	}

	return packages, nil
}

// Info retrieves detailed information about a package from NPM.
func (n *NpmManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := n.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	var info *PackageInfo
	if installed {
		// Get info from installed package
		info, err = n.getInstalledPackageInfo(ctx, name)
		if err != nil {
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
				"failed to get installed package information")
		}
	} else {
		// Get info from available package
		info, err = n.getAvailablePackageInfo(ctx, name)
		if err != nil {
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
				"failed to get available package information")
		}
	}

	info.Manager = "npm"
	info.Installed = installed
	return info, nil
}

// getInstalledPackageInfo gets information about an installed package
func (n *NpmManager) getInstalledPackageInfo(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name, "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get installed package info")
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "{}" {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("installed package '%s' not found", name))
	}

	// Parse JSON output to get version
	info := &PackageInfo{Name: name}
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"version":`) {
			info.Version = n.extractJSONValue(line, "version")
		}
	}

	// Get additional info from npm view for installed packages
	viewInfo, err := n.getPackageView(ctx, name)
	if err == nil {
		info.Description = viewInfo.Description
		info.Homepage = viewInfo.Homepage
		info.Dependencies = viewInfo.Dependencies
	}

	return info, nil
}

// getAvailablePackageInfo gets information about an available (but not installed) package
func (n *NpmManager) getAvailablePackageInfo(ctx context.Context, name string) (*PackageInfo, error) {
	return n.getPackageView(ctx, name)
}

// getPackageView gets package information using npm view
func (n *NpmManager) getPackageView(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, "npm", "view", name, "--json")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
					fmt.Sprintf("package '%s' not found", name))
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get package info")
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "{}" {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("package '%s' not found", name))
	}

	// Parse JSON output
	info := &PackageInfo{Name: name}
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"name":`) {
			info.Name = n.extractJSONValue(line, "name")
		} else if strings.Contains(line, `"version":`) {
			info.Version = n.extractJSONValue(line, "version")
		} else if strings.Contains(line, `"description":`) {
			info.Description = n.extractJSONValue(line, "description")
		} else if strings.Contains(line, `"homepage":`) {
			info.Homepage = n.extractJSONValue(line, "homepage")
		} else if strings.Contains(line, `"dependencies":`) {
			// Dependencies are in a nested object, we'll parse them separately
			info.Dependencies = n.extractDependencies(result)
		}
	}

	return info, nil
}

// extractDependencies extracts dependencies from the npm view JSON output
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
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name, "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative approach if JSON fails
		return n.getVersionFromLS(ctx, name)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("no version information found for package '%s'", name))
	}

	// Parse JSON to extract version
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"`+name+`":`) && strings.Contains(line, `"version":`) {
			version := n.extractJSONValue(line, "version")
			if version != "" {
				return version, nil
			}
		}
		// Also check for direct version field after package name
		if strings.Contains(line, `"version":`) {
			version := n.extractJSONValue(line, "version")
			if version != "" {
				return version, nil
			}
		}
	}

	// Fallback to ls approach
	return n.getVersionFromLS(ctx, name)
}

// getVersionFromLS gets version using npm ls command as fallback
func (n *NpmManager) getVersionFromLS(ctx context.Context, name string) (string, error) {
	cmd := exec.CommandContext(ctx, "npm", "ls", "-g", name, "--depth=0")
	output, err := cmd.Output()
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

// extractJSONValue extracts a value from a JSON line
func (n *NpmManager) extractJSONValue(line, key string) string {
	keyPattern := `"` + key + `":`
	if !strings.Contains(line, keyPattern) {
		return ""
	}

	parts := strings.Split(line, keyPattern)
	if len(parts) < 2 {
		return ""
	}

	valuepart := strings.TrimSpace(parts[1])
	if strings.HasPrefix(valuepart, `"`) {
		valuepart = valuepart[1:] // Remove leading quote
		if idx := strings.Index(valuepart, `"`); idx > 0 {
			return valuepart[:idx]
		}
	}
	return ""
}
