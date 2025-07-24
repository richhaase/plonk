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
)

// ExtractBinaryNameFromPath extracts the binary name from a Go module path
// Examples:
//   - github.com/user/tool -> tool
//   - github.com/user/project/cmd/tool -> tool
//   - github.com/user/tool@v1.2.3 -> tool
func ExtractBinaryNameFromPath(modulePath string) string {
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

// GoInstallManager manages Go packages using BaseManager for common functionality.
type GoInstallManager struct {
	*BaseManager
}

// NewGoInstallManager creates a new go install manager.
func NewGoInstallManager() *GoInstallManager {
	return newGoInstallManager()
}

// newGoInstallManager creates a go install manager.
func newGoInstallManager() *GoInstallManager {
	config := ManagerConfig{
		BinaryName:  "go",
		VersionArgs: []string{"version"},
		ListArgs: func() []string {
			return []string{"env", "GOBIN"}
		},
		InstallArgs: func(pkg string) []string {
			modulePath, version := parseModulePath(pkg)
			moduleSpec := fmt.Sprintf("%s@%s", modulePath, version)
			return []string{"install", moduleSpec}
		},
		UninstallArgs: func(pkg string) []string {
			// go doesn't have uninstall, we'll handle this manually
			return []string{"version", "-m"}
		},
	}

	// Add go-specific error patterns
	errorMatcher := NewCommonErrorMatcher()
	errorMatcher.AddPattern(ErrorTypeNotFound, "cannot find module", "no matching versions", "malformed module path")
	errorMatcher.AddPattern(ErrorTypeNetwork, "connection", "timeout")
	errorMatcher.AddPattern(ErrorTypeBuild, "build failed", "compilation")

	base := NewBaseManager(config)
	base.ErrorMatcher = errorMatcher

	return &GoInstallManager{
		BaseManager: base,
	}
}

// IsAvailable checks if go is installed and has a supported version.
func (g *GoInstallManager) IsAvailable(ctx context.Context) (bool, error) {
	// Use BaseManager's IsAvailable but add version check
	available, err := g.BaseManager.IsAvailable(ctx)
	if !available || err != nil {
		return available, err
	}

	// Check if version is >= 1.16 (when go install was improved)
	cmd := exec.CommandContext(ctx, g.GetBinary(), "version")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("go version check failed: %w", err)
	}

	versionStr := string(output)
	if !strings.Contains(versionStr, "go1.") {
		return false, fmt.Errorf("unsupported go version: Go 1.16 or later required")
	}

	return true, nil
}

// getGoBinDir returns the directory where go installs binaries
func (g *GoInstallManager) getGoBinDir(ctx context.Context) (string, error) {
	// First try GOBIN
	cmd := exec.CommandContext(ctx, g.GetBinary(), "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOBIN: %w", err)
	}

	gobin := strings.TrimSpace(string(output))
	if gobin != "" {
		return gobin, nil
	}

	// Fall back to GOPATH/bin
	cmd = exec.CommandContext(ctx, g.GetBinary(), "env", "GOPATH")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOPATH: %w", err)
	}

	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		// Use default
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
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
		return nil, fmt.Errorf("failed to read GOBIN directory: %w", err)
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

	checkCmd := exec.CommandContext(checkCtx, g.GetBinary(), "version", "-m", binaryPath)
	_, err := checkCmd.Output()
	// If go version -m succeeds, it's a Go binary
	return err == nil
}

// parseModulePath extracts the module path from a package specification
func parseModulePath(pkg string) (modulePath string, version string) {
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
	err := g.ExecuteInstall(ctx, name)
	if err != nil {
		return err
	}

	// Check if GOBIN is in PATH and warn if not
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
	binaryName := extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Binary doesn't exist - this is fine for uninstall
		return nil
	}

	// Verify it's a Go binary before removing
	if !g.isGoBinary(ctx, binaryPath) {
		return fmt.Errorf("'%s' is not a Go binary", binaryName)
	}

	// Remove the binary
	err = os.Remove(binaryPath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied removing %s", binaryName)
		}
		return fmt.Errorf("failed to remove binary %s: %w", name, err)
	}

	return nil
}

// extractBinaryName extracts the binary name from a module path
func extractBinaryName(modulePath string) string {
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

	binaryName := extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if file exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return false, nil
	}

	// Verify it's a Go binary
	return g.isGoBinary(ctx, binaryPath), nil
}

// SupportsSearch returns false as Go doesn't have a built-in package search command.
func (g *GoInstallManager) SupportsSearch() bool {
	return false
}

// Search searches for Go modules.
func (g *GoInstallManager) Search(ctx context.Context, query string) ([]string, error) {
	// Go doesn't have a built-in search command
	return nil, fmt.Errorf("go does not have a built-in search command. Search for Go packages at https://pkg.go.dev/search?q=%s", query)
}

// Info retrieves detailed information about a Go binary.
func (g *GoInstallManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	binDir, err := g.getGoBinDir(ctx)
	if err != nil {
		return nil, err
	}

	binaryName := extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return nil, err
	}

	if !installed {
		return nil, fmt.Errorf("binary '%s' not found", binaryName)
	}

	// Get module information using go version -m
	cmd := exec.CommandContext(ctx, g.GetBinary(), "version", "-m", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get module information for %s: %w", name, err)
	}

	info := &PackageInfo{
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

	binaryName := extractBinaryName(name)
	binaryPath := filepath.Join(binDir, binaryName)

	// Check if installed
	installed, err := g.IsInstalled(ctx, name)
	if err != nil {
		return "", err
	}
	if !installed {
		return "", fmt.Errorf("binary '%s' is not installed", binaryName)
	}

	// Get version using go version -m
	cmd := exec.CommandContext(ctx, g.GetBinary(), "version", "-m", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version information for %s: %w", name, err)
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

	return "", fmt.Errorf("could not extract version for binary '%s'", binaryName)
}
