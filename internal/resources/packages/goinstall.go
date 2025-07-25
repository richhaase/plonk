// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

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

// GoInstallManager manages Go packages.
type GoInstallManager struct {
	binary string
}

// NewGoInstallManager creates a new go install manager.
func NewGoInstallManager() *GoInstallManager {
	return &GoInstallManager{
		binary: "go",
	}
}

// getGoBinDir returns the directory where go installs binaries
func (g *GoInstallManager) getGoBinDir(ctx context.Context) (string, error) {
	// First try GOBIN
	cmd := exec.CommandContext(ctx, g.binary, "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOBIN: %w", err)
	}

	gobin := strings.TrimSpace(string(output))
	if gobin != "" {
		return gobin, nil
	}

	// Fall back to GOPATH/bin
	cmd = exec.CommandContext(ctx, g.binary, "env", "GOPATH")
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
		return nil, err
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

	checkCmd := exec.CommandContext(checkCtx, g.binary, "version", "-m", binaryPath)
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
	err := g.handleInstall(ctx, name)
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
	cmd := exec.CommandContext(ctx, g.binary, "version", "-m", binaryPath)
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

// InstalledVersion retrieves the installed version of a Go binary
func (g *GoInstallManager) InstalledVersion(ctx context.Context, name string) (string, error) {
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
	cmd := exec.CommandContext(ctx, g.binary, "version", "-m", binaryPath)
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

// IsAvailable checks if go is installed and accessible
func (g *GoInstallManager) IsAvailable(ctx context.Context) (bool, error) {
	if !CheckCommandAvailable(g.binary) {
		return false, nil
	}

	err := VerifyBinary(ctx, g.binary, []string{"version"})
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

// SupportsSearch returns false as go install doesn't support package search
func (g *GoInstallManager) SupportsSearch() bool {
	return false
}

// handleInstall handles the install operation for Go packages
func (g *GoInstallManager) handleInstall(ctx context.Context, name string) error {
	modulePath, version := parseModulePath(name)
	moduleSpec := fmt.Sprintf("%s@%s", modulePath, version)

	output, err := ExecuteCommandCombined(ctx, g.binary, "install", moduleSpec)
	if err != nil {
		return g.handleInstallError(err, output, name)
	}
	return nil
}

// handleInstallError processes install command errors
func (g *GoInstallManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := strings.ToLower(string(output))

	if exitCode, ok := ExtractExitCode(err); ok {
		// Check for known error patterns
		if strings.Contains(outputStr, "cannot find module") ||
			strings.Contains(outputStr, "cannot find package") ||
			strings.Contains(outputStr, "404 not found") ||
			strings.Contains(outputStr, "unknown revision") {
			return fmt.Errorf("package '%s' not found", packageName)
		}

		if strings.Contains(outputStr, "already exists") ||
			strings.Contains(outputStr, "already installed") {
			// Package is already installed - this is typically fine
			return nil
		}

		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "access is denied") {
			return fmt.Errorf("permission denied installing %s", packageName)
		}

		if strings.Contains(outputStr, "database is locked") ||
			strings.Contains(outputStr, "lock") && strings.Contains(outputStr, "held") {
			return fmt.Errorf("package manager database is locked")
		}

		if strings.Contains(outputStr, "network error") ||
			strings.Contains(outputStr, "cannot download") ||
			strings.Contains(outputStr, "connection refused") ||
			strings.Contains(outputStr, "timeout") {
			return fmt.Errorf("network error during installation")
		}

		if strings.Contains(outputStr, "build failed") ||
			strings.Contains(outputStr, "compilation error") ||
			strings.Contains(outputStr, "cannot compile") ||
			strings.Contains(outputStr, "build constraints exclude") {
			return fmt.Errorf("failed to build package '%s'", packageName)
		}

		if strings.Contains(outputStr, "incompatible") ||
			strings.Contains(outputStr, "version conflict") ||
			strings.Contains(outputStr, "requires go") {
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)
		}

		// Only treat non-zero exit codes as errors
		if exitCode != 0 {
			return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
		}
		// Exit code 0 with no recognized error pattern - success
		return nil
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute install command: %w", err)
}
