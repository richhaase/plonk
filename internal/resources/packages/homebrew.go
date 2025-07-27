// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct {
	binary string
}

// NewHomebrewManager creates a new homebrew manager.
func NewHomebrewManager() *HomebrewManager {
	return &HomebrewManager{
		binary: "brew",
	}
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, h.binary, "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	return SplitLines(output), nil
}

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, h.binary, "install", name)
	if err != nil {
		return h.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, h.binary, "uninstall", name)
	if err != nil {
		return h.handleUninstallError(err, output, name)
	}
	return nil
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
	cmd := exec.CommandContext(ctx, h.binary, "search", query)
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
	cmd := exec.CommandContext(ctx, h.binary, "info", name)
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
	info.Manager = "brew"
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

// InstalledVersion retrieves the installed version of a package
func (h *HomebrewManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := h.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Get version using brew list --versions
	cmd := exec.CommandContext(ctx, h.binary, "list", "--versions", name)
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

// IsAvailable checks if homebrew is installed and accessible
func (h *HomebrewManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(h.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, h.binary, []string{"--version"})
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

// SupportsSearch returns true as Homebrew supports package search
func (h *HomebrewManager) SupportsSearch() bool {
	return true
}

// handleInstallError processes install command errors
func (h *HomebrewManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no available formula") || strings.Contains(outputStr, "no formulae found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already installed") {
			return nil // Already installed is success
		}
		if strings.Contains(outputStr, "permission denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}
		if strings.Contains(outputStr, "because it is required by") || strings.Contains(outputStr, "still has dependents") {
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)
		}

		if exitCode != 0 {
			return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors
func (h *HomebrewManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no such keg") || strings.Contains(outputStr, "not installed") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}
		if strings.Contains(outputStr, "because it is required by") || strings.Contains(outputStr, "still has dependents") {
			return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)
		}

		if exitCode != 0 {
			return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute uninstall command: %w", err)
}
