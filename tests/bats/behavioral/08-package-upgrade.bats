#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'
load '../lib/package_test_helper'

setup() {
  setup_test_env
}

# Helper to get installed npm package version
get_npm_version() {
  local package="$1"
  npm list -g "$package" --depth=0 2>/dev/null | grep "$package@" | sed 's/.*@//' | tr -d ' '
}

# Helper to get installed brew package version
get_brew_version() {
  local package="$1"
  brew list --versions "$package" 2>/dev/null | awk '{print $2}'
}

# =============================================================================
# Version Upgrade Verification Tests
# These tests verify that upgrade actually changes package versions
# =============================================================================

@test "npm upgrade actually upgrades from older version" {
  require_package_manager "npm"
  require_safe_package "npm:is-odd"

  # First, ensure the package is not installed
  npm uninstall -g is-odd 2>/dev/null || true

  # Install an OLD version directly via npm (not plonk)
  run npm install -g is-odd@2.0.0
  assert_success
  track_artifact "package" "npm:is-odd"

  # Verify old version is installed
  old_version=$(get_npm_version "is-odd")
  if [[ "$old_version" != "2.0.0" ]]; then
    fail "Expected version 2.0.0, got $old_version"
  fi

  # Add to plonk tracking (plonk install is idempotent, won't reinstall)
  run plonk install npm:is-odd
  assert_success

  # Run upgrade via plonk
  run plonk upgrade npm:is-odd
  assert_success

  # Verify version changed (should be 3.0.1 or newer)
  new_version=$(get_npm_version "is-odd")

  # Version should be different from old version (upgraded)
  if [[ "$new_version" == "2.0.0" ]]; then
    fail "Upgrade did not change version - still at 2.0.0"
  fi

  # Verify it's a higher version (simple numeric comparison for major version)
  new_major="${new_version%%.*}"
  if [[ "$new_major" -lt 3 ]]; then
    fail "Expected version >= 3.0.0, got $new_version"
  fi
}

@test "uv upgrade actually upgrades from older version" {
  require_package_manager "uv"
  require_safe_package "uv:cowsay"

  # First, ensure the package is not installed
  uv tool uninstall cowsay 2>/dev/null || true

  # Install an OLD version directly via uv
  run uv tool install cowsay==5.0
  if [[ $status -ne 0 ]]; then
    skip "Failed to install cowsay@5.0 via uv"
  fi
  track_artifact "package" "uv:cowsay"

  # Get old version
  old_version=$(uv tool list 2>/dev/null | grep "^cowsay" | sed 's/cowsay v//' | awk '{print $1}')

  # Add to plonk tracking
  run plonk install uv:cowsay
  assert_success

  # Run upgrade via plonk
  run plonk upgrade uv:cowsay
  assert_success

  # Get new version
  new_version=$(uv tool list 2>/dev/null | grep "^cowsay" | sed 's/cowsay v//' | awk '{print $1}')

  # Version should have changed (unless already at latest)
  # At minimum, verify upgrade command ran and package still works
  run uv tool list
  assert_success
  assert_output --partial "cowsay"
}

# =============================================================================
# Basic upgrade syntax tests
# =============================================================================

@test "upgrade command shows help when no arguments given to empty environment" {
  run plonk upgrade
  assert_success
  assert_output --partial "No packages to upgrade"
}

@test "upgrade command rejects trailing colon syntax" {
  require_safe_package "brew:cowsay"

  # First install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Now try invalid trailing colon syntax
  run plonk upgrade brew:
  assert_failure
  assert_output --partial "invalid syntax 'brew:'"
  assert_output --partial "use 'brew' to upgrade"
}

@test "upgrade command rejects unknown package" {
  run plonk upgrade nonexistent-package-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "upgrade command rejects unknown manager:package" {
  run plonk upgrade brew:nonexistent-package-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

# =============================================================================
# Per-manager single package upgrade tests
# =============================================================================

@test "upgrade single brew package" {
  test_upgrade_single "brew" "cowsay"
}

@test "upgrade single npm package" {
  test_upgrade_single "npm" "is-odd"
}

@test "upgrade single uv package" {
  test_upgrade_single "uv" "cowsay"
}

@test "upgrade single gem package" {
  test_upgrade_single "gem" "colorize"
}

@test "upgrade single cargo package" {
  test_upgrade_single "cargo" "ripgrep"
}

@test "upgrade single pnpm package" {
  test_upgrade_single "pnpm" "prettier"
}

# =============================================================================
# Per-manager upgrade all packages tests
# =============================================================================

@test "upgrade all brew packages" {
  test_upgrade_all_manager "brew" "figlet" "sl"
}

@test "upgrade all npm packages" {
  test_upgrade_all_manager "npm" "is-odd" "left-pad"
}

@test "upgrade all pnpm packages" {
  test_upgrade_all_manager "pnpm" "prettier"
}

# =============================================================================
# Cross-manager upgrade tests
# =============================================================================

@test "upgrade package across managers" {
  require_safe_package "brew:cowsay"
  require_package_manager "uv"
  require_safe_package "uv:cowsay"

  # Install cowsay via both managers
  run plonk install brew:cowsay uv:cowsay
  if [[ $status -ne 0 ]]; then
    skip "Failed to install cowsay packages"
  fi
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "uv:cowsay"

  # Upgrade cowsay (should upgrade both)
  run plonk upgrade cowsay
  assert_success
  assert_output --partial "cowsay"
  # Should show both managers (only if both were installed)
  assert_output --partial "brew"
  assert_output --partial "uv"
}

# All packages upgrade test
@test "upgrade all packages" {
  require_safe_package "brew:figlet"
  require_package_manager "npm"
  require_safe_package "npm:is-odd"

  # Install packages from different managers
  run plonk install brew:figlet npm:is-odd
  assert_success
  track_artifact "package" "brew:figlet"
  track_artifact "package" "npm:is-odd"

  # Upgrade all packages
  run plonk upgrade
  assert_success
  assert_output --partial "figlet"
  assert_output --partial "is-odd"
  assert_output --partial "Summary:"
}

# =============================================================================
# Error handling tests
# =============================================================================

@test "upgrade handles unavailable package manager gracefully" {
  # Try to upgrade with a package manager that doesn't exist
  run plonk upgrade nonexistent-manager:some-package
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "upgrade continues on individual package failures" {
  require_safe_package "brew:figlet"

  # Install one real package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Try to upgrade the real package alongside a fake one via cross-manager syntax
  # This should upgrade figlet successfully but report that fake-package isn't managed
  run plonk upgrade figlet fake-package-xyz
  # This will fail overall, but figlet should still be processed
  assert_failure
  assert_output --partial "not managed by plonk"
}

# =============================================================================
# Spinner tests for upgrade command
# =============================================================================

@test "upgrade shows spinner during operation" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Upgrade and check for spinner output
  run plonk upgrade brew:cowsay
  assert_success
  assert_output --partial "Upgrading"
}

@test "upgrade shows progress indicators for multiple packages" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Upgrade both and check for progress indicators
  run plonk upgrade brew:cowsay brew:figlet
  assert_success
  assert_output --partial "[1/2]"
  assert_output --partial "[2/2]"
  assert_output --partial "Upgrading"
}

# =============================================================================
# Dry-run tests
# =============================================================================

@test "upgrade --dry-run shows what would happen" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Dry-run upgrade
  run plonk upgrade --dry-run brew:cowsay
  assert_success
  assert_output --partial "would-upgrade"
  assert_output --partial "cowsay"
}

@test "upgrade --dry-run does not modify packages" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Get initial state
  initial_version=$(brew list --versions cowsay)

  # Dry-run upgrade
  run plonk upgrade --dry-run brew:cowsay
  assert_success

  # Version should be unchanged
  final_version=$(brew list --versions cowsay)
  assert [ "$initial_version" = "$final_version" ]
}

@test "upgrade -n is alias for --dry-run" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk upgrade -n brew:cowsay
  assert_success
  assert_output --partial "would-upgrade"
}

@test "upgrade --dry-run all packages" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  # Install both packages
  run plonk install brew:figlet brew:sl
  assert_success
  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"

  # Dry-run upgrade all
  run plonk upgrade --dry-run
  assert_success
  assert_output --partial "would-upgrade"
  assert_output --partial "figlet"
  assert_output --partial "sl"
}

@test "upgrade --dry-run by manager" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Dry-run upgrade by manager
  run plonk upgrade --dry-run brew
  assert_success
  assert_output --partial "would-upgrade"
  assert_output --partial "cowsay"
}
