// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/spf13/cobra"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate the plonk configuration file for syntax and structural errors.

This command checks:
- YAML syntax correctness
- Required fields presence
- Package name format
- File path format
- Configuration structure

Examples:
  plonk config validate           # Validate current configuration
  plonk config validate -o json   # Output validation results as JSON`,
	RunE: runConfigValidate,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get config directory
	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		outputData := ConfigValidateOutput{
			ConfigPath: configPath,
			Valid:      false,
			Errors:     []string{"Configuration file not found"},
			Message:    "No configuration file found. Run 'plonk init' to create one.",
		}
		return RenderOutput(outputData, format)
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Validate configuration
	validator := config.NewSimpleValidator()
	result := validator.ValidateConfigFromYAML(content)

	// Build output
	outputData := ConfigValidateOutput{
		ConfigPath: configPath,
		Valid:      result.Valid,
		Errors:     result.Errors,
		Warnings:   result.Warnings,
		Message:    result.GetSummary(),
	}

	return RenderOutput(outputData, format)
}

// ConfigValidateOutput represents the output structure for config validate command
type ConfigValidateOutput struct {
	ConfigPath string   `json:"config_path" yaml:"config_path"`
	Valid      bool     `json:"valid" yaml:"valid"`
	Errors     []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Message    string   `json:"message" yaml:"message"`
}

// TableOutput generates human-friendly table output for config validate
func (c ConfigValidateOutput) TableOutput() string {
	output := fmt.Sprintf("Configuration: %s\n", c.ConfigPath)

	if c.Valid {
		output += "✅ " + c.Message
		if len(c.Warnings) > 0 {
			output += "\n\nWarnings:"
			for _, warning := range c.Warnings {
				output += fmt.Sprintf("\n  ⚠️  %s", warning)
			}
		}
	} else {
		output += "❌ " + c.Message
		if len(c.Errors) > 0 {
			output += "\n\nErrors:"
			for _, err := range c.Errors {
				output += fmt.Sprintf("\n  ❌ %s", err)
			}
		}
		if len(c.Warnings) > 0 {
			output += "\n\nWarnings:"
			for _, warning := range c.Warnings {
				output += fmt.Sprintf("\n  ⚠️  %s", warning)
			}
		}
	}

	return output + "\n"
}

// StructuredData returns the structured data for serialization
func (c ConfigValidateOutput) StructuredData() any {
	return c
}
