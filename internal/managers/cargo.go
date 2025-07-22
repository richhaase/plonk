// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/executor"
)

// CargoManager manages Rust packages via cargo using BaseManager for common functionality.
type CargoManager struct {
	*BaseManager
}

// NewCargoManager creates a new cargo manager with the default executor.
func NewCargoManager() *CargoManager {
	return newCargoManager(nil)
}

// NewCargoManagerWithExecutor creates a new cargo manager with a custom executor for testing.
func NewCargoManagerWithExecutor(exec executor.CommandExecutor) *CargoManager {
	return newCargoManager(exec)
}

// newCargoManager creates a cargo manager with the given executor.
func newCargoManager(exec executor.CommandExecutor) *CargoManager {
	config := ManagerConfig{
		BinaryName:  "cargo",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"install", "--list"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", pkg}
		},
	}

	// Add cargo-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "no crates found", "could not find")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "binary `", "` already exists")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "not installed")

	var base *BaseManager
	if exec == nil {
		base = NewBaseManager(config)
	} else {
		base = NewBaseManagerWithExecutor(config, exec)
	}
	base.ErrorMatcher = errorMatcher

	return &CargoManager{
		BaseManager: base,
	}
}

// ListInstalled lists all installed cargo packages.
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := c.ExecuteList(ctx)
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
	return c.ExecuteInstall(ctx, name)
}

// Uninstall removes a cargo package.
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	return c.ExecuteUninstall(ctx, name)
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
	output, err := c.Executor.Execute(ctx, c.GetBinary(), "search", query)
	if err != nil {
		// cargo search returns a non-zero exit code if no packages are found.
		if _, ok := err.(interface{ ExitCode() int }); ok {
			outputStr := string(output)
			if strings.Contains(outputStr, "no crates found") {
				return []string{}, nil
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
			"failed to search for cargo package")
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
	output, err := c.Executor.Execute(ctx, c.GetBinary(), "search", name, "--limit", "1")
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get info for cargo package")
	}

	info := c.parseInfoOutput(output, name)
	if info == nil {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("package '%s' not found", name)).
			WithSuggestionMessage(fmt.Sprintf("Search available packages: cargo search %s", name))
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
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check package installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("package '%s' is not installed", name))
	}

	// Use cargo install --list to get version information
	output, err := c.Executor.Execute(ctx, c.GetBinary(), "install", "--list")
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	version := c.extractVersion(output, name)
	if version == "" {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("could not find version information for package '%s'", name))
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
