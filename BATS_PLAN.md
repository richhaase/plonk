# BATS Testing Plan for Plonk

## Overview

This document outlines the strategy for implementing BATS (Bash Automated Testing System) tests for plonk. The focus is on **behavioral testing** of the CLI interface - verifying what users see and experience when using plonk commands.

## Testing Philosophy

### What BATS Will Test
- ✅ Command parsing and argument handling
- ✅ Output formatting (table/json/yaml)
- ✅ Error messages and help text
- ✅ Command workflows and sequences
- ✅ Exit codes and failure handling
- ✅ Interactive prompts and user feedback
- ✅ Real package installation/uninstallation
- ✅ Real dotfile management

### What BATS Won't Test
- ❌ Internal logic (covered by Go unit tests)
- ❌ Package manager internals
- ❌ State management implementation details
- ❌ Mocked operations (we test real behavior)

## Core Principles

1. **No Mocking** - Test against real plonk behavior with actual package managers
2. **UI/UX Focus** - Verify command outputs, error messages, and user workflows
3. **Safe Isolation** - Use isolated plonk config while keeping system intact
4. **Real Operations** - Actually install/uninstall packages and manage dotfiles
5. **Developer Safety** - Use "safe" test packages and dotfiles with clear warnings

## Test Environment Strategy

### Environment Setup
```bash
# Isolate plonk's configuration only
export PLONK_DIR="$BATS_TEST_TMPDIR/plonk-config"

# Keep HOME intact - package managers need it
# Real home, real package managers, isolated plonk state
```

### Safe Test Resources

#### Safe Packages
Small, harmless packages that are unlikely to interfere with development:
```bash
SAFE_PACKAGES=(
  "brew:jq"          # JSON processor, ~1MB
  "brew:tree"        # Directory tree viewer, ~50KB
  "npm:is-odd"       # Tiny npm test package, ~4KB
  "npm:left-pad"     # Famous tiny package, ~4KB
  "pip:cowsay"       # ASCII cow generator, harmless
  "gem:lolcat"       # Rainbow text, fun and safe
)
```

#### Safe Dotfiles
Test-specific dotfiles unlikely to conflict with real configs:
```bash
SAFE_DOTFILES=(
  ".plonk-test-rc"
  ".plonk-test-profile"
  ".plonk-test-gitconfig"
  ".config/plonk-test/config.yaml"
  ".config/plonk-test/settings.json"
)
```

## Test Organization

```
tests/bats/
├── README.md                    # ⚠️ WARNINGS and usage instructions
├── config/
│   ├── safe-packages.list       # Default safe package list
│   ├── safe-dotfiles.list       # Default safe dotfile list
│   └── fixtures/                # Test configs and dotfiles
│       ├── basic-config.yaml
│       ├── multi-package.yaml
│       └── dotfiles/
├── lib/
│   ├── test_helper.bash         # Core test utilities
│   ├── assertions.bash          # Custom assertions
│   └── cleanup.bash             # Cleanup utilities
├── behavioral/
│   ├── 01-basic-commands.bats   # help, version, status
│   ├── 02-config-flow.bats      # config show/edit workflows
│   ├── 03-package-install.bats  # install/uninstall behavior
│   ├── 04-dotfile-flow.bats     # add/rm dotfile behavior
│   ├── 05-apply-behavior.bats   # apply command with failures
│   ├── 06-search-info.bats      # search and info output
│   ├── 07-output-formats.bats   # json/yaml/table formatting
│   └── 08-error-handling.bats   # error messages and recovery
└── cleanup/
    └── 99-cleanup-all.bats      # Final cleanup test

```

## Cleanup Strategy

### Automatic Tracking
```bash
# Track all created artifacts
setup_test_env() {
  export PLONK_DIR="$BATS_TEST_TMPDIR/plonk-config"
  export CLEANUP_FILE="$BATS_TEST_TMPDIR/cleanup.list"
  mkdir -p "$PLONK_DIR"
  : > "$CLEANUP_FILE"
}

# Record artifacts for cleanup
track_artifact() {
  echo "$1" >> "$CLEANUP_FILE"
}
```

### Cleanup Levels
1. **Test-level**: Clean up dotfiles after each test
2. **File-level**: Clean up all artifacts after test file
3. **Suite-level**: Optional full cleanup including packages

### User Control
```bash
# Environment variables for cleanup control
PLONK_TEST_CLEANUP_PACKAGES=1    # Remove test packages after suite
PLONK_TEST_CLEANUP_DOTFILES=1    # Remove test dotfiles (default: yes)
PLONK_TEST_SKIP_DANGEROUS=1      # Skip potentially destructive tests
```

## Example Test Patterns

### Basic Behavioral Test
```bash
@test "status shows correct counts for managed items" {
  setup_test_env

  # Setup: Install a package
  run plonk install brew:jq
  assert_success
  track_artifact "package:brew:jq"

  # Test: Verify status output
  run plonk status
  assert_success
  assert_output --partial "1 managed"
  assert_output --partial "jq"
}
```

### Error Behavior Test
```bash
@test "install shows helpful error for non-existent package" {
  setup_test_env

  run plonk install brew:definitely-not-real-xyz123
  assert_failure

  # Verify error message quality
  assert_output --partial "not found"
  refute_output --partial "panic"
  refute_output --partial "stack trace"
}
```

### Workflow Test
```bash
@test "apply continues after partial failure" {
  setup_test_env

  # Create config with valid and invalid packages
  cat > "$PLONK_DIR/plonk.yaml" <<EOF
packages:
  - jq                    # Valid
  - fake-package-xyz      # Invalid
  - tree                  # Valid
EOF

  run plonk apply
  assert_failure  # Non-zero exit due to partial failure

  # Verify behavioral output
  assert_output --partial "✓ jq"
  assert_output --partial "✗ fake-package-xyz"
  assert_output --partial "✓ tree"
  assert_output --partial "2 succeeded, 1 failed"

  track_artifact "package:brew:jq"
  track_artifact "package:brew:tree"
}
```

## Safety Measures

### Developer Warnings
- Clear README with warnings about system modification
- Safe defaults that minimize conflict risk
- Cleanup instructions and automation
- Environment variables for safety controls

### CI/CD Considerations
```yaml
# GitHub Actions example
- name: Run BATS tests
  env:
    PLONK_TEST_CLEANUP_PACKAGES: "1"
    PLONK_TEST_SAFE_PACKAGES: "brew:jq"  # Minimal set for CI
  run: |
    bats tests/bats/behavioral/
```

### Progressive Testing
1. Start with read-only tests (status, info, help)
2. Add state-changing tests (install, add)
3. Add cleanup/removal tests (uninstall, rm)
4. Full workflow tests (apply with failures)

## Implementation Status

### Phase 1: Foundation ✅
- Test helper infrastructure
- Basic command tests (help, version, status)
- Cleanup mechanisms (default enabled)

### Phase 2: Core Behaviors ✅
- Package install/uninstall for all managers
- Output format testing (table format)
- Single binary build per test suite using setup_suite.bash

### Phase 3: Complex Workflows (Next)
- Dotfile add/rm operations
- Apply with partial failures
- Search across managers
- Error recovery scenarios

### Phase 4: Edge Cases
- Network failures
- Permission issues
- Conflicting operations

## Success Criteria

1. **Coverage**: All user-facing commands have behavioral tests
2. **Safety**: No test damages a developer's system unintentionally
3. **Clarity**: Test output clearly shows what succeeded/failed
4. **Speed**: Full suite runs in under 2 minutes
5. **Reliability**: Tests are deterministic and repeatable

## Maintenance Guidelines

1. **New Features**: Add behavioral tests with the feature
2. **Bug Fixes**: Add regression test showing the fix
3. **Safe Lists**: Review and update quarterly
4. **CI Integration**: Run subset of safe tests on every PR
5. **Documentation**: Keep README warnings up to date

## Notes

- BATS version 1.8.0+ recommended for test filtering features
- Consider parallel test execution for speed (with careful isolation)
- Regular "test health" reviews to remove flaky tests
- Community input on safe package/dotfile selections welcome
