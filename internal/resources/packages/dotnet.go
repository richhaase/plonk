// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DotnetManager manages .NET Global Tools.
type DotnetManager struct {
	binary string
}

// NewDotnetManager creates a new .NET Global Tools manager.
func NewDotnetManager() *DotnetManager {
	return &DotnetManager{
		binary: "dotnet",
	}
}

// ListInstalled lists all globally installed .NET tools.
func (d *DotnetManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, d.binary, "tool", "list", "-g")
	if err != nil {
		// dotnet tool list can return non-zero exit codes when no tools are installed
		if exitCode, ok := ExtractExitCode(err); ok {
			// Only treat it as an error if the exit code indicates a real failure
			if exitCode > 1 {
				return nil, fmt.Errorf("dotnet tool list command failed with severe error: %w", err)
			}
			// Exit code 1 might just mean no tools installed - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, err
		}
	}

	return d.parseListOutput(output), nil
}

// parseListOutput parses dotnet tool list output to extract package names
func (d *DotnetManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string
	var inDataSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Look for the header separator (dashes)
		if strings.Contains(line, "-------") {
			inDataSection = true
			continue
		}

		// Skip header line
		if strings.Contains(line, "Package Id") && strings.Contains(line, "Version") {
			continue
		}

		// If we're in the data section, extract package names
		if inDataSection {
			// Line format: "package-name    version    commands"
			// Extract the first column (package name)
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				packageName := fields[0]
				if packageName != "" {
					packages = append(packages, packageName)
				}
			}
		}
	}

	// Sort packages for consistent output
	sort.Strings(packages)
	return packages
}

// Install installs a global .NET tool.
func (d *DotnetManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, d.binary, "tool", "install", "-g", name)
	if err != nil {
		return d.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a global .NET tool.
func (d *DotnetManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, d.binary, "tool", "uninstall", "-g", name)
	if err != nil {
		return d.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific tool is installed globally.
func (d *DotnetManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installedTools, err := d.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}

	// Check if the tool name is in the list of installed tools
	for _, tool := range installedTools {
		if tool == name {
			return true, nil
		}
	}

	return false, nil
}

// Search is not supported by .NET CLI - returns empty results.
// .NET tools are discovered through NuGet.org search instead.
func (d *DotnetManager) Search(ctx context.Context, query string) ([]string, error) {
	// .NET CLI doesn't have a built-in search command for tools
	// Tools are typically found through NuGet.org or documentation
	return []string{}, nil
}

// Info retrieves detailed information about a .NET tool.
func (d *DotnetManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if tool is installed first
	installed, err := d.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "dotnet",
		Installed: installed,
	}

	if installed {
		// Get version from installed tools list
		version, err := d.getInstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}
	}

	// For non-installed tools, we can't get info without additional NuGet API calls
	// which would add complexity. For now, return basic info.
	if !installed {
		info.Description = "Use 'dotnet tool install -g " + name + "' to install this .NET global tool"
	}

	return info, nil
}

// getInstalledVersion extracts the version of an installed tool
func (d *DotnetManager) getInstalledVersion(ctx context.Context, name string) (string, error) {
	output, err := ExecuteCommand(ctx, d.binary, "tool", "list", "-g")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	var inDataSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for the header separator
		if strings.Contains(line, "-------") {
			inDataSection = true
			continue
		}

		// Skip header line
		if strings.Contains(line, "Package Id") {
			continue
		}

		if inDataSection {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] == name {
				return fields[1], nil // Version is the second field
			}
		}
	}

	return "", fmt.Errorf("version not found for tool %s", name)
}

// InstalledVersion retrieves the installed version of a global .NET tool
func (d *DotnetManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if tool is installed
	installed, err := d.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check tool installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("tool '%s' is not installed globally", name)
	}

	return d.getInstalledVersion(ctx, name)
}

// SelfInstall attempts to install .NET via available package managers
func (d *DotnetManager) SelfInstall(ctx context.Context) error {
	// Check if already available (idempotent)
	if available, _ := d.IsAvailable(ctx); available {
		return nil
	}

	// Execute official Microsoft .NET installer script
	script := `curl -sSL https://dot.net/v1/dotnet-install.sh | bash`
	return executeInstallScript(ctx, script, ".NET SDK")
}

// Dependencies returns package managers this manager depends on for self-installation
func (d *DotnetManager) Dependencies() []string {
	return []string{} // .NET is independent - uses official Microsoft installer script
}

// Upgrade upgrades one or more packages to their latest versions
func (d *DotnetManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// First get all installed tools
		installed, err := d.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to list installed tools: %w", err)
		}

		// Upgrade each tool individually
		var upgradeErrors []string
		for _, tool := range installed {
			output, err := ExecuteCommandCombined(ctx, d.binary, "tool", "update", "-g", tool)
			if err != nil {
				upgradeErr := d.handleUpgradeError(err, output, tool)
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
	var upgradeErrors []string
	for _, pkg := range packages {
		output, err := ExecuteCommandCombined(ctx, d.binary, "tool", "update", "-g", pkg)
		if err != nil {
			upgradeErr := d.handleUpgradeError(err, output, pkg)
			upgradeErrors = append(upgradeErrors, upgradeErr.Error())
			continue
		}
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("failed to upgrade tools: %s", strings.Join(upgradeErrors, "; "))
	}
	return nil
}

func init() {
	RegisterManager("dotnet", func() PackageManager {
		return NewDotnetManager()
	})
}

// IsAvailable checks if dotnet CLI is installed and accessible
func (d *DotnetManager) IsAvailable(ctx context.Context) (bool, error) {
	// First check if dotnet is in PATH
	if CheckCommandAvailable(d.binary) {
		err := VerifyBinary(ctx, d.binary, []string{"--version"})
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

	// If not in PATH, check standard installation locations
	dotnetPaths := d.getStandardInstallPaths()
	for _, dotnetPath := range dotnetPaths {
		if _, err := os.Stat(dotnetPath); err == nil {
			// Found dotnet binary, verify it works
			err := VerifyBinary(ctx, dotnetPath, []string{"--version"})
			if err != nil {
				// Check for context cancellation
				if IsContextError(err) {
					return false, err
				}
				// This installation is not functional, try next
				continue
			}
			// Update binary path to use the found installation
			d.binary = dotnetPath
			return true, nil
		}
	}

	return false, nil
}

// CheckHealth performs a comprehensive health check of the .NET Global Tools installation
func (d *DotnetManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     ".NET Global Tools Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  ".NET SDK is available and properly configured",
	}

	// Check basic availability first
	available, err := d.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = ".NET SDK availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking .NET SDK: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "fail"
		check.Message = ".NET SDK is required but not available"
		check.Issues = []string{".NET SDK is required for managing global tools"}
		check.Suggestions = []string{
			"Install .NET SDK: https://dotnet.microsoft.com/download",
			"Or via Homebrew: brew install --cask dotnet",
			"After installation, ensure dotnet is in your PATH",
		}
		return check, nil
	}

	// Get .NET global tools directory (predictable path)
	binDir := d.getBinDirectory()

	// Check if bin directory is in PATH
	pathCheck := checkDirectoryInPath(binDir)
	check.Details = append(check.Details, fmt.Sprintf(".NET global tools directory: %s", binDir))

	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = ".NET global tools directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Status = "warn"
		check.Message = ".NET global tools directory does not exist"
		check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
	} else {
		check.Details = append(check.Details, ".NET global tools directory is in PATH")
	}

	return check, nil
}

// getBinDirectory returns the predictable .NET global tools directory
func (d *DotnetManager) getBinDirectory() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".dotnet", "tools")
}

// getStandardInstallPaths returns standard locations where .NET might be installed
func (d *DotnetManager) getStandardInstallPaths() []string {
	homeDir, _ := os.UserHomeDir()
	return []string{
		filepath.Join(homeDir, ".dotnet", "dotnet"),     // User installation (Linux/macOS)
		"/usr/local/share/dotnet/dotnet",                // System-wide installation (macOS)
		"/usr/share/dotnet/dotnet",                      // System-wide installation (Linux)
		"C:\\Program Files\\dotnet\\dotnet.exe",         // System-wide installation (Windows)
		filepath.Join(homeDir, ".dotnet", "dotnet.exe"), // User installation (Windows)
	}
}

// handleUpgradeError processes upgrade command errors
func (d *DotnetManager) handleUpgradeError(err error, output []byte, toolName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "is not installed") ||
			strings.Contains(outputStr, "No such tool") ||
			strings.Contains(outputStr, "Tool") && strings.Contains(outputStr, "is not installed") {
			return fmt.Errorf("tool '%s' not found or not installed", toolName)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "no updates available") ||
			strings.Contains(outputStr, "is already up to date") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "Access denied") ||
			strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied upgrading %s", toolName)
		}
		if strings.Contains(outputStr, "Unable to find package") ||
			strings.Contains(outputStr, "NU1101") {
			return fmt.Errorf("tool '%s' not found in registry", toolName)
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
func (d *DotnetManager) handleInstallError(err error, output []byte, toolName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "Unable to find package") ||
			strings.Contains(outputStr, "NU1101") ||
			strings.Contains(outputStr, "No packages exist with this id") {
			return fmt.Errorf("tool '%s' not found", toolName)
		}
		if strings.Contains(outputStr, "Invalid project-package combination") ||
			strings.Contains(outputStr, "NU1212") ||
			strings.Contains(outputStr, "DotnetToolReference project style can only contain references of the DotnetTool type") {
			return fmt.Errorf("package '%s' is not a .NET global tool", toolName)
		}
		if strings.Contains(outputStr, "already installed") {
			return fmt.Errorf("tool '%s' is already installed", toolName)
		}
		if strings.Contains(outputStr, "Access denied") ||
			strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied installing %s", toolName)
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

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}

// handleUninstallError processes uninstall command errors
func (d *DotnetManager) handleUninstallError(err error, output []byte, toolName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "is not installed") ||
			strings.Contains(outputStr, "No such tool") ||
			strings.Contains(outputStr, "Tool") && strings.Contains(outputStr, "is not installed") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "Access denied") ||
			strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", toolName)
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
