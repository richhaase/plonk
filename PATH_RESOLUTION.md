# Path Resolution Refactor: Completion Plan

## 1. Overview & Goal

**The Problem:** While a new `PathResolver` was introduced in `internal/paths`, the codebase still contains multiple, inconsistent, and potentially unsafe implementations of path resolution and manipulation logic outside of this centralized package. This violates the principle of a single source of truth and reintroduces the very complexity and security risks this refactor aimed to eliminate.

**The Goal:** To ensure that **all path resolution, validation, and manipulation logic** within the `plonk` codebase is exclusively delegated to the `internal/paths` package. This means:
*   No manual tilde (`~`) expansion.
*   No manual `filepath.Join`, `filepath.Abs` for resolving user-provided paths.
*   No manual `os.Stat` checks for path existence or type (file/dir) when the `PathResolver` can provide this.
*   No custom `sourceToTarget` or `TargetToSource` logic outside `internal/paths`.
*   No custom path filtering/skipping logic outside `internal/paths`.

**Guiding Principles:**
*   **Delegation:** If `internal/paths` can do it, it *must* do it.
*   **Consistency:** All path handling must behave identically across the application.
*   **Security:** Centralized validation in `internal/paths` must be the only gatekeeper for path safety.
*   **No UX Change:** The user experience of the CLI must remain identical.
*   **Test-Driven:** All changes must be verified by existing tests (`just test`, `just test-ux`).

## 2. Detailed Action Items (File by File)

The following sections detail the specific changes required in each identified file.

### 2.1. `internal/dotfiles/operations.go`

**Problem:** The `GetDestinationPath` function contains manual tilde expansion for "backward compatibility." This is a direct violation of the single source of truth.

**Action:**
1.  **Remove Manual Expansion:** Modify `GetDestinationPath` to use `p.pathResolver.GetDestinationPath` (or a similar method if one is added to `PathResolver` for this specific purpose).
2.  **Remove Comment:** Delete the misleading "backward compatibility" comment.

**Example (Conceptual Change):**

```diff
 // GetDestinationPath returns the full destination path for a dotfile
 func (m *Manager) GetDestinationPath(destination string) string {
-	// For backward compatibility with state files,
-	// we need to keep expanding without validation
-	if strings.HasPrefix(destination, "~/") {
-		return filepath.Join(m.homeDir, destination[2:])
-	}
-	return destination
+	// Delegate to the centralized PathResolver
+	resolvedPath, err := m.pathResolver.GetDestinationPath(destination)
+	if err != nil {
+		// Handle error appropriately, perhaps log and return original or a known error
+		// For now, assuming GetDestinationPath in PathResolver handles all cases
+		return destination // Fallback if PathResolver fails, but ideally it shouldn't for valid inputs
+	}
+	return resolvedPath
 }
```

### 2.2. `internal/services/dotfile_operations.go`

**Problem:** The `AddSingleFile` function contains a manual fallback for `resolver.GeneratePaths` if an error occurs. This means the `PathResolver` is not being fully trusted or utilized.

**Action:**
1.  **Remove Fallback Logic:** Eliminate the `if err != nil` block that manually constructs `relPath` and `destPath`.
2.  **Ensure `PathResolver` Handles All Cases:** The `resolver.GeneratePaths` function in `internal/paths/resolver.go` should be robust enough to handle all valid inputs without error. If it's not, enhance `resolver.go` first.

**Example (Conceptual Change):**

```diff
 // AddSingleFile adds a single file to dotfile management
 func AddSingleFile(ctx context.Context, options AddSingleFileOptions) operations.OperationResult {
 	result := operations.OperationResult{
 		Name: options.FilePath,
 	}

 	// Generate source and destination paths
 	resolver := paths.NewPathResolver(options.HomeDir, options.ConfigDir)
-	_, destPath, err := resolver.GeneratePaths(options.FilePath)
-	if err != nil {
-		// Fallback to simple relative path
-		relPath, _ := filepath.Rel(options.HomeDir, options.FilePath)
-		destPath = relPath
-	}
+	sourcePath, destPath, err := resolver.GeneratePaths(options.FilePath)
+	if err != nil {
+		result.Status = "failed"
+		result.Error = errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "add", "failed to generate paths")
+		return result
+	}

 	if options.DryRun {
 		result.Status = "would-add"
 		return result
 	}

 	// Copy file to plonk config directory
-	sourcePath := filepath.Join(options.ConfigDir, source) // 'source' is undefined in this context
-
-	// ... rest of the function
+	// Use sourcePath from resolver.GeneratePaths
+	// ... rest of the function
 }
```
*(Note: The original `source` variable in `AddSingleFile` was not defined, indicating a potential bug or incomplete refactor. The corrected example assumes `GeneratePaths` returns both source and destination.)*

### 2.3. `internal/commands/shared.go`

**Problem:** This file still contains `resolveDotfilePath` and `generatePaths` functions, which are redundant wrappers around `PathResolver` and include problematic fallback logic. It also contains `copyFileWithAttributes`, which is a utility function that doesn't belong here.

**Action:**
1.  **Remove `resolveDotfilePath`:**
    *   Delete the `resolveDotfilePath` function.
    *   Find all callers of `resolveDotfilePath` and replace them with direct calls to `paths.NewPathResolver(...).ResolveDotfilePath(...)`.
2.  **Remove `generatePaths`:**
    *   Delete the `generatePaths` function.
    *   Find all callers of `generatePaths` and replace them with direct calls to `paths.NewPathResolver(...).GeneratePaths(...)`. Ensure error handling is robust.
3.  **Move `copyFileWithAttributes`:**
    *   Move `copyFileWithAttributes` to `internal/dotfiles/fileops.go` (or a new `internal/util/fileops.go` if it's more general).
    *   Update all callers to use the new location.

### 2.4. `internal/config/yaml_config.go`

**Problem:** This file is a major offender, containing multiple forms of manual path manipulation and validation that should be centralized.

**Action:**
1.  **Refactor `GetDefaultConfigDirectory()`:**
    *   Modify this function to use `paths.NewPathResolverFromDefaults()` and its methods to determine the default config directory.
    *   Remove all manual `os.Getenv("HOME")`, `strings.HasPrefix(envDir, "~/")`, and `filepath.Join` logic.

    **Example (Conceptual Change):**
    ```diff
     // GetDefaultConfigDirectory returns the default config directory, checking PLONK_DIR environment variable first
     func GetDefaultConfigDirectory() string {
    -	// Check for PLONK_DIR environment variable
    -	if envDir := os.Getenv("PLONK_DIR"); envDir != "" {
    -		// Expand ~ if present
    -		if strings.HasPrefix(envDir, "~/") {
    -			return filepath.Join(os.Getenv("HOME"), envDir[2:])
    -		}
    -		return envDir
    -	}
    -
    -	// Default location
    -	return filepath.Join(os.Getenv("HOME"), ".config", "plonk")
    +	// Delegate to PathResolver for robust and consistent resolution
    +	resolver, err := paths.NewPathResolverFromDefaults()
    +	if err != nil {
    +		// Handle error, perhaps log and return a sensible default or panic if unrecoverable
    +		// For now, returning a hardcoded default as a fallback, but ideally PathResolver handles this.
    +		home := os.Getenv("HOME")
    +		if home == "" {
    +			home = "/tmp" // Fallback for testing or extreme cases
    +		}
    +		return filepath.Join(home, ".config", "plonk")
    +	}
    +	return resolver.ConfigDir() // Assuming PathResolver stores and exposes the resolved configDir
     }
    ```

2.  **Refactor `ConfigAdapter.GetDotfileTargets()`:**
    *   This method currently walks the `configDir` and manually resolves paths, skips files, and converts source to target. This entire block of logic needs to be replaced.
    *   It should delegate to `paths.PathResolver.ExpandDirectory` and `paths.PathResolver.GeneratePaths` (or similar methods that handle source/target conversion).
    *   The `shouldSkipDotfile` function should be integrated into the `PathResolver`'s directory expansion or a dedicated path filtering utility within `internal/paths`.

    **Example (Conceptual Change):**
    ```diff
     // GetDotfileTargets returns a map of source -> destination paths for dotfiles
     func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
         result := make(map[string]string)

    -	// Auto-discover dotfiles from configured directory
    -	configDir := GetDefaultConfigDirectory()
    -	resolvedConfig := c.config.Resolve()
    -	ignorePatterns := resolvedConfig.GetIgnorePatterns()
    -
    -	// Walk the directory to find all files
    -	_ = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
    -		if err != nil {
    -			return nil // Skip files we can't read
    -		}
    -
    -		// Get relative path from config dir
    -		relPath, err := filepath.Rel(configDir, path)
    -		if err != nil {
    -			return nil
    -		}
    -
    -		// Skip certain files and directories
    -		if shouldSkipDotfile(relPath, info, ignorePatterns) {
    -			if info.IsDir() {
    -				return filepath.SkipDir
    -			}
    -			return nil
    -		}
    -
    -		// Skip directories themselves (we'll get the files inside)
    -		if info.IsDir() {
    -			return nil
    -		}
    -
    -		// Add to results with proper mapping
    -		source := relPath
    -		target := sourceToTarget(source)
    -		result[source] = target
    -
    -		return nil
    -	})
    +	// Use PathResolver to expand directory and generate paths
    +	resolver, err := paths.NewPathResolverFromDefaults() // Or pass resolver from higher up
    +	if err != nil {
    +		// Handle error, log it, and return empty map or error
    +		return result
    +	}
    +
    +	// Assuming PathResolver.ExpandDirectory can take ignore patterns or a filter
    +	// Or, filter after expansion if PathResolver doesn't support it directly
    +	entries, err := resolver.ExpandDirectory(resolver.ConfigDir()) // Assuming ConfigDir is exposed
    +	if err != nil {
    +		// Handle error, log it, and return empty map or error
    +		return result
    +	}
    +
    +	for _, entry := range entries {
    +		// Assuming PathResolver.GeneratePaths can convert FullPath to source/destination
    +		source, destination, err := resolver.GeneratePaths(entry.FullPath)
    +		if err != nil {
    +			// Log error for this entry and continue
    +			continue
    +		}
    +		result[source] = destination
    +	}

         return result
     }
    ```

3.  **Remove `shouldSkipDotfile`, `sourceToTarget`, `TargetToSource`:**
    *   These functions should be deleted from `yaml_config.go`.
    *   Their logic should be fully integrated into the `internal/paths` package, either as methods on `PathResolver` or as standalone helper functions within `internal/paths` if they are truly generic path utilities.

## 3. Verification Steps

Upon completion of the above actions, the following verification steps must be performed:

1.  **Run Unit Tests:** Execute `just test`. All unit tests must pass.
2.  **Run UX Integration Tests:** Execute `just test-ux`. All integration tests must pass, confirming no user-facing behavior changes.
3.  **Code Inspection:** Manually inspect the modified files to ensure:
    *   No manual path manipulation (tilde expansion, `filepath.Join`, `os.Stat` for path existence/type) remains outside `internal/paths`.
    *   All path-related concerns are delegated to `internal/paths`.
    *   The `internal/paths` package is now the sole authority for these operations.
4.  **Confirm Deletion:** Ensure `shouldSkipDotfile`, `sourceToTarget`, and `TargetToSource` are removed from `yaml_config.go`.

## 4. Success Criteria

This phase of the refactor will be considered complete when:
*   All action items listed in section 2 are implemented.
*   All tests (`just test` and `just test-ux`) pass.
*   A manual code inspection confirms the complete delegation of path logic to `internal/paths`.

## 5. Next Steps

Once this plan is successfully executed and verified, we can confidently proceed with other refactoring efforts, knowing that our foundational path logic is sound and consistent.
