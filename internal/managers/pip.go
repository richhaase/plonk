// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// PipManager manages Python packages via pip.
type PipManager struct{}

// NewPipManager creates a new pip manager.
func NewPipManager() *PipManager {
	return &PipManager{}
}

// IsAvailable checks if pip is installed and accessible.
func (p *PipManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("pip")
	if err != nil {
		// Try pip3 as fallback
		_, err = exec.LookPath("pip3")
		if err != nil {
			// Neither pip nor pip3 found in PATH - this is not an error condition
			return false, nil
		}
	}

	// Verify pip is actually functional by running a simple command
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// pip exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "pip binary found but not functional")
	}

	return true, nil
}

// ListInstalled lists all user-installed pip packages.
func (p *PipManager) ListInstalled(ctx context.Context) ([]string, error) {
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "list", "--user", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		// Check if --user flag is not supported (some pip installations)
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			if strings.Contains(stderr, "--user") || strings.Contains(stderr, "unknown option") {
				// Try without --user flag
				return p.listInstalledFallback(ctx)
			}
		}
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
			"failed to execute pip list command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse JSON output
	var packages []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(result), &packages); err != nil {
		// Fallback to plain text parsing if JSON fails
		return p.listInstalledFallback(ctx)
	}

	// Extract package names
	var names []string
	for _, pkg := range packages {
		if pkg.Name != "" {
			names = append(names, strings.ToLower(pkg.Name))
		}
	}

	return names, nil
}

// listInstalledFallback lists packages without JSON format (for older pip versions)
func (p *PipManager) listInstalledFallback(ctx context.Context) ([]string, error) {
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "list", "--user")
	output, err := cmd.Output()
	if err != nil {
		// Try without --user flag as last resort
		cmd = exec.CommandContext(ctx, pipCmd, "list")
		output, err = cmd.Output()
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
				"failed to execute pip list command")
		}
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	// Parse plain text output (skip header lines)
	var packages []string
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		// Skip header lines (usually first 2 lines)
		if i < 2 || strings.Contains(line, "---") {
			continue
		}
		line = strings.TrimSpace(line)
		if line != "" {
			// Extract package name (before version)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				packages = append(packages, strings.ToLower(parts[0]))
			}
		}
	}

	return packages, nil
}

// Install installs a pip package for the user.
func (p *PipManager) Install(ctx context.Context, name string) error {
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "install", "--user", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for package not found
			if strings.Contains(outputStr, "No matching distribution") || strings.Contains(outputStr, "Could not find") {
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("package '%s' not found in PyPI", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages: pip search %s (or check https://pypi.org)", name))
			}

			// Check for already installed
			if strings.Contains(outputStr, "Requirement already satisfied") {
				// Package is already installed - this is typically fine
				return nil
			}

			// Check for permission errors
			if strings.Contains(outputStr, "Permission denied") || strings.Contains(outputStr, "access is denied") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
					fmt.Sprintf("permission denied installing %s", name)).
					WithSuggestionMessage("Try using --user flag or fix pip permissions")
			}

			// Check for --user flag issues
			if strings.Contains(outputStr, "--user") && strings.Contains(outputStr, "error") {
				// Try without --user flag
				cmd = exec.CommandContext(ctx, pipCmd, "install", name)
				_, err = cmd.CombinedOutput()
				if err == nil {
					return nil
				}
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("package installation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute pip install command")
	}

	return nil
}

// Uninstall removes a pip package.
func (p *PipManager) Uninstall(ctx context.Context, name string) error {
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "uninstall", "-y", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for "not installed" condition
			if strings.Contains(outputStr, "not installed") || strings.Contains(outputStr, "Cannot uninstall") {
				// Package is not installed - this is typically fine for uninstall
				return nil
			}

			// Check for permission errors
			if strings.Contains(outputStr, "Permission denied") || strings.Contains(outputStr, "access is denied") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
					fmt.Sprintf("permission denied uninstalling %s", name)).
					WithSuggestionMessage("Try with elevated permissions or check file ownership")
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
				fmt.Sprintf("package uninstallation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute pip uninstall command")
	}

	return nil
}

// IsInstalled checks if a specific package is installed.
func (p *PipManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	// Normalize package name for comparison
	normalizedName := p.normalizeName(name)

	packages, err := p.ListInstalled(ctx)
	if err != nil {
		return false, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "check", name,
			"failed to check package installation status")
	}

	for _, pkg := range packages {
		if p.normalizeName(pkg) == normalizedName {
			return true, nil
		}
	}

	return false, nil
}

// Search searches for packages in PyPI.
func (p *PipManager) Search(ctx context.Context, query string) ([]string, error) {
	// pip search is deprecated, use PyPI API instead
	return p.searchPyPI(ctx, query)
}

// searchPyPI searches PyPI using the JSON API
func (p *PipManager) searchPyPI(ctx context.Context, query string) ([]string, error) {
	// For now, return a helpful message about pip search deprecation
	// In a full implementation, this would use the PyPI JSON API
	return nil, errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "search",
		"pip search is deprecated").
		WithSuggestionMessage(fmt.Sprintf("Visit https://pypi.org/search/?q=%s to search for packages", query))
}

// Info retrieves detailed information about a package.
func (p *PipManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "show", name)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if strings.Contains(string(exitError.Stderr), "not found") || exitError.ExitCode() == 1 {
				return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
					fmt.Sprintf("package '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Check available packages at https://pypi.org/project/%s", name))
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get package info")
	}

	// Parse pip show output
	info := &PackageInfo{
		Name:      name,
		Manager:   "pip",
		Installed: installed,
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Summary:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "Summary:"))
		} else if strings.HasPrefix(line, "Home-page:") {
			info.Homepage = strings.TrimSpace(strings.TrimPrefix(line, "Home-page:"))
		} else if strings.HasPrefix(line, "Requires:") {
			requires := strings.TrimSpace(strings.TrimPrefix(line, "Requires:"))
			if requires != "" {
				deps := strings.Split(requires, ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						info.Dependencies = append(info.Dependencies, dep)
					}
				}
			}
		}
	}

	return info, nil
}

// GetInstalledVersion retrieves the installed version of a pip package
func (p *PipManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check package installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("package '%s' is not installed", name))
	}

	// Get version using pip show
	pipCmd := p.getPipCommand()
	cmd := exec.CommandContext(ctx, pipCmd, "show", name)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			if version != "" {
				return version, nil
			}
		}
	}

	return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
		fmt.Sprintf("could not extract version for package '%s' from pip output", name))
}

// getPipCommand returns the appropriate pip command (pip or pip3)
func (p *PipManager) getPipCommand() string {
	// Try pip first
	if _, err := exec.LookPath("pip"); err == nil {
		return "pip"
	}
	// Fall back to pip3
	return "pip3"
}

// normalizeName normalizes a package name according to pip's rules
// (lowercase and replace - with _)
func (p *PipManager) normalizeName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}
