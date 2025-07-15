// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/executor"
)

// AptManager manages APT packages using BaseManager for common functionality.
type AptManager struct {
	*BaseManager
}

// NewAptManager creates a new apt manager with the default executor.
func NewAptManager() *AptManager {
	config := ManagerConfig{
		BinaryName:  "apt",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"list", "--installed"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", "-y", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"remove", "-y", pkg}
		},
		PreferJSON: false,
	}

	// Add apt-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "Unable to locate package", "E: Package", "has no installation candidate")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "is already the newest version", "already installed")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "is not installed")
	errorMatcher.AddPattern(ErrorTypePermission, "Permission denied", "you don't have enough privileges")
	errorMatcher.AddPattern(ErrorTypeLocked, "Could not get lock", "dpkg: error", "database is locked")
	errorMatcher.AddPattern(ErrorTypeDependency, "Depends:", "but it is not installable", "Broken packages")

	base := NewBaseManager(config)
	base.ErrorMatcher = errorMatcher

	return &AptManager{
		BaseManager: base,
	}
}

// NewAptManagerWithExecutor creates a new apt manager with a custom executor for testing.
func NewAptManagerWithExecutor(exec executor.CommandExecutor) *AptManager {
	config := ManagerConfig{
		BinaryName:  "apt",
		VersionArgs: []string{"--version"},
		ListArgs: func() []string {
			return []string{"list", "--installed"}
		},
		InstallArgs: func(pkg string) []string {
			return []string{"install", "-y", pkg}
		},
		UninstallArgs: func(pkg string) []string {
			return []string{"remove", "-y", pkg}
		},
		PreferJSON: false,
	}

	// Add apt-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "Unable to locate package", "E: Package", "has no installation candidate")
	errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "is already the newest version", "already installed")
	errorMatcher.AddPattern(ErrorTypeNotInstalled, "is not installed")
	errorMatcher.AddPattern(ErrorTypePermission, "Permission denied", "you don't have enough privileges")
	errorMatcher.AddPattern(ErrorTypeLocked, "Could not get lock", "dpkg: error", "database is locked")
	errorMatcher.AddPattern(ErrorTypeDependency, "Depends:", "but it is not installable", "Broken packages")

	base := NewBaseManagerWithExecutor(config, exec)
	base.ErrorMatcher = errorMatcher

	return &AptManager{
		BaseManager: base,
	}
}

// IsAvailable checks if APT is installed and this is a Debian-based system.
func (a *AptManager) IsAvailable(ctx context.Context) (bool, error) {
	// APT is only available on Linux
	if runtime.GOOS != "linux" {
		return false, nil
	}

	// Use BaseManager's IsAvailable but add OS check
	return a.BaseManager.IsAvailable(ctx)
}

// ListInstalled lists all installed APT packages.
func (a *AptManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := a.ExecuteList(ctx)
	if err != nil {
		return nil, err
	}

	return a.parseListOutput(output), nil
}

// parseListOutput parses apt list --installed output
func (a *AptManager) parseListOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "WARNING:") || strings.HasPrefix(line, "Listing...") {
			continue
		}

		// APT list format: "package/repository version architecture [status]"
		parts := strings.Fields(line)
		if len(parts) > 0 {
			// Extract package name (remove /repository suffix if present)
			packageName := parts[0]
			if slashIndex := strings.Index(packageName, "/"); slashIndex > 0 {
				packageName = packageName[:slashIndex]
			}
			packages = append(packages, packageName)
		}
	}

	return packages
}

// Install installs an APT package.
func (a *AptManager) Install(ctx context.Context, name string) error {
	return a.ExecuteInstall(ctx, name)
}

// Uninstall removes an APT package.
func (a *AptManager) Uninstall(ctx context.Context, name string) error {
	return a.ExecuteUninstall(ctx, name)
}

// IsInstalled checks if a specific package is installed.
func (a *AptManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	packages, err := a.ListInstalled(ctx)
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

// Search searches for packages in APT repositories.
func (a *AptManager) Search(ctx context.Context, query string) ([]string, error) {
	output, err := a.Executor.Execute(ctx, a.GetBinary(), "search", query)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query,
			"failed to search apt packages")
	}

	return a.parseSearchOutput(output), nil
}

// parseSearchOutput parses apt search output
func (a *AptManager) parseSearchOutput(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" || strings.Contains(result, "No packages found") {
		return []string{}
	}

	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "WARNING:") {
			continue
		}

		// APT search format: "package/repository version architecture"
		// Description lines start with space, package lines don't
		if !strings.HasPrefix(line, " ") { // Package name lines don't start with space
			parts := strings.Fields(trimmed)
			if len(parts) > 0 {
				packageName := parts[0]
				if slashIndex := strings.Index(packageName, "/"); slashIndex > 0 {
					packageName = packageName[:slashIndex]
				}
				packages = append(packages, packageName)
			}
		}
	}

	return packages
}

// Info retrieves detailed information about a package.
func (a *AptManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	output, err := a.Executor.Execute(ctx, a.GetBinary(), "show", name)
	if err != nil {
		if execErr, ok := err.(interface{ ExitCode() int }); ok && execErr.ExitCode() == 100 {
			return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
				fmt.Sprintf("package '%s' not found", name)).
				WithSuggestionMessage(fmt.Sprintf("Search available packages: apt search %s", name))
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get package info")
	}

	// Check if installed
	installed, err := a.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	info := a.parseInfoOutput(output, name)
	info.Manager = "apt"
	info.Installed = installed

	return info, nil
}

// parseInfoOutput parses apt show output
func (a *AptManager) parseInfoOutput(output []byte, name string) *PackageInfo {
	info := &PackageInfo{
		Name: name,
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Package:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Name = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Version = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Description:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Description = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Homepage:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Homepage = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Depends:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				deps := strings.TrimSpace(parts[1])
				// Parse dependencies (comma-separated, may have version constraints)
				for _, dep := range strings.Split(deps, ",") {
					dep = strings.TrimSpace(dep)
					// Remove version constraints like (>= 1.0)
					if parenIndex := strings.Index(dep, " ("); parenIndex > 0 {
						dep = dep[:parenIndex]
					}
					if dep != "" {
						info.Dependencies = append(info.Dependencies, dep)
					}
				}
			}
		} else if strings.HasPrefix(line, "Installed-Size:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.InstalledSize = strings.TrimSpace(parts[1])
			}
		}
	}

	return info
}

// GetInstalledVersion retrieves the installed version of a package
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

	// Get version using apt show
	output, err := a.Executor.Execute(ctx, a.GetBinary(), "show", name)
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	version := a.extractVersion(output)
	if version == "" {
		return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
			fmt.Sprintf("could not extract version for package '%s'", name))
	}

	return version, nil
}

// extractVersion extracts version from apt show output
func (a *AptManager) extractVersion(output []byte) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
