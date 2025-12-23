#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'
load '../lib/package_test_helper'

setup() {
  setup_test_env
}

# =============================================================================
# Uninstall spinner tests
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

  # Verify package was actually uninstalled
  run brew list sl
  assert_failure
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

  # Verify package was actually uninstalled
  run brew list figlet
  assert_failure
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

  # Verify packages were actually uninstalled
  run brew list cowsay
  assert_failure
  run brew list figlet
  assert_failure
}

# =============================================================================
# Apply spinner tests
# =============================================================================

@test "apply shows spinners for package installation" {
  require_safe_package "brew:cowsay"

  # Install and then manually remove package
  run plonk install brew:cowsay
  assert_success
  brew uninstall cowsay --force &>/dev/null || true

  # Apply should show spinner while reinstalling
  run plonk apply --packages
  assert_success
  assert_output --partial "Installing"
  assert_output --partial "✓"
  assert_output --partial "installed cowsay"

  # Verify package was actually installed
  run brew list cowsay
  assert_success

  track_artifact "package" "brew:cowsay"
}

@test "apply shows spinners for dotfile deployment" {
  local testfile=".plonk-test-spinner"
  require_safe_dotfile "$testfile"

  # Add dotfile and then remove it
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  rm -f "$HOME/$testfile"

  # Apply should show spinner while deploying
  run plonk apply --dotfiles
  assert_success
  assert_output --partial "Deploying"
  assert_output --partial "✓"
  assert_output --partial "deployed"

  # Verify dotfile was actually deployed
  run test -f "$HOME/$testfile"
  assert_success

  track_artifact "dotfile" "$testfile"
}

@test "apply shows progress indicators for multiple operations" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install packages to get them in lock file
  run plonk install brew:cowsay brew:figlet
  assert_success

  # Manually remove packages
  brew uninstall cowsay figlet --force &>/dev/null || true

  # Apply should show progress indicators
  run plonk apply --packages
  assert_success
  assert_output --partial "[1/2]"
  assert_output --partial "[2/2]"
  assert_output --partial "Installing"

  # Verify packages were actually installed
  run brew list cowsay
  assert_success
  run brew list figlet
  assert_success

  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"
}

@test "apply dry-run shows spinner without making changes" {
  require_safe_package "brew:sl"

  # Install and then manually remove package
  run plonk install brew:sl
  assert_success
  brew uninstall sl --force &>/dev/null || true

  # Dry-run apply should show spinner
  run plonk apply --dry-run
  assert_success
  assert_output --partial "Installing"
  assert_output --partial "✓"
  assert_output --partial "would-install"

  # Verify package was NOT actually installed
  run brew list sl
  assert_failure

  track_artifact "package" "brew:sl"
}

# =============================================================================
# Upgrade spinner tests
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

  # Verify package is still installed after upgrade
  run brew list cowsay
  assert_success
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

  # Verify packages are still installed after upgrade
  run brew list cowsay
  assert_success
  run brew list figlet
  assert_success
}
