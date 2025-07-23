# Phase 4: Command Migration Plan

## Overview

Migrate all commands from old pointer-based Config API to new direct struct API. This will allow removal of the compatibility layer.

## API Changes Required

### Old API → New API Mapping

1. **Config Loading**
   - `config.LoadConfigWithDefaults()` → `config.LoadNewWithDefaults()`
   - `config.LoadConfig()` → `config.LoadNew()`

2. **Nil Checks**
   - `if cfg.DefaultManager != nil` → Not needed (always has value)
   - `*cfg.DefaultManager` → `cfg.DefaultManager`

3. **Resolve Pattern**
   - `cfg.Resolve().GetDefaultManager()` → `cfg.GetDefaultManager()`
   - `cfg.Resolve()` → Not needed (config is already resolved)

## Commands to Update

### Batch 1: Simple Commands (No pointer dereferencing)
- [x] env.go - Uses LoadConfigWithDefaults, no pointer checks
- [x] ls.go - Uses LoadConfigWithDefaults, no pointer checks
- [x] rm.go - Uses LoadConfigWithDefaults, no pointer checks
- [x] sync.go - Uses LoadConfigWithDefaults, no pointer checks
- [x] config_show.go - Uses Resolve() pattern
- [x] search.go - Uses Resolve().GetDefaultManager()

### Batch 2: Commands with Pointer Checks
- [ ] install.go - Has `cfg.DefaultManager != nil` check
- [ ] doctor.go - Has `cfg.DefaultManager != nil` check and Resolve()

### Batch 3: Test Files
- [ ] zero_config_test.go - Uses LoadConfig and Resolve()

### Batch 4: Other Packages
- [ ] internal/runtime/context.go - May use config loading
- [ ] internal/core/state.go - Uses ConfigManager

## Migration Strategy

1. Start with Batch 1 (simple commands)
2. Update and test each command individually
3. Move to Batch 2 (complex commands)
4. Update test files
5. Check other packages
6. Remove compatibility layer
7. Rename NewConfig → Config

## Example Migration

### Before:
```go
cfg := config.LoadConfigWithDefaults(configDir)
if cfg.DefaultManager != nil && *cfg.DefaultManager != "" {
    manager = *cfg.DefaultManager
}
defaultManager := cfg.Resolve().GetDefaultManager()
```

### After:
```go
cfg := config.LoadNewWithDefaults(configDir)
if cfg.DefaultManager != "" {
    manager = cfg.DefaultManager
}
defaultManager := cfg.GetDefaultManager()
```
