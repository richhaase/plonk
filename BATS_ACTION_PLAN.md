# BATS Implementation Action Plan

## Overview
This document provides a detailed, step-by-step action plan for implementing BATS behavioral testing for plonk. Each phase includes specific tasks, code examples, and validation checkpoints.

## Implementation Status

### Completed Phases
- [x] **Phase 1: Foundation Setup** ✅ (Completed)
  - Created directory structure
  - Created safety documentation
  - Created safe lists
  - Created test helpers with build functionality
  - Created initial smoke test
  - Fixed portability issues (mapfile, assert functions)
  - All smoke tests passing (7/7)

- [x] **Phase 2: Basic Command Tests** ✅ (Completed)
  - Created basic command tests (help, status, aliases)
  - Created output format tests (table format only)
  - All tests passing (7/7)

- [x] **Phase 3: State-Changing Tests** ✅ (Completed)
  - Created package install tests for all managers (brew, npm, pip, gem, go, cargo)
  - Created package uninstall tests for all managers
  - Implemented setup_suite.bash for single binary build per test run
  - Removed unnecessary features (skip_if_dangerous, PLONK_TEST_MODE, etc.)
  - Changed cleanup to default enabled
  - Removed duplicate multiple-package tests (kept only brew as example)
  - **Enhanced with proper behavioral validation:**
    - Replaced system packages (jq) with obscure packages (cowsay, figlet, sl, fortune)
    - Added package-manager-specific verification (brew list, npm list -g, pip show, etc.)
    - Added lock file verification to ensure plonk state is updated
    - Removed use of 'which' command to avoid false positives from other package managers
    - Enhanced dry-run tests to verify no actual changes occur
  - All tests passing with proper behavioral validation

- [x] **Phase 4: Dotfile Tests** ✅ (Completed)
  - Created dotfile add tests (single, directory, nested, conflicts)
  - Created dotfile remove tests
  - Fixed all test assertions to match actual plonk output
  - Documented directory removal bug with failing test
  - 12/13 tests passing (1 failure documents expected behavior)

### In Progress
- [ ] **Phase 5: Apply Command Tests** 🚧 (In Progress)
  - Created initial apply tests
  - Discovered bug: plonk apply shows "All up to date" for system-installed binaries (e.g., jq at /usr/bin/jq)
  - Discovered test isolation issues - need proper cleanup between tests
  - Added cleanup to setup() but still seeing state persistence
  - User suggested using more obscure packages since jq is system-installed on macOS

### Remaining Phases
- [ ] Phase 6: Error Handling Tests
- [ ] Phase 7: Integration Tests
- [ ] Phase 8: CI/CD Integration
- [ ] Phase 9: Documentation Updates

### Known Issues/Bugs Discovered
1. **Directory removal bug** (Phase 4): `plonk rm ~/.config/myapp/` fails with "directory not empty" instead of performing `rm -rf` on the plonk-managed copy
2. **System binary detection bug** (Phase 5): `plonk apply` incorrectly treats system-installed binaries (e.g., /usr/bin/jq) as "managed" and shows "All up to date"
3. **Go package installation bug** (Phase 3): `plonk install go:package` reports success but doesn't actually install the binary. The package appears in plonk's status and lock file, but the binary is not installed to GOPATH/bin/

## Pre-Implementation Checklist

- [x] Install BATS locally: `brew install bats-core`
- [x] Review existing plonk commands to understand current behavior
- [x] Backup your plonk config if you have one: `cp -r ~/.config/plonk ~/.config/plonk.backup`
- [x] Ensure you have test package managers installed (brew, npm, etc.)

## Phase 1: Foundation Setup (Day 1-2) ✅ COMPLETED

### 1.1 Create Directory Structure

```bash
# Create the full directory tree
mkdir -p tests/bats/{config,lib,behavioral,cleanup,fixtures/{configs,dotfiles}}

# Expected structure:
tests/bats/
├── README.md
├── config/
│   ├── safe-packages.list
│   └── safe-dotfiles.list
├── lib/
│   ├── test_helper.bash
│   ├── assertions.bash
│   └── cleanup.bash
├── behavioral/
├── cleanup/
└── fixtures/
    ├── configs/
    └── dotfiles/
```

### 1.2 Create Safety Documentation

**File: `tests/bats/README.md`**
```markdown
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
4. **Provide clear test descriptions**
```

### 1.3 Create Safe Lists

**File: `tests/bats/config/safe-packages.list`**
```bash
# Safe packages for BATS testing
# Format: manager:package
# These packages are small, harmless, and unlikely to conflict with development

# Homebrew packages (macOS/Linux)
brew:cowsay      # ASCII cow, harmless fun tool
brew:figlet      # ASCII art text generator
brew:sl          # Steam locomotive typo corrector
brew:fortune     # Fortune cookie messages

# NPM packages (Node.js)
npm:is-odd       # Checks if number is odd, ~4KB, no deps
npm:left-pad     # Pads strings, ~4KB, famous tiny package

# Python packages
pip:cowsay       # ASCII cow, harmless fun tool
pip:six          # Python 2/3 compatibility, very small

# Ruby gems
gem:colorize     # Terminal color output, small utility

# Go packages
go:github.com/rakyll/hey  # HTTP load generator, small tool
```

**File: `tests/bats/config/safe-dotfiles.list`**
```bash
# Safe dotfiles for BATS testing
# These files are unlikely to conflict with real configurations

.plonk-test-rc
.plonk-test-profile
.plonk-test-gitconfig
.plonk-test-bashrc
.config/plonk-test/config.yaml
.config/plonk-test/settings.json
.config/plonk-test-app/prefs
```

### 1.4 Create Core Test Helper

**File: `tests/bats/lib/test_helper.bash`**
```bash
#!/usr/bin/env bash

# Load BATS helpers
load '/usr/local/lib/bats-support/load'
load '/usr/local/lib/bats-assert/load'

# Global variables
export PLONK_TEST_DIR="$BATS_TEST_DIRNAME/.."
export SAFE_PACKAGES_FILE="$PLONK_TEST_DIR/config/safe-packages.list"
export SAFE_DOTFILES_FILE="$PLONK_TEST_DIR/config/safe-dotfiles.list"

# Initialize test environment
setup_test_env() {
  # Create isolated plonk config directory
  export PLONK_DIR="$BATS_TEST_TMPDIR/plonk-config"
  mkdir -p "$PLONK_DIR"

  # Create cleanup tracking file
  export CLEANUP_FILE="$BATS_TEST_TMPDIR/cleanup.list"
  : > "$CLEANUP_FILE"

  # Load safe lists
  load_safe_lists

  # Set test mode flag
  export PLONK_TEST_MODE=1
}

# Load safe lists from files or environment
load_safe_lists() {
  # Load safe packages
  if [[ -n "$PLONK_TEST_SAFE_PACKAGES" ]]; then
    IFS=',' read -ra SAFE_PACKAGES <<< "$PLONK_TEST_SAFE_PACKAGES"
  else
    mapfile -t SAFE_PACKAGES < <(grep -v '^#' "$SAFE_PACKAGES_FILE" | grep -v '^$')
  fi
  export SAFE_PACKAGES

  # Load safe dotfiles
  if [[ -n "$PLONK_TEST_SAFE_DOTFILES" ]]; then
    IFS=',' read -ra SAFE_DOTFILES <<< "$PLONK_TEST_SAFE_DOTFILES"
  else
    mapfile -t SAFE_DOTFILES < <(grep -v '^#' "$SAFE_DOTFILES_FILE" | grep -v '^$')
  fi
  export SAFE_DOTFILES
}

# Track an artifact for cleanup
track_artifact() {
  local type="$1"
  local name="$2"
  echo "${type}:${name}" >> "$CLEANUP_FILE"
}

# Check if a package is in the safe list
is_safe_package() {
  local package="$1"
  for safe in "${SAFE_PACKAGES[@]}"; do
    if [[ "$package" == "$safe" ]] || [[ "$package" == *":${safe#*:}" ]]; then
      return 0
    fi
  done
  return 1
}

# Check if a dotfile is in the safe list
is_safe_dotfile() {
  local dotfile="$1"
  for safe in "${SAFE_DOTFILES[@]}"; do
    if [[ "$dotfile" == "$safe" ]] || [[ "$(basename "$dotfile")" == "$safe" ]]; then
      return 0
    fi
  done
  return 1
}


# Require a safe package or skip
require_safe_package() {
  local package="$1"
  if ! is_safe_package "$package"; then
    skip "Package $package not in safe list"
  fi
}

# Require a safe dotfile or skip
require_safe_dotfile() {
  local dotfile="$1"
  if ! is_safe_dotfile "$dotfile"; then
    skip "Dotfile $dotfile not in safe list"
  fi
}

# Create a test config file
create_test_config() {
  local content="$1"
  echo "$content" > "$PLONK_DIR/plonk.yaml"
}

# Create a test dotfile
create_test_dotfile() {
  local name="$1"
  local content="${2:-# Test file created by BATS}"

  require_safe_dotfile "$name"

  local dir=$(dirname "$name")
  if [[ "$dir" != "." ]]; then
    mkdir -p "$HOME/$dir"
  fi

  echo "$content" > "$HOME/$name"
  track_artifact "dotfile" "$name"
}

# Cleanup function for teardown
cleanup_test_artifacts() {
  if [[ ! -f "$CLEANUP_FILE" ]]; then
    return
  fi

  while IFS=: read -r type name; do
    case "$type" in
      dotfile)
        rm -rf "$HOME/$name" 2>/dev/null || true
        ;;
      package)
        if [[ "$PLONK_TEST_CLEANUP_PACKAGES" == "1" ]]; then
          plonk uninstall "$name" --force 2>/dev/null || true
        fi
        ;;
    esac
  done < "$CLEANUP_FILE"
}

# Standard teardown
teardown() {
  if [[ "$PLONK_TEST_CLEANUP_DOTFILES" != "0" ]]; then
    cleanup_test_artifacts
  fi
}

# File-level teardown
teardown_file() {
  cleanup_test_artifacts
}
```

### 1.5 Create Assertion Helpers

**File: `tests/bats/lib/assertions.bash`**
```bash
#!/usr/bin/env bash

# Assert output contains all of the provided strings
assert_output_contains_all() {
  for expected in "$@"; do
    assert_output --partial "$expected"
  done
}

# Assert output contains none of the provided strings
assert_output_contains_none() {
  for unexpected in "$@"; do
    refute_output --partial "$unexpected"
  done
}

# Assert JSON field has expected value
assert_json_field() {
  local field="$1"
  local expected="$2"
  local actual=$(echo "$output" | jq -r "$field")

  if [[ "$actual" != "$expected" ]]; then
    echo "Expected $field to be '$expected', got '$actual'" >&2
    return 1
  fi
}

# Assert JSON field exists
assert_json_field_exists() {
  local field="$1"
  local actual=$(echo "$output" | jq -r "$field")

  if [[ "$actual" == "null" ]]; then
    echo "Expected $field to exist in JSON output" >&2
    return 1
  fi
}

# Assert table contains row with values
assert_table_row() {
  local row_pattern="$1"
  if ! echo "$output" | grep -E "$row_pattern" > /dev/null; then
    echo "Expected table to contain row matching: $row_pattern" >&2
    return 1
  fi
}

# Assert exit code
assert_exit_code() {
  local expected="$1"
  if [[ "$status" -ne "$expected" ]]; then
    echo "Expected exit code $expected, got $status" >&2
    return 1
  fi
}
```

### 1.6 Create Cleanup Utilities

**File: `tests/bats/lib/cleanup.bash`**
```bash
#!/usr/bin/env bash

# Remove all test dotfiles
cleanup_all_test_dotfiles() {
  for dotfile in "${SAFE_DOTFILES[@]}"; do
    if [[ -e "$HOME/$dotfile" ]]; then
      rm -rf "$HOME/$dotfile"
      echo "Removed test dotfile: $dotfile"
    fi
  done
}

# Uninstall all test packages
cleanup_all_test_packages() {
  for package in "${SAFE_PACKAGES[@]}"; do
    if plonk status | grep -q "$package"; then
      plonk uninstall "$package" --force 2>/dev/null || true
      echo "Removed test package: $package"
    fi
  done
}

# Full cleanup of all test artifacts
cleanup_all_test_artifacts() {
  echo "Starting full test cleanup..."
  cleanup_all_test_dotfiles
  cleanup_all_test_packages
  echo "Cleanup complete"
}

# Check for test artifacts
check_for_test_artifacts() {
  local found_artifacts=0

  echo "Checking for test artifacts..."

  # Check dotfiles
  for dotfile in "${SAFE_DOTFILES[@]}"; do
    if [[ -e "$HOME/$dotfile" ]]; then
      echo "Found test dotfile: $dotfile"
      ((found_artifacts++))
    fi
  done

  # Check packages
  for package in "${SAFE_PACKAGES[@]}"; do
    if plonk status | grep -q "${package#*:}"; then
      echo "Found test package: $package"
      ((found_artifacts++))
    fi
  done

  return $found_artifacts
}
```

### 1.7 Create Initial Smoke Test

**File: `tests/bats/behavioral/00-smoke.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

@test "plonk binary exists and is executable" {
  run which plonk
  assert_success

  # Verify it's executable
  run test -x "$(which plonk)"
  assert_success
}

@test "test environment setup works correctly" {
  setup_test_env

  # Verify environment variables
  assert [ -n "$PLONK_DIR" ]
  assert [ -d "$PLONK_DIR" ]
  assert [ -n "$CLEANUP_FILE" ]
  assert [ -f "$CLEANUP_FILE" ]
}

@test "safe lists load correctly" {
  setup_test_env

  # Verify safe packages loaded
  assert [ ${#SAFE_PACKAGES[@]} -gt 0 ]

  # Verify safe dotfiles loaded
  assert [ ${#SAFE_DOTFILES[@]} -gt 0 ]

  # Test package checking
  run is_safe_package "brew:jq"
  assert_success

  run is_safe_package "brew:definitely-not-safe"
  assert_failure
}

@test "cleanup tracking works" {
  setup_test_env

  # Track some artifacts
  track_artifact "dotfile" ".test-file"
  track_artifact "package" "brew:test"

  # Verify they were tracked
  run grep "dotfile:.test-file" "$CLEANUP_FILE"
  assert_success

  run grep "package:brew:test" "$CLEANUP_FILE"
  assert_success
}

@test "plonk help command works" {
  run plonk help
  assert_success
  assert_output --partial "Usage:"
  assert_output --partial "Available Commands:"
}

@test "plonk version shows version info" {
  run plonk --version
  assert_success
  assert_output --partial "plonk"
}
```

## Phase 2: Basic Command Tests (Day 3-4) ✅ COMPLETED

### 2.1 Basic Commands Test

**File: `tests/bats/behavioral/01-basic-commands.bats`** ✅
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "plonk with no args shows help" {
  run plonk
  assert_success
  assert_output --partial "Usage:"
  assert_output --partial "Available Commands:"
}

@test "plonk status works with empty config" {
  run plonk status
  assert_success
  assert_output --partial "0 managed"
}

@test "plonk st alias works" {
  run plonk st
  assert_success
  assert_output --partial "0 managed"
}

# Removed doctor and error handling tests per user request

@test "help for specific command works" {
  run plonk help install
  assert_success
  assert_output --partial "Install packages"
  assert_output --partial "Examples:"
}
```

### 2.2 Output Format Tests

**File: `tests/bats/behavioral/02-output-formats.bats`** ✅
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "status supports table format (default)" {
  run plonk status
  assert_success
  # Look for table-like formatting
  assert_output --partial "Plonk Status"
}

# Only testing table format per user request
@test "info supports table format" {
  run plonk info jq
  assert_success
  # Table format should show package details
  assert_output --partial "Package:"
  assert_output --partial "jq"
}

@test "search supports table format" {
  run plonk search jq
  assert_success
  # Table format shows search results
  assert_output --partial "jq"
}
```

## Phase 3: State-Changing Tests (Day 5-7)

### 3.1 Package Installation Tests

**File: `tests/bats/behavioral/03-package-install.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "install single package with prefix syntax" {
  require_safe_package "brew:jq"

  run plonk install brew:jq
  assert_success
  assert_output --partial "jq"
  assert_output --partial "installed"

  track_artifact "package" "brew:jq"

  # Verify it's in status
  run plonk status
  assert_output --partial "jq"
  assert_output --partial "1 managed"
}

@test "install shows error for non-existent package" {
  run plonk install brew:definitely-not-real-xyz123
  assert_failure
  assert_output --partial "not found"
  refute_output --partial "panic"
}

@test "install multiple packages in one command" {
  require_safe_package "brew:jq"
  require_safe_package "brew:tree"

  run plonk install brew:jq brew:tree
  assert_success
  assert_output_contains_all "jq" "tree"

  track_artifact "package" "brew:jq"
  track_artifact "package" "brew:tree"

  # Verify both in status
  run plonk status
  assert_output_contains_all "jq" "tree" "2 managed"
}

@test "install with dry-run doesn't actually install" {
  require_safe_package "brew:jq"

  run plonk install brew:jq --dry-run
  assert_success
  assert_output --partial "would install"

  # Verify not actually installed
  run plonk status
  assert_output --partial "0 managed"
}

@test "install already managed package shows appropriate message" {
  require_safe_package "brew:jq"

  # First install
  run plonk install brew:jq
  assert_success
  track_artifact "package" "brew:jq"

  # Try to install again
  run plonk install brew:jq
  assert_success
  assert_output --partial "already managed"
}
```

### 3.2 Package Uninstall Tests

**File: `tests/bats/behavioral/04-package-uninstall.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "uninstall managed package" {
  require_safe_package "brew:jq"

  # Install first
  run plonk install brew:jq
  assert_success

  # Then uninstall
  run plonk uninstall brew:jq
  assert_success
  assert_output --partial "uninstalled"

  # Verify gone from status
  run plonk status
  assert_output --partial "0 managed"
}

@test "uninstall non-managed package shows error" {
  run plonk uninstall brew:not-managed-package
  assert_failure
  assert_output --partial "not managed"
}

@test "uninstall with force removes even if not managed" {
  require_safe_package "brew:tree"

  # Ensure it's not managed
  run plonk status
  refute_output --partial "tree"

  # Force uninstall (this might fail if not installed at all)
  run plonk uninstall brew:tree --force
  # Don't assert success/failure as it depends on system state
}

@test "uninstall with dry-run shows what would happen" {
  require_safe_package "brew:jq"

  # Install first
  run plonk install brew:jq
  assert_success
  track_artifact "package" "brew:jq"

  # Dry-run uninstall
  run plonk uninstall brew:jq --dry-run
  assert_success
  assert_output --partial "would uninstall"

  # Verify still managed
  run plonk status
  assert_output --partial "jq"
}
```

## Phase 4: Dotfile Tests (Day 8)

### 4.1 Dotfile Add Tests

**File: `tests/bats/behavioral/05-dotfile-add.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "add single dotfile" {
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Create source file
  create_test_dotfile "$testfile" "# BATS test file"

  # Add to plonk
  run plonk add "$HOME/$testfile"
  assert_success
  assert_output --partial "Added"
  assert_output --partial "$testfile"

  # Verify in status
  run plonk status
  assert_output --partial "$testfile"
  assert_output --partial "1 managed"
}

@test "add dotfile with dry-run" {
  local testfile=".plonk-test-profile"
  require_safe_dotfile "$testfile"

  create_test_dotfile "$testfile"

  run plonk add "$HOME/$testfile" --dry-run
  assert_success
  assert_output --partial "would add"

  # Verify not actually added
  run plonk status
  assert_output --partial "0 managed"
}

@test "add non-existent dotfile shows error" {
  run plonk add "$HOME/.definitely-does-not-exist-xyz"
  assert_failure
  assert_output --partial "not found"
}

@test "add directory of dotfiles" {
  local testdir=".config/plonk-test"
  require_safe_dotfile "$testdir/config.yaml"

  # Create test directory with files
  mkdir -p "$HOME/$testdir"
  echo "test: true" > "$HOME/$testdir/config.yaml"
  echo "# Test settings" > "$HOME/$testdir/settings.json"

  track_artifact "dotfile" "$testdir"

  # Add directory
  run plonk add "$HOME/$testdir"
  assert_success
  assert_output --partial "Added"

  # Verify files tracked
  run plonk status
  assert_output --partial "config.yaml"
}
```

### 4.2 Dotfile Remove Tests

**File: `tests/bats/behavioral/06-dotfile-rm.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "remove managed dotfile" {
  local testfile=".plonk-test-gitconfig"
  require_safe_dotfile "$testfile"

  # Add a dotfile first
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success

  # Remove it
  run plonk rm "$testfile"
  assert_success
  assert_output --partial "Removed"

  # Verify gone from status
  run plonk status
  assert_output --partial "0 managed"
}

@test "remove with dry-run shows what would happen" {
  local testfile=".plonk-test-bashrc"
  require_safe_dotfile "$testfile"

  # Add first
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success

  # Dry-run remove
  run plonk rm "$testfile" --dry-run
  assert_success
  assert_output --partial "would remove"

  # Verify still managed
  run plonk status
  assert_output --partial "$testfile"
}

@test "remove non-managed dotfile shows error" {
  run plonk rm ".not-managed-file"
  assert_failure
  assert_output --partial "not managed"
}
```

## Phase 5: Complex Workflows (Day 9-10)

### 5.1 Apply Command Tests

**File: `tests/bats/behavioral/07-apply-behavior.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "apply with all successful operations" {
  require_safe_package "brew:jq"
  require_safe_package "brew:tree"

  # Create config with valid packages
  create_test_config "packages:
  - jq
  - tree"

  run plonk apply
  assert_success
  assert_output_contains_all "✓ jq" "✓ tree"
  assert_output --partial "2 succeeded"

  track_artifact "package" "brew:jq"
  track_artifact "package" "brew:tree"
}

@test "apply continues after package failure" {
  require_safe_package "brew:jq"

  # Create config with valid and invalid packages
  create_test_config "packages:
  - jq
  - definitely-fake-xyz
  - tree"

  require_safe_package "brew:tree"

  run plonk apply
  assert_failure  # Should fail due to partial failure

  # Verify output shows continuation
  assert_output --partial "✓ jq"
  assert_output --partial "✗ definitely-fake-xyz"
  assert_output --partial "✓ tree"
  assert_output --partial "2 succeeded, 1 failed"

  track_artifact "package" "brew:jq"
  track_artifact "package" "brew:tree"

  # Verify successful packages were installed
  run plonk status
  assert_output_contains_all "jq" "tree"
}

@test "apply with dry-run shows what would happen" {
  require_safe_package "brew:jq"

  create_test_config "packages:
  - jq"

  run plonk apply --dry-run
  assert_success
  assert_output --partial "would install"
  assert_output --partial "jq"

  # Verify nothing actually installed
  run plonk status
  assert_output --partial "0 managed"
}

@test "apply with empty config succeeds" {
  create_test_config ""

  run plonk apply
  assert_success
  assert_output --partial "All up to date"
}

@test "apply with mixed packages and dotfiles" {
  require_safe_package "brew:jq"
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Create dotfile
  create_test_dotfile "$testfile"

  # Create config
  create_test_config "packages:
  - jq

dotfiles:
  - $testfile"

  run plonk apply
  assert_success
  assert_output --partial "✓ jq"
  assert_output --partial "✓ $testfile"

  track_artifact "package" "brew:jq"
}
```

### 5.2 Search and Info Tests

**File: `tests/bats/behavioral/08-search-info.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "search finds packages across managers" {
  run plonk search git
  assert_success
  assert_output --partial "git"
  # Should show results from multiple managers
}

@test "search with prefix searches specific manager" {
  run plonk search brew:git
  assert_success
  assert_output --partial "git"
  assert_output --partial "brew"
}

@test "search for non-existent package" {
  run plonk search definitely-not-a-real-package-xyz123
  assert_success  # Search doesn't fail, just returns no results
  assert_output --partial "No packages found"
}

@test "info shows details for available package" {
  run plonk info brew:jq
  assert_success
  assert_output --partial "jq"
  assert_output --partial "Available"
}

@test "info shows managed status for installed package" {
  require_safe_package "brew:jq"

  # Install package
  run plonk install brew:jq
  assert_success
  track_artifact "package" "brew:jq"

  # Check info
  run plonk info jq
  assert_success
  assert_output --partial "Managed by plonk"
}

@test "info for non-existent package shows not found" {
  run plonk info brew:definitely-fake-xyz
  assert_success  # Command succeeds but shows not found
  assert_output --partial "not found"
}
```

## Phase 6: Error Handling Tests (Day 11)

**File: `tests/bats/behavioral/09-error-handling.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "invalid config file shows clear error" {
  # Create invalid YAML
  create_test_config "packages:
  - jq
  invalid yaml here: ["

  run plonk apply
  assert_failure
  assert_output --partial "invalid"
  refute_output --partial "panic"
}

@test "permission denied handled gracefully" {
  skip "Requires special setup to test permission errors"

  # Would need to create read-only directories, etc.
  # Complex to test reliably across systems
}

@test "missing package manager shows helpful message" {
  # Try to use a manager that might not be installed
  run plonk install cargo:ripgrep

  # Should either succeed (if cargo installed) or show helpful error
  if [[ $status -ne 0 ]]; then
    assert_output --partial "not available"
    # Should suggest installation
  fi
}

```

## Phase 7: Integration Tests (Day 12)

### 7.1 Cleanup Test

**File: `tests/bats/cleanup/99-cleanup-all.bats`**
```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/cleanup'
load '../lib/assertions'

@test "check for test artifacts before cleanup" {
  setup_test_env

  run check_for_test_artifacts
  # Exit code is number of artifacts found
  echo "Found $status test artifacts"
}

@test "cleanup all test dotfiles" {
  setup_test_env

  if [[ "$PLONK_TEST_CLEANUP_DOTFILES" == "0" ]]; then
    skip "Dotfile cleanup disabled"
  fi

  run cleanup_all_test_dotfiles
  assert_success
}

@test "cleanup all test packages" {
  setup_test_env

  if [[ "$PLONK_TEST_CLEANUP_PACKAGES" != "1" ]]; then
    skip "Package cleanup not requested (set PLONK_TEST_CLEANUP_PACKAGES=1)"
  fi

  run cleanup_all_test_packages
  assert_success
}

@test "verify cleanup completed" {
  setup_test_env

  run check_for_test_artifacts
  if [[ $status -eq 0 ]]; then
    echo "All test artifacts cleaned up successfully"
  else
    echo "Warning: $status test artifacts remain"
  fi
}
```

### 7.2 Test Fixtures

**File: `tests/bats/fixtures/configs/basic.yaml`**
```yaml
# Basic test configuration
packages:
  - jq
  - tree

dotfiles:
  - .plonk-test-rc
```

**File: `tests/bats/fixtures/configs/complex.yaml`**
```yaml
# Complex test configuration
default_manager: brew

packages:
  - jq
  - tree
  - name: typescript
    manager: npm
  - name: cowsay
    manager: pip

dotfiles:
  - .plonk-test-rc
  - .config/plonk-test/

hooks:
  pre_apply:
    - command: echo "Starting apply"
  post_apply:
    - command: echo "Apply complete"
```

**File: `tests/bats/fixtures/dotfiles/.plonk-test-rc`**
```bash
# Test RC file for BATS
# This file is safe to create/delete

export PLONK_TEST_MARKER="bats-test"
```

## Phase 8: CI Integration (Day 13)

### 8.1 GitHub Actions Workflow

**File: `.github/workflows/bats-tests.yml`**
```yaml
name: BATS Behavioral Tests

on:
  pull_request:
    paths:
      - 'internal/**'
      - 'tests/bats/**'
      - '.github/workflows/bats-tests.yml'
  push:
    branches:
      - main
  workflow_dispatch:

jobs:
  bats-tests:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        test-suite: [basic, behavioral]
      fail-fast: false

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build plonk
        run: go build -o plonk cmd/plonk/main.go

      - name: Add plonk to PATH
        run: echo "$PWD" >> $GITHUB_PATH

      - name: Install BATS
        run: |
          if [[ "${{ runner.os }}" == "macOS" ]]; then
            brew install bats-core
          else
            sudo apt-get update
            sudo apt-get install -y bats
          fi

      - name: Install test package managers
        run: |
          if [[ "${{ runner.os }}" == "Linux" ]]; then
            # Install npm
            curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            sudo apt-get install -y nodejs

            # pip should already be available
            python3 -m pip --version || sudo apt-get install -y python3-pip
          fi

      - name: Run basic tests
        if: matrix.test-suite == 'basic'
        env:
          PLONK_TEST_SKIP_DANGEROUS: "1"
        run: |
          bats tests/bats/behavioral/0[0-2]-*.bats

      - name: Run behavioral tests (safe subset)
        if: matrix.test-suite == 'behavioral'
        env:
          PLONK_TEST_SKIP_DANGEROUS: "0"
          PLONK_TEST_CLEANUP_PACKAGES: "1"
          PLONK_TEST_CLEANUP_DOTFILES: "1"
          PLONK_TEST_SAFE_PACKAGES: "brew:jq"  # Minimal for CI
          PLONK_TEST_SAFE_DOTFILES: ".plonk-test-rc"
        run: |
          bats tests/bats/behavioral/

      - name: Cleanup check
        if: always()
        run: |
          bats tests/bats/cleanup/99-cleanup-all.bats
```

### 8.2 Makefile Integration

**Add to `Makefile`:**
```makefile
# BATS testing targets
.PHONY: test-bats test-bats-safe test-bats-full test-bats-clean

# Run safe BATS tests only
test-bats-safe:
	@echo "Running safe BATS tests (no system modifications)..."
	PLONK_TEST_SKIP_DANGEROUS=1 bats tests/bats/behavioral/0[0-2]-*.bats

# Run all BATS tests (WITH SYSTEM MODIFICATIONS)
test-bats-full:
	@echo "⚠️  WARNING: This will modify your system! ⚠️"
	@echo "Installing packages: $(shell cat tests/bats/config/safe-packages.list | grep -v '^#')"
	@echo "Creating dotfiles: $(shell cat tests/bats/config/safe-dotfiles.list | grep -v '^#')"
	@echo ""
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		bats tests/bats/behavioral/; \
	else \
		echo "Aborted."; \
		exit 1; \
	fi

# Run BATS tests with auto-cleanup
test-bats:
	@echo "Running BATS tests with auto-cleanup..."
	PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 \
		bats tests/bats/behavioral/

# Clean up all test artifacts
test-bats-clean:
	@echo "Cleaning up all BATS test artifacts..."
	PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 \
		bats tests/bats/cleanup/99-cleanup-all.bats

# Run specific BATS test file
test-bats-file:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make test-bats-file FILE=tests/bats/behavioral/01-basic-commands.bats"; \
		exit 1; \
	fi
	bats $(FILE)
```

### 8.3 Just Recipe Integration

**Add to `justfile`:**
```just
# Run safe BATS tests (no system modifications)
test-bats-safe:
    PLONK_TEST_SKIP_DANGEROUS=1 bats tests/bats/behavioral/0[0-2]-*.bats

# Run all BATS tests with cleanup
test-bats:
    PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 bats tests/bats/behavioral/

# Clean up test artifacts
test-bats-clean:
    PLONK_TEST_CLEANUP_PACKAGES=1 bats tests/bats/cleanup/99-cleanup-all.bats

# Run BATS tests in watch mode
test-bats-watch:
    watchexec -e bats,bash -- just test-bats-safe
```

## Phase 9: Documentation Updates (Day 14)

### 9.1 Update Main README

Add to `README.md`:
```markdown
## Testing

Plonk uses two testing approaches:

1. **Go Unit Tests**: Test internal logic and functions
   ```bash
   go test ./...
   ```

2. **BATS Behavioral Tests**: Test CLI behavior and user experience
   ```bash
   # Safe tests only (no system modifications)
   make test-bats-safe

   # Full test suite (MODIFIES SYSTEM - see tests/bats/README.md)
   make test-bats
   ```

⚠️ **WARNING**: BATS tests install real packages and create real files. See `tests/bats/README.md` for details.
```

### 9.2 Update REFACTOR.md

Add test completion status:
```markdown
### Phase 14: Additional UX Improvements ✅ COMPLETE
- [x] Apply command partial failure handling
- [x] Help text updates
- [x] Error message consistency
- [x] Command behavior polish
- [x] BATS test suite implementation
  - [x] Test infrastructure created
  - [x] Behavioral tests for all commands
  - [x] Safety mechanisms implemented
  - [x] CI integration configured
```

### 9.3 Create PHASE_14_COMPLETION.md

**File: `PHASE_14_COMPLETION.md`**
```markdown
# Phase 14 Completion Report

## Overview
Phase 14 has been successfully completed with all objectives met.

## Completed Items

### 1. Apply Command Partial Failure Handling ✅
- Modified orchestrator to continue on errors
- Enhanced output to show detailed success/failure
- Exit codes reflect partial failures

### 2. Help Text and Documentation Updates ✅
- Updated all command help text
- Removed references to old commands
- Updated README with new syntax

### 3. Error Message Consistency ✅
- Standardized error formatting functions
- Consistent error patterns across commands
- Helpful error messages with solutions

### 4. Command Behavior Polish ✅
- Removed obsolete code
- Consistent output handling
- Improved user experience

### 5. BATS Test Suite Implementation ✅
- Created comprehensive behavioral test suite
- Implemented safety mechanisms
- Added CI/CD integration
- Documented test approach

## Test Coverage

### Behavioral Tests Created:
- Basic commands (help, version, status)
- Output formats (table, JSON, YAML)
- Package operations (install, uninstall)
- Dotfile operations (add, rm)
- Apply command with failures
- Search and info commands
- Error handling scenarios

### Safety Features:
- Isolated test environment
- Safe package/dotfile lists
- Cleanup mechanisms
- Skip dangerous tests option
- Clear warnings in documentation

## Metrics
- Test files created: 10
- Test cases: ~50
- Lines of test code: ~1,500
- Documentation: Comprehensive

## Next Steps
- Move to Final Phase: Testing & Documentation
- Regular test maintenance
- Community feedback on safe lists
```

## Validation Checklist

### Week 1 Checkpoint
- [ ] Test infrastructure works
- [ ] Basic tests pass
- [ ] No system damage
- [ ] Documentation clear

### Week 2 Checkpoint
- [ ] All commands have tests
- [ ] CI integration works
- [ ] Cleanup verified
- [ ] Team sign-off

### Final Validation
- [ ] Full suite < 2 minutes
- [ ] No artifact leaks
- [ ] Safe by default
- [ ] Ready for use

## Maintenance Notes

1. **Adding New Tests**:
   - Always use safe lists
   - Track artifacts for cleanup
   - Test locally first
   - Update CI if needed

2. **Updating Safe Lists**:
   - Review quarterly
   - Get team consensus
   - Test thoroughly
   - Update documentation

3. **Debugging Failures**:
   - Check environment vars
   - Verify package managers
   - Look for artifacts
   - Run cleanup if needed

This detailed action plan provides everything needed to implement comprehensive BATS testing for plonk while maintaining safety and focusing on behavioral testing of the CLI interface.
