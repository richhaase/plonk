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

// PipManager manages Python packages via pip using BaseManager for common functionality.
type PipManager struct {
	*BaseManager
}

// NewPipManager creates a new pip manager with the default executor.
func NewPipManager() *PipManager {
	return newPipManager(nil)
}

// NewPipManagerWithExecutor creates a new pip manager with a custom executor for testing.
func NewPipManagerWithExecutor(exec executor.CommandExecutor) *PipManager {
	return newPipManager(exec)
}

// newPipManager creates a pip manager with the given executor.
func newPipManager(exec executor.CommandExecutor) *PipManager {
	config := ManagerConfig{
		BinaryName:       "pip",
		FallbackBinaries: []string{"pip3"},
		VersionArgs:      []string{"--version"},
		ListArgs: func() []string {
			return []string{"list", "--user", "--format=json"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", "--user", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", "-y", pkg}
		},
		PreferJSON: false, // We already include --format=json in ListArgs
	}

	// Add pip-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "Could not find", "No matching distribution", "ERROR: No matching distribution")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "Requirement already satisfied", "already satisfied")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "WARNING: Skipping", "not installed", "Cannot uninstall")
	errorMatcher.AddPattern(ErrorTypePermission, "Permission denied", "access is denied")

	var base *BaseManager
	if exec == nil {
		base = NewBaseManager(config)
	} else {
		base = NewBaseManagerWithExecutor(config, exec)
	}
	base.ErrorMatcher = errorMatcher

	return &PipManager{
		BaseManager: base,
	}
}

// IsAvailable inherits from BaseManager which handles pip/pip3 fallback

// ListInstalled lists all user-installed pip packages.
func (p *PipManager) ListInstalled(ctx context.Context) ([]string, error) {
	pipCmd := p.GetBinary()
	args := p.Config.ListArgs()

	output, err := p.Executor.Execute(ctx, pipCmd, args...)
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

	return p.parseListOutput(output)
}

// parseListOutput parses pip list JSON output
func (p *PipManager) parseListOutput(output []byte) ([]string, error) {
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
		return p.parseListOutputPlainText(output), nil
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

// parseListOutputPlainText parses non-JSON pip list output
func (p *PipManager) parseListOutputPlainText(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
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

	return packages
}

// listInstalledFallback lists packages without JSON format (for older pip versions)
func (p *PipManager) listInstalledFallback(ctx context.Context) ([]string, error) {
	pipCmd := p.GetBinary()
	output, err := p.Executor.Execute(ctx, pipCmd, "list", "--user")
	if err != nil {
		// Try without --user flag as last resort
		output, err = p.Executor.Execute(ctx, pipCmd, "list")
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
				"failed to execute pip list command")
		}
	}

	return p.parseListOutputPlainText(output), nil
}

// Install installs a pip package for the user with custom retry logic for --user flag issues.
func (p *PipManager) Install(ctx context.Context, name string) error {
	pipCmd := p.GetBinary()
	args := p.Config.InstallArgs(name)

	output, err := p.Executor.ExecuteCombined(ctx, pipCmd, args...)
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions using ErrorMatcher
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			errorType := p.ErrorMatcher.MatchError(outputStr)

			switch errorType {
			case ErrorTypeNotFound:
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("package '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages or check the package name"))

			case ErrorTypeAlreadyInstalled:
				// Package is already installed - this is typically fine
				return nil

			case ErrorTypePermission:
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
					fmt.Sprintf("permission denied installing %s", name)).
					WithSuggestionMessage("Try with elevated permissions or check file permissions")

			default:
				// Check for --user flag issues
				if strings.Contains(outputStr, "--user") && strings.Contains(outputStr, "error") {
					// Try without --user flag
					_, retryErr := p.Executor.ExecuteCombined(ctx, pipCmd, "install", name)
					if retryErr == nil {
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
func (p *PipManager) Uninstall(ctx context.Context, name string) error {
	return p.ExecuteUninstall(ctx, name)
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

// SupportsSearch returns true as pip has a search command.
func (p *PipManager) SupportsSearch() bool {
	return true
}

// Search searches for packages in PyPI using pip search.
func (p *PipManager) Search(ctx context.Context, query string) ([]string, error) {
	pipCmd := p.GetBinary()
	output, err := p.Executor.Execute(ctx, pipCmd, "search", query)
	if err != nil {
		// Check if search command is not available (some pip versions/configurations)
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "ERROR") && strings.Contains(outputStr, "XMLRPC") {
				// PyPI disabled XMLRPC search API, return helpful message
				return nil, errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "search",
					"PyPI search API is currently disabled").
					WithSuggestionMessage(fmt.Sprintf("Visit https://pypi.org/search/?q=%s to search for packages", query))
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
			"failed to search pip packages")
	}

	return p.parseSearchOutput(output), nil
}

// parseSearchOutput parses pip search output
func (p *PipManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// pip search format: "package-name (version) - Description"
		// Extract just the package name
		if parenIndex := strings.Index(line, " ("); parenIndex > 0 {
			packageName := strings.TrimSpace(line[:parenIndex])
			if packageName != "" {
				packages = append(packages, packageName)
			}
		}
	}

	return packages
}

// Info retrieves detailed information about a package.
func (p *PipManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	pipCmd := p.GetBinary()
	output, err := p.Executor.Execute(ctx, pipCmd, "show", name)
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

	info := p.parseInfoOutput(output, name)
	info.Manager = "pip"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses pip show output
func (p *PipManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{
		Name: name,
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

	return info
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
	pipCmd := p.GetBinary()
	output, err := p.Executor.Execute(ctx, pipCmd, "show", name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	version := p.extractVersion(output)
	if version == "" {
		return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
			fmt.Sprintf("could not extract version for package '%s' from pip output", name))
	}

	return version, nil
}

// extractVersion extracts version from pip show output
func (p *PipManager) extractVersion(output []byte) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			if version != "" {
				return version
			}
		}
	}
	return ""
}

// normalizeName normalizes a package name according to pip's rules
// (lowercase and replace - with _)
func (p *PipManager) normalizeName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}
