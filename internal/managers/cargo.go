package managers

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// CargoManager manages cargo packages.
type CargoManager struct{}

// NewCargoManager creates a new cargo manager.
func NewCargoManager() *CargoManager {
	return &CargoManager{}
}

// IsAvailable checks if cargo is installed and accessible.
func (c *CargoManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("cargo")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify cargo is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "cargo", "--version")
	err = cmd.Run()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// cargo exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "cargo binary found but not functional")
	}
	return true, nil
}

// ListInstalled lists all installed cargo packages.
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list")
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "list-installed", "failed to list installed cargo packages")
	}

	lines := strings.Split(string(out), "\n")
	var packages []string
	for _, line := range lines {
		if strings.HasSuffix(line, ":") {
			continue // Skip header lines
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			packages = append(packages, fields[0])
		}
	}

	return packages, nil
}

// Install installs a cargo package.
func (c *CargoManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "cargo", "install", name)
	if err := cmd.Run(); err != nil {
		return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name, "failed to install cargo package")
	}
	return nil
}

// Uninstall removes a cargo package.
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "cargo", "uninstall", name)
	if err := cmd.Run(); err != nil {
		return errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", name, "failed to uninstall cargo package")
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
	cmd := exec.CommandContext(ctx, "cargo", "search", query)
	out, err := cmd.Output()
	if err != nil {
		// cargo search returns a non-zero exit code if no packages are found.
		if strings.Contains(string(out), "no crates found") {
			return []string{}, nil
		}
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "search", query, "failed to search for cargo package")
	}

	var packages []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 {
			packages = append(packages, fields[0])
		}
	}

	return packages, nil
}

// Info retrieves detailed information about a package.
func (c *CargoManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	cmd := exec.CommandContext(ctx, "cargo", "search", name, "--limit", "1")
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name, "failed to get info for cargo package")
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if !scanner.Scan() {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info", fmt.Sprintf("package '%s' not found", name)).WithSuggestionMessage(fmt.Sprintf("Search available packages: cargo search %s", name))
	}

	line := scanner.Text()
	fields := strings.SplitN(line, " = ", 2)
	if len(fields) < 2 {
		return nil, errors.NewError(errors.ErrInternal, errors.DomainPackages, "info", fmt.Sprintf("unexpected output from cargo search for package '%s'", name))
	}

	packageName := strings.TrimSpace(fields[0])
	if packageName != name {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info", fmt.Sprintf("package '%s' not found", name)).WithSuggestionMessage(fmt.Sprintf("Search available packages: cargo search %s", name))
	}

	var description string
	if len(fields) > 1 {
		description = strings.Trim(fields[1], `"`)
	}

	installed, err := c.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	return &PackageInfo{
		Name:        name,
		Description: description,
		Installed:   installed,
		Manager:     "cargo",
	}, nil
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
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get package version information")
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("no version information found for package '%s'", name))
	}

	// Parse output to find the package
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines like "packagename v1.2.3:"
		if strings.HasPrefix(line, name+" v") {
			// Extract version from "packagename v1.2.3:"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				version := parts[1]
				// Remove 'v' prefix and trailing colon if present
				if strings.HasPrefix(version, "v") {
					version = version[1:]
				}
				if strings.HasSuffix(version, ":") {
					version = version[:len(version)-1]
				}
				return version, nil
			}
		}
	}

	return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
		fmt.Sprintf("could not find version information for package '%s'", name))
}
