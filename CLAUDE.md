# Plonk Zero-Config Implementation

## Overview

This document tracks the implementation of the Zero-Config approach for Plonk configuration management. The goal is to make configuration completely optional with sensible defaults, allowing users to run `plonk pkg add git` immediately without any setup.

## Design Principles

1. **Zero-Config by Default**: Plonk works out of the box without any configuration
2. **Optional Overrides**: Users can optionally customize settings via `plonk.yaml`
3. **Centralized Defaults**: Single source of truth for all default values
4. **Clear Resolution**: Explicit separation between defaults, user overrides, and resolved values
5. **Future-Proof**: Easy to add new settings without breaking existing installations

## Current State Analysis

After the successful lock file implementation, the configuration system has these characteristics:
- **Separated Concerns**: Package state moved to `plonk.lock`, config only handles settings
- **Good Defaults**: All settings have sensible defaults in various `Get*()` methods
- **Scattered Logic**: Defaults are duplicated across multiple methods and locations
- **Required Config**: Currently requires a config file to exist, even if minimal

## Target Architecture

### File Structure

```
~/.config/plonk/
├── plonk.yaml (optional - user overrides only)
└── plonk.lock (managed - package state)
```

### Configuration Types

```go
// User configuration (optional file)
type Config struct {
    Settings       *Settings `yaml:"settings,omitempty"`
    IgnorePatterns []string  `yaml:"ignore_patterns,omitempty"`
}

type Settings struct {
    DefaultManager    *string   `yaml:"default_manager,omitempty"`
    OperationTimeout  *int      `yaml:"operation_timeout,omitempty"`
    PackageTimeout    *int      `yaml:"package_timeout,omitempty"`
    DotfileTimeout    *int      `yaml:"dotfile_timeout,omitempty"`
    ExpandDirectories *[]string `yaml:"expand_directories,omitempty"`
}

// Centralized defaults
type ConfigDefaults struct {
    DefaultManager    string
    OperationTimeout  int
    PackageTimeout    int
    DotfileTimeout    int
    ExpandDirectories []string
    IgnorePatterns    []string
}

// Resolved configuration (defaults + user overrides)
type ResolvedConfig struct {
    DefaultManager    string
    OperationTimeout  int
    PackageTimeout    int
    DotfileTimeout    int
    ExpandDirectories []string
    IgnorePatterns    []string
}
```

## Implementation Plan

### Phase 1: Refactor Configuration Structure
- [ ] Convert all Settings fields to pointers (*string, *int, *[]string)
- [ ] Create ConfigDefaults struct with all default values
- [ ] Create ResolvedConfig struct for final computed values
- [ ] Implement GetDefaults() function as single source of truth
- [ ] Update YAML tags to use omitempty for all fields

### Phase 2: Implement Resolution Logic
- [ ] Create Config.Resolve() method to merge defaults + user overrides
- [ ] Implement individual resolution methods (getDefaultManager, getOperationTimeout, etc.)
- [ ] Remove existing Get*() methods from Config struct
- [ ] Update validation to work with pointer fields
- [ ] Ensure nil pointer handling throughout resolution

### Phase 3: Make Configuration Optional
- [ ] Update LoadConfig() to handle missing config file gracefully
- [ ] Return default-only ResolvedConfig when no config file exists
- [ ] Remove config file existence requirements from commands
- [ ] Update error handling to not fail on missing config
- [ ] Add GetOrCreateConfig() helper for commands that need to save config

### Phase 4: Update Command Integration
- [ ] Update all commands to use ResolvedConfig instead of Config
- [ ] Replace config.Settings references with resolvedConfig fields
- [ ] Update config adapters to work with ResolvedConfig
- [ ] Ensure backward compatibility for existing config files
- [ ] Test all commands work without any config file

### Phase 5: Testing and Documentation
- [ ] Add tests for zero-config scenarios
- [ ] Test config resolution with partial user overrides
- [ ] Update documentation to emphasize optional nature
- [ ] Update examples to show minimal/empty config files
- [ ] Add "Getting Started" guide with zero setup

## Technical Details

### Default Values

```go
func GetDefaults() ConfigDefaults {
    return ConfigDefaults{
        DefaultManager:    "homebrew",
        OperationTimeout:  300, // 5 minutes
        PackageTimeout:    180, // 3 minutes
        DotfileTimeout:    60,  // 1 minute
        ExpandDirectories: []string{
            ".config", ".ssh", ".aws", ".kube", 
            ".docker", ".gnupg", ".local",
        },
        IgnorePatterns: []string{
            ".DS_Store", ".git", "*.backup", "*.tmp", "*.swp",
        },
    }
}
```

### Resolution Logic

```go
func (c *Config) Resolve() *ResolvedConfig {
    defaults := GetDefaults()
    
    return &ResolvedConfig{
        DefaultManager:    c.getDefaultManager(defaults.DefaultManager),
        OperationTimeout:  c.getOperationTimeout(defaults.OperationTimeout),
        PackageTimeout:    c.getPackageTimeout(defaults.PackageTimeout),
        DotfileTimeout:    c.getDotfileTimeout(defaults.DotfileTimeout),
        ExpandDirectories: c.getExpandDirectories(defaults.ExpandDirectories),
        IgnorePatterns:    c.getIgnorePatterns(defaults.IgnorePatterns),
    }
}

func (c *Config) getDefaultManager(defaultValue string) string {
    if c.Settings != nil && c.Settings.DefaultManager != nil {
        return *c.Settings.DefaultManager
    }
    return defaultValue
}
```

### Configuration Loading

```go
func LoadConfig(configDir string) (*ResolvedConfig, error) {
    // Try to load user config
    userConfig, err := tryLoadUserConfig(configDir)
    if err != nil && !os.IsNotExist(err) {
        return nil, err // Real error, not just missing file
    }
    
    // If no config file, use defaults only
    if userConfig == nil {
        userConfig = &Config{}
    }
    
    // Resolve final configuration
    return userConfig.Resolve(), nil
}
```

## User Experience

### Before (Current)
```bash
# User must create config first
plonk init  # or manually create plonk.yaml
plonk pkg add git
```

### After (Zero-Config)
```bash
# Works immediately
plonk pkg add git  # Uses homebrew by default

# Optional customization
echo "settings:\n  default_manager: npm" > ~/.config/plonk/plonk.yaml
plonk pkg add typescript  # Now uses npm by default
```

## Migration Strategy

### Backward Compatibility
- Existing `plonk.yaml` files continue to work unchanged
- No breaking changes to configuration format
- Validation remains the same (just handles nil values)

### Upgrade Path
1. **Immediate**: All existing installations get zero-config benefits
2. **Gradual**: Users can simplify configs by removing default values
3. **Documentation**: Update guides to show minimal configs

## Benefits

### For New Users
- **Instant Gratification**: `plonk pkg add git` works immediately
- **Zero Learning Curve**: No configuration knowledge required
- **Progressive Enhancement**: Add config only when needed

### For Existing Users
- **Simplified Configs**: Remove redundant default values
- **Cleaner Documentation**: Focus on customization, not setup
- **Better Defaults**: Centralized, well-tested default values

### For Developers
- **Maintainable**: Single source of truth for defaults
- **Testable**: Easy to test both zero-config and custom scenarios
- **Extensible**: Simple to add new settings with defaults

## Implementation Notes

### Field Pointer Strategy
Using pointers allows distinguishing between:
- **Not specified** (`nil`) - use default
- **Explicitly set to zero** (`&0`) - use zero value
- **Explicitly set** (`&value`) - use user value

### Validation Updates
```go
func (v *SimpleValidator) ValidateConfig(config *Config) *ValidationResult {
    // Validate only non-nil fields
    if config.Settings != nil {
        if config.Settings.DefaultManager != nil {
            // Validate manager value
        }
        if config.Settings.OperationTimeout != nil {
            // Validate timeout range
        }
    }
    return result
}
```

### Error Handling
```go
func LoadConfig(configDir string) (*ResolvedConfig, error) {
    // Distinguish between "file doesn't exist" and "file is corrupted"
    userConfig, err := tryLoadUserConfig(configDir)
    if err != nil {
        if os.IsNotExist(err) {
            // Not an error - use defaults
            userConfig = &Config{}
        } else {
            // Real error - corrupted file, permissions, etc.
            return nil, errors.Wrap(err, ...)
        }
    }
    
    return userConfig.Resolve(), nil
}
```

## Testing Strategy

### Unit Tests
- [ ] Test resolution with nil settings
- [ ] Test resolution with partial settings
- [ ] Test resolution with full custom settings
- [ ] Test validation with pointer fields
- [ ] Test config loading with missing file

### Integration Tests
- [ ] Test commands work without config file
- [ ] Test commands work with minimal config
- [ ] Test config creation/editing workflows
- [ ] Test migration from existing configs

### End-to-End Tests
- [ ] Fresh installation workflow
- [ ] Zero-config package management
- [ ] Progressive configuration setup

## Documentation Updates

### README.md
- Update "Quick Start" to show zero-config workflow
- Emphasize optional nature of configuration
- Show minimal configuration examples

### CONFIGURATION.md
- Lead with "Configuration is optional" message
- Show examples of minimal overrides
- Document all available settings with defaults

### CLI.md
- Remove configuration requirements from command descriptions
- Update examples to work without config

## Future Enhancements

### Environment Variable Support
```go
func GetDefaults() ConfigDefaults {
    return ConfigDefaults{
        DefaultManager: getEnvOrDefault("PLONK_DEFAULT_MANAGER", "homebrew"),
        // ...
    }
}
```

### Multiple Config Sources
- Global config: `/etc/plonk/plonk.yaml`
- User config: `~/.config/plonk/plonk.yaml`
- Project config: `./plonk.yaml`

### Dynamic Defaults
```go
func GetDefaults() ConfigDefaults {
    defaults := ConfigDefaults{...}
    
    // Platform-specific defaults
    if runtime.GOOS == "linux" {
        defaults.DefaultManager = "apt"
    }
    
    return defaults
}
```

## Success Criteria

### User Experience
- [ ] New users can run `plonk pkg add git` without any setup
- [ ] Existing configs continue to work unchanged
- [ ] Configuration file is truly optional
- [ ] Clear documentation about default values

### Code Quality
- [ ] Single source of truth for all defaults
- [ ] Clean separation between user config and resolved config
- [ ] Robust error handling for missing/corrupted configs
- [ ] Comprehensive test coverage

### Maintainability
- [ ] Easy to add new settings with defaults
- [ ] Clear patterns for config resolution
- [ ] Minimal code duplication
- [ ] Self-documenting default values

## Phase 1 Implementation Details

### Step 1: Update Configuration Types
**File**: `internal/config/yaml_config.go`

**Current Structure**:
```go
type Settings struct {
    DefaultManager    string   `yaml:"default_manager" validate:"required,oneof=homebrew npm cargo"`
    OperationTimeout  int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
    // ...
}
```

**New Structure**:
```go
type Settings struct {
    DefaultManager    *string   `yaml:"default_manager,omitempty" validate:"omitempty,oneof=homebrew npm cargo"`
    OperationTimeout  *int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
    PackageTimeout    *int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`
    DotfileTimeout    *int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`
    ExpandDirectories *[]string `yaml:"expand_directories,omitempty"`
}
```

### Step 2: Create Default Values Structure
**File**: `internal/config/defaults.go` (new file)

```go
package config

type ConfigDefaults struct {
    DefaultManager    string
    OperationTimeout  int
    PackageTimeout    int
    DotfileTimeout    int
    ExpandDirectories []string
    IgnorePatterns    []string
}

func GetDefaults() ConfigDefaults {
    return ConfigDefaults{
        DefaultManager:    "homebrew",
        OperationTimeout:  300,
        PackageTimeout:    180,
        DotfileTimeout:    60,
        ExpandDirectories: []string{
            ".config", ".ssh", ".aws", ".kube",
            ".docker", ".gnupg", ".local",
        },
        IgnorePatterns: []string{
            ".DS_Store", ".git", "*.backup", "*.tmp", "*.swp",
        },
    }
}
```

### Step 3: Create Resolved Configuration Type
**File**: `internal/config/resolved.go` (new file)

```go
package config

type ResolvedConfig struct {
    DefaultManager    string
    OperationTimeout  int
    PackageTimeout    int
    DotfileTimeout    int
    ExpandDirectories []string
    IgnorePatterns    []string
}

// Timeout helper methods
func (r *ResolvedConfig) GetOperationTimeout() int { return r.OperationTimeout }
func (r *ResolvedConfig) GetPackageTimeout() int   { return r.PackageTimeout }
func (r *ResolvedConfig) GetDotfileTimeout() int   { return r.DotfileTimeout }
```

### Step 4: Implement Resolution Logic
**File**: `internal/config/yaml_config.go`

```go
func (c *Config) Resolve() *ResolvedConfig {
    defaults := GetDefaults()
    
    return &ResolvedConfig{
        DefaultManager:    c.getDefaultManager(defaults.DefaultManager),
        OperationTimeout:  c.getOperationTimeout(defaults.OperationTimeout),
        PackageTimeout:    c.getPackageTimeout(defaults.PackageTimeout),
        DotfileTimeout:    c.getDotfileTimeout(defaults.DotfileTimeout),
        ExpandDirectories: c.getExpandDirectories(defaults.ExpandDirectories),
        IgnorePatterns:    c.getIgnorePatterns(defaults.IgnorePatterns),
    }
}

func (c *Config) getDefaultManager(defaultValue string) string {
    if c.Settings != nil && c.Settings.DefaultManager != nil {
        return *c.Settings.DefaultManager
    }
    return defaultValue
}

func (c *Config) getOperationTimeout(defaultValue int) int {
    if c.Settings != nil && c.Settings.OperationTimeout != nil {
        return *c.Settings.OperationTimeout
    }
    return defaultValue
}

// Similar methods for other settings...
```

This foundation provides the structure for the zero-config approach while maintaining backward compatibility and setting up clean patterns for the remaining phases.