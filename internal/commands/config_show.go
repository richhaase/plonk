// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	plonkoutput "github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func getConfigPath(configDir string) string {
	return filepath.Join(configDir, "plonk.yaml")
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display effective configuration",
	Long: `Display the effective plonk configuration (defaults merged with user settings).

Shows the complete configuration that plonk is actually using, including all default
values merged with any user-specified overrides from the config file.

Examples:
  plonk config show           # Show effective configuration
  plonk config show -o json   # Show as JSON
  plonk config show -o yaml   # Show as YAML (default)`,
	RunE: runConfigShow,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get config directory
	configDir := config.GetDefaultConfigDirectory()
	configPath := getConfigPath(configDir)

	// Load configuration (this handles missing files gracefully due to zero-config)
	cfg := config.LoadWithDefaults(configDir)

	// Build output data
	outputData := ConfigShowOutput{
		ConfigPath: configPath,
		Config:     cfg,
		Checker:    config.NewUserDefinedChecker(configDir),
		ConfigDir:  configDir,
	}

	return RenderOutput(outputData, format)
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string                     `json:"config_path" yaml:"config_path"`
	Config     *config.Config             `json:"config" yaml:"config"`
	Checker    *config.UserDefinedChecker `json:"-" yaml:"-"` // Not included in JSON/YAML
	ConfigDir  string                     `json:"-" yaml:"-"` // Not included in JSON/YAML
}

// TableOutput generates human-friendly table output for config show
func (c ConfigShowOutput) TableOutput() string {
	output := fmt.Sprintf("# Configuration for plonk\n")
	output += fmt.Sprintf("# Config file: %s\n\n", c.ConfigPath)

	if c.Config == nil {
		return output + "No configuration loaded\n"
	}

	// Helper to format a field with optional user-defined annotation
	formatField := func(name string, value interface{}) string {
		// Marshal just this field to YAML
		fieldMap := map[string]interface{}{name: value}
		data, _ := yaml.Marshal(fieldMap)
		line := strings.TrimSpace(string(data))

		// Check if user-defined and add annotation
		if c.Checker != nil && c.Checker.IsFieldUserDefined(name, value) {
			line += "  " + plonkoutput.ColorInfo("(user-defined)")
		}

		return line + "\n"
	}

	// Format each field
	output += formatField("default_manager", c.Config.DefaultManager)
	output += formatField("operation_timeout", c.Config.OperationTimeout)
	output += formatField("package_timeout", c.Config.PackageTimeout)
	output += formatField("dotfile_timeout", c.Config.DotfileTimeout)

	// Add blank line before lists
	output += "\n"
	output += formatField("expand_directories", c.Config.ExpandDirectories)

	output += "\n"
	output += formatField("ignore_patterns", c.Config.IgnorePatterns)

	// Handle optional nested structures
	if len(c.Config.Dotfiles.UnmanagedFilters) > 0 {
		output += "\n"
		output += formatField("dotfiles", c.Config.Dotfiles)
	}

	if len(c.Config.Hooks.PreApply) > 0 || len(c.Config.Hooks.PostApply) > 0 {
		output += "\n"
		output += formatField("hooks", c.Config.Hooks)
	}

	return output
}

// StructuredData returns the structured data for serialization
func (c ConfigShowOutput) StructuredData() any {
	return c
}
