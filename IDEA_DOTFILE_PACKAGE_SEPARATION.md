# IDEA: Dotfile vs Package Command Separation

## Current State

- `add` and `rm` commands work with dotfiles only
- `install` and `uninstall` commands work with packages only
- Clear separation between the two resource types
- User note: "I'm open to adjusting this, but I think there's value in a flatter command structure if possible"

## Issues to Consider

1. **Mental Model**: Users must remember which command for which resource
2. **Command Count**: More commands in the top-level namespace
3. **Future Resources**: How will this scale with AI Lab resources?
4. **Consistency**: Different verbs for similar actions (add vs install)

## Potential Solutions

### Option A: Keep Current Separation (Status Quo)
```bash
plonk add ~/.zshrc          # dotfiles
plonk install ripgrep       # packages
plonk rm ~/.zshrc           # dotfiles
plonk uninstall ripgrep     # packages
```

**Pros**:
- Clear separation
- Familiar verbs (install for packages)
- No ambiguity

**Cons**:
- Must remember which command for which type
- Doesn't scale well to new resource types

### Option B: Unified Commands with Type Detection
```bash
plonk add ~/.zshrc          # detects file path → dotfile
plonk add ripgrep           # detects no path → package
plonk remove ~/.zshrc       # dotfile
plonk remove ripgrep        # package
```

**Pros**:
- Fewer commands
- Consistent verbs
- Simpler mental model

**Cons**:
- Ambiguity (what about `./ripgrep` file vs `ripgrep` package?)
- Magic behavior might confuse
- Breaking change

### Option C: Unified Commands with Explicit Types
```bash
plonk add dotfile ~/.zshrc
plonk add package ripgrep
plonk remove dotfile ~/.zshrc
plonk remove package ripgrep
```

**Pros**:
- Explicit and clear
- Extensible to new types
- Consistent pattern

**Cons**:
- More verbose
- Breaking change

### Option D: Subcommand Structure
```bash
plonk dotfiles add ~/.zshrc
plonk packages install ripgrep
plonk dotfiles remove ~/.zshrc
plonk packages uninstall ripgrep
```

**Pros**:
- Very clear grouping
- Room for type-specific subcommands
- Follows pattern of `git remote`, `docker container`

**Cons**:
- More verbose
- Deeper command structure

### Option E: Keep Flat but Consistent Verbs
```bash
plonk add-dotfile ~/.zshrc
plonk add-package ripgrep
# or
plonk dadd ~/.zshrc        # dotfile-add
plonk padd ripgrep         # package-add
```

**Pros**:
- Flat structure maintained
- Clear what each does
- Could add aliases for common operations

**Cons**:
- Unusual command naming
- More commands in namespace

## Questions for Discussion

1. How important is keeping a flat command structure?
2. Do users think in terms of "adding" dotfiles and "installing" packages, or is this distinction artificial?
3. With AI Lab resources coming (docker-compose, etc.), should we design for extensibility now?
4. Would command aliases help? (e.g., `install` as alias for `add package`)

## Usage Patterns to Consider

Current typical workflows:
```bash
# Setting up new machine
plonk init
plonk sync

# Adding things incrementally
plonk add ~/.vimrc
plonk install ripgrep

# Cleanup
plonk rm ~/.oldrc
plonk uninstall unused-package
```

How would these look with different approaches?

## Recommendation Placeholder

_To be filled after discussion_
