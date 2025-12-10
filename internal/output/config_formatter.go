// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"gopkg.in/yaml.v3"
)

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string      `json:"config_path" yaml:"config_path"`
	Config     interface{} `json:"config" yaml:"config"`
	Checker    interface{} `json:"-" yaml:"-"` // Not included in JSON/YAML
	ConfigDir  string      `json:"-" yaml:"-"` // Not included in JSON/YAML
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

	// If we have a real config and checker, highlight user-defined fields.
	if cfg, ok := c.Config.(*config.Config); ok {
		if checker, ok := c.Checker.(*config.UserDefinedChecker); ok {
			highlighted, err := formatConfigWithHighlights(cfg, checker)
			if err == nil {
				output += highlighted
				return output
			}
		}
	}

	// Fallback: marshal the entire config to YAML without highlighting.
	data, err := yaml.Marshal(c.Config)
	if err != nil {
		return output + "Error formatting configuration\n"
	}

	output += string(data)
	return output
}

// StructuredData returns the structured data for serialization
func (f ConfigShowFormatter) StructuredData() any {
	return f.Data
}

// formatConfigWithHighlights formats the config as YAML and adds color
// highlighting for user-defined fields in table output, while leaving the
// YAML structure unchanged.
func formatConfigWithHighlights(cfg *config.Config, checker *config.UserDefinedChecker) (string, error) {
	// Compute non-default fields and managers.
	nonDefaultFields := checker.GetNonDefaultFields(cfg)
	nonDefaultManagers := checker.GetNonDefaultManagers(cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	customKeys := make(map[string]struct{}, len(nonDefaultFields))
	for k := range nonDefaultFields {
		customKeys[k] = struct{}{}
	}

	customManagers := make(map[string]struct{}, len(nonDefaultManagers))
	for k := range nonDefaultManagers {
		customManagers[k] = struct{}{}
	}

	// Build sets of custom list items for selected fields where it's helpful
	// to see per-item differences (rather than just the entire field).
	defaults := config.GetDefaults()
	customExpandDirs := make(map[string]struct{})
	removedExpandDirs := make(map[string]struct{})
	if len(nonDefaultFields) > 0 {
		currentSet := make(map[string]struct{}, len(cfg.ExpandDirectories))
		for _, v := range cfg.ExpandDirectories {
			currentSet[v] = struct{}{}
		}
		for _, v := range defaults.ExpandDirectories {
			if _, ok := currentSet[v]; !ok {
				removedExpandDirs[v] = struct{}{}
			}
		}
		for _, v := range cfg.ExpandDirectories {
			if _, isDefault := removedExpandDirs[v]; !isDefault {
				// Item is either default or added; added if not in defaults.
				foundInDefaults := false
				for _, dv := range defaults.ExpandDirectories {
					if dv == v {
						foundInDefaults = true
						break
					}
				}
				if !foundInDefaults {
					customExpandDirs[v] = struct{}{}
				}
			}
		}
	}

	customIgnorePatterns := make(map[string]struct{})
	removedIgnorePatterns := make(map[string]struct{})
	if len(nonDefaultFields) > 0 {
		currentSet := make(map[string]struct{}, len(cfg.IgnorePatterns))
		for _, v := range cfg.IgnorePatterns {
			currentSet[v] = struct{}{}
		}
		for _, v := range defaults.IgnorePatterns {
			if _, ok := currentSet[v]; !ok {
				removedIgnorePatterns[v] = struct{}{}
			}
		}
		for _, v := range cfg.IgnorePatterns {
			if _, isRemoved := removedIgnorePatterns[v]; !isRemoved {
				foundInDefaults := false
				for _, dv := range defaults.IgnorePatterns {
					if dv == v {
						foundInDefaults = true
						break
					}
				}
				if !foundInDefaults {
					customIgnorePatterns[v] = struct{}{}
				}
			}
		}
	}

	var out strings.Builder
	inManagers := false
	inExpandDirs := false
	inIgnorePatterns := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Preserve blank lines as-is.
		if trimmed == "" {
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		// Detect top-level keys (no leading spaces).
		if len(line) > 0 && (line[0] != ' ' && line[0] != '\t') {
			// If we are leaving a list block, emit removed items as comments.
			if inExpandDirs && len(removedExpandDirs) > 0 {
				for item := range removedExpandDirs {
					comment := fmt.Sprintf("# removed: - %s", item)
					out.WriteString(ColorRemoved(comment))
					out.WriteString("\n")
				}
			}
			if inIgnorePatterns && len(removedIgnorePatterns) > 0 {
				for item := range removedIgnorePatterns {
					comment := fmt.Sprintf("# removed: - %s", item)
					out.WriteString(ColorRemoved(comment))
					out.WriteString("\n")
				}
			}

			// Update managers context.
			if strings.HasPrefix(trimmed, "managers:") {
				inManagers = true
				inExpandDirs = false
				inIgnorePatterns = false
				out.WriteString(line)
				out.WriteString("\n")
				continue
			}
			inManagers = false
			inExpandDirs = false
			inIgnorePatterns = false

			if strings.HasPrefix(trimmed, "expand_directories:") {
				inExpandDirs = true
			}
			if strings.HasPrefix(trimmed, "ignore_patterns:") {
				inIgnorePatterns = true
			}

			// Colorize top-level fields that differ from defaults.
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) > 0 {
				key := parts[0]
				if _, isCustom := customKeys[key]; isCustom {
					out.WriteString(ColorInfo(line))
					out.WriteString("\n")
					continue
				}
			}

			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		// Within managers block, highlight custom/overridden managers by name.
		if inManagers {
			// Expect lines like "  npm:" for manager names.
			if strings.HasPrefix(line, "  ") && !strings.HasPrefix(strings.TrimSpace(line), "-") {
				managerLine := strings.TrimSpace(line)
				parts := strings.SplitN(managerLine, ":", 2)
				if len(parts) > 0 {
					managerName := parts[0]
					if _, isCustomMgr := customManagers[managerName]; isCustomMgr {
						out.WriteString(ColorInfo(line))
						out.WriteString("\n")
						continue
					}
				}
			}
		}

		// Within expand_directories / ignore_patterns, highlight custom list items.
		if inExpandDirs || inIgnorePatterns {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "- ") {
				item := strings.TrimSpace(strings.TrimPrefix(trim, "- "))
				if inExpandDirs {
					if _, isCustom := customExpandDirs[item]; isCustom {
						out.WriteString(ColorAdded(line))
						out.WriteString("\n")
						continue
					}
				}
				if inIgnorePatterns {
					if _, isCustom := customIgnorePatterns[item]; isCustom {
						out.WriteString(ColorAdded(line))
						out.WriteString("\n")
						continue
					}
				}
			}
		}

		// Default: no highlighting.
		out.WriteString(line)
		out.WriteString("\n")
	}

	return out.String(), nil
}
