// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// PipxManager manages pipx packages.
type PipxManager struct {
	binary string
}

// NewPipxManager creates a new pipx manager.
func NewPipxManager() *PipxManager {
	return &PipxManager{
		binary: "pipx",
	}
}

// ListInstalled lists all installed pipx packages.
func (p *PipxManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "list", "--short")
	if err != nil {
		// pipx list can return non-zero exit codes when no packages are installed
		if exitCode, ok := ExtractExitCode(err); ok {
			// Only treat it as an error if the exit code indicates a real failure
			if exitCode > 1 {
				return nil, fmt.Errorf("pipx list command failed with severe error: %w", err)
			}
			// Exit code 1 might just mean no packages installed - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, err
		}
	}

	return p.parseListOutput(output), nil
}

// parseListOutput parses pipx list output to extract package names
func (p *PipxManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string

	// pipx list --short format: one package per line
	// Format: "package-name 1.2.3"
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract package name (first field)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packageName := parts[0]
			if packageName != "" {
				packages = append(packages, packageName)
			}
		}
	}

	// Sort packages for consistent output
	sort.Strings(packages)
	return packages
}

// Install installs a pipx package.
func (p *PipxManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "install", name)
	if err != nil {
		return p.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a pipx package.
func (p *PipxManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "uninstall", name)
	if err != nil {
		return p.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (p *PipxManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installedPackages, err := p.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	// Check if the package name is in the list of installed packages
	for _, pkg := range installedPackages {
		if pkg == name {
			return true, nil
		}
	}

	return false, nil
}

// Search searches for packages in the Python Package Index.
func (p *PipxManager) Search(ctx context.Context, query string) ([]string, error) {
	// pipx doesn't have a built-in search command
	// Users typically search PyPI directly or use pip search alternatives
	return []string{}, nil
}

// Info retrieves detailed information about a pipx package.
func (p *PipxManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "pipx",
		Installed: installed,
	}

	if installed {
		// Get version from list output for installed packages
		version, err := p.getInstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}

		// Get detailed info using pipx list (without --short)
		output, err := ExecuteCommand(ctx, p.binary, "list")
		if err == nil {
			p.parseInfoFromListOutput(string(output), name, info)
		}
	}

	// For non-installed packages, provide basic info
	if !installed {
		info.Description = "Use 'pipx install " + name + "' to install this Python application"
	}

	return info, nil
}

// parseInfoFromListOutput extracts additional info from pipx list output
func (p *PipxManager) parseInfoFromListOutput(output, packageName string, info *PackageInfo) {
	lines := strings.Split(output, "\n")
	var inPackageSection bool

	for _, line := range lines {
		// Look for package section
		if strings.Contains(line, packageName) && strings.Contains(line, "package") {
			inPackageSection = true
			continue
		}

		// Exit package section when we hit another package or empty line
		if inPackageSection && (strings.Contains(line, "package") || strings.TrimSpace(line) == "") {
			if !strings.Contains(line, packageName) {
				break
			}
		}

		// Extract info from package section
		if inPackageSection {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- ") {
				// Extract executable names
				executable := strings.TrimPrefix(line, "- ")
				if executable != "" {
					info.Description = fmt.Sprintf("Python application with executable: %s", executable)
				}
			}
		}
	}
}

// getInstalledVersion extracts the version of an installed package
func (p *PipxManager) getInstalledVersion(ctx context.Context, name string) (string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "list", "--short")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == name {
			return parts[1], nil // Version is the second field
		}
	}

	return "", fmt.Errorf("version not found for package %s", name)
}

// InstalledVersion retrieves the installed version of a pipx package
func (p *PipxManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	return p.getInstalledVersion(ctx, name)
}

// SelfInstall attempts to install pipx via available package managers
func (p *PipxManager) SelfInstall(ctx context.Context) error {
	// Check if already available (idempotent)
	if available, _ := p.IsAvailable(ctx); available {
		return nil
	}

	// Install pipx via Homebrew (canonical method)
	if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
		return p.installViaHomebrew(ctx)
	}

	return fmt.Errorf("pipx installation requires Homebrew - install Homebrew first from https://brew.sh")
}

// installViaHomebrew installs pipx via Homebrew
func (p *PipxManager) installViaHomebrew(ctx context.Context) error {
	return executeInstallCommand(ctx, "brew", []string{"install", "pipx"}, "pipx")
}

// Upgrade upgrades one or more packages to their latest versions
func (p *PipxManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// Upgrade all packages using pipx upgrade-all
		output, err := ExecuteCommandCombined(ctx, p.binary, "upgrade-all")
		if err != nil {
			return p.handleUpgradeError(err, output, "all packages")
		}
		return nil
	}

	// Upgrade specific packages
	var upgradeErrors []string
	for _, pkg := range packages {
		output, err := ExecuteCommandCombined(ctx, p.binary, "upgrade", pkg)
		if err != nil {
			upgradeErr := p.handleUpgradeError(err, output, pkg)
			upgradeErrors = append(upgradeErrors, upgradeErr.Error())
			continue
		}
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("failed to upgrade packages: %s", strings.Join(upgradeErrors, "; "))
	}
	return nil
}

// Dependencies returns package managers this manager depends on for self-installation
func (p *PipxManager) Dependencies() []string {
	return []string{"brew"} // pipx requires brew to install pipx
}

func init() {
	RegisterManager("pipx", func() PackageManager {
		return NewPipxManager()
	})
}

// IsAvailable checks if pipx is installed and accessible
func (p *PipxManager) IsAvailable(ctx context.Context) (bool, error) {
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

// CheckHealth performs a comprehensive health check of the pipx installation
func (p *PipxManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "Pipx Package Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "pipx is available and properly configured",
	}

	// Check basic availability first
	available, err := p.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = "pipx availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking pipx: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "pipx is not available"
		check.Issues = []string{"pipx command not found or not functional"}
		check.Suggestions = []string{
			"Install pipx via pip: pip3 install --user pipx",
			"Or via Homebrew: brew install pipx",
			"After installation, ensure pipx is in your PATH",
		}
		return check, nil
	}

	// Discover pipx binary directory dynamically
	binDir, err := p.getBinDirectory(ctx)
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not determine pipx binary directory"
		check.Issues = []string{fmt.Sprintf("Error discovering binary directory: %v", err)}
		return check, nil
	}

	// Check if bin directory is in PATH
	pathCheck := checkDirectoryInPath(binDir)
	check.Details = append(check.Details, fmt.Sprintf("pipx binary directory: %s", binDir))

	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = "pipx binary directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Status = "warn"
		check.Message = "pipx binary directory does not exist"
		check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
	} else {
		check.Details = append(check.Details, "pipx binary directory is in PATH")
	}

	return check, nil
}

// getBinDirectory discovers the pipx binary directory using pipx environment
func (p *PipxManager) getBinDirectory(ctx context.Context) (string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "environment")
	if err != nil {
		return "", fmt.Errorf("failed to get pipx environment: %w", err)
	}

	// Parse pipx environment output to find PIPX_BIN_DIR
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PIPX_BIN_DIR=") {
			binDir := strings.TrimPrefix(line, "PIPX_BIN_DIR=")
			binDir = strings.Trim(binDir, "\"'")
			// Skip empty values (user-set environment variables are often empty)
			// and return the first non-empty value (derived computed value)
			if binDir != "" {
				return binDir, nil
			}
		}
	}

	// Fallback to default location
	homeDir, _ := filepath.Abs("~")
	return filepath.Join(homeDir, ".local", "bin"), nil
}

// handleUpgradeError processes upgrade command errors
func (p *PipxManager) handleUpgradeError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "No such package") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "No apps associated with package") {
			return fmt.Errorf("package '%s' not found or not installed", packageName)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "up to date") ||
			strings.Contains(outputStr, "Nothing to upgrade") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied upgrading %s", packageName)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("package upgrade failed: %s", errorOutput)
			}
			return fmt.Errorf("package upgrade failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute upgrade command: %w", err)
}

// handleInstallError processes install command errors
func (p *PipxManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "No such package") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "404") ||
			strings.Contains(outputStr, "Could not find") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already installed") ||
			strings.Contains(outputStr, "already exists") {
			return fmt.Errorf("package '%s' is already installed", packageName)
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
				return fmt.Errorf("package installation failed: %s", errorOutput)
			}
			return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}

// handleUninstallError processes uninstall command errors
func (p *PipxManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "No such package") ||
			strings.Contains(outputStr, "No apps associated with package") {
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
				return fmt.Errorf("package uninstallation failed: %s", errorOutput)
			}
			return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}
