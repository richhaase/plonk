# Task 008: Implement visudo-style Config Edit Command

## Overview
Enhance the `plonk config edit` command to work like `visudo`:
1. Show merged runtime configuration (defaults + user overrides)
2. Allow editing in user's editor
3. Validate on save
4. Write only non-default values to plonk.yaml
5. Support retry on validation failure

## Design Specifications

### Workflow
1. **Generate temp file** with full runtime config (defaults merged with user values)
2. **Add helpful header** with instructions
3. **Mark user-defined values** with comments
4. **Open in editor** ($VISUAL, $EDITOR, or vim)
5. **On save**: validate the edited config
6. **If invalid**: show all errors with line numbers, prompt to (e)dit, (r)evert, or (q)uit
7. **If valid**: extract non-default values and write minimal plonk.yaml
8. **Clean up** temp file in all cases

### Temp File Format
```yaml
# Plonk Configuration Editor
# - Delete any line to revert to default
# - Only values different from defaults will be saved to plonk.yaml
# - Save and exit to apply, or exit without saving to cancel

default_manager: brew
operation_timeout: 300
package_timeout: 600  # (user-defined)
dotfile_timeout: 30

expand_directories:
  - .config
  - .local
  - .ssh  # (user-defined)

ignore_patterns:
  - "*.swp"
  - ".DS_Store"
  - "*~"
  - ".git"
```

### Implementation Plan

#### Phase 1: Create Edit Loop Infrastructure

**File**: `/internal/commands/config_edit.go` (new file)

```go
package commands

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/richhaase/plonk/internal/config"
    "gopkg.in/yaml.v3"
)

// editConfig implements the visudo-style edit workflow
func editConfig(configDir string) error {
    // 1. Generate temp file with merged config
    tempFile, err := createTempConfigFile(configDir)
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tempFile)

    // 2. Edit loop
    for {
        // Open in editor
        if err := openInEditor(tempFile); err != nil {
            return fmt.Errorf("failed to open editor: %w", err)
        }

        // Parse and validate
        editedConfig, err := parseAndValidateConfig(tempFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "\nValidation failed:\n%v\n", err)

            // Prompt for action
            action := promptAction()
            switch action {
            case 'e':
                continue // Edit again
            case 'r':
                return nil // Revert (don't save)
            case 'q':
                return fmt.Errorf("configuration invalid, changes discarded")
            }
        }

        // Success - save only non-defaults
        return saveNonDefaultValues(configDir, editedConfig)
    }
}
```

#### Phase 2: Temp File Generation

```go
// createTempConfigFile creates a temp file with merged runtime config
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
    tempFile.WriteString(header)

    // Generate YAML with annotations
    if err := writeAnnotatedConfig(tempFile, cfg, configDir); err != nil {
        os.Remove(tempFile.Name())
        return "", err
    }

    tempFile.Close()
    return tempFile.Name(), nil
}
```

#### Phase 3: Config Annotation

```go
// writeAnnotatedConfig writes the config with (user-defined) annotations
func writeAnnotatedConfig(w io.Writer, cfg *config.Config, configDir string) error {
    defaults := config.GetDefaults()
    userConfig, _ := config.Load(configDir) // May fail if no user config

    // Helper to check if value is user-defined
    isUserDefined := func(field string) bool {
        if userConfig == nil {
            return false
        }
        // Compare field values between user config and defaults
        return !reflect.DeepEqual(
            getFieldValue(userConfig, field),
            getFieldValue(defaults, field),
        )
    }

    // Write each field with annotation if user-defined
    fmt.Fprintf(w, "default_manager: %s", cfg.DefaultManager)
    if isUserDefined("DefaultManager") {
        fmt.Fprint(w, "  # (user-defined)")
    }
    fmt.Fprintln(w)

    // ... continue for all fields
}
```

#### Phase 4: Validation Integration

```go
// parseAndValidateConfig reads and validates the temp file
func parseAndValidateConfig(filename string) (*config.Config, error) {
    // Read file
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Parse YAML with line number tracking
    var cfg config.Config
    decoder := yaml.NewDecoder(bytes.NewReader(data))
    decoder.SetStrict(true)

    if err := decoder.Decode(&cfg); err != nil {
        // Enhanced error with line numbers
        return nil, enhanceYAMLError(err, filename)
    }

    // Run plonk's existing validation
    if err := cfg.Validate(); err != nil {
        // Enhance with line numbers where possible
        return nil, enhanceValidationError(err, filename, data)
    }

    return &cfg, nil
}
```

#### Phase 5: Save Non-Default Values

```go
// saveNonDefaultValues writes only non-default values to plonk.yaml
func saveNonDefaultValues(configDir string, cfg *config.Config) error {
    defaults := config.GetDefaults()

    // Create a new config with only non-default values
    nonDefaults := &config.Config{}

    // Compare each field
    if cfg.DefaultManager != defaults.DefaultManager {
        nonDefaults.DefaultManager = cfg.DefaultManager
    }

    if cfg.OperationTimeout != defaults.OperationTimeout {
        nonDefaults.OperationTimeout = cfg.OperationTimeout
    }

    // For lists, save entire list if ANY element differs
    if !reflect.DeepEqual(cfg.ExpandDirectories, defaults.ExpandDirectories) {
        nonDefaults.ExpandDirectories = cfg.ExpandDirectories
    }

    if !reflect.DeepEqual(cfg.IgnorePatterns, defaults.IgnorePatterns) {
        nonDefaults.IgnorePatterns = cfg.IgnorePatterns
    }

    // Marshal to minimal YAML (no comments)
    data, err := yaml.Marshal(nonDefaults)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    // Write to plonk.yaml
    configPath := filepath.Join(configDir, "plonk.yaml")
    if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }

    fmt.Printf("Configuration saved to %s\n", configPath)
    return nil
}
```

#### Phase 6: User Interaction

```go
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

// openInEditor opens the file in user's preferred editor
func openInEditor(filename string) error {
    editor := os.Getenv("VISUAL")
    if editor == "" {
        editor = os.Getenv("EDITOR")
    }
    if editor == "" {
        editor = "vim"
    }

    cmd := exec.Command(editor, filename)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}
```

### Key Implementation Details

1. **Reflection Usage**: Use reflection to compare field values between configs
2. **YAML Preservation**: Use yaml.v3 for better control over formatting
3. **Error Enhancement**: Add line numbers to both YAML and validation errors
4. **Minimal Output**: Only write fields that differ from defaults
5. **List Handling**: If any element in a list differs, write the entire list

### Testing Scenarios

1. **Empty Config**: Starting with no plonk.yaml
2. **Partial Config**: User has some values set
3. **Invalid YAML**: Syntax errors
4. **Invalid Values**: Bad manager names, negative timeouts
5. **List Modifications**: Adding/removing from defaults
6. **Editor Crash**: Ensure temp file cleanup
7. **Cancel Scenarios**: Ctrl+C, quit without saving

### Success Criteria

- [ ] Full runtime config shown in editor
- [ ] User-defined values marked with `# (user-defined)`
- [ ] All validation errors shown with line numbers
- [ ] Edit/revert/quit loop works correctly
- [ ] Only non-default values saved to plonk.yaml
- [ ] Temp files always cleaned up
- [ ] Existing config.Validate() fully utilized
