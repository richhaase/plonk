# Gem (Ruby) Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding Ruby gem support to plonk. The implementation will follow plonk's existing patterns while handling gem-specific behaviors around Ruby version managers and gem installation paths.

**Status**: Planning Phase

## Design Principles

1. **Focus on Global Gems Only** - Track only gems installed system-wide or user-wide
2. **Version Manager Agnostic** - Work with rbenv, rvm, chruby, or system Ruby
3. **Executable Gems Only** - Focus on gems that provide CLI tools
4. **User Installation Preferred** - Use `--user-install` when possible

## Key Challenges and Solutions

### 1. Multiple Ruby Environments
**Challenge**: Users may have system Ruby, rbenv, rvm, chruby, asdf, etc.

**Solution**:
- Use whatever `gem` is in PATH (like pip approach)
- Don't try to detect or manage Ruby versions
- State reconciliation naturally handles environment switches
- Work with currently active Ruby

### 2. Installation Scope
**Challenge**: Gems can be installed system-wide, user-wide, or in gemsets (rvm)

**Solution**:
- Prefer user installation with `gem install --user-install`
- Use `gem list --local` to detect all gems
- Document that plonk tracks the active Ruby's gems
- Ignore vendored gems (project-specific)

### 3. Executable Detection
**Challenge**: Not all gems provide executables, plonk should only track CLI tools

**Solution**:
- Check gem specification for executables
- Use `gem contents <gem> --executables` to list binaries
- Only track gems that provide executables
- Filter display to show executable gems

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/gem.go`
```go
type GemManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for gem binary and Ruby
- `ListInstalled()` - Use `gem list --local --no-versions` + filter for executables
- `Install()` - Use `gem install --user-install <gem>`
- `Uninstall()` - Use `gem uninstall <gem>`
- `IsInstalled()` - Check with `gem list --local <gem>`
- `Search()` - Use `gem search <query>`
- `Info()` - Use `gem specification <gem>` for details
- `GetInstalledVersion()` - Parse from `gem list` output

#### 1.2 Register in Manager Registry
- Add "gem" to `internal/managers/registry.go`
- Ensure proper initialization in factory method

#### 1.3 Handle gem-specific edge cases
- Deal with --user-install path configuration
- Handle gems with multiple executables
- Support version specifications during install
- Handle gem dependencies gracefully

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/gem_test.go`)
- Mock command executor for all gem commands
- Test all interface methods
- Test error conditions (gem not found, gem not found, etc.)
- Test version parsing

#### 2.2 Integration Tests
- Test with real gem if available
- Test with common executable gems (bundler, rails, rubocop)
- Test user vs system installation detection
- Test with different Ruby version managers

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add gem to PACKAGE_MANAGERS.md
- Update CLI.md with gem examples
- Document Ruby version manager behavior

#### 3.2 Error Messages
- Add gem-specific error messages and suggestions
- Handle common Ruby/gem issues (permissions, gem not found)

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `gem --version` | Verify gem is functional |
| List installed | `gem list --local --no-versions` | Filter for executable gems |
| Install | `gem install --user-install <gem>` | Prefer user installation |
| Uninstall | `gem uninstall <gem>` | May prompt for version |
| Check if installed | `gem list --local <gem>` | Check specific gem |
| Search | `gem search <query>` | Search RubyGems.org |
| Get info | `gem specification <gem>` | Detailed gem info |
| Get version | `gem list <gem>` | Parse version from output |

### Data Structures

```go
// Gem-specific information
type GemInfo struct {
    Name         string
    Version      string
    Executables  []string  // List of provided CLI tools
    Description  string
    Homepage     string
}
```

### Error Handling

Following plonk's error patterns:
```go
// gem not found
return false, nil  // Not an error, just unavailable

// Gem not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "gem",
    fmt.Sprintf("gem '%s' not found", name)).
    WithSuggestionMessage("Try: plonk search <gem> or check RubyGems.org")

// Permission error
return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "gem",
    fmt.Sprintf("permission denied installing %s", name)).
    WithSuggestionMessage("Try using --user-install or check Ruby installation")
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - gem found and functional
   - gem not found
   - Ruby not found

2. **List Tests**
   - Filter executable vs library gems
   - Handle gems with multiple versions
   - Empty gem list

3. **Install/Uninstall Tests**
   - Successful operations
   - Gem not found on RubyGems.org
   - Permission errors
   - Already installed/not installed

4. **Search/Info Tests**
   - Found gems
   - Not found gems
   - Parse gem specifications

### Mock Examples
```go
// Mock successful list command
executor.EXPECT().CommandContext(ctx, "gem", "list", "--local", "--no-versions").
    Return("bundler\nrails\nrubocop\nrake\n", nil)

// Mock gem with executables check
executor.EXPECT().CommandContext(ctx, "gem", "contents", "rubocop", "--executables").
    Return("rubocop\n", nil)

// Mock version detection
executor.EXPECT().CommandContext(ctx, "gem", "list", "rubocop").
    Return("rubocop (1.56.0, 1.55.0)\n", nil)
```

## Future Considerations

1. **Gemsets Support** - Handle rvm gemsets for isolated environments
2. **Bundler Integration** - Consider system-wide bundler-installed tools
3. **Binary Stubs** - Handle rbenv/rvm binary stubs correctly
4. **Gem Sources** - Support custom gem sources beyond RubyGems.org
5. **Platform-Specific Gems** - Handle gems with native extensions

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Comprehensive test coverage (>80%)
3. ✅ Handles multiple Ruby environments gracefully
4. ✅ Clear error messages with actionable suggestions
5. ✅ Documentation updated
6. ✅ Works with rbenv, rvm, and system Ruby
7. ✅ Follows plonk's existing patterns and conventions

## Key Differences from Other Managers

1. **Executable Focus** - Only track gems that provide CLI tools
2. **Version Managers** - Must work across rbenv, rvm, chruby
3. **User Install Flag** - Prefer --user-install for permissions
4. **Multiple Versions** - Gems can have multiple versions installed

## Common Ruby Gems to Test With

- `bundler` - Ruby dependency manager
- `rails` - Web framework (provides rails CLI)
- `rubocop` - Ruby linter
- `pry` - Enhanced Ruby REPL
- `rake` - Build tool
- `rspec` - Testing framework (provides rspec CLI)
- `solargraph` - Language server
- `jekyll` - Static site generator

## Implementation Notes from pip Experience

1. **Path Detection** - Like pip, use whatever gem is in PATH
2. **Environment Switches** - State reconciliation handles version manager switches
3. **Name Normalization** - Gems typically use underscores and hyphens consistently
4. **Error Handling** - Follow established plonk patterns
5. **User Installation** - Like pip's --user, prefer --user-install

## Timeline Estimate

- Phase 1 (Core Implementation): 2-3 hours
- Phase 2 (Testing): 2-3 hours
- Phase 3 (Documentation): 1 hour

Total: ~5-7 hours of development time
