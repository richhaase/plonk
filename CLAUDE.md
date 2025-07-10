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

### Phase 1: Zero-Config Infrastructure ✅
**Status**: COMPLETED  
**Goal**: Create foundation for optional configuration

- [x] Convert all Settings fields to pointers (*string, *int, *[]string)
- [x] Create ConfigDefaults struct with all default values  
- [x] Create ResolvedConfig struct for final computed values
- [x] Implement GetDefaults() function as single source of truth
- [x] Update YAML tags to use omitempty for all fields
- [x] Create Config.Resolve() method to merge defaults + user overrides
- [x] Implement individual resolution methods (getDefaultManager, getOperationTimeout, etc.)
- [x] Remove existing Get*() methods from Config struct
- [x] Update validation to work with pointer fields
- [x] Ensure nil pointer handling throughout resolution
- [x] Fix all command files to use Config.Resolve() API
- [x] Add helper functions (StringPtr, IntPtr, StringSlicePtr) for tests

### Phase 2: Make Configuration Optional ✅
**Status**: COMPLETED  
**Goal**: Handle missing config files gracefully with defaults

- [x] Update LoadConfig() to handle missing config file gracefully
- [x] Return default-only Config when no config file exists
- [x] Remove config file existence requirements from commands
- [x] Update error handling to not fail on missing config
- [x] Add GetOrCreateConfig() helper for commands that need to save config
- [x] Implement `plonk init` command for creating starter configurations

### Phase 3: Testing and Documentation ✅
**Status**: COMPLETED  
**Goal**: Comprehensive testing and user documentation

- [x] Add tests for zero-config scenarios
- [x] Test config resolution with partial user overrides
- [x] Update documentation to emphasize optional nature
- [x] Update examples to show minimal/empty config files
- [x] Add "Getting Started" guide with zero setup

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

## Progress Tracking

### Phase 1: Zero-Config Infrastructure ✅ COMPLETED 
**Completed**: 2025-07-10  
**Commit**: `149928f - feat: implement Zero-Config Phase 1 - pointer-based configuration infrastructure`

**Key Achievements**:
- **Pointer-Based Configuration**: All Settings fields converted to pointers to distinguish nil (not set) vs zero values
- **Centralized Defaults**: Created ConfigDefaults struct with all default values in one location
- **Resolved Configuration**: ResolvedConfig struct provides clean computed values for commands
- **Configuration Resolution**: Config.Resolve() method merges defaults with user overrides
- **Validation Updates**: Updated validation to handle optional pointer fields properly
- **Command Integration**: All commands now use cfg.Resolve().GetDefaultManager() pattern
- **Helper Functions**: Added StringPtr(), IntPtr(), StringSlicePtr() for easier testing
- **Backwards Compatibility**: Existing config files continue to work exactly as before

**Files Created**:
- `internal/config/defaults.go` - Centralized default values
- `internal/config/resolved.go` - Final computed configuration structure with helper functions

**Files Modified**:
- `internal/config/yaml_config.go` - Added pointer fields and Config.Resolve() method
- `internal/config/simple_validator.go` - Updated validation for pointer fields
- `internal/commands/*.go` - Updated all commands to use Config.Resolve() API
- Multiple test files - Updated to use pointer helpers and new validation logic

**Technical Debt Resolved**:
- Eliminated hardcoded default values scattered throughout codebase
- Removed ambiguity between "not set" vs "set to zero value"
- Created single source of truth for default configuration values
- Established clean patterns for configuration resolution

**All Tests Pass** ✅ and **Linter Clean** ✅

### Lessons Learned from Phase 1

1. **Complete Changes Over Incremental**: Making cohesive changes across all affected files in a single commit works better than breaking it into tiny pieces that leave the codebase in broken states

2. **Pointer Field Patterns**: Using helper functions like StringPtr() makes test code much cleaner and more readable than inline pointer creation

3. **Validation Complexity**: Updating validation to handle pointer fields required careful attention to nil checks and proper error message formatting

4. **Command Integration**: The Config.Resolve() pattern provides a clean API that hides pointer complexity from command code

5. **Test Maintenance**: Converting existing tests to use the new pointer-based structure requires systematic attention to comparison operations and error message formatting

### Phase 2: Make Configuration Optional ✅ COMPLETED  
**Completed**: 2025-07-10  
**Commits**: Multiple commits implementing zero-config functionality

**Key Achievements**:
- **Zero-Config LoadConfig()**: Modified LoadConfig() to return empty config (all defaults) when plonk.yaml doesn't exist
- **Graceful Error Handling**: Commands no longer fail due to missing config files
- **GetOrCreateConfig() Helper**: Added utility function for commands that need to save configuration  
- **plonk init Command**: Created comprehensive init command with helpful config template and comments
- **End-to-End Testing**: Verified zero-config behavior works across all major commands (status, doctor, config show)

**Files Created**:
- `internal/commands/init.go` - New plonk init command implementation

**Files Modified**:
- `internal/config/yaml_config.go` - Updated LoadConfig() for zero-config behavior, added GetOrCreateConfig()
- `internal/commands/dot_add.go` - Updated to use GetOrCreateConfig() helper
- `internal/commands/apply_test.go` - Updated test to expect zero-config behavior
- `internal/config/yaml_config_test.go` - Updated test to verify zero-config works
- `internal/errors/types.go` - Added ErrFileExists error type for init command

**Zero-Config User Experience Achieved**:
- Users can install Plonk and immediately use it without any configuration
- `plonk status`, `plonk doctor`, `plonk config show` all work without config files
- `plonk init` provides easy way to create config when customization is desired
- Doctor command provides helpful guidance but doesn't block functionality

**Technical Implementation**:
- LoadConfig() returns `&Config{}` when file missing (resolves to all defaults)
- Commands use `cfg.Resolve()` pattern to get computed values
- Backwards compatibility maintained - existing configs work exactly as before

### Lessons Learned from Phase 2

1. **Zero-Config Pattern**: The approach of returning empty structs that resolve to defaults is clean and maintainable
2. **User Experience Focus**: Adding `plonk init` immediately after implementing zero-config provides complete workflow
3. **Helpful Config Templates**: Generated config files with comments significantly improve user experience
4. **Test Updates Required**: Zero-config changes require updating tests that expect errors for missing files

### Next Steps: Phase 3

Phase 3 will focus on comprehensive testing of zero-config scenarios and updating documentation to emphasize the optional nature of configuration.