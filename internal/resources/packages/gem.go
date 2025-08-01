// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// GemManager manages Ruby gems.
type GemManager struct {
	binary string
}

// NewGemManager creates a new gem manager.
func NewGemManager() *GemManager {
	return &GemManager{
		binary: "gem",
	}
}

// ListInstalled lists all installed gems.
func (g *GemManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := ExecuteCommand(ctx, g.binary, "list", "--local", "--no-versions")
	if err != nil {
		return nil, err
	}

	return g.parseListOutput(output), nil
}

// parseListOutput parses gem list output
func (g *GemManager) parseListOutput(output []byte) []string {
	lines := SplitLines(output)
	var gems []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "***") {
			gems = append(gems, line)
		}
	}
	return gems
}

// Install installs a Ruby gem.
func (g *GemManager) Install(ctx context.Context, name string) error {
	// First try with --user-install
	output, err := ExecuteCommandCombined(ctx, g.binary, "install", name, "--user-install")
	if err != nil {
		outputStr := string(output)

		// Check if we should retry without --user-install
		if strings.Contains(outputStr, "--user-install") || strings.Contains(outputStr, "Use --user-install") {
			// Try without --user-install
			output2, err2 := ExecuteCommandCombined(ctx, g.binary, "install", name)
			if err2 == nil {
				return nil
			}
			// If this also fails, handle the new error
			return g.handleGemInstallError(err2, output2, name)
		}

		// Handle the original error
		return g.handleGemInstallError(err, output, name)
	}

	return nil
}

// handleGemInstallError provides gem-specific error handling with better messages
func (g *GemManager) handleGemInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for Ruby version mismatch
	if strings.Contains(outputStr, "requires Ruby version") {
		// Extract version requirements from error message
		var suggestion string
		if idx := strings.Index(outputStr, "Try installing it with"); idx > 0 {
			// Extract the suggested command
			cmdStart := idx + len("Try installing it with")
			if cmdEnd := strings.Index(outputStr[cmdStart:], "\n"); cmdEnd > 0 {
				suggestion = strings.TrimSpace(outputStr[cmdStart : cmdStart+cmdEnd])
			}
		}

		msg := fmt.Sprintf("gem '%s' requires a different Ruby version", packageName)

		if suggestion != "" {
			return fmt.Errorf("%s. Try: %s", msg, suggestion)
		}

		// Extract current Ruby version
		if idx := strings.Index(outputStr, "The current ruby version is"); idx > 0 {
			versionStart := idx + len("The current ruby version is")
			remainingStr := outputStr[versionStart:]
			if versionEnd := strings.Index(remainingStr, "."); versionEnd > 0 {
				// Find the end of the version number (usually something like "2.6.10")
				endIdx := versionEnd + 10
				if endIdx > len(remainingStr) {
					endIdx = len(remainingStr)
				}
				currentVersion := strings.TrimSpace(remainingStr[:endIdx])
				return fmt.Errorf("%s. Your Ruby version is %s. Check gem requirements or use rbenv/rvm to switch Ruby versions", msg, currentVersion)
			}
		}

		return fmt.Errorf("%s. Check Ruby version requirements or use rbenv/rvm to manage Ruby versions", msg)
	}

	// Fall back to base error handling
	return g.handleInstallError(err, output, packageName)
}

// Uninstall removes a Ruby gem.
func (g *GemManager) Uninstall(ctx context.Context, name string) error {
	output, err := ExecuteCommandCombined(ctx, g.binary, "uninstall", name, "-x", "-a", "-I")
	if err != nil {
		return g.handleUninstallError(err, output, name)
	}
	return nil
}

// IsInstalled checks if a specific gem is installed.
func (g *GemManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	output, err := ExecuteCommand(ctx, g.binary, "list", "--local", name)
	if err != nil {
		return false, fmt.Errorf("failed to check gem installation status for %s: %w", name, err)
	}

	result := strings.TrimSpace(string(output))
	// Check if the output contains the gem name
	// gem list output format: "gemname (version1, version2)"
	return strings.HasPrefix(result, name+" ") || result == name, nil
}

// Search searches for gems in RubyGems.org.
func (g *GemManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := ExecuteCommand(ctx, g.binary, "search", query)
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitCode, ok := ExtractExitCode(err); ok {
			// For gem search, exit code 1 usually means no results found
			if exitCode == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, fmt.Errorf("gem search command failed for %s: %w", query, err)
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, fmt.Errorf("failed to execute gem search command: %w", err)
	}

	return g.parseSearchOutput(output), nil
}

// parseSearchOutput parses gem search output
func (g *GemManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
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

	return gems
}

// Info retrieves detailed information about a gem.
func (g *GemManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if gem is installed first
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check gem installation status for %s: %w", name, err)
	}

	// Get gem specification
	output, err := ExecuteCommand(ctx, g.binary, "specification", name)
	if err != nil {
		if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
			return nil, fmt.Errorf("gem '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get gem info for %s: %w", name, err)
	}

	info := g.parseInfoOutput(output, name)
	info.Manager = "gem"
	info.Installed = installed

	// Get dependencies if installed
	if installed {
		info.Dependencies = g.getDependencies(ctx, name)
	}

	return info, nil
}

// parseInfoOutput parses gem specification output (YAML format)
func (g *GemManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{
		Name: name,
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

	return info
}

// getDependencies gets the dependencies of an installed gem
func (g *GemManager) getDependencies(ctx context.Context, name string) []string {
	output, err := ExecuteCommand(ctx, g.binary, "dependency", name)
	if err != nil {
		return []string{}
	}

	var dependencies []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Skip the gem itself and empty lines
		if line == "" || strings.HasPrefix(line, "Gem "+name) {
			continue
		}
		// Dependencies are listed with indentation (check before trimming)
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

// InstalledVersion retrieves the installed version of a gem
func (g *GemManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if gem is installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check gem installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("gem '%s' is not installed", name)
	}

	// Get version using gem list with specific gem
	output, err := ExecuteCommand(ctx, g.binary, "list", "--local", name)
	if err != nil {
		return "", fmt.Errorf("failed to get gem version information for %s: %w", name, err)
	}

	version := g.extractVersion(output, name)
	if version == "" {
		return "", fmt.Errorf("could not extract version for gem '%s' from output", name)
	}

	return version, nil
}

// extractVersion extracts version from gem list output
func (g *GemManager) extractVersion(output []byte, name string) string {
	result := strings.TrimSpace(string(output))
	// Parse version from output like "gemname (1.2.3)" or "gemname (1.2.3, 1.2.2)"
	if idx := strings.Index(result, "("); idx > 0 && idx < len(result)-1 {
		versionPart := result[idx+1:]
		if endIdx := strings.Index(versionPart, ")"); endIdx > 0 {
			versions := versionPart[:endIdx]
			// If multiple versions, return the first (latest)
			if commaIdx := strings.Index(versions, ","); commaIdx > 0 {
				return strings.TrimSpace(versions[:commaIdx])
			}
			return strings.TrimSpace(versions)
		}
	}
	return ""
}

// IsAvailable checks if gem is installed and accessible
func (g *GemManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(g.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, g.binary, []string{"--version"})
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

// SupportsSearch returns true as gem supports package search
func (g *GemManager) SupportsSearch() bool {
	return true
}

// handleInstallError processes install command errors
func (g *GemManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "could not find a valid gem") ||
			strings.Contains(outputStr, "could not find gem") ||
			strings.Contains(outputStr, "unknown gem") {
			return fmt.Errorf("package '%s' not found", packageName)
		}

		if strings.Contains(outputStr, "already installed") {
			// Package is already installed - this is typically fine
			return nil
		}

		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "you don't have write permissions") ||
			strings.Contains(outputStr, "insufficient permissions") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}

		if strings.Contains(outputStr, "database is locked") ||
			strings.Contains(outputStr, "gem is locked") {
			return fmt.Errorf("package manager database is locked")
		}

		if strings.Contains(outputStr, "network error") ||
			strings.Contains(outputStr, "could not download") ||
			strings.Contains(outputStr, "unable to download") ||
			strings.Contains(outputStr, "connection refused") {
			return fmt.Errorf("network error during installation")
		}

		if strings.Contains(outputStr, "failed to build") ||
			strings.Contains(outputStr, "error building") ||
			strings.Contains(outputStr, "extconf failed") ||
			strings.Contains(outputStr, "make failed") {
			return fmt.Errorf("failed to build package '%s'", packageName)
		}

		if strings.Contains(outputStr, "dependency conflict") ||
			strings.Contains(outputStr, "incompatible") ||
			strings.Contains(outputStr, "conflicts with") {
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
	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors
func (g *GemManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "is not installed") ||
			strings.Contains(outputStr, "cannot uninstall") ||
			strings.Contains(outputStr, "gem not installed") {
			// Package is not installed - this is typically fine for uninstall
			return nil
		}

		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "you don't have write permissions") ||
			strings.Contains(outputStr, "insufficient permissions") {
			return fmt.Errorf("permission denied uninstalling %s", packageName)
		}

		if strings.Contains(outputStr, "database is locked") ||
			strings.Contains(outputStr, "gem is locked") {
			return fmt.Errorf("package manager database is locked")
		}

		if strings.Contains(outputStr, "dependency") && strings.Contains(outputStr, "requires") ||
			strings.Contains(outputStr, "is depended upon") {
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
	return fmt.Errorf("failed to execute uninstall command: %w", err)
}
