// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CargoManager manages Rust packages via cargo.
type CargoManager struct {
	binary string
	exec   CommandExecutor
}

// NewCargoManager creates a new cargo manager with default executor.
func NewCargoManager() *CargoManager {
	return NewCargoManagerWithExecutor(nil)
}

// NewCargoManagerWithExecutor creates a new cargo manager with the provided executor.
// If executor is nil, uses the default executor.
func NewCargoManagerWithExecutor(executor CommandExecutor) *CargoManager {
	if executor == nil {
		executor = defaultExecutor
	}
	return &CargoManager{
		binary: "cargo",
		exec:   executor,
	}
}

// ListInstalled lists all installed cargo packages.
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteWith(ctx, c.exec, c.binary, "install", "--list")
	if err != nil {
		return nil, err
	}

	return c.parseListOutput(output), nil
}

// parseListOutput parses cargo install --list output
func (c *CargoManager) parseListOutput(output []byte) []string {
	lines := strings.Split(string(output), "\n")
	var packages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for package header lines like "packagename v1.2.3:"
		if strings.Contains(line, " v") && strings.HasSuffix(line, ":") {
			// Extract package name from "packagename v1.2.3:"
			fields := strings.Fields(line)
			if len(fields) > 0 {
				packages = append(packages, fields[0])
			}
		}
	}

	return packages
}

// Install installs a cargo package.
func (c *CargoManager) Install(ctx context.Context, name string) error {
	output, err := CombinedOutputWith(ctx, c.exec, c.binary, "install", name)
	if err != nil {
		return c.handleInstallError(err, []byte(output), name)
	}
	return nil
}

// Uninstall removes a cargo package.
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	output, err := CombinedOutputWith(ctx, c.exec, c.binary, "uninstall", name)
	if err != nil {
		return c.handleUninstallError(err, []byte(output), name)
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (c *CargoManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	packages, err := c.ListInstalled(ctx)
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

// Search searches for packages in the cargo registry.
func (c *CargoManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := ExecuteWith(ctx, c.exec, c.binary, "search", query)
	if err != nil {
		// cargo search returns a non-zero exit code if no packages are found.
		if _, ok := ExtractExitCode(err); ok {
			outputStr := string(output)
			if strings.Contains(outputStr, "no crates found") {
				return []string{}, nil
			}
		}
		return nil, fmt.Errorf("failed to search for cargo package %s: %w", query, err)
	}

	return c.parseSearchOutput(output), nil
}

// parseSearchOutput parses cargo search output
func (c *CargoManager) parseSearchOutput(output []byte) []string {
	var packages []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		// Parse lines like: "serde = \"1.0.136\""
		fields := strings.Fields(line)
		if len(fields) > 0 {
			packages = append(packages, fields[0])
		}
	}

	return packages
}

// Info retrieves detailed information about a package.
func (c *CargoManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Use search with limit 1 to get package info
	output, err := ExecuteWith(ctx, c.exec, c.binary, "search", name, "--limit", "1")
	if err != nil {
		return nil, fmt.Errorf("failed to get info for cargo package %s: %w", name, err)
	}

	info := c.parseInfoOutput(output, name)
	if info == nil {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	// Check if installed
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	info.Installed = installed
	info.Manager = "cargo"

	return info, nil
}

// parseInfoOutput parses cargo search output for package info
func (c *CargoManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if !scanner.Scan() {
		return nil
	}

	line := scanner.Text()
	// Parse line like: "serde = \"Serialization framework for Rust\""
	fields := strings.SplitN(line, " = ", 2)
	if len(fields) < 2 {
		return nil
	}

	packageName := strings.TrimSpace(fields[0])
	if packageName != name {
		return nil
	}

	var description string
	if len(fields) > 1 {
		description = strings.Trim(fields[1], `"`)
	}

	// Extract version if present in the description
	// Some outputs include version like: "serde = \"1.0.136\"  # Serialization..."
	var version string
	if strings.Contains(description, "  # ") {
		parts := strings.SplitN(description, "  # ", 2)
		version = strings.Trim(parts[0], `" `)
		if len(parts) > 1 {
			description = parts[1]
		}
	} else {
		// The version might be the entire description for simple outputs
		// Check if it looks like a version (starts with digit)
		trimmed := strings.Trim(description, `"`)
		if len(trimmed) > 0 && (trimmed[0] >= '0' && trimmed[0] <= '9') {
			version = trimmed
			description = ""
		}
	}

	return &PackageInfo{
		Name:        name,
		Version:     version,
		Description: description,
	}
}

// InstalledVersion retrieves the installed version of a Cargo package
func (c *CargoManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Use cargo install --list to get version information
	output, err := ExecuteWith(ctx, c.exec, c.binary, "install", "--list")
	if err != nil {
		return "", fmt.Errorf("failed to get package version information for %s: %w", name, err)
	}

	version := c.extractVersion(output, name)
	if version == "" {
		return "", fmt.Errorf("could not find version information for package '%s'", name)
	}

	return version, nil
}

// extractVersion extracts version from cargo install --list output
func (c *CargoManager) extractVersion(output []byte, name string) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines like "packagename v1.2.3:"
		if strings.HasPrefix(line, name+" v") {
			// Extract version from "packagename v1.2.3:"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				version := parts[1]
				// Remove 'v' prefix and trailing colon if present
				version = strings.TrimPrefix(version, "v")
				version = strings.TrimSuffix(version, ":")
				return version
			}
		}
	}
	return ""
}

// IsAvailable checks if cargo is installed and accessible
func (c *CargoManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailableWith(c.exec, c.binary) {
		return false, nil
	}

	err := VerifyBinaryWith(ctx, c.exec, c.binary, []string{"--version"})
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

// CheckHealth performs a comprehensive health check of the Cargo installation
func (c *CargoManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     "Cargo Manager",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Cargo is available and properly configured",
	}

	// Check availability
	available, err := c.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "warn"
		check.Message = "Cargo availability check failed"
		check.Issues = []string{fmt.Sprintf("Error checking cargo: %v", err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = "Cargo is not available"
		check.Issues = []string{"cargo command not found"}
		check.Suggestions = []string{
			"Install Rust (includes Cargo): curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
			"Or install via Homebrew: brew install rust",
		}
		return check, nil
	}

	// Discover cargo bin directory
	binDir, err := c.getBinDirectory()
	if err != nil {
		check.Status = "warn"
		check.Message = "Could not determine Cargo bin directory"
		check.Issues = []string{fmt.Sprintf("Error discovering bin directory: %v", err)}
		return check, nil
	}

	check.Details = append(check.Details, fmt.Sprintf("Cargo bin directory: %s", binDir))

	// Check PATH
	pathCheck := checkDirectoryInPath(binDir)
	if !pathCheck.inPath && pathCheck.exists {
		check.Status = "warn"
		check.Message = "Cargo bin directory exists but not in PATH"
		check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
		check.Suggestions = pathCheck.suggestions
	} else if !pathCheck.exists {
		check.Details = append(check.Details, "Cargo bin directory does not exist (no packages installed)")
	} else {
		check.Details = append(check.Details, "Cargo bin directory is in PATH")
	}

	return check, nil
}

// getBinDirectory discovers the Cargo bin directory
func (c *CargoManager) getBinDirectory() (string, error) {
	// Check CARGO_HOME first
	if cargoHome := os.Getenv("CARGO_HOME"); cargoHome != "" {
		return filepath.Join(cargoHome, "bin"), nil
	}

	// Default to ~/.cargo/bin
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cargo", "bin"), nil
}

// Upgrade upgrades one or more packages to their latest versions
func (c *CargoManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		// Get all installed packages and upgrade them
		installed, err := c.ListInstalled(ctx)
		if err != nil {
			return fmt.Errorf("failed to get installed packages: %w", err)
		}
		packages = installed
	}

	if len(packages) == 0 {
		return nil // No packages to upgrade
	}

	// Upgrade packages by reinstalling them (cargo install reinstalls with latest version)
	for _, pkg := range packages {
		output, err := CombinedOutputWith(ctx, c.exec, c.binary, "install", pkg)
		if err != nil {
			return c.handleUpgradeError(err, []byte(output), pkg)
		}
	}
	return nil
}

// Dependencies returns package managers this manager depends on for self-installation
func (c *CargoManager) Dependencies() []string {
	return []string{} // Cargo is independent - uses official rustup installer script
}

func init() {
	RegisterManagerV2("cargo", func(exec CommandExecutor) PackageManager {
		return NewCargoManagerWithExecutor(exec)
	})
}

// handleUpgradeError processes upgrade command errors
func (c *CargoManager) handleUpgradeError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no matching package found") ||
			strings.Contains(outputStr, "not find package") ||
			strings.Contains(outputStr, "could not find") {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		if strings.Contains(outputStr, "already installed") {
			return nil // Already up-to-date is success
		}
		if strings.Contains(outputStr, "permission denied") {
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
func (c *CargoManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "no matching package found") ||
			strings.Contains(outputStr, "not find package") ||
			strings.Contains(outputStr, "could not find") {
			return fmt.Errorf("package '%s' not found", packageName)
		}

		if strings.Contains(outputStr, "already installed") {
			// Package is already installed - this is typically fine
			return nil
		}

		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}

		if strings.Contains(outputStr, "database is locked") ||
			strings.Contains(outputStr, "cargo is locked") {
			return fmt.Errorf("package manager database is locked")
		}

		if strings.Contains(outputStr, "network error") ||
			strings.Contains(outputStr, "failed to fetch") ||
			strings.Contains(outputStr, "connection timed out") {
			return fmt.Errorf("network error during installation")
		}

		if strings.Contains(outputStr, "failed to compile") ||
			strings.Contains(outputStr, "build failed") ||
			strings.Contains(outputStr, "compilation error") {
			return fmt.Errorf("failed to build package '%s'", packageName)
		}

		if strings.Contains(outputStr, "dependency conflict") ||
			strings.Contains(outputStr, "incompatible") ||
			strings.Contains(outputStr, "version conflict") {
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)
		}

		// Only treat non-zero exit codes as errors
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
		// Exit code 0 with no recognized error pattern - success
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}

// handleUninstallError processes uninstall command errors
func (c *CargoManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "not installed") ||
			strings.Contains(outputStr, "package `"+strings.ToLower(packageName)+"` is not installed") {
			// Package is not installed - this is typically fine for uninstall
			return nil
		}

		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access denied") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}

		if strings.Contains(outputStr, "database is locked") ||
			strings.Contains(outputStr, "cargo is locked") {
			return fmt.Errorf("package manager database is locked")
		}

		if strings.Contains(outputStr, "required by") ||
			strings.Contains(outputStr, "still depends on") {
			return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)
		}

		// Only treat non-zero exit codes as errors
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
		// Exit code 0 with no recognized error pattern - success
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return err
}
