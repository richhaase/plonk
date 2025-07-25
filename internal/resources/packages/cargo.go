// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CargoManager manages Rust packages via cargo.
type CargoManager struct {
	binary string
}

// NewCargoManager creates a new cargo manager.
func NewCargoManager() *CargoManager {
	return &CargoManager{
		binary: "cargo",
	}
}

// ListInstalled lists all installed cargo packages.
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, c.binary, "install", "--list")
	if err != nil {
		return nil, fmt.Errorf("failed to list installed cargo packages: %w", err)
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
	output, err := ExecuteCommandCombined(ctx, c.binary, "install", name)
	if err != nil {
		return c.handleInstallError(err, output, name)
	}
	return nil
}

// Uninstall removes a cargo package.
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, c.binary, "uninstall", name)
	if err != nil {
		return c.handleUninstallError(err, output, name)
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
	cmd := exec.CommandContext(ctx, c.binary, "search", query)
	output, err := cmd.Output()
	if err != nil {
		// cargo search returns a non-zero exit code if no packages are found.
		if _, ok := err.(interface{ ExitCode() int }); ok {
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
	cmd := exec.CommandContext(ctx, c.binary, "search", name, "--limit", "1")
	output, err := cmd.Output()
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

// GetInstalledVersion retrieves the installed version of a Cargo package
func (c *CargoManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check package installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("package '%s' is not installed", name)
	}

	// Use cargo install --list to get version information
	cmd := exec.CommandContext(ctx, c.binary, "install", "--list")
	output, err := cmd.Output()
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
	if !CheckCommandAvailable(c.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, c.binary, []string{"--version"})
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

// SupportsSearch returns true as Cargo supports package search
func (c *CargoManager) SupportsSearch() bool {
	return true
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
			return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
		}
		// Exit code 0 with no recognized error pattern - success
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute install command: %w", err)
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
			return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
		}
		// Exit code 0 with no recognized error pattern - success
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute uninstall command: %w", err)
}
