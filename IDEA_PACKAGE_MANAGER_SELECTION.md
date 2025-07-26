# IDEA: Simplify Package Manager Selection

## Current State

- Users can specify package manager with `--<manager>` flag
- Without flag, uses `default_manager` from config
- If package exists in multiple managers, default is used even if not optimal

Example:
```bash
plonk install ripgrep          # Uses default_manager
plonk install --brew ripgrep   # Forces homebrew
plonk install --cargo ripgrep  # Forces cargo
```

## Issues to Consider

1. **User Must Know**: Which manager has the package
2. **Suboptimal Choices**: Default might not be the best source
3. **Verbose Flags**: `--homebrew` is long to type
4. **No Discovery**: No easy way to see which managers offer a package

## Potential Solutions

### Option A: Smart Auto-Detection
- Query all managers for the package
- Use priority order: prefer binary installers (brew/cargo) over language-specific (npm/pip)
- Show what will be installed before proceeding

```bash
$ plonk install ripgrep
Found ripgrep in:
  → homebrew (recommended - pre-built binary)
  - cargo (source build)
Installing from homebrew...
```

**Pros**: Smart defaults, educational
**Cons**: Slower (multiple queries), might surprise users

### Option B: Manager Prefixes
- Use shorthand prefixes: `brew:ripgrep`, `npm:typescript`
- No prefix = use default or auto-detect

```bash
plonk install brew:ripgrep npm:typescript pip:black
```

**Pros**: Explicit, compact, allows mixed installs
**Cons**: New syntax to learn

### Option C: Interactive Selection
- When ambiguous, prompt user:

```bash
$ plonk install ripgrep
Multiple sources available:
  1) homebrew (pre-built binary)
  2) cargo (build from source)
Select [1-2, or 'a' for all]:
```

**Pros**: User control, educational
**Cons**: Breaks automation, interrupts flow

### Option D: Config-Based Rules
- Add `preferred_managers` config section:

```yaml
preferred_managers:
  ripgrep: homebrew
  typescript: npm
  "*": homebrew  # fallback
```

**Pros**: Predictable, customizable
**Cons**: More config complexity

### Option E: Combined Approach
1. Check `preferred_managers` config first
2. If not specified, auto-detect with smart defaults
3. Allow override with prefix syntax
4. Add `--interactive` flag for prompting

## Questions for Discussion

1. How often do users need to override the default manager?
2. Is the current `--<manager>` syntax too verbose?
3. Should we optimize for explicit control or smart defaults?
4. How important is backwards compatibility here?

## Related Commands to Consider

- `plonk search <package>` - Should show which managers have it
- `plonk info <package>` - Should show all available sources
- `plonk which <package>` - New command to show what would be installed?

## Recommendation Placeholder

_To be filled after discussion_
