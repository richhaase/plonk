// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// NpmManager manages NPM packages.
type NpmManager struct{}

// NewNpmManager creates a new NPM manager.
func NewNpmManager() *NpmManager {
	return &NpmManager{}
}

// IsAvailable checks if NPM is installed and accessible.
func (n *NpmManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("npm")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}
	
	// Verify npm is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "npm", "--version")
	err = cmd.Run()
	if err != nil {
		// npm exists but is not functional - this is an error
		return false, fmt.Errorf("npm binary found but not functional: %w", err)
	}
	
	return true, nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--parseable")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// npm list can return non-zero exit codes even when working correctly
			// (e.g., when there are peer dependency warnings)
			// Only treat it as an error if the exit code indicates a real failure
			if exitError.ExitCode() > 1 {
				return nil, fmt.Errorf("failed to list npm packages: %w", err)
			}
			// Exit code 1 might just be warnings - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, fmt.Errorf("failed to execute npm list: %w", err)
		}
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output to extract package names
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "/") {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if pkg != "" && pkg != "lib" {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages, nil
}

// Install installs a global NPM package.
func (n *NpmManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		
		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// NPM often returns exit code 1 for warnings, check actual content
			if exitError.ExitCode() == 1 {
				// Check for "already installed" warnings
				if strings.Contains(outputStr, "already installed") || strings.Contains(outputStr, "up to date") {
					// Package is already installed - this is typically fine
					return nil
				}
				
				// Check for package not found
				if strings.Contains(outputStr, "404") || strings.Contains(outputStr, "Not found") || strings.Contains(outputStr, "E404") {
					return fmt.Errorf("package '%s' not found in npm registry", name)
				}
				
				// Check for permission errors
				if strings.Contains(outputStr, "EACCES") || strings.Contains(outputStr, "permission denied") {
					return fmt.Errorf("permission denied installing %s: try running with sudo or fix npm permissions\nOutput: %s", name, outputStr)
				}
			}
			
			// Other exit errors with more context
			return fmt.Errorf("failed to install %s (exit code %d): %w\nOutput: %s", name, exitError.ExitCode(), err, outputStr)
		}
		
		// Non-exit errors (command not found, context cancellation, etc.)
		return fmt.Errorf("failed to execute npm install for %s: %w\nOutput: %s", name, err, outputStr)
	}
	
	return nil
}

// Uninstall removes a global NPM package.
func (n *NpmManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "uninstall", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		
		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// NPM often returns exit code 1 for warnings, check actual content
			if exitError.ExitCode() == 1 {
				// Check for "not installed" warnings
				if strings.Contains(outputStr, "not installed") || strings.Contains(outputStr, "up to date") {
					// Package is not installed - this is typically fine for uninstall
					return nil
				}
				
				// Check for permission errors
				if strings.Contains(outputStr, "EACCES") || strings.Contains(outputStr, "permission denied") {
					return fmt.Errorf("permission denied uninstalling %s: try running with sudo or fix npm permissions\nOutput: %s", name, outputStr)
				}
				
				// Check for dependency issues (less common in npm global packages)
				if strings.Contains(outputStr, "ENOENT") || strings.Contains(outputStr, "cannot remove") {
					return fmt.Errorf("cannot uninstall %s: package files may be corrupted or missing\nOutput: %s", name, outputStr)
				}
			}
			
			// Other exit errors with more context
			return fmt.Errorf("failed to uninstall %s (exit code %d): %w\nOutput: %s", name, exitError.ExitCode(), err, outputStr)
		}
		
		// Non-exit errors (command not found, context cancellation, etc.)
		return fmt.Errorf("failed to execute npm uninstall for %s: %w\nOutput: %s", name, err, outputStr)
	}
	
	return nil
}

// IsInstalled checks if a specific package is installed globally.
func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (npm not found, permission issues, etc.)
		return false, fmt.Errorf("failed to check package %s: %w", name, err)
	}
	return true, nil
}

// Search searches for packages in NPM registry.
func (n *NpmManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "search", query, "--json")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// For npm search, exit code 1 usually means no results found
			if exitError.ExitCode() == 1 {
				return []string{}, nil
			}
			// Other exit codes indicate real errors
			return nil, fmt.Errorf("failed to search npm packages: %w", err)
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, fmt.Errorf("failed to execute npm search: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		// No packages found - this is normal, not an error
		return []string{}, nil
	}

	// Parse JSON output to extract package names
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, `"name":`) {
			// Extract package name from JSON line like: "name": "package-name",
			parts := strings.Split(line, `"name":`)
			if len(parts) > 1 {
				namepart := strings.TrimSpace(parts[1])
				if strings.HasPrefix(namepart, `"`) && strings.Contains(namepart, `"`) {
					// Extract the name between quotes
					namepart = namepart[1:] // Remove leading quote
					if idx := strings.Index(namepart, `"`); idx > 0 {
						packageName := namepart[:idx]
						if packageName != "" {
							packages = append(packages, packageName)
						}
					}
				}
			}
		}
	}

	return packages, nil
}

