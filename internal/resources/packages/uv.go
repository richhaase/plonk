// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UvManager manages UV tool packages.
type UvManager struct {
	binary string
	exec   CommandExecutor
}

// NewUvManager creates a new UV manager with default executor.
func NewUvManager() *UvManager {
	return NewUvManagerWithExecutor(nil)
}

// NewUvManagerWithExecutor creates a new UV manager with the provided executor.
// If executor is nil, uses the default executor.
func NewUvManagerWithExecutor(executor CommandExecutor) *UvManager {
	if executor == nil {
		executor = defaultExecutor
	}
	return &UvManager{
		binary: "uv",
		exec:   executor,
	}
}

// ListInstalled lists all installed UV tools.
func (u *UvManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteWith(ctx, u.exec, u.binary, "tool", "list")
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
	output, err := CombinedOutputWith(ctx, u.exec, u.binary, "tool", "install", name)
	if err != nil {
		return u.handleInstallError(err, []byte(output), name)
	}
	return nil
}

// Uninstall removes a UV tool.
func (u *UvManager) Uninstall(ctx context.Context, name string) error {
	output, err := CombinedOutputWith(ctx, u.exec, u.binary, "tool", "uninstall", name)
	if err != nil {
		return u.handleUninstallError(err, []byte(output), name)
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
	output, err := ExecuteWith(ctx, u.exec, u.binary, "tool", "list")
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

// Upgrade upgrades one or more packages to their latest versions
func (u *UvManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// First get all installed tools
		installed, err := u.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to list installed tools: %w", err)
		}

		// Upgrade each tool individually
		var upgradeErrors []string
		for _, tool := range installed {
			output, err := CombinedOutputWith(ctx, u.exec, u.binary, "tool", "upgrade", tool)
			if err != nil {
				upgradeErr := u.handleUpgradeError(err, []byte(output), tool)
				upgradeErrors = append(upgradeErrors, upgradeErr.Error())
				continue
			}
		}

		if len(upgradeErrors) > 0 {
			return fmt.Errorf("some tools failed to upgrade: %s", strings.Join(upgradeErrors, "; "))
		}
		return nil
	}

	// Upgrade specific packages
	args := append([]string{"tool", "upgrade"}, packages...)
	output, err := CombinedOutputWith(ctx, u.exec, u.binary, args...)
	if err != nil {
		return u.handleUpgradeError(err, []byte(output), strings.Join(packages, ", "))
	}
	return nil
}

// Dependencies returns package managers this manager depends on for self-installation
func (u *UvManager) Dependencies() []string {
	return []string{} // UV is independent - uses official installer script
}

func init() {
	RegisterManager("uv", func() PackageManager {
		return NewUvManager()
	})
}

// IsAvailable checks if uv is installed and accessible
func (u *UvManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailableWith(u.exec, u.binary) {
		return false, nil
	}

	err := VerifyBinaryWith(ctx, u.exec, u.binary, []string{"--version"})
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

// CheckHealth performs a comprehensive health check of the UV tool installation
func (u *UvManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "UV Tool Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "UV is available and properly configured",
	}

	// Check basic availability first
	available, err := u.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = "UV availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking UV: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "UV is not available"
		check.Issues = []string{"UV command not found or not functional"}
		check.Suggestions = []string{
			"Install UV: curl -LsSf https://astral.sh/uv/install.sh | sh",
			"Or via pipx: pipx install uv",
			"After installation, ensure uv is in your PATH",
		}
		return check, nil
	}

	// Discover UV tool directory dynamically
	binDir, err := u.getBinDirectory(ctx)
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not determine UV tool directory"
		check.Issues = []string{fmt.Sprintf("Error discovering tool directory: %v", err)}
		return check, nil
	}

	// Check if bin directory is in PATH
	pathCheck := checkDirectoryInPath(binDir)
	check.Details = append(check.Details, fmt.Sprintf("UV tool directory: %s", binDir))

	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = "UV tool directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Status = "warn"
		check.Message = "UV tool directory does not exist"
		check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
	} else {
		check.Details = append(check.Details, "UV tool directory is in PATH")
	}

	return check, nil
}

// getBinDirectory returns the UV tool bin directory where executables are symlinked
func (u *UvManager) getBinDirectory(ctx context.Context) (string, error) {
	// UV tools are installed in ~/.local/share/uv/tools/<tool-name>/
	// but executables are symlinked to ~/.local/bin/<executable-name>
	// This is the directory that should be in PATH for UV tools to work
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "bin"), nil
}

// handleUpgradeError processes upgrade command errors
func (u *UvManager) handleUpgradeError(err error, output []byte, packages string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "No such tool") ||
			strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "not found") {
			return fmt.Errorf("one or more tools not found: %s", packages)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "Nothing to upgrade") ||
			strings.Contains(outputStr, "up to date") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied upgrading %s", packages)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("tool upgrade failed: %s", errorOutput)
			}
			return fmt.Errorf("tool upgrade failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute upgrade command: %w", err)
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
