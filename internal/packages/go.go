// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// GoSimple implements Manager for Go packages
type GoSimple struct {
	mu        sync.Mutex
	installed map[string]bool
}

// NewGoSimple creates a new Go manager
func NewGoSimple() *GoSimple {
	return &GoSimple{}
}

// IsInstalled checks if a go package is installed by looking for its binary
func (g *GoSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Load installed list on first call
	if g.installed == nil {
		if err := g.loadInstalled(); err != nil {
			return false, err
		}
	}

	// Extract binary name from package path
	// e.g., "golang.org/x/tools/gopls" -> "gopls"
	binaryName := name
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		binaryName = parts[len(parts)-1]
	}
	// Remove @version suffix if present
	if idx := strings.Index(binaryName, "@"); idx != -1 {
		binaryName = binaryName[:idx]
	}

	return g.installed[binaryName], nil
}

// loadInstalled scans the Go bin directory for installed binaries
func (g *GoSimple) loadInstalled() error {
	installed := make(map[string]bool)

	binDir := goBinDir()
	if binDir == "" {
		return fmt.Errorf("failed to determine go bin directory: GOBIN not set and home directory unavailable")
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Go bin directory doesn't exist yet - this is normal for fresh installs
			// Only set the cache after successful loading
			g.installed = installed
			return nil
		}
		return fmt.Errorf("failed to read go bin directory %s: %w", binDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			installed[entry.Name()] = true
		}
	}

	// Only set the cache after successful loading
	g.installed = installed
	return nil
}

// Install installs a go package
func (g *GoSimple) Install(ctx context.Context, name string) error {
	// Add @latest if no version specified
	pkg := name
	if !strings.Contains(name, "@") {
		pkg = name + "@latest"
	}

	cmd := exec.CommandContext(ctx, "go", "install", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go install failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// Update cache after successful install
	g.markInstalled(name)
	return nil
}

// markInstalled updates the cache to mark a package as installed
func (g *GoSimple) markInstalled(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.installed != nil {
		g.installed[name] = true
	}
}

// goBinDir returns the directory where go install puts binaries
func goBinDir() string {
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		gopath = filepath.Join(home, "go")
	}

	return filepath.Join(gopath, "bin")
}
