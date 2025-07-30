// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	plonkoutput "github.com/richhaase/plonk/internal/output"
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
	RunE: runConfigEdit,
	Args: cobra.NoArgs,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	configDir := config.GetConfigDir()

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
		fmt.Printf("Opening configuration with %s...\n", editor)
		if err := openInEditor(editor, tempFile); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		// Parse and validate
		editedConfig, validationErr := parseAndValidateConfig(tempFile)
		if validationErr != nil {
			fmt.Fprintf(os.Stderr, "\n%s\n%s\n", plonkoutput.ColorError("Configuration validation failed:"), validationErr)

			// Prompt for action
			action := promptAction()
			switch action {
			case 'e':
				continue // Edit again
			case 'r':
				fmt.Println("Changes reverted.")
				return nil // Revert (don't save)
			case 'q':
				return fmt.Errorf("configuration invalid, changes discarded")
			}
		}

		// Success - save only non-defaults
		if err := saveNonDefaultValues(configDir, editedConfig); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("%s Configuration saved (only non-default values)\n", plonkoutput.Success())
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

	cmd := exec.Command(parts[0], append(parts[1:], filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// createTempConfigFile creates a temp file with the merged runtime config
func createTempConfigFile(configDir string) (string, error) {
	// Load current runtime config (defaults + user overrides)
	cfg := config.LoadWithDefaults(configDir)

	// Create temp file
	tempFile, err := ioutil.TempFile("", "plonk-config-*.yaml")
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

	// Generate YAML with annotations
	if err := writeAnnotatedConfig(tempFile, cfg, configDir); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	tempFile.Close()
	return tempFile.Name(), nil
}

// writeAnnotatedConfig writes the config with (user-defined) annotations
func writeAnnotatedConfig(w *os.File, cfg *config.Config, configDir string) error {
	// Get defaults for comparison
	defaults := config.GetDefaults()

	// Load user config to check what's actually user-defined
	userConfig, _ := config.Load(configDir) // May fail if no user config exists

	// Helper to check if a value is user-defined
	isUserDefined := func(actualValue, defaultValue interface{}) bool {
		// If we don't have a user config, nothing is user-defined
		if userConfig == nil {
			return false
		}
		// Use reflect.DeepEqual to compare values
		return !reflect.DeepEqual(actualValue, defaultValue)
	}

	// Helper to write a field with optional annotation
	writeField := func(name string, value interface{}, isUserDef bool) {
		// Marshal just this field
		fieldMap := map[string]interface{}{name: value}
		data, _ := yaml.Marshal(fieldMap)
		// Remove the trailing newline as we'll add it ourselves
		line := strings.TrimSpace(string(data))

		fmt.Fprint(w, line)
		if isUserDef {
			fmt.Fprint(w, "  # (user-defined)")
		}
		fmt.Fprintln(w)
	}

	// Write each field
	writeField("default_manager", cfg.DefaultManager,
		isUserDefined(cfg.DefaultManager, defaults.DefaultManager))

	writeField("operation_timeout", cfg.OperationTimeout,
		isUserDefined(cfg.OperationTimeout, defaults.OperationTimeout))

	writeField("package_timeout", cfg.PackageTimeout,
		isUserDefined(cfg.PackageTimeout, defaults.PackageTimeout))

	writeField("dotfile_timeout", cfg.DotfileTimeout,
		isUserDefined(cfg.DotfileTimeout, defaults.DotfileTimeout))

	// For lists, check if the entire list differs
	writeField("expand_directories", cfg.ExpandDirectories,
		isUserDefined(cfg.ExpandDirectories, defaults.ExpandDirectories))

	writeField("ignore_patterns", cfg.IgnorePatterns,
		isUserDefined(cfg.IgnorePatterns, defaults.IgnorePatterns))

	// Handle nested structures
	if !reflect.DeepEqual(cfg.Dotfiles, defaults.Dotfiles) {
		writeField("dotfiles", cfg.Dotfiles, true)
	} else if len(cfg.Dotfiles.UnmanagedFilters) > 0 {
		writeField("dotfiles", cfg.Dotfiles, false)
	}

	if !reflect.DeepEqual(cfg.Hooks, defaults.Hooks) {
		writeField("hooks", cfg.Hooks, true)
	} else if len(cfg.Hooks.PreApply) > 0 || len(cfg.Hooks.PostApply) > 0 {
		writeField("hooks", cfg.Hooks, false)
	}

	return nil
}

// parseAndValidateConfig reads and validates the temp file
func parseAndValidateConfig(filename string) (*config.Config, error) {
	// Read file
	data, err := ioutil.ReadFile(filename)
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
	if cfg.DefaultManager == "" {
		cfg.DefaultManager = config.GetDefaults().DefaultManager
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = config.GetDefaults().OperationTimeout
	}
	if cfg.PackageTimeout == 0 {
		cfg.PackageTimeout = config.GetDefaults().PackageTimeout
	}
	if cfg.DotfileTimeout == 0 {
		cfg.DotfileTimeout = config.GetDefaults().DotfileTimeout
	}
	if len(cfg.ExpandDirectories) == 0 {
		cfg.ExpandDirectories = config.GetDefaults().ExpandDirectories
	}
	if len(cfg.IgnorePatterns) == 0 {
		cfg.IgnorePatterns = config.GetDefaults().IgnorePatterns
	}

	return &cfg, nil
}

// saveNonDefaultValues writes only non-default values to plonk.yaml
func saveNonDefaultValues(configDir string, cfg *config.Config) error {
	defaults := config.GetDefaults()

	// Create a map to hold only non-default values
	nonDefaults := make(map[string]interface{})

	// Compare each field and add non-defaults to map
	if cfg.DefaultManager != defaults.DefaultManager {
		nonDefaults["default_manager"] = cfg.DefaultManager
	}

	if cfg.OperationTimeout != defaults.OperationTimeout {
		nonDefaults["operation_timeout"] = cfg.OperationTimeout
	}

	if cfg.PackageTimeout != defaults.PackageTimeout {
		nonDefaults["package_timeout"] = cfg.PackageTimeout
	}

	if cfg.DotfileTimeout != defaults.DotfileTimeout {
		nonDefaults["dotfile_timeout"] = cfg.DotfileTimeout
	}

	// For lists, save entire list if ANY element differs
	if !reflect.DeepEqual(cfg.ExpandDirectories, defaults.ExpandDirectories) {
		nonDefaults["expand_directories"] = cfg.ExpandDirectories
	}

	if !reflect.DeepEqual(cfg.IgnorePatterns, defaults.IgnorePatterns) {
		nonDefaults["ignore_patterns"] = cfg.IgnorePatterns
	}

	// Handle nested structures
	if !reflect.DeepEqual(cfg.Dotfiles, defaults.Dotfiles) {
		nonDefaults["dotfiles"] = cfg.Dotfiles
	}

	if !reflect.DeepEqual(cfg.Hooks, defaults.Hooks) {
		nonDefaults["hooks"] = cfg.Hooks
	}

	// If everything is default, write empty file
	configPath := filepath.Join(configDir, "plonk.yaml")
	if len(nonDefaults) == 0 {
		// Write empty file (or minimal comment)
		return ioutil.WriteFile(configPath, []byte(""), 0644)
	}

	// Marshal to minimal YAML
	data, err := yaml.Marshal(nonDefaults)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to plonk.yaml
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// promptAction prompts user for edit/revert/quit decision
func promptAction() rune {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nWhat would you like to do? (e)dit again, (r)evert changes, (q)uit: ")

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

		fmt.Println("Please enter 'e', 'r', or 'q'")
	}
}
