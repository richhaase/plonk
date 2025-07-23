# Configuration Simplification Plan (Updated)

## Status: Phase 0 Complete, Ready for Phase 1

This is an updated version of the configuration simplification plan that reflects the actual configuration structure found in the codebase.

**Progress:**
- ✅ Phase 0: Build new simplified config system - COMPLETE
- ⏳ Phase 1: Create compatibility layer - READY TO START
- ⏳ Phase 2: Replace implementation atomically - PENDING
- ⏳ Phase 3: Remove compatibility layer and old files - PENDING

## 1. Overview & Goal

**The Problem:** The current configuration system spans over 3000 lines across 15+ files to manage a simple YAML configuration with only 6 fields. It uses multiple layers of structs (`Config`, `ResolvedConfig`, `ConfigDefaults`), manual validation, complex loaders, managers, and services.

**The Goal:** Replace the entire configuration system with a **single, idiomatic Go solution** that:
1. Maintains exact compatibility with existing `plonk.yaml` files
2. Provides the same zero-config behavior (gracefully handles missing files)
3. Uses standard `yaml` and `validate` struct tags
4. Reduces 3000+ lines to under 200 lines

**Current Configuration Fields** (from actual codebase):
- `default_manager`: Default package manager (homebrew, npm, cargo, etc.)
- `operation_timeout`: General operation timeout in seconds
- `package_timeout`: Package operation timeout in seconds
- `dotfile_timeout`: Dotfile operation timeout in seconds
- `expand_directories`: Directories to expand in dot list output
- `ignore_patterns`: Patterns to ignore during dotfile discovery

## 2. The New Architecture: Simple and Idiomatic

### The Single Config File: `internal/config/config.go`

```go
package config

import (
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
    "github.com/go-playground/validator/v10"
)

// Config represents the plonk configuration
type Config struct {
    DefaultManager    string   `yaml:"default_manager,omitempty" validate:"omitempty,oneof=homebrew npm pip gem go cargo"`
    OperationTimeout  int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
    PackageTimeout    int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`
    DotfileTimeout    int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`
    ExpandDirectories []string `yaml:"expand_directories,omitempty"`
    IgnorePatterns    []string `yaml:"ignore_patterns,omitempty"`
}

// defaults for zero-config support
var defaults = Config{
    DefaultManager:   "homebrew",
    OperationTimeout: 300, // 5 minutes
    PackageTimeout:   180, // 3 minutes
    DotfileTimeout:   60,  // 1 minute
    ExpandDirectories: []string{
        ".config", ".ssh", ".aws", ".kube",
        ".docker", ".gnupg", ".local",
    },
    IgnorePatterns: []string{
        ".DS_Store", ".git", "*.backup",
        "*.tmp", "*.swp", "plonk.lock",
    },
}

// Load reads and validates configuration from the standard location
func Load(configDir string) (*Config, error) {
    configPath := filepath.Join(configDir, "plonk.yaml")
    return LoadFromPath(configPath)
}

// LoadFromPath reads and validates configuration from a specific path
func LoadFromPath(configPath string) (*Config, error) {
    // Start with defaults
    cfg := defaults

    // Read file if it exists
    data, err := os.ReadFile(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Zero-config: return defaults if file doesn't exist
            return &cfg, nil
        }
        return nil, err
    }

    // Unmarshal YAML over defaults
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    // Validate
    validate := validator.New()
    if err := validate.Struct(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// LoadWithDefaults provides zero-config behavior matching current LoadConfigWithDefaults
func LoadWithDefaults(configDir string) *Config {
    cfg, err := Load(configDir)
    if err != nil {
        // Return defaults on any error
        return &defaults
    }
    return cfg
}

// Resolve returns self for API compatibility
// In the new system, Config IS the resolved config
func (c *Config) Resolve() *Config {
    return c
}

// GetDefaultManager returns the default package manager
func (c *Config) GetDefaultManager() string {
    return c.DefaultManager
}

// GetOperationTimeout returns operation timeout in seconds
func (c *Config) GetOperationTimeout() int {
    return c.OperationTimeout
}

// GetPackageTimeout returns package timeout in seconds
func (c *Config) GetPackageTimeout() int {
    return c.PackageTimeout
}

// GetDotfileTimeout returns dotfile timeout in seconds
func (c *Config) GetDotfileTimeout() int {
    return c.DotfileTimeout
}

// GetExpandDirectories returns directories to expand
func (c *Config) GetExpandDirectories() []string {
    return c.ExpandDirectories
}

// GetIgnorePatterns returns patterns to ignore
func (c *Config) GetIgnorePatterns() []string {
    return c.IgnorePatterns
}
```

## 3. Migration Strategy

### Phase 0: Build New System (Safe - No Changes to Existing Code) ✅ COMPLETE

1. ✅ Created `internal/config/config_new.go` with the implementation above
2. ✅ Created comprehensive tests in `internal/config/config_new_test.go`
3. ✅ Test coverage includes:
   - Loading valid config files
   - Zero-config behavior (missing files return defaults)
   - Validation errors for invalid values
   - All getter methods work correctly
   - `Resolve()` returns self

**Phase 0 Testing Issue - RESOLVED:**
The initial implementation included a test `TestNewConfig_DefaultsMatch` that violated Phase 0 isolation by referencing `GetDefaults()` from the old configuration system. This has been addressed by commenting out the test with a clear note that it will be re-enabled in Phase 1 when verifying compatibility with the old system.

**Verification:** The new configuration system has been tested in complete isolation:
- All tests pass when run independently from the old system
- No dependencies on existing configuration code
- Ready for Phase 1 integration

### Phase 1: Create Compatibility Layer

Before migrating commands, create a compatibility shim:

```go
// In internal/config/compat.go (temporary file)

// ResolvedConfig is now just an alias for Config
type ResolvedConfig = Config

// ConfigManager wraps the new simple API for compatibility
type ConfigManager struct {
    configDir string
}

func NewConfigManager(configDir string) *ConfigManager {
    return &ConfigManager{configDir: configDir}
}

func (m *ConfigManager) LoadOrCreate() (*Config, error) {
    return Load(m.configDir)
}

// LoadConfigWithDefaults matches existing function signature
func LoadConfigWithDefaults(configDir string) *Config {
    return LoadWithDefaults(configDir)
}
```

### Phase 2: Replace Implementation (Atomic Switch)

1. Rename current files to `.old` suffix
2. Rename `config_new.go` to `config.go`
3. Add compatibility layer
4. Run all tests - they should pass with no changes to commands

### Phase 3: Gradual Cleanup

1. Remove compatibility layer usage from commands one at a time
2. Delete compatibility layer
3. Delete all `.old` files

## 4. Files to Delete (After Migration)

```
internal/config/
├── adapters.go (154 lines)
├── adapters_test.go
├── defaults.go (43 lines)
├── interfaces.go (39 lines)
├── interfaces_test.go
├── loader.go (134 lines)
├── loader_test.go
├── mock_config.go (generated)
├── resolved.go (69 lines)
├── schema.go (unused)
├── simple_validator.go (180 lines)
├── simple_validator_test.go
├── yaml_config.go (300+ lines)
├── yaml_config_test.go
└── zero_config_test.go
```

Total: 3000+ lines → ~200 lines (93% reduction)

## 5. Risk Mitigation

1. **Zero-Config Behavior**: The new `LoadWithDefaults` exactly matches current behavior
2. **API Compatibility**: `Resolve()` and all getters maintained
3. **Validation Differences**: Struct tags provide same validation as manual validator
4. **Testing**: Comprehensive test suite before any migration
5. **Atomic Switch**: Replace implementation in one commit, all tests must pass

## 6. Success Criteria

- [ ] All existing tests pass without modification
- [ ] `plonk.yaml` files work identically
- [ ] Zero-config behavior preserved
- [ ] Line count reduced by >90%
- [ ] Single file implementation
- [ ] Standard library approach (yaml tags, validate tags)
