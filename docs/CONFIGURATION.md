# Configuration Reference

Complete configuration file format and environment variable reference for plonk.

## Configuration File Location

**Default location:** `~/.config/plonk/plonk.yaml`

**Environment override:** `$PLONK_DIR/plonk.yaml`

## Configuration Structure

### Complete Example

```yaml
settings:
  default_manager: homebrew
  operation_timeout: 300
  package_timeout: 180
  dotfile_timeout: 60

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"

homebrew:
  - git
  - curl
  - neovim

npm:
  - typescript
  - prettier
```

### Minimal Example

```yaml
settings:
  default_manager: homebrew

homebrew:
  - git
  - curl
```

## Settings Section

### Required Settings

#### `default_manager`
**Type:** string  
**Required:** Yes  
**Values:** `homebrew`, `npm`  
**Purpose:** Default package manager for new packages

```yaml
settings:
  default_manager: homebrew
```

### Optional Settings

#### `operation_timeout`
**Type:** integer  
**Default:** 300  
**Units:** seconds  
**Purpose:** Overall operation timeout

```yaml
settings:
  operation_timeout: 600  # 10 minutes
```

#### `package_timeout`
**Type:** integer  
**Default:** 180  
**Units:** seconds  
**Purpose:** Individual package operation timeout

```yaml
settings:
  package_timeout: 300  # 5 minutes
```

#### `dotfile_timeout`
**Type:** integer  
**Default:** 60  
**Units:** seconds  
**Purpose:** Dotfile operation timeout

```yaml
settings:
  dotfile_timeout: 120  # 2 minutes
```

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
  - "build/"            # Build directories
  - "node_modules/"     # Node.js modules
  - "*.cache"           # Cache files
  - "secrets.*"         # Secret files
  - "!important.backup" # Exception - include this backup file
```

## Package Managers

### Homebrew

**Key:** `homebrew`  
**Type:** array of strings  
**Purpose:** Homebrew packages and casks

```yaml
homebrew:
  - git                 # CLI tools
  - curl
  - neovim
  - docker
  - font-hack-nerd-font # Casks (auto-detected)
  - google-chrome
```

**Package resolution:**
- Checks formulae first
- Falls back to casks if not found
- Case-sensitive matching

### NPM

**Key:** `npm`  
**Type:** array of strings  
**Purpose:** Global NPM packages

```yaml
npm:
  - typescript
  - prettier
  - eslint
  - "@types/node"      # Scoped packages
  - "lodash-cli"       # CLI tools
```

**Package resolution:**
- Global installation (`npm install -g`)
- Supports scoped packages (`@scope/package`)
- Version pinning not supported

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
# ERROR: default_manager required
settings: {}
```

#### Invalid Manager Name
```yaml
# ERROR: invalid manager
settings:
  default_manager: invalid_manager
```

#### Invalid Timeout Values
```yaml
# ERROR: negative timeout
settings:
  operation_timeout: -1
```

## Configuration Examples

### Developer Workstation

```yaml
settings:
  default_manager: homebrew
  operation_timeout: 600

ignore_patterns:
  - .DS_Store
  - .git
  - "*.log"
  - "*.cache"

homebrew:
  - git
  - gh
  - docker
  - kubernetes-cli
  - terraform
  - aws-cli
  - jq
  - curl
  - neovim
  - fzf

npm:
  - typescript
  - prettier
  - eslint
  - "@aws-cdk/cli"
  - nodemon
```

### Minimal Setup

```yaml
settings:
  default_manager: homebrew

homebrew:
  - git
  - curl
```

### NPM-focused Setup

```yaml
settings:
  default_manager: npm

npm:
  - typescript
  - prettier
  - eslint
  - nodemon
  - "@types/node"

homebrew:
  - git
  - node
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