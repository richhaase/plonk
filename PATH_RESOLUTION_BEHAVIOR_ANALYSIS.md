# Path Resolution Behavior Analysis

## Summary of Findings

Based on the comprehensive test suite in `internal/paths/resolution_behavior_test.go`, here are the key behavioral differences between the path resolution implementations:

## Implementation Behaviors

### 1. PathResolver.ResolveDotfilePath (paths/resolver.go)
**Most comprehensive and secure implementation**

- **Tilde Expansion**: ✅ Handles `~/` correctly
- **Security**: ✅ Validates all paths must be within home directory
- **Relative Paths**: ❌ Tries current directory first, then fails with validation error
- **Absolute Paths**: ✅ Accepts if within home, rejects if outside
- **Edge Cases**: ✅ Handles empty strings, dots, path traversal attempts
- **Error Handling**: ✅ Returns descriptive errors

### 2. dotfiles.Manager.ExpandPath (dotfiles/operations.go)
**Minimal implementation - only tilde expansion**

- **Tilde Expansion**: ✅ Only expands `~/` prefix
- **Security**: ❌ No validation whatsoever
- **Relative Paths**: ❌ Returns as-is (no expansion)
- **Absolute Paths**: ❌ Returns as-is (no validation)
- **Edge Cases**: ❌ No special handling
- **Error Handling**: ❌ Never returns errors

### 3. services.ResolveDotfilePath (services/dotfile_operations.go)
**Inconsistent implementation**

- **Tilde Expansion**: ✅ Handles `~/` correctly
- **Security**: ❌ No validation
- **Relative Paths**: ⚠️ Always joins with home directory (different from PathResolver)
- **Absolute Paths**: ❌ Returns as-is (no validation)
- **Edge Cases**: ⚠️ Joins empty string with home directory
- **Error Handling**: ❌ Never returns errors

### 4. yaml_config tilde expansion
**Config-specific implementation**

- **Tilde Expansion**: ✅ Only expands `~/` prefix
- **Security**: ❌ No validation
- **Relative Paths**: ❌ Returns as-is
- **Absolute Paths**: ❌ Returns as-is
- **Edge Cases**: ❌ No special handling
- **Error Handling**: ❌ Never returns errors

## Critical Differences

### 1. Security Validation
- **Only PathResolver** validates that paths are within the home directory
- Other implementations allow dangerous paths like `/etc/passwd`

### 2. Relative Path Handling
- **PathResolver**: Tries current directory first, then validates (usually fails)
- **services.ResolveDotfilePath**: Always joins with home directory
- **Others**: Return relative paths unchanged

### 3. Error Handling
- **Only PathResolver** returns errors
- All other implementations silently accept invalid input

### 4. Edge Case Handling
Examples of different behaviors:

| Input | PathResolver | dotfiles.Manager | services | yaml_config |
|-------|--------------|------------------|----------|-------------|
| `~` | ERROR | `~` | `/home/user/~` | `~` |
| `""` | ERROR | `""` | `/home/user` | `""` |
| `.` | ERROR | `.` | `/home/user` | `.` |
| `..` | ERROR | `..` | `/home` | `..` |
| `/etc/passwd` | ERROR | `/etc/passwd` | `/etc/passwd` | `/etc/passwd` |

## Migration Risks

When consolidating to PathResolver, the following behavioral changes will occur:

1. **Security Enforcement**: Paths outside home directory will be rejected
2. **Relative Path Changes**: Current directory relative paths will behave differently
3. **Error Introduction**: Code that expects no errors will need error handling
4. **Empty String Handling**: Empty strings will be rejected instead of defaulting to home

## Recommendations

1. **Add Compatibility Mode**: PathResolver should have a "legacy mode" flag for gradual migration
2. **Create Migration Shims**: Temporary functions that maintain old behavior while logging warnings
3. **Update Tests First**: Ensure all tests explicitly document expected behavior
4. **Phase Migration**: Migrate one component at a time with feature flags

## Next Steps

With this behavioral analysis complete, Phase 1.1 is finished. The test suite provides:
- Complete behavioral documentation
- Regression test coverage
- Clear understanding of migration risks

Phase 2 can now proceed with confidence that we understand exactly how each implementation behaves.
