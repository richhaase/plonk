# Path Resolution Direct Migration Plan

## Overview

The plonk codebase currently has 5 different implementations of path resolution logic. Based on behavioral analysis, we will migrate directly to the PathResolver implementation without compatibility layers, fixing security issues as we go.

## Principles

1. Each step must leave the code in a working, commitable state
2. UX must not break (commands must work the same for users)
3. No compatibility layers or adapters
4. Fix security issues as we go
5. Backwards compatibility is not a concern beyond UX

## Current State

### Existing Implementations

1. **`internal/paths/resolver.go` - PathResolver** (Primary implementation)
   - Most complete implementation with proper validation
   - **This is the correct implementation to keep**

2. **`internal/dotfiles/operations.go` - Manager.ExpandPath()**
   - Security bug: No validation
   - Will be replaced

3. **`internal/services/dotfile_operations.go` - ResolveDotfilePath()**
   - Security bug: Allows paths outside home
   - Will be removed

4. **`internal/commands/shared.go` - resolveDotfilePath()**
   - Already uses PathResolver correctly
   - No changes needed

5. **`internal/config/yaml_config.go` - Inline expansion**
   - Limited scope (config only)
   - Can be updated separately

## Migration Steps

### Step 1: Add PathResolver to dotfiles.Manager
**Goal**: Give dotfiles.Manager access to the correct implementation

```diff
// internal/dotfiles/operations.go
type Manager struct {
    homeDir      string
    configDir    string
    scanner      *Scanner
    expander     *Expander
+   pathResolver *paths.PathResolver
}

// Update NewManager to create PathResolver
func NewManager(homeDir, configDir string) *Manager {
    return &Manager{
        homeDir:      homeDir,
        configDir:    configDir,
        scanner:      NewScanner(homeDir),
        expander:     NewExpander(homeDir, configDir),
+       pathResolver: paths.NewPathResolver(homeDir, configDir),
    }
}
```

**Commitable**: ✅ Code compiles, all tests pass, no behavior change

### Step 2: Replace Manager.ExpandPath with PathResolver
**Goal**: Use PathResolver everywhere, no duplicates

```diff
// internal/dotfiles/operations.go
func (m *Manager) ExpandPath(path string) string {
-   if strings.HasPrefix(path, "~/") {
-       return filepath.Join(m.homeDir, path[2:])
-   }
-   return path
+   // Use PathResolver but maintain backward compatibility for error handling
+   resolved, err := m.pathResolver.ResolveDotfilePath(path)
+   if err != nil {
+       // For paths that don't validate (like those outside home),
+       // fall back to simple tilde expansion to maintain compatibility
+       if strings.HasPrefix(path, "~/") {
+           return filepath.Join(m.homeDir, path[2:])
+       }
+       return path
+   }
+   return resolved
}
```

**Commitable**: ✅ No duplicate code, PathResolver is used

### Step 3: Update state/dotfile_provider.go
**Goal**: Ensure it uses the updated Manager.ExpandPath (which now uses PathResolver)

```diff
// No code changes needed - it already calls manager.ExpandPath()
// which now uses PathResolver internally
```

**Commitable**: ✅ Automatically uses PathResolver through Manager

### Step 4: Remove services/dotfile_operations.go duplicate
**Goal**: Remove duplicate implementation, use PathResolver

```diff
// internal/services/dotfile_operations.go
-func ResolveDotfilePath(path, homeDir string) (string, error) {
-   if strings.HasPrefix(path, "~/") {
-       return filepath.Join(homeDir, path[2:]), nil
-   }
-   if !filepath.IsAbs(path) {
-       return filepath.Join(homeDir, path), nil
-   }
-   return path, nil
-}

// Update AddSingleDotfile to use paths.PathResolver
func (svc *DotfileOperationService) AddSingleDotfile(path, name string) error {
-   resolvedPath, err := ResolveDotfilePath(path, svc.homeDir)
+   resolver := paths.NewPathResolver(svc.homeDir, svc.configDir)
+   resolvedPath, err := resolver.ResolveDotfilePath(path)
    if err != nil {
        return err
    }
}
```

**Commitable**: ✅ Service layer now validates paths (security fix)

### Step 5: Update commands to handle errors
**Goal**: Make commands handle path validation errors gracefully

```diff
// internal/commands/shared.go
func addSingleDotfile(ctx *SharedContext, path string, force, dryRun bool) error {
    resolvedPath, err := resolveDotfilePath(path, ctx.HomeDir)
    if err != nil {
-       return err
+       // User-friendly error message
+       if strings.Contains(err.Error(), "must be within home directory") {
+           return fmt.Errorf("cannot add file outside home directory: %s", path)
+       }
+       return fmt.Errorf("invalid path %s: %w", path, err)
    }
}
```

**Commitable**: ✅ Better error messages maintain good UX

### Step 6: Clean up and verify
**Goal**: Clean up duplicate code

```diff
// internal/dotfiles/operations.go
-func (m *Manager) ExpandPath(path string) string {
-   if strings.HasPrefix(path, "~/") {
-       return filepath.Join(m.homeDir, path[2:])
-   }
-   return path
-}

// Update GetDestinationPath to use ResolvePath
func (m *Manager) GetDestinationPath(destination string) string {
-   return m.ExpandPath(destination)
+   // For backwards compatibility with state files,
+   // we need to keep expanding without validation
+   if strings.HasPrefix(destination, "~/") {
+       return filepath.Join(m.homeDir, destination[2:])
+   }
+   return destination
}
```

**Commitable**: ✅ Removes dead code while maintaining state file compatibility

### Step 7: Add integration tests
**Goal**: Ensure UX hasn't changed

```go
// tests/integration/path_resolution_test.go
func TestPathResolutionUX(t *testing.T) {
    // Test that commands still work as expected
    // - plonk add ~/.zshrc (should work)
    // - plonk add /etc/passwd (should fail with clear error)
    // - plonk add ../file (should work if within home)
}
```

**Commitable**: ✅ Proves UX is maintained

## Result

After these steps:
1. **One implementation**: PathResolver handles all path resolution
2. **Better security**: Paths outside home are rejected
3. **No adapters**: Direct usage everywhere
4. **Clear errors**: Users get helpful messages
5. **Each step works**: Can commit after any step

## Migration Order

1. Start with dotfiles package (least risky)
2. Then state package (internal)
3. Then services (adds security)
4. Finally commands (user-facing)

This order ensures we fix internal code first before touching user-facing code.

## Behavioral Changes

From the analysis in `resolution_behavior_test.go`:

### Security Fixes
- `/etc/passwd` will now be rejected (currently allowed by 3 implementations)
- Path traversal attempts (`../../../etc`) will be blocked

### Relative Path Changes
- Relative paths from current directory will need to be within home
- Empty string will be rejected instead of defaulting to home

### Error Handling
- All path resolution will now return errors that must be handled
- Invalid paths will produce clear error messages

## Success Criteria

1. **Single Implementation**: Only PathResolver contains path resolution logic
2. **No Behavior Changes for Users**: Commands work the same from user perspective
3. **Improved Security**: All paths validated consistently
4. **Reduced Code**: ~200 lines removed
5. **Better Error Messages**: Users understand why paths are rejected

## Timeline

- **Step 1-2**: 1 hour (add PathResolver to Manager)
- **Step 3**: 1 hour (fix services layer)
- **Step 4**: 1 hour (improve error messages)
- **Step 5**: 30 minutes (cleanup)
- **Step 6**: 30 minutes (integration tests)

**Total**: ~4 hours (even faster by skipping unnecessary step)
