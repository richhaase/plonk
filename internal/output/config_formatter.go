// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"slices"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"gopkg.in/yaml.v3"
)

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string
	Config     *config.Config
}

// ConfigShowFormatter formats config show output
type ConfigShowFormatter struct {
	Data ConfigShowOutput
}

// NewConfigShowFormatter creates a new formatter
func NewConfigShowFormatter(data ConfigShowOutput) ConfigShowFormatter {
	return ConfigShowFormatter{Data: data}
}

// TableOutput generates human-friendly table output for config show
func (f ConfigShowFormatter) TableOutput() string {
	c := f.Data
	output := "# Configuration for plonk\n"
	output += fmt.Sprintf("# Config file: %s\n\n", c.ConfigPath)

	if c.Config == nil {
		return output + "No configuration loaded\n"
	}

	data, err := formatConfigWithHighlights(c.Config)
	if err != nil {
		return output + "Error formatting configuration\n"
	}
	return output + data
}

// formatConfigWithHighlights formats the config as YAML and adds color
// highlighting for user-defined fields in table output, while leaving the
// YAML structure unchanged.
func formatConfigWithHighlights(cfg *config.Config) (string, error) {
	// Compute non-default fields.
	nonDefaultFields := config.GetNonDefaultFields(cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	defaults := config.GetDefaults()
	addedDirs, removedDirs := listChanges(cfg.ExpandDirectories, defaults.ExpandDirectories)
	addedPatterns, removedPatterns := listChanges(cfg.IgnorePatterns, defaults.IgnorePatterns)

	var out strings.Builder
	var added, removed map[string]bool
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			out.WriteString(line + "\n")
			continue
		}

		// Leaving a list emits its removed defaults before the next top-level key.
		if line[0] != ' ' && line[0] != '\t' {
			for item := range removed {
				out.WriteString(ColorRemoved(fmt.Sprintf("# removed: - %s", item)) + "\n")
			}
			added, removed = nil, nil
			switch {
			case strings.HasPrefix(trimmed, "expand_directories:"):
				added, removed = addedDirs, removedDirs
			case strings.HasPrefix(trimmed, "ignore_patterns:"):
				added, removed = addedPatterns, removedPatterns
			}

			key, _, _ := strings.Cut(trimmed, ":")
			if _, custom := nonDefaultFields[key]; custom {
				line = ColorInfo(line)
			}
		} else if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if added[item] {
				line = ColorAdded(line)
			}
		}
		out.WriteString(line + "\n")
	}
	return out.String(), nil
}

func listChanges(current, defaults []string) (added, removed map[string]bool) {
	added, removed = make(map[string]bool), make(map[string]bool)
	for _, item := range current {
		if !slices.Contains(defaults, item) {
			added[item] = true
		}
	}
	for _, item := range defaults {
		if !slices.Contains(current, item) {
			removed[item] = true
		}
	}
	return added, removed
}
