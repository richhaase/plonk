// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// AptManager manages packages via APT (Advanced Package Tool) on Debian-based systems.
type AptManager struct{}

// NewAptManager creates a new APT manager.
func NewAptManager() *AptManager {
	return &AptManager{}
}

// IsAvailable checks if APT is installed and this is a Debian-based system.
func (a *AptManager) IsAvailable(ctx context.Context) (bool, error) {
	// APT is only available on Linux
	if runtime.GOOS != "linux" {
		return false, nil
	}

	// Check if apt command exists
	_, err := exec.LookPath("apt")
	if err != nil {
		// apt not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify apt is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "apt", "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// apt exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "apt binary found but not functional")
	}

	return true, nil
}

// ListInstalled lists all manually installed packages (not dependencies).
func (a *AptManager) ListInstalled(ctx context.Context) ([]string, error) {
	// Use apt-mark showmanual to list only manually installed packages
	cmd := exec.CommandContext(ctx, "apt-mark", "showmanual")
	output, err := cmd.Output()
	if err != nil {
		// Check if apt-mark is not available
		if exitError, ok := err.(*exec.ExitError); ok {
			if strings.Contains(string(exitError.Stderr), "command not found") {
				// Fall back to dpkg query
				return a.listInstalledFallback(ctx)
			}
		}
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
			"failed to execute apt-mark showmanual command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	// Parse the output - one package per line
	lines := strings.Split(result, "\n")
	var packages []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			packages = append(packages, line)
		}
	}

	return packages, nil
}

// listInstalledFallback uses dpkg to list installed packages as a fallback
func (a *AptManager) listInstalledFallback(ctx context.Context) ([]string, error) {
	// Use dpkg to get list of installed packages
	cmd := exec.CommandContext(ctx, "dpkg", "--get-selections")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list",
			"failed to execute dpkg --get-selections command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	// Parse dpkg output (package-name<tab>install)
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "install" {
			packages = append(packages, fields[0])
		}
	}

	return packages, nil
}

// Install installs a package via apt-get.
func (a *AptManager) Install(ctx context.Context, name string) error {
	// Use apt-get for scripting (more stable interface than apt)
	cmd := exec.CommandContext(ctx, "apt-get", "install", "-y", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for permission denied
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "are you root?") ||
				strings.Contains(outputStr, "Could not open lock file") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "install",
					"sudo access required for apt operations").
					WithSuggestionMessage("Run plonk with sudo or as root")
			}

			// Check for package not found
			if strings.Contains(outputStr, "Unable to locate package") ||
				strings.Contains(outputStr, "has no installation candidate") {
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("package '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages: apt-cache search %s", name))
			}

			// Check for already installed
			if strings.Contains(outputStr, "is already the newest version") {
				// Package is already installed - this is typically fine
				return nil
			}

			// Check for APT lock
			if strings.Contains(outputStr, "Could not get lock") ||
				strings.Contains(outputStr, "Unable to lock") {
				return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "install",
					"apt database is locked").
					WithSuggestionMessage("Wait for other apt processes to complete or remove lock file")
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("package installation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute apt-get install command")
	}

	return nil
}

// Uninstall removes a package via apt-get.
func (a *AptManager) Uninstall(ctx context.Context, name string) error {
	// Use apt-get remove (not purge) to leave config files
	cmd := exec.CommandContext(ctx, "apt-get", "remove", "-y", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for permission denied
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "are you root?") ||
				strings.Contains(outputStr, "Could not open lock file") {
				return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
					"sudo access required for apt operations").
					WithSuggestionMessage("Run plonk with sudo or as root")
			}

			// Check for "not installed" condition
			if strings.Contains(outputStr, "is not installed") ||
				strings.Contains(outputStr, "Unable to locate package") {
				// Package is not installed - this is typically fine for uninstall
				return nil
			}

			// Check for APT lock
			if strings.Contains(outputStr, "Could not get lock") ||
				strings.Contains(outputStr, "Unable to lock") {
				return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "uninstall",
					"apt database is locked").
					WithSuggestionMessage("Wait for other apt processes to complete or remove lock file")
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name,
				fmt.Sprintf("package uninstallation failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "uninstall", name,
			"failed to execute apt-get remove command")
	}

	return nil
}

// IsInstalled checks if a specific package is installed.
func (a *AptManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	// Use dpkg -l to check if package is installed
	cmd := exec.CommandContext(ctx, "dpkg", "-l", name)
	output, err := cmd.Output()
	if err != nil {
		// dpkg returns exit code 1 if package is not found
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "check", name,
			"failed to check package installation status")
	}

	// Parse dpkg output to check installation status
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for lines starting with "ii" (installed) or "hi" (hold installed)
		if strings.HasPrefix(line, "ii ") || strings.HasPrefix(line, "hi ") {
			// Check if this line is for our package
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// Search searches for packages using apt-cache search.
func (a *AptManager) Search(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "search", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "search",
			"failed to execute apt-cache search command")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	// Parse search results (format: package-name - description)
	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		// Extract package name (before the " - ")
		if idx := strings.Index(line, " - "); idx > 0 {
			packageName := strings.TrimSpace(line[:idx])
			if packageName != "" {
				packages = append(packages, packageName)
			}
		}
	}

	return packages, nil
}

// Info retrieves detailed information about a package.
func (a *AptManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	// Check if package is installed first
	installed, err := a.IsInstalled(ctx, name)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to check package installation status")
	}

	// Use apt-cache show to get package info
	cmd := exec.CommandContext(ctx, "apt-cache", "show", name)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if strings.Contains(string(exitError.Stderr), "No packages found") {
				return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
					fmt.Sprintf("package '%s' not found", name)).
					WithSuggestionMessage(fmt.Sprintf("Search available packages: apt-cache search %s", name))
			}
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get package info")
	}

	// Parse apt-cache show output
	info := &PackageInfo{
		Name:      name,
		Manager:   "apt",
		Installed: installed,
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Description:") || strings.HasPrefix(line, "Description-en:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "Description-en:"), "Description:"))
		} else if strings.HasPrefix(line, "Homepage:") {
			info.Homepage = strings.TrimSpace(strings.TrimPrefix(line, "Homepage:"))
		} else if strings.HasPrefix(line, "Depends:") {
			depends := strings.TrimSpace(strings.TrimPrefix(line, "Depends:"))
			if depends != "" {
				// Simple parsing of dependencies (doesn't handle versions)
				deps := strings.Split(depends, ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					// Remove version constraints
					if idx := strings.IndexAny(dep, " (<>="); idx > 0 {
						dep = dep[:idx]
					}
					if dep != "" {
						info.Dependencies = append(info.Dependencies, dep)
					}
				}
			}
		} else if strings.HasPrefix(line, "Installed-Size:") {
			info.InstalledSize = strings.TrimSpace(strings.TrimPrefix(line, "Installed-Size:")) + " kB"
		}
	}

	return info, nil
}

// GetInstalledVersion retrieves the installed version of a package.
func (a *AptManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	// First check if package is installed
	installed, err := a.IsInstalled(ctx, name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to check package installation status")
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("package '%s' is not installed", name))
	}

	// Use dpkg-query to get version
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", name)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
			fmt.Sprintf("could not extract version for package '%s'", name))
	}

	return version, nil
}
