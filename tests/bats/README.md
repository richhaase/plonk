# ⚠️ BATS Tests for Plonk - SYSTEM MODIFICATION WARNING ⚠️

## 🚨 THESE TESTS MODIFY YOUR REAL SYSTEM 🚨

### What These Tests Do
- **INSTALL REAL PACKAGES** using your package managers (brew, npm, pip, etc.)
- **CREATE REAL DOTFILES** in your home directory
- **MODIFY PLONK STATE** (using isolated config in temp directory)

### Before Running Tests
1. **BACKUP YOUR PLONK CONFIG**: `cp -r ~/.config/plonk ~/.config/plonk.backup`
2. **REVIEW THE SAFE LISTS**: Check `config/safe-packages.list` and `config/safe-dotfiles.list`
3. **UNDERSTAND THE RISKS**: These tests will install software on your system

### Running Tests Safely

```bash
# Basic run (leaves test packages installed)
bats tests/bats/behavioral/

# Run with automatic cleanup of packages
PLONK_TEST_CLEANUP_PACKAGES=1 bats tests/bats/

# Run with custom safe lists
export PLONK_TEST_SAFE_PACKAGES="brew:jq"
export PLONK_TEST_SAFE_DOTFILES=".plonk-test-rc"
bats tests/bats/

# Skip destructive tests
PLONK_TEST_SKIP_DANGEROUS=1 bats tests/bats/

# Dry run - see what would be tested
PLONK_TEST_DRY_RUN=1 bats tests/bats/
```

### After Running Tests

The tests may leave artifacts on your system:
- Installed packages (jq, tree, etc.)
- Test dotfiles (.plonk-test-rc, etc.)

To restore your intended system state:
```bash
# Option 1: Run cleanup test
bats tests/bats/cleanup/99-cleanup-all.bats

# Option 2: Manually restore your configuration
plonk apply

# Option 3: Uninstall test packages manually
plonk uninstall brew:jq brew:tree npm:is-odd pip:cowsay
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PLONK_TEST_CLEANUP_PACKAGES` | `0` | Set to `1` to auto-remove test packages |
| `PLONK_TEST_CLEANUP_DOTFILES` | `1` | Set to `0` to keep test dotfiles |
| `PLONK_TEST_SKIP_DANGEROUS` | `0` | Set to `1` to skip destructive tests |
| `PLONK_TEST_SAFE_PACKAGES` | See safe-packages.list | Comma-separated list of allowed packages |
| `PLONK_TEST_SAFE_DOTFILES` | See safe-dotfiles.list | Comma-separated list of allowed dotfiles |
| `PLONK_TEST_VERBOSE` | `0` | Set to `1` for verbose output |

### Test Development Guidelines

1. **Only use packages/dotfiles from the safe lists**
2. **Always track created artifacts for cleanup**
3. **Test on a non-critical system first**
4. **Mark destructive tests with `skip_if_dangerous`**
5. **Provide clear test descriptions**

### Directory Structure

```
tests/bats/
├── README.md              # This file
├── config/                # Configuration files
│   ├── safe-packages.list # Allowed test packages
│   └── safe-dotfiles.list # Allowed test dotfiles
├── lib/                   # Test utilities
│   ├── test_helper.bash   # Core test functions
│   ├── assertions.bash    # Custom assertions
│   └── cleanup.bash       # Cleanup utilities
├── behavioral/            # Main test suites
│   ├── 00-smoke.bats     # Basic setup verification
│   ├── 01-basic-commands.bats
│   └── ...
├── cleanup/              # Cleanup tests
│   └── 99-cleanup-all.bats
└── fixtures/             # Test data
    ├── configs/          # Sample config files
    └── dotfiles/         # Sample dotfiles
```

### Troubleshooting

**Tests fail with "command not found"**
- Ensure plonk is in your PATH
- Run `go build` to create the binary

**Tests fail with "package manager not available"**
- Install required package managers (brew, npm, etc.)
- Or skip tests for unavailable managers

**Tests leave artifacts**
- Run the cleanup test: `bats tests/bats/cleanup/99-cleanup-all.bats`
- Set `PLONK_TEST_CLEANUP_PACKAGES=1` for auto-cleanup

**Permission errors**
- Some tests may require sudo (though we try to avoid this)
- Run tests as your normal user, not root

### Contributing

When adding new tests:
1. Use descriptive test names
2. Always use safe packages/dotfiles
3. Track all artifacts for cleanup
4. Test locally before committing
5. Update safe lists if needed (with team review)
