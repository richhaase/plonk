// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"debug/buildinfo"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func uninstallCommand(ctx context.Context, binary string, args ...string) error {
	cmd := exec.CommandContext(ctx, binary, args...)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %s: %w", binary, strings.Join(args, " "), strings.TrimSpace(string(result)), err)
	}
	return nil
}

func (b *BrewSimple) Uninstall(ctx context.Context, name string) error {
	if err := uninstallCommand(ctx, "brew", "uninstall", "--", name); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.installed = nil
	return nil
}

func (c *CargoSimple) Uninstall(ctx context.Context, name string) error {
	if err := uninstallCommand(ctx, "cargo", "uninstall", "--", name); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.installed = nil
	return nil
}

func (p *PNPMSimple) Uninstall(ctx context.Context, name string) error {
	if err := uninstallCommand(ctx, "pnpm", "remove", "-g", "--", pnpmPackageName(name)); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.installed = nil
	return nil
}

func (u *UVSimple) Uninstall(ctx context.Context, name string) error {
	if err := uninstallCommand(ctx, "uv", "tool", "uninstall", "--", name); err != nil {
		return err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.installed = nil
	return nil
}

// Go has no uninstall command. Remove the binary only when its build metadata
// matches the requested import path; a basename collision must not delete a
// different tool. Rooted operations keep the deletion inside the Go bin directory.
func (g *GoSimple) Uninstall(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	importPath, _, _ := strings.Cut(name, "@")
	binary := path.Base(importPath)
	if binary == "." || binary == ".." || binary == "/" || binary == "" || !strings.Contains(importPath, "/") {
		return fmt.Errorf("invalid Go package path: %s", name)
	}
	binDir := goBinDir()
	if binDir == "" {
		return fmt.Errorf("cannot determine Go bin directory")
	}
	root, err := os.OpenRoot(binDir)
	if err != nil {
		return err
	}
	defer root.Close()
	file, err := root.Open(binary)
	if err != nil {
		return err
	}
	info, err := buildinfo.Read(file)
	_ = file.Close()
	if err != nil {
		return fmt.Errorf("cannot verify installed Go binary %s: %w", binary, err)
	}
	if info.Path != importPath {
		return fmt.Errorf("installed binary %s belongs to %s, not %s", binary, info.Path, importPath)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := root.Remove(binary); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.installed = nil
	return nil
}
