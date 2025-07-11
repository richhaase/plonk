# Configuration Reference

**🚀 Configuration is completely optional!** Plonk works with sensible defaults out of the box.

This reference shows available configuration options when you want to customize behavior.

## File Overview

Plonk uses up to two files to manage your environment:

- **`plonk.yaml`**: Optional configuration file (create with `plonk init`)
- **`plonk.lock`**: Lock file (automatically managed package state)

## Configuration File (`plonk.yaml`) - Optional

**Default location:** `~/.config/plonk/plonk.yaml`

**Environment override:** `$PLONK_DIR/plonk.yaml`

**Create with:** `plonk init` (generates helpful template with comments)

### Zero-Config Defaults

When no configuration file exists, these defaults are used:

```yaml
# These are the built-in defaults - no config file needed!
default_manager: homebrew      # Primary package manager
operation_timeout: 300         # 5 minutes
package_timeout: 180           # 3 minutes
dotfile_timeout: 60            # 1 minute
expand_directories:
    - .config
    - .ssh
    - .aws
    - .kube
    - .docker
    - .gnupg
    - .local

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"
```

### Complete Example

Only specify settings you want to override from the defaults above:

```yaml
default_manager: homebrew
operation_timeout: 300
package_timeout: 180
dotfile_timeout: 60
expand_directories:
    - .config
    - .ssh
    - .aws
    - .kube
    - .docker
    - .gnupg
    - .local

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"
```

### Minimal Examples

**Example 1: Just change package manager**
```yaml
default_manager: npm
# Everything else uses defaults
```

**Example 2: Custom ignore patterns**
```yaml
ignore_patterns:
  - .DS_Store
  - "*.log"
  - node_modules/
# All settings use defaults
```

**Example 3: Custom timeouts only**
```yaml
operation_timeout: 600    # 10 minutes instead of 5
package_timeout: 300      # 5 minutes instead of 3
# Other settings and ignore patterns use defaults
```

## Lock File (`plonk.lock`)

**Location:** `~/.config/plonk/plonk.lock` (or `$PLONK_DIR/plonk.lock`)

**Note:** This file is automatically managed by plonk. Do not edit manually.

### Example Lock File

```yaml
version: 1
packages:
  homebrew:
    - name: git
      installed_at: "2024-01-15T10:30:00Z"
      version: "2.43.0"
    - name: neovim
      installed_at: "2024-01-15T10:31:00Z"
      version: "0.9.5"
  npm:
    - name: typescript
      installed_at: "2024-01-15T10:32:00Z"
      version: "5.3.3"
  cargo:
    - name: ripgrep
      installed_at: "2024-01-15T10:35:00Z"
      version: "14.1.0"
```

### Minimal Example

```yaml
default_manager: homebrew
```

## Settings Section

### Required Settings

#### `default_manager`
**Type:** string
**Required:** Yes
**Values:** `homebrew`, `npm`, `cargo`
**Purpose:** Default package manager for new packages

```yaml
default_manager: homebrew
```

### Optional Settings

#### `operation_timeout`
**Type:** integer
**Default:** 300
**Units:** seconds
**Purpose:** Overall operation timeout

```yaml
operation_timeout: 600  # 10 minutes
```

#### `package_timeout`
**Type:** integer
**Default:** 180
**Units:** seconds
**Purpose:** Individual package operation timeout

```yaml
package_timeout: 300  # 5 minutes
```

#### `dotfile_timeout`
**Type:** integer
**Default:** 60
**Units:** seconds
**Purpose:** Dotfile operation timeout

```yaml
dotfile_timeout: 120  # 2 minutes
```

#### `expand_directories`
**Type:** array of strings
**Default:** `[".config", ".ssh", ".aws", ".kube", ".docker", ".gnupg", ".local"]`
**Purpose:** Directories to expand in `plonk dot list` output

```yaml
expand_directories:
    - .config
    - .ssh
    - .aws
    - .kube
    - .docker
    - .gnupg
    - .local
```

**Behavior:**
- Directories in this list show individual files and subdirectories
- Other directories appear as single entries
- Helps users see detailed contents of important dotfile directories
- Expansion is limited to 2 levels deep for performance

## Ignore Patterns

**Type:** array of strings
**Format:** Gitignore-style patterns
**Purpose:** Files/directories to ignore during dotfile auto-discovery

### Pattern Rules

- `*` - Matches any sequence of characters
- `?` - Matches any single character
- `[abc]` - Matches any character in brackets
- `**` - Matches directories recursively
- `!pattern` - Negates pattern (include despite other ignores)

### Common Patterns

```yaml
ignore_patterns:
  - .DS_Store           # macOS system files
  - .git                # Git directories
  - "*.backup"          # Backup files
  - "*.tmp"             # Temporary files
  - "*.swp"             # Vim swap files
  - "*.orig"            # Merge conflict files
  - ".#*"               # Emacs lock files
  - "*~"                # Backup files
  - "*.log"             # Log files
```

### Advanced Pattern Examples

```yaml
ignore_patterns:
  - "bin/"              # Binary output directories
  - "node_modules/"     # Node.js modules
  - "*.cache"           # Cache files
  - "secrets.*"         # Secret files
  - "!important.backup" # Exception - include this backup file
```

## Package Management

Packages are automatically tracked in the lock file (`plonk.lock`) when you use package commands:

```bash
# Add packages (updates lock file automatically)
plonk pkg add git
plonk pkg add typescript --manager npm
plonk pkg add ripgrep --manager cargo

# Remove packages
plonk pkg remove git
plonk pkg remove typescript --uninstall

# List packages
plonk pkg list
```

**Supported managers:**
- **Homebrew**: Formulae and casks
- **NPM**: Global packages
- **Cargo**: Rust packages

## Dotfile Management

### Auto-Discovery

Dotfiles are automatically discovered from the config directory:

```
~/.config/plonk/
├── zshrc              → ~/.zshrc
├── vimrc              → ~/.vimrc
├── gitconfig          → ~/.gitconfig
├── config/
│   └── nvim/          → ~/.config/nvim/
│       └── init.lua
└── ssh/
    └── config         → ~/.ssh/config
```

### Path Mapping Rules

1. **Simple files:** `filename` → `~/.filename`
2. **Nested directories:** `path/to/file` → `~/.path/to/file`
3. **Directory preservation:** Directory structure is maintained

### Ignored Files

Files matching `ignore_patterns` are automatically excluded:

```yaml
ignore_patterns:
  - .DS_Store           # System files ignored
  - "*.backup"          # Backup files ignored
```

## Environment Variables

### `PLONK_DIR`

**Purpose:** Override default configuration directory
**Default:** `~/.config/plonk`
**Examples:**

```bash
# Custom location
export PLONK_DIR=~/dotfiles/plonk

# Relative to home (expands ~)
export PLONK_DIR=~/dev/plonk-config

# Absolute path
export PLONK_DIR=/etc/plonk
```

### `EDITOR`

**Purpose:** Editor for `plonk config edit`
**Default:** System default
**Examples:**

```bash
export EDITOR=vim
export EDITOR=code
export EDITOR="emacs -nw"
```

## Validation Rules

### Configuration Validation

```bash
plonk config validate
```

**Checks:**
- YAML syntax validity
- Required fields present
- Valid manager names
- Timeout value ranges
- Pattern syntax validity

### Common Validation Errors

#### Missing Default Manager
```yaml
# ERROR: default_manager can be nil - uses default value
```

#### Invalid Manager Name
```yaml
# ERROR: invalid manager
default_manager: invalid_manager
```

#### Invalid Timeout Values
```yaml
# ERROR: negative timeout
operation_timeout: -1
```

## Configuration Examples

### Developer Workstation

```yaml
default_manager: homebrew
operation_timeout: 600
expand_directories:
    - .config
    - .ssh
    - .aws
    - .kube
    - .docker

ignore_patterns:
  - .DS_Store
  - .git
  - "*.log"
  - "*.cache"
```

### Minimal Setup

```yaml
default_manager: homebrew
```

### NPM-focused Setup

```yaml
default_manager: npm
```

## Migration and Upgrades

### Configuration Migration

Plonk automatically handles configuration migration between versions. No manual migration required.

### Backup Recommendations

```bash
# Backup current config
cp ~/.config/plonk/plonk.yaml ~/.config/plonk/plonk.yaml.backup

# Validate before applying
plonk config validate && plonk apply --dry-run
```

## See Also

- [CLI.md](CLI.md) - Configuration commands
- [ARCHITECTURE.md](ARCHITECTURE.md) - Configuration architecture
- [DEVELOPMENT.md](DEVELOPMENT.md) - Contributing configuration changes
