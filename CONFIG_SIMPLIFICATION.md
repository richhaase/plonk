# Configuration Simplification Plan

## 1. Overview & Goal

**The Problem:** The current configuration system, primarily in `internal/config`, is severely over-engineered. It spans over 1500 lines across more than 10 files to perform what should be a simple task: loading a YAML file. It uses multiple layers of structs (`Config`, `ResolvedConfig`, `ConfigDefaults`), manual validation, and complex loaders, making it a major source of complexity and a barrier to understanding the codebase.

**The Goal:** Replace the entire bespoke configuration system with a **single, idiomatic Go solution**. This will involve:
1.  Creating one new `Config` struct that uses standard `yaml` and `validate` struct tags.
2.  Creating one new `Load(path)` function to handle all loading, defaulting, and validation logic.
3.  The **complete deletion** of the old configuration files, including `loader.go`, `resolved.go`, `simple_validator.go`, `schema.go`, `defaults.go`, and their associated tests.

**Guiding Principles:**
*   The user-facing behavior for configuration (what keys are in `plonk.yaml`, how they work) should not change.
*   The internal implementation will be completely replaced.
*   The process will be incremental and verifiable at each step.

## 2. The New Architecture: Simple, Standard, and Idiomatic

The new system will live entirely within a new `internal/config/config.go` file.

### The Single `Config` Struct

We will define one struct that is the single source of truth for configuration.

```go
// In internal/config/config.go

package config

// Config represents the entire plonk configuration.
// It uses struct tags for YAML parsing and validation.
type Config struct {
	Packages []Package `yaml:"packages" validate:"dive"`
	Dotfiles []Dotfile `yaml:"dotfiles" validate:"dive"`
	Settings Settings  `yaml:"settings"`
}

type Package struct {
	Name    string `yaml:"name" validate:"required"`
	Manager string `yaml:"manager" validate:"required,oneof=homebrew npm pip gem go cargo"`
	// ... other package fields
}

type Dotfile struct {
	Source  string `yaml:"source" validate:"required"`
	Dest    string `yaml:"dest" validate:"required"`
	Method  string `yaml:"method" validate:"oneof=link copy"`
	// ... other dotfile fields
}

type Settings struct {
	// ... settings fields
}
```

### The Single `Load` Function

A single function will orchestrate the entire loading process.

```go
// In internal/config/config.go

import (
	"os"
	"gopkg.in/yaml.v3"
	"github.com/go-playground/validator/v10"
)

// Load reads, parses, defaults, and validates the configuration from a given path.
func Load(configPath string) (*Config, error) {
	// 1. Read file bytes
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		// Handle file not found, etc.
		return nil, err
	}

	// 2. Unmarshal YAML into the struct
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	// 3. Apply Defaults
	// Example: if a dotfile method is missing, default it to "link"
	for i := range cfg.Dotfiles {
		if cfg.Dotfiles[i].Method == "" {
			cfg.Dotfiles[i].Method = "link"
		}
	}
	// ... other defaults

	// 4. Validate the struct
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

This architecture replaces over 1500 lines of code with less than 100, using standard, well-understood Go libraries and idioms.

## 3. Detailed Refactoring Phases

### Phase 0: Build the New System in Isolation

We will build and test the new system completely before touching any of the old code.

1.  **Create New Files:**
    *   Create `internal/config/config.go`.
    *   Create `internal/config/config_test.go`.
2.  **Implement the New `Config` Struct and `Load` function:** Populate the new files with the code described in section 2.
3.  **Write Comprehensive Tests:** In `config_test.go`, write tests that cover:
    *   Loading a valid `plonk.yaml`.
    *   Correct application of all default values.
    *   Validation errors for missing required fields (e.g., a package with no `name`).
    *   Validation errors for invalid values (e.g., a dotfile with an invalid `method`).
    *   Handling of a non-existent config file.
4.  **Verify:** At the end of this phase, we will have a fully functional and tested new configuration system that is not yet used by the application.

### Phase 1: Incremental Migration

We will now migrate the application to the new system, one command at a time.

1.  **Target a Single Command:** Choose a simple command that reads configuration, such as `plonk ls`.
2.  **Replace the Call:** In the command's implementation (e.g., `internal/commands/ls.go`), find the line that calls the old config loader (e.g., `getResolvedConfig()`). Replace it with a call to our new `config.Load()`.
3.  **Verify:** Run `just test` and `just test-ux`. Manually run `plonk ls` to ensure its behavior is identical to before.
4.  **Commit:** Commit the successful migration for that single command.
5.  **Repeat:** Continue this process for all other commands (`status`, `sync`, `add`, etc.) until no part of the codebase uses the old configuration system.

### Phase 2: Cleanup

1.  **Identify Obsolete Files:** The following files in `internal/config/` will now be unused: `loader.go`, `resolved.go`, `simple_validator.go`, `schema.go`, `defaults.go`, `adapters.go`, `interfaces.go`, `yaml_config.go`, and all of their corresponding `_test.go` files.
2.  **Delete Old Files:** Delete all of the obsolete files identified in the previous step.
3.  **Final Verification:** Run `just test` and `just test-ux`. The Go compiler will immediately fail if any references to the old files remain. Fix any compilation errors and run tests one last time to confirm the success of the entire refactor.

## 4. Risk Analysis & Mitigation

*   **Risk: Mismatched Default Values.**
    *   **Description:** The new, hardcoded defaults in the `Load` function might not perfectly match the logic in the old `defaults.go`.
    *   **Mitigation:** During Phase 0, the agent responsible must carefully analyze `defaults.go` and replicate its logic exactly. The new unit tests created in `config_test.go` must explicitly assert that every default is applied correctly for all edge cases.

*   **Risk: Mismatched Validation Logic.**
    *   **Description:** The `validate` struct tags may not enforce the exact same rules as the old manual validator in `simple_validator.go`.
    *   **Mitigation:** Similar to defaults, the agent must analyze the old validation code and translate every rule into a `validate` tag. The new unit tests must include specific cases that would have failed the old validator to prove the new one catches them too.

*   **Risk: Forgetting to Migrate a Command.**
    *   **Description:** A command or helper function that uses the old config system is missed during the migration.
    *   **Mitigation:** This risk is fully mitigated by the process. In Phase 2, when we delete the old files, the Go compiler will refuse to build if any references to the old files remain. This provides a perfect safety net, pointing us directly to any code that was missed.

This plan provides a safe, methodical, and verifiable path to drastically simplifying the project's configuration system, removing a significant source of technical debt.
