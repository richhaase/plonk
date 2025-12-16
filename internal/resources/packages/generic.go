// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	jsonpath "github.com/PaesslerAG/jsonpath"

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

	if len(g.config.Available.Command) == 0 {
		return false, fmt.Errorf("no availability check configured for manager (missing 'available' config)")
	}

	cmd := g.config.Available.Command
	_, err := g.exec.Execute(ctx, cmd[0], cmd[1:]...)
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

// parseOutput parses command output based on strategy
func (g *GenericManager) parseOutput(data []byte, cfg config.ListConfig) ([]string, error) {
	trimmed := strings.TrimSpace(string(data))

	// Support both legacy "parse" and newer "parse_strategy" fields.
	parseMode := cfg.Parse
	if parseMode == "" {
		parseMode = cfg.ParseStrategy
	}

	var result []string
	var err error

	switch parseMode {
	case "lines", "":
		result = g.normalize(g.parseLines(data), cfg.Normalize)
	case "json":
		result, err = g.parseJSON(data, cfg.JSONField)
		if err == nil {
			result = g.normalize(result, cfg.Normalize)
		}
	case "json-map":
		result, err = g.parseJSONMap(data, cfg.JSONField)
		if err == nil {
			result = g.normalize(result, cfg.Normalize)
		}
	case "jsonpath":
		result, err = g.parseJSONPath(data, cfg)
		if err == nil {
			result = g.normalize(result, cfg.Normalize)
		}
	default:
		return nil, fmt.Errorf("unknown parse strategy: %s (use 'lines', 'json', 'json-map', or 'jsonpath')", parseMode)
	}

	if err != nil {
		return nil, err
	}

	if len(result) == 0 && trimmed != "" {
		return nil, fmt.Errorf("list output parsed with strategy %s but no package names were extracted", parseMode)
	}

	return result, nil
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

// parseJSONPath extracts keys and/or values using JSONPath selectors.
func (g *GenericManager) parseJSONPath(data []byte, cfg config.ListConfig) ([]string, error) {
	var root interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	names := make(map[string]struct{})
	addName := func(name string) {
		if name == "" {
			return
		}
		names[name] = struct{}{}
	}

	if cfg.KeysFrom != "" {
		res, err := jsonpath.Get(cfg.KeysFrom, root)
		if err != nil {
			return nil, fmt.Errorf("jsonpath keys_from %q: %w", cfg.KeysFrom, err)
		}
		items, ok := toInterfaceSlice(res)
		if ok {
			for _, item := range items {
				m, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("jsonpath keys_from %q expected object(s), got %T", cfg.KeysFrom, item)
				}
				for key := range m {
					addName(key)
				}
			}
		} else {
			m, ok := res.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("jsonpath keys_from %q did not resolve to object(s)", cfg.KeysFrom)
			}
			for key := range m {
				addName(key)
			}
		}
	}

	if cfg.ValuesFrom != "" {
		res, err := jsonpath.Get(cfg.ValuesFrom, root)
		if err != nil {
			return nil, fmt.Errorf("jsonpath values_from %q: %w", cfg.ValuesFrom, err)
		}
		items, ok := toInterfaceSlice(res)
		if ok {
			for _, item := range items {
				s, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf("jsonpath values_from %q expected string(s), got %T", cfg.ValuesFrom, item)
				}
				addName(s)
			}
		} else {
			s, ok := res.(string)
			if !ok {
				return nil, fmt.Errorf("jsonpath values_from %q did not resolve to string(s)", cfg.ValuesFrom)
			}
			addName(s)
		}
	}

	if len(names) == 0 {
		return []string{}, nil
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

// normalize applies optional normalization to package names.
func (g *GenericManager) normalize(values []string, mode string) []string {
	if mode == "lower" {
		out := make([]string, len(values))
		for i, v := range values {
			out[i] = strings.ToLower(v)
		}
		return out
	}
	return values
}

// toInterfaceSlice attempts to convert a value to []interface{} preserving nil detection.
func toInterfaceSlice(v interface{}) ([]interface{}, bool) {
	list, ok := v.([]interface{})
	if !ok {
		return nil, false
	}
	return list, true
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
