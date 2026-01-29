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
)

// GoSimple implements Manager for Go packages
type GoSimple struct{}

// NewGoSimple creates a new Go manager
func NewGoSimple() *GoSimple {
	return &GoSimple{}
}

// IsInstalled checks if a go package is installed by looking for its binary
func (g *GoSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	binDir := goBinDir()
	if binDir == "" {
		return false, nil
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

	binPath := filepath.Join(binDir, binaryName)
	_, err := os.Stat(binPath)
	return err == nil, nil
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
	return nil
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
