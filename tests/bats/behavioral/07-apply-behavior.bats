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

@test "apply detects unmanaged packages as needing installation" {
  # This test demonstrates the system binary detection bug
  # plonk apply incorrectly shows "All up to date" for packages not managed by plonk

  require_safe_package "brew:cowsay"

  # Ensure package is NOT installed via brew
  run brew list cowsay
  if [[ $status -eq 0 ]]; then
    brew uninstall cowsay --force || true
  fi

  # Create config requesting the package
  create_test_config "packages:
  - cowsay"

  # Apply should install the package
  run plonk apply
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "✓"
  refute_output --partial "All up to date"

  track_artifact "package" "brew:cowsay"

  # Verify package was actually installed
  run brew list cowsay
  assert_success
}

@test "apply with packages from multiple managers" {
  require_safe_package "brew:cowsay"
  require_safe_package "npm:is-odd"

  # Check if npm is available
  run which npm
  if [[ $status -ne 0 ]]; then
    skip "npm not available"
  fi

  # Create config with packages from different managers
  create_test_config "packages:
  - cowsay     # brew
  - is-odd     # npm"

  run plonk apply
  assert_success
  assert_output_contains_all "✓" "cowsay" "is-odd"

  track_artifact "package" "brew:cowsay"
  track_artifact "package" "npm:is-odd"

  # Verify packages are actually installed
  run brew list cowsay
  assert_success
  run npm list -g is-odd
  assert_success

  # Verify in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "cowsay" "is-odd"

  # Verify both in status
  run plonk status
  assert_output_contains_all "cowsay" "is-odd"
}

@test "apply continues after package failure" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  # Create config with valid and invalid packages
  create_test_config "packages:
  - figlet
  - definitely-fake-xyz
  - sl"

  run plonk apply
  assert_failure  # Should fail due to partial failure

  # Verify output shows continuation
  assert_output --partial "✓"
  assert_output --partial "✗"
  assert_output --partial "figlet"
  assert_output --partial "definitely-fake-xyz"
  assert_output --partial "sl"
  assert_output --partial "2 succeeded, 1 failed"

  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"

  # Verify successful packages were actually installed
  run brew list figlet
  assert_success
  run brew list sl
  assert_success

  # Verify successful packages in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "figlet" "sl"
  refute_output --partial "definitely-fake-xyz"

  # Verify successful packages were installed
  run plonk status
  assert_output_contains_all "figlet" "sl"
}

@test "apply with dry-run shows what would happen" {
  require_safe_package "brew:fortune"

  create_test_config "packages:
  - fortune"

  run plonk apply --dry-run
  assert_success
  assert_output --partial "would"
  assert_output --partial "fortune"

  # Verify nothing actually installed
  run brew list fortune
  assert_failure

  # Verify no lock file created
  run test -f "$PLONK_DIR/plonk.lock"
  assert_failure

  # Verify nothing in status
  run plonk status
  refute_output --partial "fortune"
}

@test "apply with empty config succeeds" {
  create_test_config ""

  run plonk apply
  assert_success
  assert_output --partial "up to date"
}

@test "apply with invalid YAML shows error" {
  # Create invalid YAML
  create_test_config "packages:
  - cowsay
  invalid yaml here: ["

  run plonk apply
  assert_failure
  assert_output --partial "invalid"
  refute_output --partial "panic"
}

@test "apply with mixed packages and dotfiles including directories" {
  require_safe_package "brew:cowsay"
  local testfile=".plonk-test-rc"
  local testdir=".config/plonk-test"
  require_safe_dotfile "$testfile"
  require_safe_dotfile "$testdir/config.yaml"

  # Create regular dotfile
  create_test_dotfile "$testfile" "# Test RC file"

  # Create directory with nested files
  mkdir -p "$HOME/$testdir"
  echo "test: true" > "$HOME/$testdir/config.yaml"
  echo "nested: true" > "$HOME/$testdir/settings.json"

  # Create config with both
  create_test_config "packages:
  - cowsay

dotfiles:
  - $testfile
  - $testdir"

  run plonk apply
  assert_success
  assert_output_contains_all "✓" "cowsay" "$testfile" "$testdir"

  track_artifact "package" "brew:cowsay"
  track_artifact "dotfile" "$testfile"
  track_artifact "dotfile" "$testdir"

  # Verify package actually installed
  run brew list cowsay
  assert_success

  # Verify dotfiles actually exist in plonk dir
  run test -f "$PLONK_DIR/$(basename $testfile)"
  assert_success
  run test -d "$PLONK_DIR/$(basename $testdir)"
  assert_success

  # Verify all in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "cowsay" "$testfile" "config.yaml"

  # Verify all in status
  run plonk status
  assert_output_contains_all "cowsay" "$testfile" "config.yaml" "settings.json"
}

@test "apply with partial failure in mixed resources" {
  require_safe_package "brew:figlet"
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Create dotfile
  create_test_dotfile "$testfile"

  # Create config with mix of valid and invalid items
  create_test_config "packages:
  - figlet
  - fake-package-xyz

dotfiles:
  - $testfile
  - /nonexistent/.fake-dotfile"

  run plonk apply
  assert_failure  # Should fail due to partial failures

  # Verify it processed all items
  assert_output_contains_all "figlet" "fake-package-xyz" "$testfile" "fake-dotfile"
  assert_output --partial "failed"

  track_artifact "package" "brew:figlet"
  track_artifact "dotfile" "$testfile"

  # Verify successful package was actually installed
  run brew list figlet
  assert_success

  # Verify successful dotfile exists in plonk dir
  run test -f "$PLONK_DIR/$(basename $testfile)"
  assert_success

  # Verify successful items were processed
  run plonk status
  assert_output_contains_all "figlet" "$testfile"
}
