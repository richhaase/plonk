// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/interfaces"
)

// GemManager manages Ruby gems.
type GemManager struct{}

// NewGemManager creates a new gem manager.
func NewGemManager() *GemManager {
	return &GemManager{}
}

// IsAvailable checks if gem is installed and accessible.
func (g *GemManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("gem")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify gem is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "gem", "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// gem exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "gem binary found but not functional")
	}

	return true, nil
}

// ListInstalled lists all installed gems that provide executables.
func (g *GemManager) ListInstalled(ctx context.Context) ([]string, error) {
	// Get all installed gems
	cmd := exec.CommandContext(ctx, "gem", "list", "--local", "--no-versions")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
			"failed to execute gem list command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No gems installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output to get gem names
	var gems []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "***") {
			// Check if this gem provides executables
			if g.hasExecutables(ctx, line) {
				gems = append(gems, line)
			}
		}
	}

	return gems, nil
}

// hasExecutables checks if a gem provides executable files
func (g *GemManager) hasExecutables(ctx context.Context, gemName string) bool {
	// Use a short timeout for this check since we're doing it for many gems
	checkCtx, cancel := context.WithTimeout(ctx, 2*1000*1000*1000) // 2 seconds
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "gem", "contents", gemName, "--executables")
	output, err := cmd.Output()
	if err != nil {
		// If there's an error, assume no executables
		return false
	}

	// If there's any output, the gem has executables
	return strings.TrimSpace(string(output)) != ""
}

// Install installs a Ruby gem.
func (g *GemManager) Install(ctx context.Context, name string) error {
	// Try to install with --user-install first for better permissions handling
	cmd := exec.CommandContext(ctx, "gem", "install", name, "--user-install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for gem not found
			if strings.Contains(outputStr, "Could not find a valid gem") ||
				strings.Contains(outputStr, "ERROR:  Could not find") {
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("gem '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available gems: gem search %s", name))
			}

			// Check for already installed
			if strings.Contains(outputStr, "already installed") {
				// Gem is already installed - this is typically fine
				return nil
			}

			// Check for permission errors or --user-install issues
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "Errno::EACCES") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
					fmt.Sprintf("permission denied installing %s", name)).
					WithSuggestionMessage("Try without --user-install or check Ruby installation")
			}

			// If --user-install failed for other reasons, try without it
			if strings.Contains(outputStr, "--user-install") {
				cmd = exec.CommandContext(ctx, "gem", "install", name)
				_, err = cmd.CombinedOutput()
				if err == nil {
					return nil
				}
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("gem installation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute gem install command")
	}

	return nil
}

// Uninstall removes a Ruby gem.
func (g *GemManager) Uninstall(ctx context.Context, name string) error {
	// Use -x to remove executables and -a to remove all versions
	cmd := exec.CommandContext(ctx, "gem", "uninstall", name, "-x", "-a", "-I")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for "not installed" condition
			if strings.Contains(outputStr, "is not installed") ||
				strings.Contains(outputStr, "ERROR:  While executing gem") {
				// Gem is not installed - this is typically fine for uninstall
				return nil
			}

			// Check for permission errors
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "Errno::EACCES") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
					fmt.Sprintf("permission denied uninstalling %s", name)).
					WithSuggestionMessage("Try with elevated permissions or check gem installation")
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
				fmt.Sprintf("gem uninstallation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute gem uninstall command")
	}

	return nil
}

// IsInstalled checks if a specific gem is installed.
func (g *GemManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "gem", "list", "--local", name)
	output, err := cmd.Output()
	if err != nil {
		return false, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "check", name,
			"failed to check gem installation status")
	}

	result := strings.TrimSpace(string(output))
	// Check if the output contains the gem name
	// gem list output format: "gemname (version1, version2)"
	return strings.HasPrefix(result, name+" ") || result == name, nil
}

// Search searches for gems in RubyGems.org.
func (g *GemManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "gem", "search", query)
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// For gem search, exit code 1 usually means no results found
			if exitError.ExitCode() == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
				"gem search command failed")
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "search",
			"failed to execute gem search command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No gems found - this is normal, not an error
		return []string{}, nil
	}

	// Parse output to extract gem names
	var gems []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Extract gem name from line like "gemname (version)"
			if idx := strings.Index(line, " ("); idx > 0 {
				gemName := line[:idx]
				gems = append(gems, gemName)
			}
		}
	}

	return gems, nil
}

// Info retrieves detailed information about a gem.
func (g *GemManager) Info(ctx context.Context, name string) (*interfaces.PackageInfo, error) {
	// Check if gem is installed first
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check gem installation status")
	}

	// Get gem specification
	cmd := exec.CommandContext(ctx, "gem", "specification", name)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
					fmt.Sprintf("gem '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available gems: gem search %s", name))
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get gem info")
	}

	// Parse gem specification output (YAML format)
	info := &interfaces.PackageInfo{
		Name:      name,
		Manager:   "gem",
		Installed: installed,
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			// Remove quotes if present
			info.Version = strings.Trim(info.Version, "\"'")
		} else if strings.HasPrefix(line, "summary:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "summary:"))
			info.Description = strings.Trim(info.Description, "\"'")
		} else if strings.HasPrefix(line, "homepage:") {
			info.Homepage = strings.TrimSpace(strings.TrimPrefix(line, "homepage:"))
			info.Homepage = strings.Trim(info.Homepage, "\"'")
		}
	}

	// Get dependencies if installed
	if installed {
		deps := g.getDependencies(ctx, name)
		info.Dependencies = deps
	}

	return info, nil
}

// getDependencies gets the dependencies of an installed gem
func (g *GemManager) getDependencies(ctx context.Context, name string) []string {
	cmd := exec.CommandContext(ctx, "gem", "dependency", name)
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var dependencies []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip the gem itself and empty lines
		if line == "" || strings.HasPrefix(line, "Gem "+name) {
			continue
		}
		// Dependencies are listed with indentation
		if strings.HasPrefix(line, "  ") {
			// Extract dependency name (format: "  gemname (>= version)")
			depLine := strings.TrimSpace(line)
			if idx := strings.Index(depLine, " ("); idx > 0 {
				depName := depLine[:idx]
				dependencies = append(dependencies, depName)
			}
		}
	}

	return dependencies
}

// GetInstalledVersion retrieves the installed version of a gem
func (g *GemManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if gem is installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check gem installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("gem '%s' is not installed", name))
	}

	// Get version using gem list with specific gem
	cmd := exec.CommandContext(ctx, "gem", "list", "--local", name)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get gem version information")
	}

	result := strings.TrimSpace(string(output))
	// Parse version from output like "gemname (1.2.3)" or "gemname (1.2.3, 1.2.2)"
	if idx := strings.Index(result, "("); idx > 0 && idx < len(result)-1 {
		versionPart := result[idx+1:]
		if endIdx := strings.Index(versionPart, ")"); endIdx > 0 {
			versions := versionPart[:endIdx]
			// If multiple versions, return the first (latest)
			if commaIdx := strings.Index(versions, ","); commaIdx > 0 {
				return strings.TrimSpace(versions[:commaIdx]), nil
			}
			return strings.TrimSpace(versions), nil
		}
	}

	return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
		fmt.Sprintf("could not extract version for gem '%s' from output", name))
}
