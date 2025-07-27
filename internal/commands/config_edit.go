// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long: `Edit the plonk configuration file using your preferred editor.

This command:
- Opens the configuration file in $EDITOR (or nano if not set)
- Validates the configuration after editing
- Reports any validation errors
- Creates the configuration file if it doesn't exist

The file will be validated automatically after editing. If validation fails,
you'll see the errors and can choose to edit again or cancel.

Examples:
  plonk config edit               # Edit configuration file
  EDITOR=vim plonk config edit    # Use vim as editor`,
	RunE: runConfigEdit,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	// Get config directory and file path
	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return err
	}

	// Create default config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return err
		}
		fmt.Printf("Created default configuration file: %s\n", configPath)
	}

	// Get editor
	editor := getEditor()
	fmt.Printf("Opening configuration file with %s...\n", editor)

	// Loop until valid config or user cancels
	for {
		// Open editor
		if err := openEditor(editor, configPath); err != nil {
			return err
		}

		// Validate the edited config
		if err := validateEditedConfig(configPath); err != nil {
			fmt.Printf("\n❌ Configuration validation failed:\n%s\n", err.Error())

			// Ask if user wants to edit again
			if !promptEditAgain() {
				return fmt.Errorf("configuration validation failed")
			}
			continue
		}

		// Success
		fmt.Printf("\n✅ Configuration is valid and saved to: %s\n", configPath)
		break
	}

	return nil
}

// getEditor returns the editor to use, checking environment variables
func getEditor() string {
	// Check environment variables in order of preference
	editors := []string{
		os.Getenv("EDITOR"),
		os.Getenv("VISUAL"),
		"nano", // Fallback to nano
		"vi",   // Last resort
	}

	for _, editor := range editors {
		if editor != "" {
			// Check if the editor is available
			if _, err := exec.LookPath(strings.Fields(editor)[0]); err == nil {
				return editor
			}
		}
	}

	return "nano" // Final fallback
}

// openEditor opens the specified file in the editor
func openEditor(editor, filepath string) error {
	// Split editor command in case it has arguments
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("no editor specified: set EDITOR environment variable or specify editor with --editor flag")
	}

	cmd := exec.Command(parts[0], append(parts[1:], filepath)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// validateEditedConfig validates the configuration file after editing
func validateEditedConfig(configPath string) error {
	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Validate configuration
	validator := config.NewSimpleValidator()
	result := validator.ValidateConfigFromYAML(content)

	if !result.Valid {
		var errorMsg strings.Builder
		for _, err := range result.Errors {
			errorMsg.WriteString(fmt.Sprintf("  ❌ %s\n", err))
		}

		if len(result.Warnings) > 0 {
			errorMsg.WriteString("\nWarnings:\n")
			for _, warning := range result.Warnings {
				errorMsg.WriteString(fmt.Sprintf("  ⚠️  %s\n", warning))
			}
		}

		return fmt.Errorf("configuration validation failed: %d errors found", len(result.Errors))
	}

	// Show warnings if any
	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  ⚠️  %s\n", warning)
		}
	}

	return nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string) error {
	defaultConfig := `# Plonk Configuration
# https://github.com/your-repo/plonk

settings:
  default_manager: brew
  operation_timeout: 600  # 10 minutes

# Package definitions
brew: []

npm: []

# Dotfile definitions
dotfiles: []

# Example configuration:
# brew:
#   - git
#   - neovim
#   - firefox
#   - visual-studio-code
#
# npm:
#   - prettier
#   - typescript
#
# dotfiles:
#   - zshrc
#   - gitconfig
#   - config/nvim
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0600)
}

// promptEditAgain asks the user if they want to edit the config again
func promptEditAgain() bool {
	fmt.Print("\nDo you want to edit the configuration again? (y/N): ")

	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}
