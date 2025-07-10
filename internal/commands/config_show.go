// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/errors"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// getConfigPath returns the path to the main configuration file
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
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "config-show", "output-format", "invalid output format")
	}

	// Get config directory
	configDir := config.GetDefaultConfigDirectory()
	configPath := getConfigPath(configDir)

	// Load configuration (this handles missing files gracefully due to zero-config)
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// Handle validation errors
		return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Build output data with resolved config (merges defaults with user config)
	resolvedConfig := cfg.Resolve()

	outputData := ConfigShowOutput{
		ConfigPath: configPath,
		Config:     resolvedConfig,
	}

	return RenderOutput(outputData, format)
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string                 `json:"config_path" yaml:"config_path"`
	Config     *config.ResolvedConfig `json:"config" yaml:"config"`
}

// TableOutput generates human-friendly table output for config show
func (c ConfigShowOutput) TableOutput() string {
	output := fmt.Sprintf("# Configuration: %s\n\n", c.ConfigPath)

	if c.Config != nil {
		// Convert config to YAML for display (shows effective config with defaults)
		yamlBytes, err := yaml.Marshal(c.Config)
		if err != nil {
			return fmt.Sprintf("Error formatting configuration: %v\n", err)
		}
		output += string(yamlBytes)
	}

	return output
}

// StructuredData returns the structured data for serialization
func (c ConfigShowOutput) StructuredData() any {
	return c
}
