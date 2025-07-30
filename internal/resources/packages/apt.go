// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// AptManager manages Debian/Ubuntu packages via APT.
type AptManager struct{}

// NewAptManager creates a new APT manager.
func NewAptManager() *AptManager {
	return &AptManager{}
}

// SupportsSearch returns true as APT supports searching.
func (a *AptManager) SupportsSearch() bool {
	return true
}

// IsAvailable checks if APT is available on this system.
func (a *AptManager) IsAvailable(ctx context.Context) (bool, error) {
	// Check if APT is supported on this platform
	if !IsPackageManagerSupportedOnPlatform("apt") {
		return false, nil
	}

	// Check for required APT commands
	requiredCommands := []string{"apt-get", "apt-cache", "dpkg-query"}
	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			return false, nil
		}
	}

	return true, nil
}

// ListInstalled lists all installed APT packages (not implemented for APT).
func (a *AptManager) ListInstalled(ctx context.Context) ([]string, error) {
	// APT has thousands of system packages, listing all is not practical
	return nil, fmt.Errorf("listing all APT packages is not supported")
}

// Install installs a package using apt-get.
func (a *AptManager) Install(ctx context.Context, name string) error {
	packageName := a.formatPackageName(name)

	// Build the apt-get install command
	// Using --yes for non-interactive, --no-install-recommends to minimize installs
	args := []string{"install", "--yes", "--no-install-recommends", packageName}

	cmd := exec.CommandContext(ctx, "apt-get", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return a.handleInstallError(err, output, packageName)
	}

	return nil
}

// Uninstall removes a package using apt-get.
func (a *AptManager) Uninstall(ctx context.Context, name string) error {
	packageName := a.formatPackageName(name)

	// Build the apt-get remove command
	// Using remove (not purge) to leave config files
	// Using --yes for non-interactive
	args := []string{"remove", "--yes", packageName}

	cmd := exec.CommandContext(ctx, "apt-get", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return a.handleUninstallError(err, output, packageName)
	}

	return nil
}

// IsInstalled checks if a package is installed using dpkg-query.
func (a *AptManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	packageName := a.formatPackageName(name)

	// Use dpkg-query to check if package is installed
	// dpkg-query returns 0 if package is installed, 1 if not
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${db:Status-Status}", packageName)
	output, err := cmd.Output()

	if err != nil {
		// Check for exit code 1 (package not found)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check package status: %w", err)
	}

	// Package is installed if status is "installed"
	status := strings.TrimSpace(string(output))
	return status == "installed", nil
}

// InstalledVersion returns the version of an installed package.
func (a *AptManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	packageName := a.formatPackageName(name)

	// Use dpkg-query to get version information
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", packageName)
	output, err := cmd.Output()

	if err != nil {
		// Check for exit code 1 (package not found)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", fmt.Errorf("package '%s' is not installed", packageName)
		}
		return "", fmt.Errorf("failed to get package version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("no version information available for package '%s'", packageName)
	}

	return version, nil
}

// Info returns information about a package using apt-cache.
func (a *AptManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	packageName := a.formatPackageName(name)

	// Use apt-cache show to get package information
	cmd := exec.CommandContext(ctx, "apt-cache", "show", packageName)
	output, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 100 {
			return nil, fmt.Errorf("package '%s' not found", packageName)
		}
		return nil, fmt.Errorf("failed to get package info: %w", err)
	}

	// Parse the output
	info := &PackageInfo{
		Name:    packageName,
		Manager: "apt",
	}

	// Check if installed
	installed, _ := a.IsInstalled(ctx, packageName)
	info.Installed = installed

	// Parse apt-cache output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Version: ") {
			info.Version = strings.TrimPrefix(line, "Version: ")
		} else if strings.HasPrefix(line, "Description: ") {
			info.Description = strings.TrimPrefix(line, "Description: ")
		} else if strings.HasPrefix(line, "Homepage: ") {
			info.Homepage = strings.TrimPrefix(line, "Homepage: ")
		} else if strings.HasPrefix(line, "Description-en: ") {
			// Use English description if available
			info.Description = strings.TrimPrefix(line, "Description-en: ")
		}
	}

	// If no description found in first line, look for multi-line description
	if info.Description == "" {
		inDescription := false
		var descLines []string
		for _, line := range lines {
			if strings.HasPrefix(line, "Description: ") || strings.HasPrefix(line, "Description-en: ") {
				inDescription = true
				desc := strings.TrimPrefix(line, "Description: ")
				desc = strings.TrimPrefix(desc, "Description-en: ")
				if desc != "" {
					descLines = append(descLines, desc)
				}
			} else if inDescription && strings.HasPrefix(line, " ") {
				// Continuation of description
				descLines = append(descLines, strings.TrimSpace(line))
			} else if inDescription && !strings.HasPrefix(line, " ") {
				// End of description
				break
			}
		}
		if len(descLines) > 0 {
			info.Description = descLines[0] // Use first line of description
		}
	}

	return info, nil
}

// Search searches for packages using apt-cache.
func (a *AptManager) Search(ctx context.Context, query string) ([]string, error) {
	// Use apt-cache search to find packages
	cmd := exec.CommandContext(ctx, "apt-cache", "search", "--names-only", query)
	output, err := cmd.Output()

	if err != nil {
		// apt-cache search typically doesn't fail, just returns empty results
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	// Parse the output
	// Format: "package-name - description"
	var packages []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract package name (before the " - ")
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) > 0 && parts[0] != "" {
			packages = append(packages, parts[0])
		}
	}

	return packages, nil
}

// formatPackageName handles special APT package naming (pass-through).
func (a *AptManager) formatPackageName(name string) string {
	// Pass through as-is to support:
	// - Virtual packages
	// - Architecture suffixes (package:amd64)
	// - Any special APT naming
	return strings.TrimSpace(name)
}

// handleInstallError processes install command errors
func (a *AptManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for permission denied
	if strings.Contains(outputStr, "Permission denied") ||
		strings.Contains(outputStr, "are you root?") ||
		strings.Contains(outputStr, "E: Could not open lock file") {
		return fmt.Errorf("permission denied: apt-get install requires sudo privileges. Try: sudo plonk install apt:%s", packageName)
	}

	// Check for package not found
	if strings.Contains(outputStr, "E: Unable to locate package") ||
		strings.Contains(outputStr, "has no installation candidate") {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	// Check if already installed
	if strings.Contains(outputStr, "is already the newest version") {
		return nil // Already installed is success
	}

	// Check for dependency issues
	if strings.Contains(outputStr, "E: Broken packages") ||
		strings.Contains(outputStr, "depends on") && strings.Contains(outputStr, "but it is not going to be installed") {
		return fmt.Errorf("dependency conflict installing package '%s'", packageName)
	}

	// Check for network issues
	if strings.Contains(outputStr, "Could not resolve") ||
		strings.Contains(outputStr, "Failed to fetch") {
		return fmt.Errorf("network error: failed to download package information")
	}

	// Generic error with output
	if exitCode, ok := ExtractExitCode(err); ok && exitCode != 0 {
		// Extract meaningful error line from output
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "E:") {
				return fmt.Errorf("apt-get install failed: %s", line)
			}
		}
		return fmt.Errorf("package installation failed (exit code %d)", exitCode)
	}

	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors
func (a *AptManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for permission denied
	if strings.Contains(outputStr, "Permission denied") ||
		strings.Contains(outputStr, "are you root?") ||
		strings.Contains(outputStr, "E: Could not open lock file") {
		return fmt.Errorf("permission denied: apt-get remove requires sudo privileges. Try: sudo plonk uninstall apt:%s", packageName)
	}

	// Check if package is not installed (this is success for uninstall)
	if strings.Contains(outputStr, "is not installed") ||
		strings.Contains(outputStr, "Unable to locate package") {
		return nil // Not installed is success for uninstall
	}

	// Check for dependency issues
	if strings.Contains(outputStr, "is depended on by") ||
		strings.Contains(outputStr, "E: Broken packages") {
		return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)
	}

	// Generic error with output
	if exitCode, ok := ExtractExitCode(err); ok && exitCode != 0 {
		// Extract meaningful error line from output
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "E:") {
				return fmt.Errorf("apt-get remove failed: %s", line)
			}
		}
		return fmt.Errorf("package uninstallation failed (exit code %d)", exitCode)
	}

	return fmt.Errorf("failed to execute uninstall command: %w", err)
}
