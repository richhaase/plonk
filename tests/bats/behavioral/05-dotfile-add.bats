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
  refute_output --partial "$testfile"
}

@test "add non-existent dotfile shows error" {
  run plonk add "$HOME/.definitely-does-not-exist-xyz"
  assert_failure
  assert_output --partial "not found"
}

@test "add directory with nested structure" {
  local testdir=".config/plonk-test"
  require_safe_dotfile "$testdir/config.yaml"

  # Create nested directory structure
  mkdir -p "$HOME/$testdir/subdir/deep"
  echo "test: true" > "$HOME/$testdir/config.yaml"
  echo "# Test settings" > "$HOME/$testdir/settings.json"
  echo "# Nested file" > "$HOME/$testdir/subdir/nested.conf"
  echo "# Deep file" > "$HOME/$testdir/subdir/deep/deeply.nested"

  track_artifact "dotfile" "$testdir"

  # Add directory
  run plonk add "$HOME/$testdir"
  assert_success
  assert_output --partial "Added"

  # Verify files tracked including nested ones
  run plonk status
  assert_output --partial "config.yaml"
  assert_output --partial "nested.conf"
  assert_output --partial "deeply.nested"
}

@test "add dotfile that conflicts with existing file" {
  local testfile=".plonk-test-gitconfig"
  require_safe_dotfile "$testfile"

  # Create existing file with different content
  echo "# Original content" > "$HOME/$testfile"

  # Add to plonk (plonk should win)
  run plonk add "$HOME/$testfile"
  assert_success
  assert_output --partial "Added"

  track_artifact "dotfile" "$testfile"

  # Verify plonk is now managing it
  run plonk status
  assert_output --partial "$testfile"
}

@test "add multiple dotfiles at once" {
  local file1=".plonk-test-rc"
  local file2=".plonk-test-profile"
  require_safe_dotfile "$file1"
  require_safe_dotfile "$file2"

  create_test_dotfile "$file1" "# File 1"
  create_test_dotfile "$file2" "# File 2"

  # Add both files
  run plonk add "$HOME/$file1" "$HOME/$file2"
  assert_success
  assert_output_contains_all "$file1" "$file2" "Added"

  # Verify both in status
  run plonk status
  assert_output_contains_all "$file1" "$file2"
}

@test "add already managed dotfile shows appropriate message" {
  local testfile=".plonk-test-bashrc"
  require_safe_dotfile "$testfile"

  create_test_dotfile "$testfile"

  # First add
  run plonk add "$HOME/$testfile"
  assert_success

  # Try to add again
  run plonk add "$HOME/$testfile"
  assert_success
  assert_output --partial "already"
}
