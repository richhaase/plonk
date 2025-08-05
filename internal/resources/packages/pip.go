// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PipManager manages Python packages via pip.
type PipManager struct {
	binary string
}

// NewPipManager creates a new pip manager.
func NewPipManager() *PipManager {
	return &PipManager{
		binary: "pip3",
	}
}

// ListInstalled lists all user-installed pip packages.
func (p *PipManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "list", "--user", "--format=json")
	if err != nil {
		// Check if --user flag is not supported (some pip installations)
		if exitCode, ok := ExtractExitCode(err); ok && exitCode != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "--user") || strings.Contains(outputStr, "unknown option") || strings.Contains(outputStr, "no such option") {
				// Try without --user flag
				return p.listInstalledFallback(ctx)
			}
		}
		return nil, err
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
	output, err := ExecuteCommand(ctx, p.binary, "list", "--user")
	if err != nil {
		// Try without --user flag as last resort
		output, err = ExecuteCommand(ctx, p.binary, "list")
		if err != nil {
			return nil, err
		}
	}

	return p.parseListOutputPlainText(output), nil
}

// Install installs a pip package for the user with custom retry logic for --user flag issues.
func (p *PipManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "install", "--user", "--break-system-packages", name)
	if err != nil {
		outputStr := string(output)

		// Check for --user flag issues first
		if strings.Contains(outputStr, "--user") && strings.Contains(strings.ToLower(outputStr), "error") {
			// Try without --user flag
			output2, err2 := ExecuteCommandCombined(ctx, p.binary, "install", "--break-system-packages", name)
			if err2 == nil {
				return nil
			}
			// Use the retry error if it failed
			output = output2
			err = err2
		}

		return p.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a pip package.
func (p *PipManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "uninstall", "-y", "--break-system-packages", name)
	if err != nil {
		return p.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (p *PipManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	// Normalize package name for comparison
	normalizedName := p.normalizeName(name)

	packages, err := p.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
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
	output, err := ExecuteCommand(ctx, p.binary, "search", query)
	if err != nil {
		// Check if search command is not available (some pip versions/configurations)
		if exitCode, ok := ExtractExitCode(err); ok && exitCode != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "ERROR") && strings.Contains(outputStr, "XMLRPC") {
				// PyPI disabled XMLRPC search API, return helpful message
				return nil, fmt.Errorf("PyPI search API is currently disabled. Visit https://pypi.org/search/?q=%s to search for packages", query)
			}
		}
		return nil, fmt.Errorf("failed to search pip packages for %s: %w", query, err)
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
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	output, err := ExecuteCommand(ctx, p.binary, "show", name)
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode != 0 {
			outputStr := string(output)
			if strings.Contains(outputStr, "not found") {
				return nil, fmt.Errorf("package '%s' not found", name)
			}
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
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

// InstalledVersion retrieves the installed version of a pip package
func (p *PipManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Get version using pip show
	output, err := ExecuteCommand(ctx, p.binary, "show", name)
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	version := p.extractVersion(output)
	if version == "" {
		return "", fmt.Errorf("could not extract version for package '%s' from pip output", name)
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
				return cleanVersionString(version)
			}
		}
	}
	return ""
}

// cleanVersionString removes common version prefixes
func cleanVersionString(version string) string {
	// Remove common prefixes
	prefixes := []string{"v", "version", "Version"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(version, prefix) {
			version = strings.TrimSpace(strings.TrimPrefix(version, prefix))
			break
		}
	}

	// Remove common suffixes and extra information
	if idx := strings.Index(version, " "); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "\t"); idx > 0 {
		version = version[:idx]
	}

	return strings.TrimSpace(version)
}

// normalizeName normalizes a package name according to pip's rules
// (lowercase and replace - with _)
func (p *PipManager) normalizeName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

// IsAvailable checks if pip3 is installed and accessible
func (p *PipManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(p.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, p.binary, []string{"--version"})
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

// handleInstallError processes install command errors
func (p *PipManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for known error patterns
	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(strings.ToLower(outputStr), "could not find") || strings.Contains(outputStr, "no matching distribution") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "requirement already satisfied") || strings.Contains(outputStr, "already satisfied") {
			return nil // Already installed is success
		}
		if strings.Contains(strings.ToLower(outputStr), "permission denied") || strings.Contains(outputStr, "access is denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
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

	// Direct error return for other cases
	return err
}

// handleUninstallError processes uninstall command errors
func (p *PipManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for known error patterns
	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "warning: skipping") || strings.Contains(outputStr, "not installed") || strings.Contains(outputStr, "cannot uninstall") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") || strings.Contains(outputStr, "access is denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}

		if exitCode != 0 {
			return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	// Direct error return for other cases
	return err
}
