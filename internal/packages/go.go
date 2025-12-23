// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoManager implements PackageManager for Go's 'go install' command.
type GoManager struct {
	BaseManager
}

// NewGoManager creates a new Go manager.
func NewGoManager(exec CommandExecutor) *GoManager {
	return &GoManager{
		BaseManager: NewBaseManager(exec, "go", "version"),
	}
}

// ListInstalled lists all binaries in GOBIN.
// Go doesn't have a native list command, so we scan the bin directory.
func (g *GoManager) ListInstalled(ctx context.Context) ([]string, error) {
	binDir := g.goBinDir()
	if binDir == "" {
		return []string{}, nil
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read GOBIN directory: %w", err)
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden files and common non-package files
		if strings.HasPrefix(name, ".") {
			continue
		}
		packages = append(packages, name)
	}
	return packages, nil
}

// Install installs a Go package using 'go install'.
// Package names should be in the form "github.com/user/repo@version" or just the binary name.
func (g *GoManager) Install(ctx context.Context, name string) error {
	// If no version specified, use @latest
	pkg := name
	if !strings.Contains(name, "@") {
		pkg = name + "@latest"
	}

	output, err := g.Exec().CombinedOutput(ctx, "go", "install", pkg)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\n%s", name, err, string(output))
	}
	return nil
}

// Uninstall removes a Go binary from GOBIN.
// Go doesn't have an uninstall command, so we just delete the binary.
func (g *GoManager) Uninstall(ctx context.Context, name string) error {
	binDir := g.goBinDir()
	if binDir == "" {
		return fmt.Errorf("could not determine GOBIN directory")
	}

	// Extract binary name from package path (e.g., "github.com/user/repo" -> "repo")
	binaryName := name
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		binaryName = parts[len(parts)-1]
	}
	// Remove @version suffix if present
	if idx := strings.Index(binaryName, "@"); idx != -1 {
		binaryName = binaryName[:idx]
	}

	binPath := filepath.Join(binDir, binaryName)
	err := os.Remove(binPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Already gone - idempotent success
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades packages by reinstalling with @latest.
func (g *GoManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return fmt.Errorf("go does not support upgrading all packages at once")
	}

	for _, pkg := range packages {
		// Force @latest for upgrades
		target := pkg
		if idx := strings.Index(target, "@"); idx != -1 {
			target = target[:idx]
		}
		target = target + "@latest"

		output, err := g.Exec().CombinedOutput(ctx, "go", "install", target)
		if err != nil {
			return fmt.Errorf("failed to upgrade %s: %w\n%s", pkg, err, string(output))
		}
	}
	return nil
}

// SelfInstall installs Go using the official installer or brew.
func (g *GoManager) SelfInstall(ctx context.Context) error {
	return g.SelfInstallWithBrewFallback(ctx, g.IsAvailable, "go", "",
		"automatic Go installation not supported; visit https://go.dev/dl/ to install",
	)
}

// goBinDir returns the directory where go install puts binaries.
func (g *GoManager) goBinDir() string {
	// Check GOBIN first
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin
	}

	// Fall back to GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Default GOPATH is ~/go
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		gopath = filepath.Join(home, "go")
	}

	return filepath.Join(gopath, "bin")
}
