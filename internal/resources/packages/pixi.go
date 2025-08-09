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

// PixiManager manages pixi global packages.
type PixiManager struct {
	binary string
}

// NewPixiManager creates a new pixi manager.
func NewPixiManager() *PixiManager {
	return &PixiManager{
		binary: "pixi",
	}
}

// ListInstalled lists all installed pixi global environments.
func (p *PixiManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "global", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed environments: %w", err)
	}

	return p.parseListOutput(output), nil
}

// parseListOutput parses pixi global list output to extract environment names
func (p *PixiManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || result == "No global environments found." {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var environments []string

	// Pixi global list format: "└── environment-name: version"
	// Extract environment names from each line that starts with └──
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "└── ") {
			// Extract environment name before the colon
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				envName := strings.TrimSpace(strings.TrimPrefix(parts[0], "└── "))
				if envName != "" {
					environments = append(environments, envName)
				}
			}
		}
	}

	return environments
}

// Install installs a pixi global package.
func (p *PixiManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "global", "install", name)
	if err != nil {
		return p.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a pixi global environment.
func (p *PixiManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "global", "uninstall", name)
	if err != nil {
		return p.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific environment is installed.
func (p *PixiManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installed, err := p.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check environment installation status for %s: %w", name, err)
	}

	for _, env := range installed {
		if env == name {
			return true, nil
		}
	}
	return false, nil
}

// Search searches for packages in conda-forge.
func (p *PixiManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "search", query)
	if err != nil {
		return nil, fmt.Errorf("failed to search for packages: %w", err)
	}

	return p.parseSearchOutput(output), nil
}

// parseSearchOutput parses pixi search output to extract package names
func (p *PixiManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string

	// Look for package names in format "package-name-version-build"
	// The first line usually contains the main result
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "Name") {
			continue
		}

		// Extract package name from lines like "package-name-version-build"
		if strings.Contains(line, "-") && !strings.Contains(line, "=") {
			// Split by spaces and take first part which contains package-version-build
			parts := strings.Fields(line)
			if len(parts) > 0 {
				packageInfo := parts[0]
				// Extract just the package name (everything before the first version number)
				packageName := p.extractPackageName(packageInfo)
				if packageName != "" {
					packages = append(packages, packageName)
					break // Usually we just want the main result
				}
			}
		}
	}

	return packages
}

// extractPackageName extracts the package name from a conda package string
func (p *PixiManager) extractPackageName(packageInfo string) string {
	// Match pattern: package-name followed by version (digits)
	// Example: "ripgrep-14.1.1-h0ef69ab_1" -> "ripgrep"
	re := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]*?)-\d+`)
	matches := re.FindStringSubmatch(packageInfo)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: just return the first part if no version pattern found
	parts := strings.Split(packageInfo, "-")
	if len(parts) > 0 {
		return parts[0]
	}

	return packageInfo
}

// Info retrieves information about a pixi environment or package.
func (p *PixiManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if environment is installed first
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check environment installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "pixi",
		Installed: installed,
	}

	if installed {
		// Get version from environment details
		version, err := p.InstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}
	}

	// Try to get additional info via search if available
	searchResults, err := p.Search(ctx, name)
	if err == nil && len(searchResults) > 0 {
		info.Description = "Conda-forge package managed by pixi"
	} else {
		info.Description = "Pixi global environment"
	}

	return info, nil
}

// InstalledVersion retrieves the installed version of a pixi environment
func (p *PixiManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if environment is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check environment installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("environment '%s' is not installed", name)
	}

	// Get detailed environment info
	output, err := ExecuteCommand(ctx, p.binary, "global", "list", "--environment", name)
	if err != nil {
		return "", fmt.Errorf("failed to get environment version information for %s: %w", name, err)
	}

	// Parse output to find version for the main package
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines in format: "package-name   version   build   size"
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == name {
			return fields[1], nil
		}
	}

	// Fallback: try to get version from global list output
	globalOutput, err := ExecuteCommand(ctx, p.binary, "global", "list")
	if err != nil {
		return "", fmt.Errorf("failed to get global environment list for %s: %w", name, err)
	}

	globalLines := strings.Split(string(globalOutput), "\n")
	for _, line := range globalLines {
		if strings.Contains(line, "└── "+name+":") {
			// Extract version from "└── name: version"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				version := strings.TrimSpace(parts[1])
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version information not found for environment '%s'", name)
}

// Upgrade upgrades one or more packages to their latest versions
func (p *PixiManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// First get all installed environments
		installed, err := p.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to list installed environments: %w", err)
		}

		// Upgrade each environment individually
		var upgradeErrors []string
		for _, env := range installed {
			output, err := ExecuteCommandCombined(ctx, p.binary, "global", "update", env)
			if err != nil {
				upgradeErr := p.handleUpgradeError(err, output, env)
				upgradeErrors = append(upgradeErrors, upgradeErr.Error())
				continue
			}
		}

		if len(upgradeErrors) > 0 {
			return fmt.Errorf("some environments failed to upgrade: %s", strings.Join(upgradeErrors, "; "))
		}
		return nil
	}

	// Upgrade specific packages
	args := append([]string{"global", "update"}, packages...)
	output, err := ExecuteCommandCombined(ctx, p.binary, args...)
	if err != nil {
		return p.handleUpgradeError(err, output, strings.Join(packages, ", "))
	}
	return nil
}

// SelfInstall installs Pixi using official installer
func (p *PixiManager) SelfInstall(ctx context.Context) error {
	// Check if already available (idempotent)
	if available, _ := p.IsAvailable(ctx); available {
		return nil
	}

	// Execute official Pixi installer script
	script := `curl -fsSL https://pixi.sh/install.sh | sh`
	return executeInstallScript(ctx, script, "Pixi")
}

// Dependencies returns package managers this manager depends on for self-installation
func (p *PixiManager) Dependencies() []string {
	return []string{} // Pixi is independent - uses official installer script
}

func init() {
	RegisterManager("pixi", func() PackageManager {
		return NewPixiManager()
	})
}

// IsAvailable checks if pixi is installed and accessible
func (p *PixiManager) IsAvailable(ctx context.Context) (bool, error) {
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

// CheckHealth performs a comprehensive health check of the Pixi installation
func (p *PixiManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "Pixi Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Pixi is available and properly configured",
	}

	// Check basic availability first
	available, err := p.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = "Pixi availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking Pixi: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "Pixi is not available"
		check.Issues = []string{"Pixi command not found or not functional"}
		check.Suggestions = []string{
			"Install Pixi: curl -fsSL https://pixi.sh/install.sh | bash",
			"Or via Homebrew: brew install pixi",
			"After installation, ensure pixi is in your PATH",
		}
		return check, nil
	}

	// Discover Pixi global bin directory dynamically
	binDir, err := p.getBinDirectory(ctx)
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not determine Pixi global bin directory"
		check.Issues = []string{fmt.Sprintf("Error discovering global bin directory: %v", err)}
		return check, nil
	}

	// Check if bin directory is in PATH
	pathCheck := checkDirectoryInPath(binDir)
	check.Details = append(check.Details, fmt.Sprintf("Pixi global bin directory: %s", binDir))

	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = "Pixi global bin directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Status = "warn"
		check.Message = "Pixi global bin directory does not exist"
		check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
	} else {
		check.Details = append(check.Details, "Pixi global bin directory is in PATH")
	}

	return check, nil
}

// getBinDirectory returns the Pixi global bin directory using PIXI_HOME or default location
func (p *PixiManager) getBinDirectory(ctx context.Context) (string, error) {
	// First check if PIXI_HOME is set
	pixiHome := os.Getenv("PIXI_HOME")
	if pixiHome == "" {
		// Default to ~/.pixi
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		pixiHome = filepath.Join(homeDir, ".pixi")
	}

	binDir := filepath.Join(pixiHome, "bin")
	return binDir, nil
}

// handleUpgradeError processes upgrade command errors
func (p *PixiManager) handleUpgradeError(err error, output []byte, packages string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "No environment named") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "does not exist") {
			return fmt.Errorf("one or more environments not found: %s", packages)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "Nothing to update") ||
			strings.Contains(outputStr, "up to date") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied upgrading %s", packages)
		}
		if strings.Contains(outputStr, "failed to solve the environment") {
			return fmt.Errorf("failed to resolve dependencies for packages: %s", packages)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("environment upgrade failed: %s", errorOutput)
			}
			return fmt.Errorf("environment upgrade failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	return fmt.Errorf("failed to execute upgrade command: %w", err)
}

// handleInstallError processes install command errors
func (p *PixiManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "No candidates were found") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "Cannot solve the request") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}
		if strings.Contains(outputStr, "failed to solve the environment") {
			return fmt.Errorf("failed to resolve dependencies for package '%s'", packageName)
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

	return err
}

// handleUninstallError processes uninstall command errors
func (p *PixiManager) handleUninstallError(err error, output []byte, environmentName string) error {
	outputStr := string(output)

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "does not exist") ||
			strings.Contains(outputStr, "No environment named") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("permission denied uninstalling %s", environmentName)
		}

		if exitCode != 0 {
			// Include command output for better error messages
			if len(output) > 0 {
				// Trim the output and limit length for readability
				errorOutput := strings.TrimSpace(string(output))
				if len(errorOutput) > 500 {
					errorOutput = errorOutput[:500] + "..."
				}
				return fmt.Errorf("environment uninstallation failed: %s", errorOutput)
			}
			return fmt.Errorf("environment uninstallation failed (exit code %d): %w", exitCode, err)
		}
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}
