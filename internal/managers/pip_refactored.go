// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/executor"
)

// PipManagerV2 manages Python packages via pip using CommandExecutor for testability.
type PipManagerV2 struct {
	executor     executor.CommandExecutor
	pipCommand   string // Cached pip command (pip or pip3)
	errorMatcher *ErrorMatcher
}

// NewPipManagerV2 creates a new pip manager with the default executor.
func NewPipManagerV2() *PipManagerV2 {
	return &PipManagerV2{
		executor:     &executor.RealCommandExecutor{},
		errorMatcher: NewCommonErrorMatcher(),
	}
}

// NewPipManagerV2WithExecutor creates a new pip manager with a custom executor for testing.
func NewPipManagerV2WithExecutor(exec executor.CommandExecutor) *PipManagerV2 {
	return &PipManagerV2{
		executor:     exec,
		errorMatcher: NewCommonErrorMatcher(),
	}
}

// IsAvailable checks if pip is installed and accessible.
func (p *PipManagerV2) IsAvailable(ctx context.Context) (bool, error) {
	// Check pip first
	_, err := p.executor.LookPath("pip")
	if err == nil {
		// Verify pip is functional
		_, verifyErr := p.executor.Execute(ctx, "pip", "--version")
		if verifyErr == nil {
			p.pipCommand = "pip"
			return true, nil
		}
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// Also check if the error itself is a context error
		if verifyErr == context.Canceled || verifyErr == context.DeadlineExceeded {
			return false, verifyErr
		}
		// pip exists but is not functional - continue to try pip3
	}

	// Try pip3 as fallback
	_, err = p.executor.LookPath("pip3")
	if err != nil {
		// Neither pip nor pip3 found in PATH - this is not an error condition
		return false, nil
	}

	// Verify pip3 is functional
	_, err = p.executor.Execute(ctx, "pip3", "--version")
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// Also check if the error itself is a context error
		if err == context.Canceled || err == context.DeadlineExceeded {
			return false, err
		}
		// pip3 exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "pip binary found but not functional")
	}

	p.pipCommand = "pip3"
	return true, nil
}

// ListInstalled lists all user-installed pip packages.
func (p *PipManagerV2) ListInstalled(ctx context.Context) ([]string, error) {
	pipCmd := p.getPipCommand()
	output, err := p.executor.Execute(ctx, pipCmd, "list", "--user", "--format=json")
	if err != nil {
		// Check if --user flag is not supported (some pip installations)
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "--user") || strings.Contains(outputStr, "unknown option") {
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
func (p *PipManagerV2) listInstalledFallback(ctx context.Context) ([]string, error) {
	pipCmd := p.getPipCommand()
	output, err := p.executor.Execute(ctx, pipCmd, "list", "--user")
	if err != nil {
		// Try without --user flag as last resort
		output, err = p.executor.Execute(ctx, pipCmd, "list")
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
func (p *PipManagerV2) Install(ctx context.Context, name string) error {
	pipCmd := p.getPipCommand()
	output, err := p.executor.ExecuteCombined(ctx, pipCmd, "install", "--user", name)
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions using ErrorMatcher
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			errorType := p.errorMatcher.MatchError(outputStr)

			switch errorType {
			case ErrorTypeNotFound:
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("package '%s' not found in PyPI", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages: pip search %s (or check https://pypi.org)", name))

			case ErrorTypeAlreadyInstalled:
				// Package is already installed - this is typically fine
				return nil

			case ErrorTypePermission:
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
					fmt.Sprintf("permission denied installing %s", name)).
					WithSuggestionMessage("Try using --user flag or fix pip permissions")

			default:
				// Check for --user flag issues
				if strings.Contains(outputStr, "--user") && strings.Contains(outputStr, "error") {
					// Try without --user flag
					_, err = p.executor.ExecuteCombined(ctx, pipCmd, "install", name)
					if err == nil {
						return nil
					}
				}

				// Other exit errors with more context
				return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
					fmt.Sprintf("package installation failed (exit code %d)", execErr.ExitCode()))
			}
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute pip install command")
	}

	return nil
}

// Uninstall removes a pip package.
func (p *PipManagerV2) Uninstall(ctx context.Context, name string) error {
	pipCmd := p.getPipCommand()
	output, err := p.executor.ExecuteCombined(ctx, pipCmd, "uninstall", "-y", name)
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions using ErrorMatcher
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			errorType := p.errorMatcher.MatchError(outputStr)

			switch errorType {
			case ErrorTypeNotInstalled:
				// Package is not installed - this is typically fine for uninstall
				return nil

			case ErrorTypePermission:
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
					fmt.Sprintf("permission denied uninstalling %s", name)).
					WithSuggestionMessage("Try with elevated permissions or check file ownership")

			default:
				// Check for special "Cannot uninstall" case
				if strings.Contains(outputStr, "Cannot uninstall") {
					// Package is not installed - this is typically fine for uninstall
					return nil
				}

				// Other exit errors with more context
				return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
					fmt.Sprintf("package uninstallation failed (exit code %d)", execErr.ExitCode()))
			}
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute pip uninstall command")
	}

	return nil
}

// IsInstalled checks if a specific package is installed.
func (p *PipManagerV2) IsInstalled(ctx context.Context, name string) (bool, error) {
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
func (p *PipManagerV2) Search(ctx context.Context, query string) ([]string, error) {
	// pip search is deprecated, use PyPI API instead
	return p.searchPyPI(ctx, query)
}

// searchPyPI searches PyPI using the JSON API
func (p *PipManagerV2) searchPyPI(ctx context.Context, query string) ([]string, error) {
	// For now, return a helpful message about pip search deprecation
	// In a full implementation, this would use the PyPI JSON API
	return nil, errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "search",
		"pip search is deprecated").
		WithSuggestionMessage(fmt.Sprintf("Visit https://pypi.org/search/?q=%s to search for packages", query))
}

// Info retrieves detailed information about a package.
func (p *PipManagerV2) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	pipCmd := p.getPipCommand()
	output, err := p.executor.Execute(ctx, pipCmd, "show", name)
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "not found") {
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
func (p *PipManagerV2) GetInstalledVersion(ctx context.Context, name string) (string, error) {
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
	output, err := p.executor.Execute(ctx, pipCmd, "show", name)
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
func (p *PipManagerV2) getPipCommand() string {
	// Use cached value if available
	if p.pipCommand != "" {
		return p.pipCommand
	}

	// Try pip first
	if _, err := p.executor.LookPath("pip"); err == nil {
		p.pipCommand = "pip"
		return p.pipCommand
	}
	// Fall back to pip3
	p.pipCommand = "pip3"
	return p.pipCommand
}

// normalizeName normalizes a package name according to pip's rules
// (lowercase and replace - with _)
func (p *PipManagerV2) normalizeName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}
