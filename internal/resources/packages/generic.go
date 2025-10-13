// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
)

// GenericManager implements PackageManager using config from plonk.yaml
type GenericManager struct {
	config config.ManagerConfig
	exec   CommandExecutor
}

// NewGenericManager creates a new generic manager from config
func NewGenericManager(cfg config.ManagerConfig, exec CommandExecutor) *GenericManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return &GenericManager{
		config: cfg,
		exec:   exec,
	}
}

// IsAvailable checks if the package manager is available
func (g *GenericManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := g.exec.LookPath(g.config.Binary); err != nil {
		return false, nil
	}

	_, err := g.exec.Execute(ctx, g.config.Binary, "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all installed packages
func (g *GenericManager) ListInstalled(ctx context.Context) ([]string, error) {
	if len(g.config.List.Command) == 0 {
		return []string{}, nil
	}

	args := g.config.List.Command[1:]
	output, err := g.exec.Execute(ctx, g.config.Binary, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return g.parseOutput(output, g.config.List)
}

// Install installs a package (idempotent)
func (g *GenericManager) Install(ctx context.Context, name string) error {
	if len(g.config.Install.Command) == 0 {
		return fmt.Errorf("install command not configured for this manager")
	}

	cmd := g.expandTemplate(g.config.Install.Command, name)
	output, err := g.exec.CombinedOutput(ctx, cmd[0], cmd[1:]...)

	// Treat "already installed" even on success as idempotent no-op
	if err == nil && g.isIdempotentError(string(output), g.config.Install.IdempotentErrors) {
		return nil
	}

	if err != nil && g.isIdempotentError(string(output), g.config.Install.IdempotentErrors) {
		return nil
	}

	return err
}

// Upgrade upgrades packages (implements PackageUpgrader)
func (g *GenericManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return g.UpgradeAll(ctx)
	}

	for _, pkg := range packages {
		cmd := g.expandTemplate(g.config.Upgrade.Command, pkg)
		output, err := g.exec.CombinedOutput(ctx, cmd[0], cmd[1:]...)

		// Treat "already up-to-date" even on success as idempotent no-op
		if err == nil && g.isIdempotentError(string(output), g.config.Upgrade.IdempotentErrors) {
			continue
		}

		if err != nil && !g.isIdempotentError(string(output), g.config.Upgrade.IdempotentErrors) {
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}

	return nil
}

// UpgradeAll upgrades all packages from this manager
func (g *GenericManager) UpgradeAll(ctx context.Context) error {
	if len(g.config.UpgradeAll.Command) == 0 {
		return fmt.Errorf("upgrade all not supported for this manager")
	}

	output, err := g.exec.CombinedOutput(ctx, g.config.UpgradeAll.Command[0], g.config.UpgradeAll.Command[1:]...)

	// Treat "already up-to-date" even on success as idempotent no-op
	if err == nil && g.isIdempotentError(string(output), g.config.UpgradeAll.IdempotentErrors) {
		return nil
	}

	if err != nil && g.isIdempotentError(string(output), g.config.UpgradeAll.IdempotentErrors) {
		return nil
	}

	return err
}

// Uninstall removes a package (idempotent)
func (g *GenericManager) Uninstall(ctx context.Context, name string) error {
	cmd := g.expandTemplate(g.config.Uninstall.Command, name)
	output, err := g.exec.CombinedOutput(ctx, cmd[0], cmd[1:]...)

	if err != nil && g.isIdempotentError(string(output), g.config.Uninstall.IdempotentErrors) {
		return nil
	}

	return err
}

// IsInstalled checks if a package is installed
func (g *GenericManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	installed, err := g.ListInstalled(ctx)
	if err != nil {
		return false, err
	}

	for _, pkg := range installed {
		if pkg == name {
			return true, nil
		}
	}

	return false, nil
}

// parseOutput parses command output based on strategy
func (g *GenericManager) parseOutput(data []byte, cfg config.ListConfig) ([]string, error) {
	switch cfg.Parse {
	case "lines", "":
		return g.parseLines(data), nil
	case "json":
		return g.parseJSON(data, cfg.JSONField)
	default:
		return nil, fmt.Errorf("unknown parse strategy: %s (use 'lines' or 'json')", cfg.Parse)
	}
}

// parseLines splits by newlines and filters empty
func (g *GenericManager) parseLines(data []byte) []string {
	result := strings.TrimSpace(string(data))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Special handling for managers that return parseable filesystem paths
		// instead of package names (e.g., npm/pnpm with --parseable).
		if g.config.Binary == "npm" || g.config.Binary == "pnpm" {
			// Many npm/pnpm outputs include paths containing node_modules/...
			// Extract the package name (including scope when present).
			const marker = "node_modules/"
			if idx := strings.LastIndex(line, marker); idx != -1 {
				rest := line[idx+len(marker):]
				// For scoped packages, the next two segments form the name: @scope/name
				segs := strings.Split(rest, "/")
				if len(segs) > 0 {
					if strings.HasPrefix(segs[0], "@") {
						if len(segs) >= 2 {
							packages = append(packages, segs[0]+"/"+segs[1])
							continue
						}
						// Fallback: if incomplete, keep the segment as-is
						packages = append(packages, segs[0])
						continue
					}
					packages = append(packages, segs[0])
					continue
				}
			}
		}

		// Default: take the first whitespace-delimited token
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}

	return packages
}

// parseJSON extracts field from JSON array
func (g *GenericManager) parseJSON(data []byte, field string) ([]string, error) {
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var result []string
	for _, item := range items {
		if val, ok := item[field].(string); ok {
			result = append(result, val)
		}
	}
	return result, nil
}

// expandTemplate replaces template variables in command
func (g *GenericManager) expandTemplate(cmd []string, packageName string) []string {
	result := make([]string, len(cmd))
	for i, part := range cmd {
		result[i] = strings.ReplaceAll(part, "{{.Package}}", packageName)
	}
	return result
}

// isIdempotentError checks if error message matches idempotent patterns
func (g *GenericManager) isIdempotentError(output string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	outputLower := strings.ToLower(output)
	for _, pattern := range patterns {
		if strings.Contains(outputLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
