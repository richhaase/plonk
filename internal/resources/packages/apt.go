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
	// TODO: Implement in Phase 3
	return fmt.Errorf("not implemented")
}

// Uninstall removes a package using apt-get.
func (a *AptManager) Uninstall(ctx context.Context, name string) error {
	// TODO: Implement in Phase 3
	return fmt.Errorf("not implemented")
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
