# Pre-commit Framework Migration Guide

## 🎯 Overview

Plonk now supports the industry-standard **pre-commit framework** alongside our existing custom git hooks. This provides better developer experience, faster execution, and more comprehensive checks.

## 🚀 Quick Start (Recommended)

### For New Developers

```bash
# Install pre-commit (if not already installed)
brew install pre-commit  # macOS
# or
pip install pre-commit   # Python

# Install hooks (one-time setup)
pre-commit install

# That's it! Hooks will run automatically on commit
```

### For Existing Developers

You can continue using your existing setup, or migrate to pre-commit:

```bash
# Option A: Keep current setup (no change needed)
# Your existing hooks continue to work

# Option B: Migrate to pre-commit framework
pre-commit install
# Your old hooks are preserved as .git/hooks/pre-commit.legacy
```

## 🔄 Migration Options

### Option 1: Parallel Usage (Recommended)
- Keep both systems running
- Gradually switch to pre-commit as you get comfortable
- Old hooks preserved as backup

### Option 2: Full Migration
- Uninstall old hooks: `scripts/uninstall-hooks.sh` (when created)
- Install pre-commit: `pre-commit install`

## ⚡ Performance Benefits

| Scenario | Custom Hooks | Pre-commit Framework |
|----------|-------------|---------------------|
| **Go files only** | ~30s (all checks) | ~15s (Go checks only) |
| **Docs only** | ~30s (all checks) | ~3s (file checks only) |
| **Mixed changes** | ~30s (all checks) | ~20s (relevant checks) |
| **No changes** | ~30s (all checks) | ~1s (no execution) |

## 🛠 Available Hooks

### Local Hooks (Using Justfile)
- **go-fmt-import**: Format Go code with goimports
- **go-lint**: Run golangci-lint
- **go-test**: Run Go tests

### Community Hooks
- **check-yaml**: Validate YAML syntax
- **check-toml**: Validate TOML syntax
- **end-of-file-fixer**: Ensure files end with newline
- **trailing-whitespace**: Remove trailing whitespace
- **check-merge-conflict**: Detect merge conflict markers
- **check-added-large-files**: Prevent large files (>1MB)
- **go-mod-tidy**: Tidy Go modules
- **go-vet**: Run go vet
- **go-fmt**: Run go fmt
- **go-imports**: Run goimports

## 🎛 Configuration

The pre-commit configuration is in `.pre-commit-config.yaml`. Key features:

```yaml
# Only run hooks on relevant files
files: \.go$

# Skip hooks on certain files
exclude: ^(vendor/|.*_test\.go)

# Custom arguments
args: [-local, plonk]

# Control execution
fail_fast: false  # Run all hooks even if some fail
```

## 🔧 Common Commands

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run go-lint

# Run only on staged files (normal operation)
pre-commit run

# Update hook versions
pre-commit autoupdate

# Install hooks (after cloning repo)
pre-commit install

# Uninstall hooks
pre-commit uninstall
```

## 🆚 Comparison

| Feature | Custom Hooks | Pre-commit Framework |
|---------|-------------|---------------------|
| **Setup** | `scripts/install-hooks.sh` | `pre-commit install` |
| **Speed** | Always slow | Intelligent & fast |
| **File filtering** | None | Automatic |
| **Error messages** | Basic | Rich & contextual |
| **Updates** | Manual | `pre-commit autoupdate` |
| **Ecosystem** | Limited | 1000+ hooks available |
| **Team consistency** | Variable | Guaranteed identical |

## 🔄 Backwards Compatibility

- ✅ Existing hooks continue to work
- ✅ `just precommit` still available
- ✅ CI/CD unchanged
- ✅ No breaking changes

## 🆘 Troubleshooting

### Pre-commit not found
```bash
# Install pre-commit
brew install pre-commit
# or
pip install pre-commit
```

### Hooks not running
```bash
# Reinstall hooks
pre-commit uninstall
pre-commit install
```

### Hook failing
```bash
# Run specific hook with verbose output
pre-commit run --verbose go-lint

# Skip problematic hooks temporarily
SKIP=go-lint git commit -m "message"
```

### Reset to clean state
```bash
# Remove pre-commit, keep old hooks
pre-commit uninstall
# Old hooks are automatically restored
```

## 📚 Further Reading

- [Pre-commit Framework Documentation](https://pre-commit.com/)
- [Available Hooks](https://pre-commit.com/hooks.html)
- [Go-specific Hooks](https://github.com/dnephin/pre-commit-golang)
