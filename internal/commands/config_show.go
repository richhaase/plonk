// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
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
	Short: "Display configuration content",
	Long: `Display the current plonk configuration file content.

Shows the actual YAML configuration with proper formatting.
For validation status and summary, use 'plonk config status'.

Examples:
  plonk config show           # Show configuration content
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

	// Check if config file exists
	configExists := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configExists = false
	}

	// Load configuration (this handles missing files gracefully due to zero-config)
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// Handle validation errors - still show the config if possible
		rawContent, readErr := os.ReadFile(configPath)
		if readErr != nil {
			return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", "failed to load configuration")
		}

		outputData := ConfigShowOutput{
			ConfigPath: configPath,
			Status:     "invalid",
			Message:    err.Error(),
			RawContent: string(rawContent),
		}
		return RenderOutput(outputData, format)
	}

	// Build output data with resolved config (merges defaults with user config)
	resolvedConfig := cfg.Resolve()

	// Determine status based on whether config file exists
	var status, message string
	if !configExists {
		status = "missing"
		message = "Configuration file not found, showing defaults. Run 'plonk config init' to create one."
	} else {
		status = "valid"
		message = "Configuration is valid"
	}

	outputData := ConfigShowOutput{
		ConfigPath:     configPath,
		Status:         status,
		Message:        message,
		Config:         cfg,
		ResolvedConfig: resolvedConfig,
	}

	return RenderOutput(outputData, format)
}

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath     string                 `json:"config_path" yaml:"config_path"`
	Status         string                 `json:"status" yaml:"status"`
	Message        string                 `json:"message,omitempty" yaml:"message,omitempty"`
	Config         *config.Config         `json:"config,omitempty" yaml:"config,omitempty"`
	ResolvedConfig *config.ResolvedConfig `json:"resolved_config,omitempty" yaml:"resolved_config,omitempty"`
	RawContent     string                 `json:"raw_content,omitempty" yaml:"raw_content,omitempty"`
}

// TableOutput generates human-friendly table output for config show
func (c ConfigShowOutput) TableOutput() string {
	if c.Status == "missing" {
		output := fmt.Sprintf("# Configuration: %s\n", c.ConfigPath)
		output += fmt.Sprintf("# %s\n\n", c.Message)

		// Show resolved config even when file is missing
		if c.ResolvedConfig != nil {
			yamlBytes, err := yaml.Marshal(c.ResolvedConfig)
			if err != nil {
				return fmt.Sprintf("Error formatting configuration: %v\n", err)
			}
			output += string(yamlBytes)
		}
		return output
	}

	output := fmt.Sprintf("# Configuration: %s\n\n", c.ConfigPath)

	if c.Status == "invalid" && c.RawContent != "" {
		output += "# WARNING: Configuration has validation errors\n"
		output += fmt.Sprintf("# %s\n\n", c.Message)
		output += c.RawContent
		return output
	}

	if c.ResolvedConfig != nil {
		// Convert resolved config to YAML for display (shows effective config with defaults)
		yamlBytes, err := yaml.Marshal(c.ResolvedConfig)
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
