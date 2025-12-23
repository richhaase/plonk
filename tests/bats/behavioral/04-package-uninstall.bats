#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'
load '../lib/package_test_helper'

setup() {
  setup_test_env
}

# =============================================================================
# Per-manager uninstall tests
# =============================================================================

@test "uninstall managed brew package" {
  test_uninstall_managed "brew" "sl"
}

@test "uninstall managed npm package" {
  test_uninstall_managed "npm" "left-pad"
}

@test "uninstall managed gem package" {
  test_uninstall_managed "gem" "colorize"
}

@test "uninstall managed cargo package" {
  test_uninstall_managed "cargo" "ripgrep"
}

@test "uninstall managed uv package" {
  test_uninstall_managed "uv" "rich-cli"
}

@test "uninstall managed pnpm package" {
  test_uninstall_managed "pnpm" "prettier"
}

# =============================================================================
# General uninstall behavior tests
# =============================================================================

@test "uninstall non-managed package acts as pass-through" {
  require_safe_package "brew:fortune"

  # Install directly with brew (not via plonk)
  run brew install fortune
  assert_success

  # Verify it's not managed by plonk
  run plonk status
  refute_output --partial "fortune"

  # Uninstall via plonk - should pass through to brew
  run plonk uninstall brew:fortune
  assert_success
  assert_output --partial "removed"

  # Verify it was actually uninstalled
  run brew list fortune
  assert_failure
}

# Test multiple package uninstallation - only testing with brew since
# the logic is the same for all managers (just loops over single uninstalls)
@test "uninstall multiple packages" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Uninstall both
  run plonk uninstall brew:cowsay brew:figlet
  assert_success
  assert_output_contains_all "cowsay" "figlet" "removed"

  # Verify both gone from status
  run plonk status
  refute_output --partial "cowsay"
  refute_output --partial "figlet"
}

@test "uninstall without prefix uses manager from lock file" {
  require_safe_package "brew:sl"

  # Install with brew
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Set default manager to something else
  cat > "$PLONK_DIR/plonk.yaml" <<EOF
default_manager: npm
EOF

  # Uninstall without prefix - should use brew from lock file
  run plonk uninstall sl
  assert_success
  assert_output --partial "removed"

  # Verify it was uninstalled with brew (not npm)
  run brew list sl
  assert_failure
}

@test "uninstall succeeds when package removed from lock even if system uninstall fails" {
  require_safe_package "brew:figlet"

  # Install package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Manually uninstall from system
  run brew uninstall figlet --force

  # Now plonk uninstall should still succeed (removes from lock)
  run plonk uninstall brew:figlet
  assert_success
  assert_output --partial "removed"

  # Verify it's gone from lock file
  run grep "name: figlet" "$PLONK_DIR/plonk.lock"
  assert_failure
}

# =============================================================================
# Spinner tests for uninstall command
# =============================================================================

@test "uninstall shows spinner during operation" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Uninstall and check for spinner output
  run plonk uninstall brew:sl
  assert_success
  assert_output --partial "Uninstalling"
  assert_output --partial "removed"
}

@test "uninstall shows completion message after spinner" {
  require_safe_package "brew:figlet"

  # Install first
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Uninstall and verify completion message
  run plonk uninstall brew:figlet
  assert_success
  assert_output --partial "Uninstalling: brew:figlet"
}

@test "uninstall shows error message when removal fails" {
  # Try to uninstall a non-existent package
  run plonk uninstall brew:nonexistentpackage123456
  assert_failure
  assert_output --partial "Failed to uninstall"
}

@test "uninstall in dry-run mode shows spinner without actual removal" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Dry-run uninstall
  run plonk uninstall --dry-run brew:sl
  assert_success
  assert_output --partial "Uninstalling"
  assert_output --partial "would-remove"

  # Verify package is still installed
  run brew list sl
  assert_success
}

@test "uninstall multiple packages shows progress indicators" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Uninstall both and check for progress indicators
  run plonk uninstall brew:cowsay brew:figlet
  assert_success
  assert_output --partial "[1/2]"
  assert_output --partial "[2/2]"
  assert_output --partial "Uninstalling"
}
