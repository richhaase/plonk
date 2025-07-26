# IDEA: Search and Info Commands Enhancement

## Current State

- `plonk search <package>` - Search for packages across managers
- `plonk info <package>` - Show package details
- User note: "Search and info are critical, but maybe not well implemented"

## Issues to Consider

1. **Cross-Manager Search**: How to show results from multiple sources?
2. **Performance**: Searching all managers can be slow
3. **Result Relevance**: How to rank/sort results?
4. **Info Completeness**: What information is actually useful?
5. **Package Specification**: How to get info for specific manager's package?

## Current Problems to Address

1. Unclear which manager's version of a package you're searching
2. No way to search within a specific manager only
3. Info might not show all available sources
4. Search results might be overwhelming or poorly formatted

## Potential Solutions

### Option A: Enhanced Search Display
```bash
$ plonk search ripgrep
╭─────────────┬────────────┬─────────┬────────────────────────╮
│ Manager     │ Package    │ Version │ Description            │
├─────────────┼────────────┼─────────┼────────────────────────┤
│ homebrew    │ ripgrep    │ 13.0.0  │ Fast grep replacement  │
│ cargo       │ ripgrep    │ 13.0.0  │ Fast grep replacement  │
│ homebrew    │ ripgrep-all│ 0.9.6   │ ripgrep, but also PDFs │
╰─────────────┴────────────┴─────────┴────────────────────────╯
```

**Pros**: Clear source identification, version info
**Cons**: Could be verbose for common packages

### Option B: Focused Search Options
```bash
plonk search ripgrep              # search all
plonk search --brew ripgrep       # search only homebrew
plonk search --installed ripgrep  # search installed only
plonk search --available ripgrep  # search not installed
```

**Pros**: User control, faster when focused
**Cons**: More flags to remember

### Option C: Interactive Search
```bash
$ plonk search rip
Searching...
? Select a package:
  → ripgrep (homebrew 13.0.0) - Fast grep replacement
    ripgrep (cargo 13.0.0) - Fast grep replacement
    ripgrep-all (homebrew 0.9.6) - ripgrep, but also PDFs
    grip (pip 4.6.1) - GitHub README preview
```

**Pros**: Better UX for discovery, handles typos
**Cons**: Breaks automation, requires terminal interaction

### Option D: Unified Search and Info
```bash
$ plonk find ripgrep    # combines search and info
ripgrep - Fast grep replacement

Available from:
  • homebrew 13.0.0 [installed]
  • cargo 13.0.0

Installation:
  plonk install ripgrep        # uses homebrew (default)
  plonk install --cargo ripgrep # uses cargo

Description:
  ripgrep is a line-oriented search tool that recursively
  searches directories for a regex pattern...
```

**Pros**: One command for discovery, actionable output
**Cons**: Verbose for simple searches

### Option E: Smart Package Resolution
```bash
# Make info work like install - smart detection
$ plonk info ripgrep
Showing info for ripgrep from homebrew (installed)
Use 'plonk info --all ripgrep' to see all sources

# Search shows best match first
$ plonk search grep
Best matches:
  → ripgrep (recommended - fast, modern)

Also available:
  - grep (built-in)
  - ggrep (GNU grep)
  - agrep (approximate grep)
```

## Questions for Discussion

1. What's the primary use case for search? (discovery vs verification)
2. Should search be fuzzy by default?
3. How important is search speed vs completeness?
4. Should info show how to install/uninstall?
5. Do we need package homepage/repo links?

## Enhancement Ideas

1. **Search Aliases**: Common misspellings/alternatives
2. **Categories**: Group results by type (CLI tools, libraries, etc.)
3. **Popularity**: Show download counts or stars?
4. **Local Cache**: Speed up repeated searches
5. **Did You Mean**: Spelling correction

## Related Commands

- `plonk which <package>` - Show what would be installed
- `plonk why <package>` - Show why package is installed
- `plonk alternatives <package>` - Show similar packages

## Recommendation Placeholder

_To be filled after discussion_
