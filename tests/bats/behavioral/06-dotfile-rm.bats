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
  refute_output --partial "$testfile"
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
  assert_output --partial "Would remove"

  # Verify still managed
  run plonk status
  assert_output --partial "$testfile"
}

@test "remove non-managed dotfile shows error" {
  run plonk rm ".not-managed-file"
  assert_success  # Returns success but skips
  assert_output --partial "Skipped"
}

@test "remove directory removes all nested files" {
  local testdir=".config/plonk-test"
  require_safe_dotfile "$testdir/config.yaml"

  # Create nested structure
  mkdir -p "$HOME/$testdir/subdir"
  echo "test: true" > "$HOME/$testdir/config.yaml"
  echo "# Nested" > "$HOME/$testdir/subdir/nested.conf"

  # Add directory
  run plonk add "$HOME/$testdir"
  assert_success
  track_artifact "dotfile" "$testdir"

  # Remove the directory - should succeed and rm -rf from plonk config
  run plonk rm "$testdir"
  assert_success  # BUG: Currently fails with "directory not empty"
  assert_output --partial "Removed"

  # Verify all files gone from status
  run plonk status
  refute_output --partial "config.yaml"
  refute_output --partial "nested.conf"
}

@test "remove specific file from managed directory" {
  local testdir=".config/plonk-test"
  require_safe_dotfile "$testdir/config.yaml"

  # Create multiple files
  mkdir -p "$HOME/$testdir"
  echo "test: true" > "$HOME/$testdir/config.yaml"
  echo "# Settings" > "$HOME/$testdir/settings.json"

  # Add directory
  run plonk add "$HOME/$testdir"
  assert_success
  track_artifact "dotfile" "$testdir"

  # Remove just one file
  run plonk rm "$testdir/config.yaml"
  assert_success
  assert_output --partial "Removed"
  assert_output --partial "config.yaml"

  # Verify only that file is gone
  run plonk status
  refute_output --partial "config.yaml"
  assert_output --partial "settings.json"
}

# Test multiple file removal - only testing one case since
# the logic is the same as packages (loops over single removes)
@test "remove multiple dotfiles at once" {
  local file1=".plonk-test-rc"
  local file2=".plonk-test-profile"
  require_safe_dotfile "$file1"
  require_safe_dotfile "$file2"

  # Add both files first
  create_test_dotfile "$file1"
  create_test_dotfile "$file2"
  run plonk add "$HOME/$file1" "$HOME/$file2"
  assert_success

  # Remove both
  run plonk rm "$file1" "$file2"
  assert_success
  assert_output_contains_all "$file1" "$file2" "✓"

  # Verify both gone from status
  run plonk status
  refute_output --partial "$file1"
  refute_output --partial "$file2"
}
