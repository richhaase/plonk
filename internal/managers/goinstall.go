// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/interfaces"
)

// GoInstallManager manages Go packages installed with go install.
type GoInstallManager struct{}

// NewGoInstallManager creates a new go install manager.
func NewGoInstallManager() *GoInstallManager {
	return &GoInstallManager{}
}

// IsAvailable checks if go is installed and accessible.
func (g *GoInstallManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("go")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}

	// Verify go is actually functional and check version
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.Output()
	if err != nil {
		// If the command fails due to context cancellation, return the context error
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		// go exists but is not functional - this is an error
		return false, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "go binary found but not functional")
	}

	// Check if version is >= 1.16 (when go install was improved)
	versionStr := string(output)
	if !strings.Contains(versionStr, "go1.") {
		return false, errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "check",
			"unsupported go version").WithSuggestionMessage("Go 1.16 or later required")
	}

	return true, nil
}

// getGoBinDir returns the directory where go installs binaries
func (g *GoInstallManager) getGoBinDir(ctx context.Context) (string, error) {
	// First try GOBIN
	cmd := exec.CommandContext(ctx, "go", "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "config",
			"failed to get GOBIN")
	}

	gobin := strings.TrimSpace(string(output))
	if gobin != "" {
		return gobin, nil
	}

	// Fall back to GOPATH/bin
	cmd = exec.CommandContext(ctx, "go", "env", "GOPATH")
	output, err = cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "config",
			"failed to get GOPATH")
	}

	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		// Use default
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, errors.ErrFileIO, errors.DomainPackages, "config",
				"failed to get home directory")
		}
		gopath = filepath.Join(home, "go")
	}

	return filepath.Join(gopath, "bin"), nil
}

// ListInstalled lists all Go binaries installed with go install.
func (g *GoInstallManager) ListInstalled(ctx context.Context) ([]string, error) {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		// No bin directory means no installed packages
		return []string{}, nil
	}

	// List all files in the bin directory
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainPackages, "list",
			"failed to read GOBIN directory")
	}

	var goBinaries []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a Go binary using go version -m
		binaryPath := filepath.Join(binDir, entry.Name())
		if g.isGoBinary(ctx, binaryPath) {
			goBinaries = append(goBinaries, entry.Name())
		}
	}

	return goBinaries, nil
}

// isGoBinary checks if a file is a Go binary using go version -m
func (g *GoInstallManager) isGoBinary(ctx context.Context, binaryPath string) bool {
	// Use a short timeout for this check
	checkCtx, cancel := context.WithTimeout(ctx, 2*1000*1000*1000) // 2 seconds
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "go", "version", "-m", binaryPath)
	err := cmd.Run()
	// If go version -m succeeds, it's a Go binary
	return err == nil
}

// parseModulePath extracts the module path from a package specification
func (g *GoInstallManager) parseModulePath(pkg string) (modulePath string, version string) {
	// Handle version specification (e.g., package@version)
	parts := strings.Split(pkg, "@")
	modulePath = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	} else {
		version = "latest"
	}
	return modulePath, version
}

// Install installs a Go package.
func (g *GoInstallManager) Install(ctx context.Context, name string) error {
	modulePath, version := g.parseModulePath(name)

	// Construct the full module specification
	moduleSpec := fmt.Sprintf("%s@%s", modulePath, version)

	cmd := exec.CommandContext(ctx, "go", "install", moduleSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		// Check for specific error conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for module not found
			if strings.Contains(outputStr, "cannot find module") ||
				strings.Contains(outputStr, "no matching versions") ||
				strings.Contains(outputStr, "malformed module path") {
				return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "install",
					fmt.Sprintf("module '%s' not found", modulePath)).
					WithSuggestionMessage(fmt.Sprintf("Search on pkg.go.dev or verify module path"))
			}

			// Check for network errors
			if strings.Contains(outputStr, "connection") ||
				strings.Contains(outputStr, "timeout") {
				return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "install",
					"network error during installation").
					WithSuggestionMessage("Check internet connection and proxy settings")
			}

			// Check for build errors
			if strings.Contains(outputStr, "build failed") ||
				strings.Contains(outputStr, "compilation") {
				return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "install",
					fmt.Sprintf("failed to build %s", modulePath)).
					WithSuggestionMessage("Module may have build dependencies or compatibility issues")
			}

			// Other exit errors with more context
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
				fmt.Sprintf("go install failed (exit code %d)", exitError.ExitCode()))
		}

		// Non-exit errors (command not found, context cancellation, etc.)
		return errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "install", name,
			"failed to execute go install command")
	}

	// Check if GOBIN is in PATH
	binDir, err := g.getGoBinDir(ctx)
	if err == nil {
		path := os.Getenv("PATH")
		if !strings.Contains(path, binDir) {
			// Just a warning, not an error
			fmt.Fprintf(os.Stderr, "Warning: %s is not in PATH. You may need to add it to use installed tools.\n", binDir)
		}
	}

	return nil
}

// Uninstall removes a Go binary.
func (g *GoInstallManager) Uninstall(ctx context.Context, name string) error {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return err
	}

	// Extract binary name from module path if needed
	binaryName := g.extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Binary doesn't exist - this is fine for uninstall
		return nil
	}

	// Verify it's a Go binary before removing
	if !g.isGoBinary(ctx, binaryPath) {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "uninstall",
			fmt.Sprintf("'%s' is not a Go binary", binaryName)).
			WithSuggestionMessage("Only Go binaries installed with 'go install' can be managed")
	}

	// Remove the binary
	err = os.Remove(binaryPath)
	if err != nil {
		if os.IsPermission(err) {
			return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "uninstall",
				fmt.Sprintf("permission denied removing %s", binaryName)).
				WithSuggestionMessage("Check file permissions")
		}
		return errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainPackages, "uninstall", name,
			"failed to remove binary")
	}

	return nil
}

// extractBinaryName extracts the binary name from a module path
func (g *GoInstallManager) extractBinaryName(modulePath string) string {
	// Remove version specification if present
	modulePath = strings.Split(modulePath, "@")[0]

	// Extract the last component of the path
	parts := strings.Split(modulePath, "/")
	binaryName := parts[len(parts)-1]

	// Handle special case of .../cmd/toolname pattern
	if len(parts) >= 2 && parts[len(parts)-2] == "cmd" {
		return binaryName
	}

	// For simple cases, the binary name is usually the last component
	return binaryName
}

// IsInstalled checks if a specific Go binary is installed.
func (g *GoInstallManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return false, err
	}

	binaryName := g.extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if file exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return false, nil
	}

	// Verify it's a Go binary
	return g.isGoBinary(ctx, binaryPath), nil
}

// Search searches for Go modules.
func (g *GoInstallManager) Search(ctx context.Context, query string) ([]string, error) {
	// Go doesn't have a built-in search command
	// Return a helpful message
	return []string{}, errors.NewError(errors.ErrUnsupported, errors.DomainPackages, "search",
		"go does not have a built-in search command").
		WithSuggestionMessage(fmt.Sprintf("Search for Go packages at https://pkg.go.dev/search?q=%s", query))
}

// Info retrieves detailed information about a Go binary.
func (g *GoInstallManager) Info(ctx context.Context, name string) (*interfaces.PackageInfo, error) {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return nil, err
	}

	binaryName := g.extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	if !installed {
		return nil, errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "info",
			fmt.Sprintf("binary '%s' not found", binaryName)).
			WithSuggestionMessage(fmt.Sprintf("Install with: plonk install --go %s", name))
	}

	// Get module information using go version -m
	cmd := exec.CommandContext(ctx, "go", "version", "-m", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "info", name,
			"failed to get module information")
	}

	info := &interfaces.PackageInfo{
		Name:      binaryName,
		Manager:   "go",
		Installed: true,
	}

	// Parse the output to extract module path and version
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mod\t") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				info.Homepage = fmt.Sprintf("https://pkg.go.dev/%s", parts[1])
				info.Version = parts[2]
			}
		} else if strings.HasPrefix(line, "dep\t") {
			// Extract dependencies
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.Dependencies = append(info.Dependencies, parts[1])
			}
		}
	}

	// If we found a module path, get the description from pkg.go.dev
	if info.Homepage != "" {
		info.Description = fmt.Sprintf("Go module: %s", strings.TrimPrefix(info.Homepage, "https://pkg.go.dev/"))
	}

	return info, nil
}

// GetInstalledVersion retrieves the installed version of a Go binary
func (g *GoInstallManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return "", err
	}

	binaryName := g.extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return "", err
	}
	if !installed {
		return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "version",
			fmt.Sprintf("binary '%s' is not installed", binaryName))
	}

	// Get version using go version -m
	cmd := exec.CommandContext(ctx, "go", "version", "-m", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.WrapWithItem(err, errors.ErrCommandExecution, errors.DomainPackages, "version", name,
			"failed to get version information")
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mod\t") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}

	// Try to extract version from first line (for older binaries)
	if len(lines) > 0 {
		// Extract version from format like "toolname: go1.21.5"
		if match := regexp.MustCompile(`go\d+\.\d+\.\d+`).FindString(lines[0]); match != "" {
			return match, nil
		}
	}

	return "", errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "version",
		fmt.Sprintf("could not extract version for binary '%s'", binaryName))
}
