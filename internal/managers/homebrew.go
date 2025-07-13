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

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct{}

// NewHomebrewManager creates a new Homebrew manager.
func NewHomebrewManager() *HomebrewManager {
	return &HomebrewManager{}
}

// IsAvailable checks if Homebrew is installed and accessible.
func (h *HomebrewManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("brew")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify brew is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "brew", "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// brew exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "brew binary found but not functional")
	}

	return true, nil
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "brew", "list")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if _, ok := err.(*exec.ExitError); ok {
			// For brew list, any non-zero exit usually indicates a real problem
			// (brew list returns exit 0 even with no packages installed)
			return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list", "failed to list homebrew packages")
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list", "failed to execute brew list")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output into package list
	packages := strings.Split(result, "\n")
	// Filter out any empty strings that might result from parsing
	var filteredPackages []string
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			filteredPackages = append(filteredPackages, trimmed)
		}
	}

	return filteredPackages, nil
}

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for "already installed" - this might be a warning, not an error
			if strings.Contains(outputStr, "already installed") {
				// Package is already installed - this is typically fine
				return nil
			}

			// Check for package not found
			if strings.Contains(outputStr, "No available formula") || strings.Contains(outputStr, "No formulae found") {
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("package '%s' not found in homebrew repositories", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages: brew search %s", name))
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("package installation failed (exit code %d)", exitError.ExitCode())).
				WithSuggestionMessage(fmt.Sprintf("Check package availability: brew search %s", name))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute brew install command")
	}
	return nil
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "uninstall", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for "not installed" - this might be fine
			if strings.Contains(outputStr, "No such keg") || strings.Contains(outputStr, "not installed") {
				// Package is not installed - this is typically fine for uninstall
				return nil
			}

			// Check for dependency issues
			if strings.Contains(outputStr, "because it is required by") || strings.Contains(outputStr, "still has dependents") {
				return errors.NewError(errors.ErrPackageUninstall, errors.DomainPackages, "uninstall",
					fmt.Sprintf("cannot uninstall %s: package has dependents that require it", name)).
					WithSuggestionMessage(fmt.Sprintf("Force uninstall with: brew uninstall --ignore-dependencies %s", name))
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
				fmt.Sprintf("package uninstallation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute brew uninstall command")
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (h *HomebrewManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "brew", "list", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (brew not found, permission issues, etc.)
		return false, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "check", name,
			"failed to check package installation status")
	}
	return true, nil
}

// Search searches for packages in Homebrew repositories.
func (h *HomebrewManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "brew", "search", query)
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// For brew search, exit code 1 usually means no results found
			if exitError.ExitCode() == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
				"homebrew search command failed")
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "search",
			"failed to execute brew search command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages found - this is normal, not an error
		return []string{}, nil
	}

	// Parse output into package list
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Homebrew search can return multiple packages per line separated by spaces
			parts := strings.Fields(line)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					packages = append(packages, part)
				}
			}
		}
	}

	return packages, nil
}

// Info retrieves detailed information about a package from Homebrew.
func (h *HomebrewManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	var info *PackageInfo
	if installed {
		// Get info from installed package
		info, err = h.getInstalledPackageInfo(ctx, name)
		if err != nil {
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
				"failed to get installed package information")
		}
	} else {
		// Get info from available package
		info, err = h.getAvailablePackageInfo(ctx, name)
		if err != nil {
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
				"failed to get available package information")
		}
	}

	info.Manager = "homebrew"
	info.Installed = installed
	return info, nil
}

// getInstalledPackageInfo gets information about an installed package
func (h *HomebrewManager) getInstalledPackageInfo(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, "brew", "info", name, "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get installed package info")
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("package '%s' not found", name)).WithSuggestionMessage(fmt.Sprintf("Search available packages: brew search %s", name))
	}

	// Parse JSON output - homebrew returns an array
	info := &PackageInfo{Name: name}
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"name":`) {
			info.Name = h.extractJSONValue(line, "name")
		} else if strings.Contains(line, `"desc":`) {
			info.Description = h.extractJSONValue(line, "desc")
		} else if strings.Contains(line, `"homepage":`) {
			info.Homepage = h.extractJSONValue(line, "homepage")
		} else if strings.Contains(line, `"installed":`) && strings.Contains(line, `"version":`) {
			info.Version = h.extractJSONValue(line, "version")
		}
	}

	return info, nil
}

// getAvailablePackageInfo gets information about an available (but not installed) package
func (h *HomebrewManager) getAvailablePackageInfo(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, "brew", "info", name, "--json")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
					fmt.Sprintf("package '%s' not found", name)).WithSuggestionMessage(fmt.Sprintf("Search available packages: brew search %s", name))
			}
		}
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "info",
			"failed to get package info")
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("package '%s' not found", name)).WithSuggestionMessage(fmt.Sprintf("Search available packages: brew search %s", name))
	}

	// Parse JSON output
	info := &PackageInfo{Name: name}
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"name":`) {
			info.Name = h.extractJSONValue(line, "name")
		} else if strings.Contains(line, `"desc":`) {
			info.Description = h.extractJSONValue(line, "desc")
		} else if strings.Contains(line, `"homepage":`) {
			info.Homepage = h.extractJSONValue(line, "homepage")
		} else if strings.Contains(line, `"versions":`) && strings.Contains(line, `"stable":`) {
			info.Version = h.extractJSONValue(line, "stable")
		}
	}

	return info, nil
}

// GetInstalledVersion retrieves the installed version of a package
func (h *HomebrewManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check package installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("package '%s' is not installed", name))
	}

	// Get version using brew list with --versions flag
	cmd := exec.CommandContext(ctx, "brew", "list", "--versions", name)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("no version information found for package '%s'", name))
	}

	// Parse output: "package_name version1 version2 ..."
	// We want the latest (last) version
	parts := strings.Fields(result)
	if len(parts) < 2 {
		return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
			fmt.Sprintf("unexpected version output format for package '%s': %s", name, result))
	}

	// Return the last version (most recent)
	return parts[len(parts)-1], nil
}

// extractJSONValue extracts a value from a JSON line
func (h *HomebrewManager) extractJSONValue(line, key string) string {
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
