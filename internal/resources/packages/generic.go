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

	if len(g.config.Upgrade.Command) == 0 {
		return fmt.Errorf("upgrade command not configured for this manager")
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
	// Support both legacy "parse" and newer "parse_strategy" fields.
	parseMode := cfg.Parse
	if parseMode == "" {
		parseMode = cfg.ParseStrategy
	}

	switch parseMode {
	case "lines", "":
		return g.parseLines(data), nil
	case "json":
		return g.parseJSON(data, cfg.JSONField)
	case "json-map":
		return g.parseJSONMap(data, cfg.JSONField)
	default:
		return nil, fmt.Errorf("unknown parse strategy: %s (use 'lines', 'json', or 'json-map')", parseMode)
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

// parseJSONMap extracts keys from a top-level JSON object or a nested object
// specified by field. This is a prototype for managers like npm that return
// package names as map keys instead of array elements.
func (g *GenericManager) parseJSONMap(data []byte, field string) ([]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON map: %w", err)
	}

	// If a field is specified, drill down into that map.
	obj := raw
	if field != "" {
		val, ok := raw[field]
		if !ok {
			return []string{}, nil
		}
		nested, ok := val.(map[string]interface{})
		if !ok {
			return []string{}, nil
		}
		obj = nested
	}

	var result []string
	for key := range obj {
		result = append(result, key)
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
