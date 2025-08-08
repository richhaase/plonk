// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PnpmManager manages pnpm packages.
type PnpmManager struct {
	binary string
}

// NewPnpmManager creates a new pnpm manager.
func NewPnpmManager() *PnpmManager {
	return &PnpmManager{
		binary: "pnpm",
	}
}

// PnpmListOutput represents the structure of pnpm list -g --json output
type PnpmListOutput struct {
	Dependencies map[string]struct {
		Version string `json:"version"`
		Path    string `json:"path,omitempty"`
	} `json:"dependencies"`
}

// PnpmViewOutput represents the structure of pnpm view output
type PnpmViewOutput struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	Dependencies map[string]string `json:"dependencies"`
}

// IsAvailable checks if pnpm is installed and accessible
func (p *PnpmManager) IsAvailable(ctx context.Context) (bool, error) {
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

// ListInstalled lists all globally installed pnpm packages.
func (p *PnpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, p.binary, "list", "-g", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	var listOutput PnpmListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm list output: %w", err)
	}

	var packages []string
	for pkg := range listOutput.Dependencies {
		packages = append(packages, pkg)
	}

	sort.Strings(packages)
	return packages, nil
}

// Install installs a pnpm package globally.
func (p *PnpmManager) Install(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "add", "-g", name)
	if err != nil {
		return p.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a pnpm package.
func (p *PnpmManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, p.binary, "remove", "-g", name)
	if err != nil {
		return p.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (p *PnpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installed, err := p.ListInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	for _, pkg := range installed {
		if pkg == name {
			return true, nil
		}
	}
	return false, nil
}

// InstalledVersion retrieves the installed version of a package
func (p *PnpmManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Get detailed package list
	output, err := ExecuteCommand(ctx, p.binary, "list", "-g", "--json")
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	var listOutput PnpmListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return "", fmt.Errorf("failed to parse pnpm list output: %w", err)
	}

	if pkg, exists := listOutput.Dependencies[name]; exists {
		return pkg.Version, nil
	}

	return "", fmt.Errorf("version information not found for package '%s'", name)
}

// Search is not supported by pnpm - returns empty results.
func (p *PnpmManager) Search(ctx context.Context, query string) ([]string, error) {
	// pnpm doesn't have a built-in search command
	return []string{}, nil
}

// Info retrieves detailed information about a package.
func (p *PnpmManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed
	installed, err := p.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}

	info := &PackageInfo{
		Name:      name,
		Manager:   "pnpm",
		Installed: installed,
	}

	if installed {
		// Get version from installed packages
		version, err := p.InstalledVersion(ctx, name)
		if err == nil {
			info.Version = version
		}
	}

	// Get package info from registry
	output, err := ExecuteCommand(ctx, p.binary, "view", name, "--json")
	if err == nil {
		var viewOutput PnpmViewOutput
		if json.Unmarshal(output, &viewOutput) == nil {
			info.Description = viewOutput.Description
			info.Homepage = viewOutput.Homepage
			if !installed && viewOutput.Version != "" {
				info.Version = viewOutput.Version
			}
			// Convert dependencies map to slice
			for dep := range viewOutput.Dependencies {
				info.Dependencies = append(info.Dependencies, dep)
			}
		}
	}

	return info, nil
}

// CheckHealth performs a comprehensive health check of the pnpm installation
func (p *PnpmManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "PNPM Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "PNPM is available and properly configured",
	}

	// Check availability
	available, err := p.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "warn"
		check.Message = "PNPM availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking pnpm: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "PNPM is not available"
		check.Issues = []string{"pnpm command not found"}
		check.Suggestions = []string{
			"Install pnpm via standalone script: curl -fsSL https://get.pnpm.io/install.sh | sh",
			"See https://pnpm.io/installation for manual installation options",
		}
		return check, nil
	}

	// Discover global directory
	globalDir, err := p.getGlobalDirectory(ctx)
	if err == nil {
		check.Details = append(check.Details, fmt.Sprintf("PNPM global directory: %s", globalDir))
	}

	return check, nil
}

// getGlobalDirectory discovers the pnpm global directory
func (p *PnpmManager) getGlobalDirectory(ctx context.Context) (string, error) {
	// Get pnpm global directory
	output, err := ExecuteCommand(ctx, p.binary, "root", "-g")
	if err != nil {
		return "", fmt.Errorf("failed to get pnpm global directory: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SelfInstall installs pnpm using the single standalone script method
func (p *PnpmManager) SelfInstall(ctx context.Context) error {
	// Check if already available (idempotent)
	if available, _ := p.IsAvailable(ctx); available {
		return nil
	}

	// Use ONLY the most independent installation method - standalone script
	// This ensures predictable behavior and treats pnpm as a first-class citizen
	return p.installViaStandaloneScript(ctx)
}

// installViaStandaloneScript uses pnpm's official installation script
func (p *PnpmManager) installViaStandaloneScript(ctx context.Context) error {
	script := `curl -fsSL https://get.pnpm.io/install.sh | sh`
	return executeInstallScript(ctx, script, "pnpm")
}

// Upgrade upgrades one or more packages to their latest versions
func (p *PnpmManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// Get all installed packages first
		installed, err := p.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to list installed packages: %w", err)
		}

		// Update all packages
		var upgradeErrors []string
		for _, pkg := range installed {
			output, err := ExecuteCommandCombined(ctx, p.binary, "update", "-g", pkg)
			if err != nil {
				upgradeErr := p.handleUpgradeError(err, output, pkg)
				if upgradeErr != nil {
					upgradeErrors = append(upgradeErrors, upgradeErr.Error())
				}
				continue
			}
		}

		if len(upgradeErrors) > 0 {
			return fmt.Errorf("some packages failed to upgrade: %s", strings.Join(upgradeErrors, "; "))
		}
		return nil
	}

	// Upgrade specific packages
	var upgradeErrors []string
	for _, pkg := range packages {
		output, err := ExecuteCommandCombined(ctx, p.binary, "update", "-g", pkg)
		if err != nil {
			upgradeErr := p.handleUpgradeError(err, output, pkg)
			if upgradeErr != nil {
				upgradeErrors = append(upgradeErrors, upgradeErr.Error())
			}
			continue
		}
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("failed to upgrade packages: %s", strings.Join(upgradeErrors, "; "))
	}
	return nil
}

// handleInstallError processes install command errors
func (p *PnpmManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "err_pnpm_peer_dep_issues") {
			return fmt.Errorf("peer dependency issues installing %s", packageName)
		}
		if strings.Contains(outputStr, "err_pnpm_package_not_found") ||
			strings.Contains(outputStr, "404 not found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already exists") ||
			strings.Contains(outputStr, "already installed") {
			// Package is already installed - this is typically fine
			return nil
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access is denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}
		if strings.Contains(outputStr, "network error") ||
			strings.Contains(outputStr, "cannot download") ||
			strings.Contains(outputStr, "connection refused") ||
			strings.Contains(outputStr, "timeout") {
			return fmt.Errorf("network error during installation")
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

	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors
func (p *PnpmManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "no such package") {
			return nil // Not installed is success for uninstall
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access is denied") {
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

	return fmt.Errorf("failed to execute uninstall command: %w", err)
}

// handleUpgradeError processes upgrade command errors
func (p *PnpmManager) handleUpgradeError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "err_pnpm_package_not_found") ||
			strings.Contains(outputStr, "404 not found") ||
			strings.Contains(outputStr, "not found") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already up-to-date") ||
			strings.Contains(outputStr, "already installed") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access is denied") {
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

// Dependencies returns package managers this manager depends on for self-installation
func (p *PnpmManager) Dependencies() []string {
	return []string{} // pnpm is independent - uses official installer script
}

func init() {
	RegisterManager("pnpm", func() PackageManager {
		return NewPnpmManager()
	})
}
