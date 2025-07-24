// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// HomebrewManager manages Homebrew packages using BaseManager for common functionality.
type HomebrewManager struct {
	*BaseManager
}

// NewHomebrewManager creates a new homebrew manager.
func NewHomebrewManager() *HomebrewManager {
	return newHomebrewManager()
}

// newHomebrewManager creates a homebrew manager.
func newHomebrewManager() *HomebrewManager {
	config := ManagerConfig{
		BinaryName:  "brew",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"list"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", pkg}
		},
	}

	// Add homebrew-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "No available formula", "No formulae found")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "already installed")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "No such keg", "not installed")
	errorMatcher.AddPattern(ErrorTypeDependency, "because it is required by", "still has dependents")

	base := NewBaseManager(config)
	base.ErrorMatcher = errorMatcher

	return &HomebrewManager{
		BaseManager: base,
	}
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := h.ExecuteList(ctx)
	if err != nil {
		return nil, err
	}

	return h.parseListOutput(output), nil
}

// parseListOutput parses brew list output
func (h *HomebrewManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
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

	return filteredPackages
}

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
	return h.ExecuteInstall(ctx, name)
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error {
	return h.ExecuteUninstall(ctx, name)
}

// IsInstalled checks if a specific package is installed.
func (h *HomebrewManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	packages, err := h.ListInstalled(ctx)
	if err != nil {
		return false, err
	}

	for _, pkg := range packages {
		if pkg == name {
			return true, nil
		}
	}

	return false, nil
}

// Search searches for packages in Homebrew repositories.
func (h *HomebrewManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, h.GetBinary(), "search", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search homebrew packages for %s: %w", query, err)
	}

	return h.parseSearchOutput(output), nil
}

// parseSearchOutput parses brew search output
func (h *HomebrewManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || strings.Contains(result, "No formula found") {
		return []string{}
	}

	// Parse output into package list
	packages := strings.Split(result, "\n")
	var filteredPackages []string
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			// Skip header lines and empty lines
			if !strings.HasPrefix(trimmed, "=") && !strings.HasPrefix(trimmed, "If you meant") {
				filteredPackages = append(filteredPackages, trimmed)
			}
		}
	}

	return filteredPackages
}

// Info retrieves detailed information about a package.
func (h *HomebrewManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, h.GetBinary(), "info", name)
	output, err := cmd.Output()
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
			return nil, fmt.Errorf("package '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
	}

	// Check if installed
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	info := h.parseInfoOutput(output, name)
	info.Manager = "homebrew"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses brew info output
func (h *HomebrewManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{
		Name: name,
	}

	lines := strings.Split(string(output), "\n")
	versionFound := false

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// First line usually contains name and version
		if strings.HasPrefix(line, name) && strings.Contains(line, ":") {
			// Extract version from lines like "package: stable 1.2.3"
			parts := strings.Fields(line)
			for j, part := range parts {
				if part == "stable" && j+1 < len(parts) {
					info.Version = parts[j+1]
					versionFound = true
					break
				}
			}

			// Check for description on the same line after version
			if colonIndex := strings.Index(line, ":"); colonIndex > 0 && colonIndex < len(line)-1 {
				desc := strings.TrimSpace(line[colonIndex+1:])
				// Remove version info if present
				if versionFound && strings.Contains(desc, "stable") {
					// Look for description after stable version
					stableIndex := strings.Index(desc, "stable")
					if stableIndex >= 0 {
						remainingDesc := desc[stableIndex:]
						parts := strings.Fields(remainingDesc)
						if len(parts) > 2 { // "stable", "version", "description..."
							info.Description = strings.TrimSpace(strings.Join(parts[2:], " "))
						}
					}
				} else if desc != "" && !strings.Contains(desc, "stable") {
					info.Description = desc
				}
			}
		}

		// Look for description on the next line if not found yet
		if info.Description == "" && i > 0 && versionFound {
			if line != "" && !strings.HasPrefix(line, "From:") && !strings.HasPrefix(line, "URL:") &&
				!strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "=") {
				info.Description = line
			}
		}

		// Look for homepage
		if strings.HasPrefix(line, "From:") || strings.HasPrefix(line, "URL:") {
			// Extract URL from lines like "From: https://..."
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "http") {
					info.Homepage = part
					break
				}
			}
		}
	}

	return info
}

// GetInstalledVersion retrieves the installed version of a package
func (h *HomebrewManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Get version using brew list --versions
	cmd := exec.CommandContext(ctx, h.GetBinary(), "list", "--versions", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	version := h.extractVersion(output, name)
	if version == "" {
		return "", fmt.Errorf("could not extract version for package '%s'", name)
	}

	return version, nil
}

// extractVersion extracts version from brew list --versions output
func (h *HomebrewManager) extractVersion(output []byte, name string) string {
	result := strings.TrimSpace(string(output))
	// Output format: "package 1.2.3 1.2.2" (latest version first)
	if strings.HasPrefix(result, name+" ") {
		parts := strings.Fields(result)
		if len(parts) >= 2 {
			return parts[1] // Return the first version (latest)
		}
	}
	return ""
}
