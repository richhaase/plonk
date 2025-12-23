#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'
load '../lib/package_test_helper'

setup() {
  setup_test_env
}

# =============================================================================
# Single package install tests (per manager)
# =============================================================================

@test "install single brew package" {
  test_install_single "brew" "cowsay"
}

@test "install npm package" {
  test_install_single "npm" "is-odd"
}

@test "install gem package" {
  test_install_single "gem" "colorize"
}

@test "install cargo package" {
  test_install_single "cargo" "ripgrep"
}

@test "install uv package" {
  test_install_single "uv" "cowsay"
}

@test "install pnpm package" {
  test_install_single "pnpm" "prettier"
}

# =============================================================================
# Multiple package installation
# =============================================================================

# Test multiple package installation - only testing with brew since
# the logic is the same for all managers (just loops over single installs)
@test "install multiple brew packages" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  run plonk install brew:figlet brew:sl
  assert_success
  assert_output_contains_all "figlet" "sl"

  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"

  # Verify they're actually installed by brew
  run brew list figlet
  assert_success
  run brew list sl
  assert_success

  # Verify both in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "figlet" "sl"

  # Verify both in status
  run plonk status
  assert_output_contains_all "figlet" "sl"
}

# =============================================================================
# Dry-run tests
# =============================================================================

@test "install --dry-run shows what would happen" {
  require_safe_package "brew:lolcat"

  run plonk install --dry-run brew:lolcat
  assert_success
  assert_output --partial "lolcat"
  assert_output --partial "would-add"
}

@test "install --dry-run does not modify lock file" {
  require_safe_package "brew:lolcat"

  # Get initial lock file state
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    initial_content=$(cat "$PLONK_DIR/plonk.lock")
  else
    initial_content=""
  fi

  run plonk install --dry-run brew:lolcat
  assert_success

  # Lock file should not have changed
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    final_content=$(cat "$PLONK_DIR/plonk.lock")
  else
    final_content=""
  fi

  assert [ "$initial_content" = "$final_content" ]
}

@test "install --dry-run does not install package" {
  require_safe_package "brew:fortune"

  # Make sure package is not installed
  run brew uninstall fortune 2>/dev/null || true

  run plonk install --dry-run brew:fortune
  assert_success

  # Package should not be installed
  run brew list fortune
  assert_failure
}

@test "install -n is alias for --dry-run" {
  require_safe_package "brew:lolcat"

  run plonk install -n brew:lolcat
  assert_success
  assert_output --partial "would-add"
}

@test "install --dry-run with multiple packages shows all" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  run plonk install --dry-run brew:figlet brew:sl
  assert_success
  assert_output --partial "figlet"
  assert_output --partial "sl"
}
