#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# =============================================================================
# Install dry-run tests
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

# =============================================================================
# Uninstall dry-run tests
# =============================================================================

# Note: Uninstall dry-run spinner behavior is tested in 11-spinner-and-progress.bats

# =============================================================================
# Dotfile add dry-run tests
# =============================================================================

@test "add dotfile with dry-run" {
  local testfile=".plonk-test-profile"
  require_safe_dotfile "$testfile"

  create_test_dotfile "$testfile"

  run plonk add "$HOME/$testfile" --dry-run
  assert_success
  assert_output --partial "Would add"

  # Verify not actually added
  run plonk status
  refute_output --partial "$testfile"
}

# =============================================================================
# Dotfile rm dry-run tests
# =============================================================================

@test "rm with dry-run shows what would happen" {
  local testfile=".plonk-test-bashrc"
  require_safe_dotfile "$testfile"

  # Add first
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success

  # Dry-run remove
  run plonk rm "$testfile" --dry-run
  assert_success
  assert_output --partial "Would remove"

  # Verify still managed
  run plonk status
  assert_output --partial "$testfile"
}

# =============================================================================
# Apply dry-run tests
# =============================================================================

@test "apply with dry-run shows what would happen" {
  require_safe_package "brew:fortune"

  # Install package to lock file
  run plonk install brew:fortune
  assert_success

  # Manually uninstall
  brew uninstall fortune --force &>/dev/null || true

  # Run apply with dry-run
  run plonk apply --dry-run
  assert_success
  assert_output --partial "would"
  assert_output --partial "fortune"

  # Verify package was NOT actually installed
  run brew list fortune
  assert_failure

  track_artifact "package" "brew:fortune"
}

# =============================================================================
# Upgrade dry-run tests
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
