// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// UvManager manages UV tool packages.
type UvManager struct {
	binary string
}

// NewUvManager creates a new UV manager.
func NewUvManager() *UvManager {
	return &UvManager{
		binary: "uv",
	}
}

// ListInstalled lists all installed UV tools.
func (u *UvManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, u.binary, "tool", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed tools: %w", err)
	}

	return u.parseListOutput(output), nil
}

// parseListOutput parses UV tool list output to extract package names
func (u *UvManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || result == "No tools installed" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string

	// UV tool list format: "package-name v1.2.3"
	// Extract package names from each line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		// Split on space and take first part (package name)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packageName := parts[0]
			if packageName != "" {
				packages = append(packages, packageName)
			}
		}
	}

	return packages
}

// Install installs a UV tool.
func (u *UvManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, u.binary, "tool", "install", name)
	if err != nil {
		return u.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a UV tool.
func (u *UvManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, u.binary, "tool", "uninstall", name)
	if err != nil {
		return u.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific tool is installed.
func (u *UvManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installed, err := u.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}

	for _, pkg := range installed {
		if pkg == name {
			return true, nil
		}
	}
	return false, nil
}

// Search is not supported by UV tools - returns empty results.
func (u *UvManager) Search(ctx context.Context, query string) ([]string, error) {
	// UV doesn't have a built-in search capability for tools
	return []string{}, nil
}

// Info retrieves information about a UV tool.
func (u *UvManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if tool is installed first
	installed, err := u.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "uv",
		Installed: installed,
	}

	if installed {
		// Get version from tool list for installed tools
		version, err := u.InstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}
	}

	// UV tools don't have detailed package info like description, homepage etc.
	// since they're meant for executable tools, not library packages
	info.Description = "UV tool package"

	return info, nil
}

// InstalledVersion retrieves the installed version of a UV tool
func (u *UvManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if tool is installed
	installed, err := u.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("tool '%s' is not installed", name)
	}

	// Get detailed tool list output
	output, err := ExecuteCommand(ctx, u.binary, "tool", "list")
	if err != nil {
		return "", fmt.Errorf("failed to get tool version information for %s: %w", name, err)
	}

	// Parse output to find version for specific tool
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, name+" ") {
			// Extract version using regex to handle "v1.2.3" format
			versionRegex := regexp.MustCompile(`v(\d+\.\d+(?:\.\d+)?(?:[^\s]*)?)`)
			matches := versionRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				return matches[1], nil
			}
		}
	}

	return "", fmt.Errorf("version information not found for tool '%s'", name)
}

func init() {
	RegisterManager("uv", func() PackageManager {
		return NewUvManager()
	})
}

// IsAvailable checks if uv is installed and accessible
func (u *UvManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(u.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, u.binary, []string{"--version"})
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

// SupportsSearch returns false as UV does not support tool search
func (u *UvManager) SupportsSearch() bool {
	return false
}

// handleInstallError processes install command errors
func (u *UvManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "No such package") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "404") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
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
				return fmt.Errorf("tool installation failed: %s", errorOutput)
			}
			return fmt.Errorf("tool installation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return err
}

// handleUninstallError processes uninstall command errors
func (u *UvManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "No tool named") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("tool uninstallation failed: %s", errorOutput)
			}
			return fmt.Errorf("tool uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}
