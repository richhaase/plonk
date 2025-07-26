#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env

  # Clean up any packages/dotfiles from previous tests
  plonk uninstall brew:jq brew:tree npm:is-odd --force &>/dev/null || true
  plonk rm .plonk-test-rc .config/plonk-test .plonk-test-profile .plonk-test-bashrc --force &>/dev/null || true
  rm -rf "$HOME/.plonk-test-rc" "$HOME/.config/plonk-test" "$HOME/.plonk-test-profile" "$HOME/.plonk-test-bashrc" &>/dev/null || true
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

  # Verify both in status
  run plonk status
  assert_output_contains_all "cowsay" "is-odd"
}

@test "apply continues after package failure" {
  require_safe_package "brew:jq"
  require_safe_package "brew:tree"

  # Create config with valid and invalid packages
  create_test_config "packages:
  - jq
  - definitely-fake-xyz
  - tree"

  run plonk apply
  assert_failure  # Should fail due to partial failure

  # Verify output shows continuation
  assert_output --partial "✓"
  assert_output --partial "✗"
  assert_output --partial "jq"
  assert_output --partial "definitely-fake-xyz"
  assert_output --partial "tree"
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
  assert_output --partial "would"
  assert_output --partial "jq"

  # Verify nothing actually installed
  run plonk status
  refute_output --partial "jq"
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
  - jq
  invalid yaml here: ["

  run plonk apply
  assert_failure
  assert_output --partial "invalid"
  refute_output --partial "panic"
}

@test "apply with mixed packages and dotfiles including directories" {
  require_safe_package "brew:jq"
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
  - jq

dotfiles:
  - $testfile
  - $testdir"

  run plonk apply
  assert_success
  assert_output_contains_all "✓" "jq" "$testfile" "$testdir"

  track_artifact "package" "brew:jq"
  track_artifact "dotfile" "$testfile"
  track_artifact "dotfile" "$testdir"

  # Verify all in status
  run plonk status
  assert_output_contains_all "jq" "$testfile" "config.yaml" "settings.json"
}

@test "apply with partial failure in mixed resources" {
  require_safe_package "brew:jq"
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Create dotfile
  create_test_dotfile "$testfile"

  # Create config with mix of valid and invalid items
  create_test_config "packages:
  - jq
  - fake-package-xyz

dotfiles:
  - $testfile
  - /nonexistent/.fake-dotfile"

  run plonk apply
  assert_failure  # Should fail due to partial failures

  # Verify it processed all items
  assert_output_contains_all "jq" "fake-package-xyz" "$testfile" "fake-dotfile"
  assert_output --partial "failed"

  track_artifact "package" "brew:jq"
  track_artifact "dotfile" "$testfile"

  # Verify successful items were processed
  run plonk status
  assert_output_contains_all "jq" "$testfile"
}
