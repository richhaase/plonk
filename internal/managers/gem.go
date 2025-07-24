// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GemManager manages Ruby gems using BaseManager for common functionality.
type GemManager struct {
	*BaseManager
}

// NewGemManager creates a new gem manager.
func NewGemManager() *GemManager {
	return newGemManager()
}

// newGemManager creates a gem manager.
func newGemManager() *GemManager {
	config := ManagerConfig{
		BinaryName:  "gem",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"list", "--local", "--no-versions"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", pkg, "--user-install"}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"uninstall", pkg, "-x", "-a", "-I"}
		},
	}

	// Add gem-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "Could not find a valid gem", "ERROR:  Could not find")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "already installed")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "is not installed")
	errorMatcher.AddPattern(ErrorTypePermission, "Errno::EACCES", "Gem::FilePermissionError")
	errorMatcher.AddPattern(ErrorTypeDependency, "requires Ruby version", "ruby version is")

	base := NewBaseManager(config)
	base.ErrorMatcher = errorMatcher

	return &GemManager{
		BaseManager: base,
	}
}

// ListInstalled lists all installed gems.
func (g *GemManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := g.ExecuteList(ctx)
	if err != nil {
		return nil, err
	}

	return g.parseListOutput(output), nil
}

// parseListOutput parses gem list output
func (g *GemManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	// Parse output to get gem names
	var gems []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "***") {
			gems = append(gems, line)
		}
	}

	return gems
}

// Install installs a Ruby gem.
func (g *GemManager) Install(ctx context.Context, name string) error {
	// First try with --user-install
	args := g.Config.InstallArgs(name)
	cmd := exec.CommandContext(ctx, g.GetBinary(), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check if we should retry without --user-install
		if strings.Contains(outputStr, "--user-install") || strings.Contains(outputStr, "Use --user-install") {
			// Try without --user-install
			retryCmd := exec.CommandContext(ctx, g.GetBinary(), "install", name)
			output2, err2 := retryCmd.CombinedOutput()
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
			if versionEnd := strings.Index(outputStr[versionStart:], "."); versionEnd > 0 {
				currentVersion := strings.TrimSpace(outputStr[versionStart : versionStart+versionEnd+10])
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
	return g.ExecuteUninstall(ctx, name)
}

// IsInstalled checks if a specific gem is installed.
func (g *GemManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, g.GetBinary(), "list", "--local", name)
	output, err := cmd.Output()
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
	cmd := exec.CommandContext(ctx, g.GetBinary(), "search", query)
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if execErr, ok := err.(interface{ ExitCode() int }); ok {
			// For gem search, exit code 1 usually means no results found
			if execErr.ExitCode() == 1 {
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
	cmd := exec.CommandContext(ctx, g.GetBinary(), "specification", name)
	output, err := cmd.Output()
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 1 {
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
	cmd := exec.CommandContext(ctx, g.GetBinary(), "dependency", name)
	output, err := cmd.Output()
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

// GetInstalledVersion retrieves the installed version of a gem
func (g *GemManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if gem is installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check gem installation status for %s: %w", name, err)
	}
	if !installed {
		return "", fmt.Errorf("gem '%s' is not installed", name)
	}

	// Get version using gem list with specific gem
	cmd := exec.CommandContext(ctx, g.GetBinary(), "list", "--local", name)
	output, err := cmd.Output()
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
