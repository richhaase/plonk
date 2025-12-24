# BATS Tests for Plonk

## Running Tests (Docker - Recommended)

**Always use Docker** to run BATS tests. This isolates test effects from your local system:

```bash
# Build Docker image and run all tests
just docker-test-all

# Run all tests (if image already built)
just docker-test

# Run smoke tests only (fast verification)
just docker-test-smoke

# Run specific test file
just docker-test-file tests/bats/behavioral/02-package-install.bats

# Interactive shell for debugging
just docker-shell
```

---

## ⚠️ Local Execution Warning ⚠️

> **Only run tests locally if you:**
> - Are certain you understand the risks
> - Accept that tests WILL modify your system
> - Have backed up your plonk configuration

### What Local Tests Do
- **INSTALL REAL PACKAGES** via brew, npm, cargo, uv, etc.
- **CREATE REAL DOTFILES** in your home directory
- **MODIFY SYSTEM STATE** that persists after tests complete

### Local Execution (Not Recommended)

If you must run locally:

```bash
# 1. BACKUP FIRST
cp -r ~/.config/plonk ~/.config/plonk.backup

# 2. Review what will be installed
cat tests/bats/config/safe-packages.list
cat tests/bats/config/safe-dotfiles.list

# 3. Run tests
bats tests/bats/behavioral/

# 4. Cleanup if needed
bats tests/bats/cleanup/99-cleanup-all.bats
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PLONK_TEST_CLEANUP_PACKAGES` | `1` | Set to `0` to keep test packages |
| `PLONK_TEST_CLEANUP_DOTFILES` | `1` | Set to `0` to keep test dotfiles |
| `PLONK_TEST_SAFE_PACKAGES` | See safe-packages.list | Comma-separated list of allowed packages |
| `PLONK_TEST_SAFE_DOTFILES` | See safe-dotfiles.list | Comma-separated list of allowed dotfiles |

### Test Development Guidelines

1. **Only use packages/dotfiles from the safe lists**
2. **Always track created artifacts for cleanup**
3. **Test on a non-critical system first**
4. **Provide clear test descriptions**

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
└── cleanup/              # Cleanup tests
    └── 99-cleanup-all.bats
```

### Troubleshooting

**Tests fail with "command not found"**
- Ensure plonk is in your PATH
- Run `go build` to create the binary

**Tests fail with "package manager not available"**
- Install required package managers (brew, npm, etc.)
- Or skip tests for unavailable managers

**Tests leave artifacts**
- Tests should cleanup automatically
- If not, run: `bats tests/bats/cleanup/99-cleanup-all.bats`
- Or set `PLONK_TEST_CLEANUP_PACKAGES=0` to intentionally keep packages

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
