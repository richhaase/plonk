#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env

  # Clean up any packages/dotfiles from previous tests
  plonk uninstall brew:cowsay brew:figlet brew:sl brew:fortune npm:is-odd npm:left-pad --force &>/dev/null || true
  plonk rm .plonk-test-rc .config/plonk-test .plonk-test-profile .plonk-test-bashrc --force &>/dev/null || true
  rm -rf "$HOME/.plonk-test-rc" "$HOME/.config/plonk-test" "$HOME/.plonk-test-profile" "$HOME/.plonk-test-bashrc" &>/dev/null || true
}

@test "apply reinstalls packages from lock file" {
  require_safe_package "brew:cowsay"

  # First, install a package to get it in the lock file
  run plonk install brew:cowsay
  assert_success

  # Verify it's in the lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "cowsay"

  # Manually uninstall the package (simulating a fresh system)
  brew uninstall cowsay --force &>/dev/null || true

  # Verify package is NOT installed
  run brew list cowsay
  assert_failure

  # Now run apply - it should reinstall from lock file
  run plonk apply
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "✓"

  # Verify package was reinstalled
  run brew list cowsay
  assert_success

  track_artifact "package" "brew:cowsay"
}

@test "apply with packages from multiple managers" {
  require_safe_package "brew:figlet"
  require_safe_package "npm:is-odd"

  # Check if npm is available
  run which npm
  if [[ $status -ne 0 ]]; then
    skip "npm not available"
  fi

  # Install packages to get them in lock file
  run plonk install brew:figlet
  assert_success
  run plonk install npm:is-odd
  assert_success

  # Verify both in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "figlet" "is-odd"

  # Manually uninstall packages
  brew uninstall figlet --force &>/dev/null || true
  npm uninstall -g is-odd &>/dev/null || true

  # Run apply
  run plonk apply
  assert_success
  assert_output_contains_all "✓" "figlet" "is-odd"

  # Verify packages were reinstalled
  run brew list figlet
  assert_success
  run npm list -g is-odd
  assert_success

  track_artifact "package" "brew:figlet"
  track_artifact "package" "npm:is-odd"
}

@test "apply reports when packages are already installed" {
  require_safe_package "brew:sl"

  # Install package
  run plonk install brew:sl
  assert_success

  # Run apply when package is already installed
  run plonk apply
  assert_success
  assert_output --partial "All up to date"

  track_artifact "package" "brew:sl"
}

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

@test "apply with empty lock file succeeds" {
  # Ensure lock file is empty or doesn't exist
  rm -f "$PLONK_DIR/plonk.lock"

  run plonk apply
  assert_success
  assert_output --partial "up to date"
}

@test "apply with mixed packages and dotfiles" {
  require_safe_package "brew:cowsay"
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Add package and dotfile
  run plonk install brew:cowsay
  assert_success

  create_test_dotfile "$testfile" "# Test RC file"
  run plonk add "$HOME/$testfile"
  assert_success

  # Verify dotfile is in plonk config (without leading dot)
  local stored_name="${testfile#.}"  # Remove leading dot
  run test -f "$PLONK_DIR/$stored_name"
  assert_success

  # Manually remove both from the system
  brew uninstall cowsay --force &>/dev/null || true
  rm -f "$HOME/$testfile"

  # Run apply
  run plonk apply
  assert_success
  assert_output_contains_all "✓" "cowsay" "$testfile"

  # Verify both were restored
  run brew list cowsay
  assert_success

  # This fails due to a bug in plonk - dotfile apply reports success but doesn't restore the file
  run test -f "$HOME/$testfile"
  assert_success

  track_artifact "package" "brew:cowsay"
  track_artifact "dotfile" "$testfile"
}

@test "apply with --packages flag only applies packages" {
  require_safe_package "brew:figlet"
  local testfile=".plonk-test-profile"
  require_safe_dotfile "$testfile"

  # Add package and dotfile
  run plonk install brew:figlet
  assert_success

  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success

  # Remove both
  brew uninstall figlet --force &>/dev/null || true
  rm -f "$HOME/$testfile"

  # Apply packages only
  run plonk apply --packages
  assert_success
  assert_output --partial "figlet"
  refute_output --partial "$testfile"

  # Verify package was installed but dotfile was not
  run brew list figlet
  assert_success
  run test -f "$HOME/$testfile"
  assert_failure

  track_artifact "package" "brew:figlet"
  track_artifact "dotfile" "$testfile"
}

@test "apply with --dotfiles flag only applies dotfiles" {
  require_safe_package "brew:sl"
  local testfile=".plonk-test-bashrc"
  require_safe_dotfile "$testfile"

  # Add package and dotfile
  run plonk install brew:sl
  assert_success

  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success

  # Remove both
  brew uninstall sl --force &>/dev/null || true
  rm -f "$HOME/$testfile"

  # Apply dotfiles only
  run plonk apply --dotfiles
  assert_success
  assert_output --partial "$testfile"
  refute_output --partial "sl"

  # Verify dotfile was created but package was not installed
  run test -f "$HOME/$testfile"
  assert_success
  run brew list sl
  assert_failure

  track_artifact "package" "brew:sl"
  track_artifact "dotfile" "$testfile"
}

# Spinner tests for apply command
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
