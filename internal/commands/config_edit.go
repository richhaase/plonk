// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long: `Edit the plonk configuration file using your preferred editor.

This command works like visudo:
- Shows the full runtime configuration (defaults + your overrides)
- Opens it in your preferred editor ($VISUAL, $EDITOR, or vim)
- Validates the configuration after editing
- Saves only non-default values to plonk.yaml
- Supports edit/revert/quit on validation errors

Only values that differ from defaults are saved to keep your config minimal.

Examples:
  plonk config edit               # Edit configuration file
  EDITOR=vim plonk config edit    # Use vim as editor`,
	RunE:         runConfigEdit,
	SilenceUsage: true,
	Args:         cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return editConfigVisudoStyle(configDir)
}

// editConfigVisudoStyle implements the visudo-style edit workflow
func editConfigVisudoStyle(configDir string) error {
	// Create temp file with merged runtime config
	tempFile, err := createTempConfigFile(configDir)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile) // Always clean up

	// Get editor
	editor := getEditor()

	// Edit loop
	for {
		// Open in editor
		output.Printf("Opening configuration with %s...\n", editor)
		if err := openInEditor(editor, tempFile); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		// Parse and validate
		editedConfig, validationErr := parseAndValidateConfig(tempFile)
		if validationErr != nil {
			fmt.Fprintf(os.Stderr, "\n%s\n%s\n", output.ColorError("Configuration validation failed:"), validationErr)

			// Prompt for action
			action := promptAction()
			switch action {
			case 'e':
				continue // Edit again
			case 'r':
				output.Println("Changes reverted.")
				return nil // Revert (don't save)
			case 'q':
				return fmt.Errorf("configuration invalid, changes discarded")
			}
		}

		// Success - save only non-defaults
		if err := saveNonDefaultValues(configDir, editedConfig); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		output.Printf("%s Configuration saved (only non-default values)\n", output.Success())
		gitops.AutoCommit(configDir, "config edit", nil)
		return nil
	}
}

// getEditor returns the editor to use, checking VISUAL then EDITOR
func getEditor() string {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}
	return editor
}

// openInEditor opens the file in user's preferred editor
func openInEditor(editor, filename string) error {
	// Split editor command in case it has arguments
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("invalid editor: %s", editor)
	}

	//nolint:gosec // G204: editor from $VISUAL/$EDITOR env vars - standard Unix pattern
	cmd := exec.Command(parts[0], append(parts[1:], filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// createTempConfigFile creates a temp file with the merged runtime config
func createTempConfigFile(configDir string) (string, error) {
	configPath := getConfigPath(configDir)
	cfg, loadErr := config.LoadFromPath(configPath)
	useRaw := false

	// Create temp file
	tempFile, err := os.CreateTemp("", "plonk-config-*.yaml")
	if err != nil {
		return "", err
	}

	// Write header
	header := `# Plonk Configuration Editor
# - Delete any line to revert to default
# - Only values different from defaults will be saved to plonk.yaml
# - Save and exit to apply, or exit without saving to cancel

`
	if _, err := tempFile.WriteString(header); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	if loadErr == nil {
		if err := writeFullConfig(tempFile, cfg); err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	} else {
		if !os.IsNotExist(loadErr) {
			raw, readErr := os.ReadFile(configPath)
			if readErr != nil {
				os.Remove(tempFile.Name())
				return "", fmt.Errorf("failed to load existing config: %w", loadErr)
			}
			useRaw = true
			if _, err := tempFile.Write(raw); err != nil {
				os.Remove(tempFile.Name())
				return "", err
			}
		} else {
			defaultCfg := config.LoadWithDefaults(configDir)
			if err := writeFullConfig(tempFile, defaultCfg); err != nil {
				os.Remove(tempFile.Name())
				return "", err
			}
		}
	}

	if useRaw {
		tempFile.Close()
		return tempFile.Name(), nil
	}

	tempFile.Close()
	return tempFile.Name(), nil
}

// writeFullConfig writes the full config to the provided file in YAML format.
// This is used by config edit to present the same effective configuration that
// config show displays, without any annotations.
func writeFullConfig(w *os.File, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// parseAndValidateConfig reads and validates the temp file
func parseAndValidateConfig(filename string) (*config.Config, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Remove header comments before parsing
	lines := strings.Split(string(data), "\n")
	var configLines []string
	for _, line := range lines {
		// Skip header comment lines (but keep inline comments)
		if strings.HasPrefix(strings.TrimSpace(line), "#") && !strings.Contains(line, "# (user-defined)") {
			continue
		}
		// Remove the (user-defined) annotations before parsing
		line = strings.Replace(line, "  # (user-defined)", "", 1)
		configLines = append(configLines, line)
	}
	cleanData := []byte(strings.Join(configLines, "\n"))

	// Use plonk's validator which provides detailed errors
	validator := config.NewSimpleValidator()
	result := validator.ValidateConfigFromYAML(cleanData)

	if !result.Valid {
		// Build detailed error message with all errors
		var errorMsg strings.Builder
		for i, err := range result.Errors {
			if i > 0 {
				errorMsg.WriteString("\n")
			}
			errorMsg.WriteString(fmt.Sprintf("  - %s", err))
		}
		return nil, fmt.Errorf("%s", errorMsg.String())
	}

	// Parse again to get the actual config object
	var cfg config.Config
	if err := yaml.Unmarshal(cleanData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Apply defaults to ensure we have complete config
	config.ApplyDefaults(&cfg)

	return &cfg, nil
}

// saveNonDefaultValues writes only non-default values to plonk.yaml
func saveNonDefaultValues(configDir string, cfg *config.Config) error {
	// Create checker to get non-default fields
	checker := config.NewUserDefinedChecker(configDir)

	// Get only non-default top-level values
	nonDefaults := checker.GetNonDefaultFields(cfg)

	// If everything is default, write empty file
	configPath := filepath.Join(configDir, "plonk.yaml")
	if len(nonDefaults) == 0 {
		// Write empty file (or minimal comment)
		return os.WriteFile(configPath, []byte(""), 0644)
	}

	// Marshal to minimal YAML
	data, err := yaml.Marshal(nonDefaults)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to plonk.yaml
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// promptAction prompts user for edit/revert/quit decision
func promptAction() rune {
	reader := bufio.NewReader(os.Stdin)

	for {
		output.Printf("\nWhat would you like to do? (e)dit again, (r)evert changes, (q)uit: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if len(input) > 0 {
			switch input[0] {
			case 'e', 'r', 'q':
				return rune(input[0])
			}
		}

		output.Println("Please enter 'e', 'r', or 'q'")
	}
}
