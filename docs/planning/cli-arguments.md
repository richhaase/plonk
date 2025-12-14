# Plonk CLI Arguments Audit

This document provides a comprehensive audit of all command-line arguments and options in the Plonk CLI tool.

**Audit Date:** 2025-12-14 (Updated)
**Source Files:** `internal/commands/*.go`

> **Note:** The CLI was recently simplified by removing redundant flags:
> - `--unmanaged` removed from `status`, `packages`, and `dotfiles` commands
> - `--packages` and `--dotfiles` removed from `status` command (use `plonk packages` and `plonk dotfiles` subcommands instead)

---

## Table of Contents

1. [Global Flags](#global-flags)
2. [Root Command](#root-command)
3. [Package Management Commands](#package-management-commands)
   - [install](#install)
   - [uninstall](#uninstall)
   - [upgrade](#upgrade)
4. [Dotfile Management Commands](#dotfile-management-commands)
   - [add](#add)
   - [rm](#rm)
   - [clone](#clone)
5. [Status & Visibility Commands](#status--visibility-commands)
   - [status](#status)
   - [packages](#packages)
   - [dotfiles](#dotfiles)
   - [diff](#diff)
6. [Configuration Commands](#configuration-commands)
   - [config](#config)
   - [config show](#config-show)
   - [config edit](#config-edit)
7. [System Commands](#system-commands)
   - [apply](#apply)
   - [doctor](#doctor)
8. [Consistency Analysis](#consistency-analysis)
9. [Shell Completion](#shell-completion)

---

## Global Flags

These flags are available on all commands via `PersistentFlags()`:

| Flag       | Short | Type   | Default | Description                       |
| ---------- | ----- | ------ | ------- | --------------------------------- |
| `--output` | `-o`  | string | `table` | Output format (table\|json\|yaml) |

**Source:** `root.go:44`

**Shell Completion:** Registered via `RegisterFlagCompletionFunc` for `--output` flag with values: `table`, `json`, `yaml`

---

## Root Command

**File:** `root.go`

```
plonk [flags]
```

### Description

A developer environment manager that manages development environments by installing packages and managing dotfiles across multiple package managers.

### Local Flags

| Flag        | Short | Type | Default | Description              |
| ----------- | ----- | ---- | ------- | ------------------------ |
| `--version` | `-v`  | bool | `false` | Show version information |

### Behavior

- If `--version` is passed, prints version and exits
- Otherwise, displays help

---

## Package Management Commands

### install

**File:** `install.go`

```
plonk install <packages...>
```

### Description

Install packages on your system and add them to your lock file for management.

### Arguments

| Argument   | Required    | Description                                          |
| ---------- | ----------- | ---------------------------------------------------- |
| `packages` | Yes (min 1) | Package specifications in format `[manager:]package` |

### Local Flags

| Flag        | Short | Type | Default | Description                                         |
| ----------- | ----- | ---- | ------- | --------------------------------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Show what would be installed without making changes |

### Package Specification Format

- `package` - Uses default manager
- `manager:package` - Uses specified manager (e.g., `brew:htop`, `npm:typescript`)

### Shell Completion

- Examples are dynamically generated based on configured package managers

### Inherited Flags

- `--output`, `-o` (global)

---

### uninstall

**File:** `uninstall.go`

```
plonk uninstall <packages...>
```

### Description

Uninstall packages from your system and remove them from your lock file.

### Arguments

| Argument   | Required    | Description                                          |
| ---------- | ----------- | ---------------------------------------------------- |
| `packages` | Yes (min 1) | Package specifications in format `[manager:]package` |

### Local Flags

| Flag        | Short | Type | Default | Description                                       |
| ----------- | ----- | ---- | ------- | ------------------------------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Show what would be removed without making changes |

### Inherited Flags

- `--output`, `-o` (global)

---

### upgrade

**File:** `upgrade.go`

```
plonk upgrade [manager:package|package|manager] ...
```

### Description

Upgrade packages managed by plonk to their latest versions. Only upgrades packages tracked in the lock file.

### Arguments

| Argument  | Required | Description                                             |
| --------- | -------- | ------------------------------------------------------- |
| `targets` | No       | Package/manager specifications (if empty, upgrades all) |

### Target Formats

- (no args) - Upgrade all managed packages
- `package` - Upgrade specific package across all managers
- `manager` - Upgrade all packages for a specific manager
- `manager:package` - Upgrade specific package from specific manager

### Local Flags

| Flag        | Short | Type | Default | Description                                          |
| ----------- | ----- | ---- | ------- | ---------------------------------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Show what would be upgraded without making changes   |

### Inherited Flags

- `--output`, `-o` (global)

---

## Dotfile Management Commands

### add

**File:** `add.go`

```
plonk add [files...]
```

### Description

Add dotfiles to plonk management by copying them to the configuration directory (`$PLONK_DIR`).

### Arguments

| Argument | Required    | Description                                           |
| -------- | ----------- | ----------------------------------------------------- |
| `files`  | Conditional | File paths (required unless `--sync-drifted` is used) |

### Local Flags

| Flag             | Short | Type | Default | Description                                         |
| ---------------- | ----- | ---- | ------- | --------------------------------------------------- |
| `--dry-run`      | `-n`  | bool | `false` | Show what would be added without making changes     |
| `--sync-drifted` | `-y`  | bool | `false` | Sync all drifted files from $HOME back to $PLONKDIR |

### Path Resolution

Plonk accepts paths in multiple formats:

- Absolute paths: `/home/user/.vimrc`
- Tilde paths: `~/.vimrc`
- Relative paths: `.vimrc` (tries current dir, then home)
- Plain names: `vimrc` (tries current dir, then home with dot prefix)

### Shell Completion

- `ValidArgsFunction` provides completion for common dotfile paths

### Inherited Flags

- `--output`, `-o` (global)

---

### rm

**File:** `rm.go`

```
plonk rm <files...>
```

### Description

Remove dotfiles from plonk management by deleting them from the configuration directory. Original files in home directory are NOT affected.

### Arguments

| Argument | Required    | Description                          |
| -------- | ----------- | ------------------------------------ |
| `files`  | Yes (min 1) | File paths to remove from management |

### Local Flags

| Flag        | Short | Type | Default | Description                                       |
| ----------- | ----- | ---- | ------- | ------------------------------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Show what would be removed without making changes |

### Shell Completion

- `ValidArgsFunction` provides completion for common dotfile paths

### Inherited Flags

- `--output`, `-o` (global)

---

### clone

**File:** `clone.go`

```
plonk clone <git-repo>
```

### Description

Clone an existing dotfiles repository and intelligently set up plonk.

### Arguments

| Argument   | Required        | Description                            |
| ---------- | --------------- | -------------------------------------- |
| `git-repo` | Yes (exactly 1) | Git repository URL or GitHub shorthand |

### Local Flags

| Flag         | Short | Type | Default | Description                                        |
| ------------ | ----- | ---- | ------- | -------------------------------------------------- |
| `--yes`      |       | bool | `false` | Non-interactive mode - answer yes to all prompts   |
| `--no-apply` |       | bool | `false` | Skip running 'plonk apply' after setup             |
| `--dry-run`  | `-n`  | bool | `false` | Show what would be cloned without making changes   |

### Git Repository Formats

- GitHub shorthand: `user/repo`
- HTTPS URL: `https://github.com/user/repo.git`
- SSH URL: `git@github.com:user/repo.git`
- Git protocol: `git://github.com/user/repo.git`

### Note

Uses package-level variables `cloneYes` and `cloneNoApply` instead of flag parsing in RunE. These flags do not use short forms.

### Inherited Flags

- `--output`, `-o` (global) - but not used in output

---

## Status & Visibility Commands

### status

**File:** `status.go`

```
plonk status [flags]
```

### Description

Display a detailed list of all plonk-managed items and their status.

### Aliases

- `st`

### Local Flags

| Flag        | Short | Type | Default | Description                 |
| ----------- | ----- | ---- | ------- | --------------------------- |
| `--missing` |       | bool | `false` | Show only missing resources |

### Inherited Flags

- `--output`, `-o` (global)

---

### packages

**File:** `packages.go`

```
plonk packages [flags]
```

### Description

Display the status of all plonk-managed packages.

### Aliases

- `p`

### Local Flags

| Flag        | Short | Type | Default | Description                |
| ----------- | ----- | ---- | ------- | -------------------------- |
| `--missing` |       | bool | `false` | Show only missing packages |

### Inherited Flags

- `--output`, `-o` (global)

---

### dotfiles

**File:** `dotfiles.go`

```
plonk dotfiles [flags]
```

### Description

Display the status of all plonk-managed dotfiles.

### Aliases

- `d`

### Local Flags

| Flag        | Short | Type | Default | Description                |
| ----------- | ----- | ---- | ------- | -------------------------- |
| `--missing` |       | bool | `false` | Show only missing dotfiles |

### Inherited Flags

- `--output`, `-o` (global)

---

### diff

**File:** `diff.go`

```
plonk diff [file]
```

### Description

Show differences between source and deployed dotfiles that have drifted.

### Arguments

| Argument | Required   | Description                    |
| -------- | ---------- | ------------------------------ |
| `file`   | No (max 1) | Specific file to show diff for |

### Local Flags

None

### Behavior

- With no arguments: shows diffs for all drifted dotfiles
- With file argument: shows diff for that specific file only
- Uses `cfg.DiffTool` or defaults to `git diff --no-index`

### Inherited Flags

- `--output`, `-o` (global) - but not used (diff outputs directly)

---

## Configuration Commands

### config

**File:** `config.go`

```
plonk config <subcommand>
```

### Description

Parent command for configuration management. Has no direct functionality.

### Subcommands

- `show` - Display current configuration
- `edit` - Edit configuration file

---

### config show

**File:** `config_show.go`

```
plonk config show
```

### Description

Display the effective plonk configuration (defaults merged with user settings).

### Arguments

- None allowed (`cobra.NoArgs`)

### Local Flags

None

### Inherited Flags

- `--output`, `-o` (global)

---

### config edit

**File:** `config_edit.go`

```
plonk config edit
```

### Description

Edit the plonk configuration file using your preferred editor (`$VISUAL`, `$EDITOR`, or `vim`). Works like visudo with validation.

### Arguments

- None allowed (`cobra.NoArgs`)

### Local Flags

None

### Behavior

- Shows full runtime configuration (defaults + user overrides)
- Opens in preferred editor
- Validates configuration after editing
- Saves only non-default values to plonk.yaml
- Supports edit/revert/quit on validation errors

### Inherited Flags

- `--output`, `-o` (global) - but not used

---

## System Commands

### apply

**File:** `apply.go`

```
plonk apply [files...]
```

### Description

Apply configuration to reconcile system state - installs missing packages and manages dotfiles.

### Arguments

| Argument | Required | Description                                                                 |
| -------- | -------- | --------------------------------------------------------------------------- |
| `files`  | No       | Specific dotfiles to apply (if specified, only those dotfiles are deployed) |

### Local Flags

| Flag         | Short | Type | Default | Description                                       |
| ------------ | ----- | ---- | ------- | ------------------------------------------------- |
| `--dry-run`  | `-n`  | bool | `false` | Show what would be applied without making changes |
| `--packages` |       | bool | `false` | Apply packages only                               |
| `--dotfiles` |       | bool | `false` | Apply dotfiles only                               |

### Flag Behavior

- `--packages` and `--dotfiles`: **Mutually exclusive** (enforced via `MarkFlagsMutuallyExclusive`)
- Cannot specify files with `--packages` or `--dotfiles` flags

### Inherited Flags

- `--output`, `-o` (global)

---

### doctor

**File:** `doctor.go`

```
plonk doctor
```

### Description

Perform health checks to ensure your system is properly configured for plonk.

### Arguments

- None

### Local Flags

None

### Shows

- System information (OS, arch, etc.)
- Package manager availability
- Configuration file status and location
- Environment variables (PLONK_DIR, etc.)
- Any issues that would prevent plonk from working

### Inherited Flags

- `--output`, `-o` (global)

---

## Consistency Analysis

### Flag Naming Conventions

| Pattern            | Commands Using                              | Notes      |
| ------------------ | ------------------------------------------- | ---------- |
| `--dry-run` / `-n` | install, uninstall, upgrade, add, rm, clone, apply | Consistent |
| `--packages`       | apply                              | Consistent |
| `--dotfiles`       | apply                              | Consistent |
| `--missing`        | status, packages, dotfiles         | Consistent |

### Missing Flags Analysis

| Command | Missing Flag | Recommendation            |
| ------- | ------------ | ------------------------- |
| `diff`  | `--dry-run`  | N/A (read-only operation) |

### Short Flag Usage

| Short | Long Form        | Command(s)                                     |
| ----- | ---------------- | ---------------------------------------------- |
| `-n`  | `--dry-run`      | install, uninstall, upgrade, add, rm, clone, apply |
| `-o`  | `--output`       | global (all commands)                          |
| `-v`  | `--version`      | root                                           |
| `-y`  | `--sync-drifted` | add                                            |

### Mutual Exclusivity

| Command | Mutually Exclusive Flags   | Enforcement                          |
| ------- | -------------------------- | ------------------------------------ |
| `apply` | `--packages`, `--dotfiles` | Cobra's `MarkFlagsMutuallyExclusive` |

### Aliases

| Command    | Aliases |
| ---------- | ------- |
| `status`   | `st`    |
| `packages` | `p`     |
| `dotfiles` | `d`     |

No aliases for: install, uninstall, upgrade, add, rm, clone, diff, doctor, apply, config

### Argument Validation

| Command       | Args Validator                |
| ------------- | ----------------------------- |
| `install`     | `MinimumNArgs(1)`             |
| `uninstall`   | `MinimumNArgs(1)`             |
| `upgrade`     | None (accepts any)            |
| `add`         | None (conditional)            |
| `rm`          | `MinimumNArgs(1)`             |
| `clone`       | `ExactArgs(1)`                |
| `diff`        | `MaximumNArgs(1)`             |
| `config show` | `NoArgs`                      |
| `config edit` | `NoArgs`                      |
| `doctor`      | None                          |
| `status`      | None                          |
| `packages`    | None                          |
| `dotfiles`    | None                          |
| `apply`       | None (accepts optional files) |

---

## Shell Completion

### Registered Completions

| Command/Flag         | Completion Type | Values                       |
| -------------------- | --------------- | ---------------------------- |
| `--output` (global)  | Static          | `table`, `json`, `yaml`      |
| `add` args           | Dynamic         | Common dotfile paths         |
| `rm` args            | Dynamic         | Common dotfile paths         |
| `install` examples   | Dynamic         | Based on configured managers |
| `uninstall` examples | Dynamic         | Based on configured managers |
| `upgrade` examples   | Dynamic         | Based on configured managers |

### Common Dotfile Suggestions (add/rm)

```
~/.zshrc, ~/.bashrc, ~/.bash_profile, ~/.profile,
~/.vimrc, ~/.vim/, ~/.nvim/,
~/.gitconfig, ~/.gitignore_global,
~/.tmux.conf, ~/.tmux/,
~/.ssh/config, ~/.ssh/,
~/.aws/config, ~/.aws/credentials,
~/.config/, ~/.config/nvim/, ~/.config/fish/, ~/.config/alacritty/,
~/.docker/config.json,
~/.zprofile, ~/.zshenv,
~/.inputrc, ~/.editorconfig
```

---

## Command Summary Table

| Command       | Args            | Flags                                                  | Aliases | Has dry-run |
| ------------- | --------------- | ------------------------------------------------------ | ------- | ----------- |
| `plonk`       | -               | `--version`, `--output`                                | -       | -           |
| `install`     | `<packages...>` | `--dry-run`                                            | -       | Yes         |
| `uninstall`   | `<packages...>` | `--dry-run`                                            | -       | Yes         |
| `upgrade`     | `[targets...]`  | `--dry-run`                                            | -       | Yes         |
| `add`         | `[files...]`    | `--dry-run`, `--sync-drifted`                          | -       | Yes         |
| `rm`          | `<files...>`    | `--dry-run`                                            | -       | Yes         |
| `clone`       | `<git-repo>`    | `--yes`, `--no-apply`, `--dry-run`                     | -       | Yes         |
| `status`      | -               | `--missing`                                            | `st`    | -           |
| `packages`    | -               | `--missing`                                            | `p`     | -           |
| `dotfiles`    | -               | `--missing`                                            | `d`     | -           |
| `diff`        | `[file]`        | -                                                      | -       | -           |
| `apply`       | `[files...]`    | `--dry-run`, `--packages`, `--dotfiles`                | -       | Yes         |
| `doctor`      | -               | -                                                      | -       | -           |
| `config`      | -               | -                                                      | -       | -           |
| `config show` | -               | -                                                      | -       | -           |
| `config edit` | -               | -                                                      | -       | -           |

---

## Environment Variables

While not CLI arguments, these environment variables affect behavior:

| Variable    | Purpose                           | Used By             |
| ----------- | --------------------------------- | ------------------- |
| `PLONK_DIR` | Override default config directory | config package      |
| `NO_COLOR`  | Disable colored output            | output.InitColors() |
| `VISUAL`    | Preferred editor (first choice)   | config edit         |
| `EDITOR`    | Preferred editor (second choice)  | config edit         |

---

## Recommendations

1. **Consider adding aliases** to commonly used commands (e.g., `i` for install, `u` for upgrade)

## Design Notes

### CLI Simplification Philosophy

The CLI follows a principle of using **dedicated subcommands** rather than filter flags where appropriate:

- Use `plonk packages` instead of `plonk status --packages`
- Use `plonk dotfiles` instead of `plonk status --dotfiles`
- Use `plonk packages --missing` instead of `plonk status --packages --missing`

This approach:
- Reduces flag proliferation and cognitive load
- Makes commands more discoverable via `plonk --help`
- Keeps each command focused on a single responsibility
- Simplifies implementation and testing
