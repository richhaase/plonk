#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Basic clone functionality tests

@test "clone command shows error when no arguments given" {
  run plonk clone
  assert_failure
  assert_output --partial "accepts 1 arg"
}

@test "clone with invalid git URL shows error" {
  run plonk clone "not-a-valid-url"
  assert_failure
  assert_output --partial "unsupported git URL format"
}

@test "clone with empty repo argument shows error" {
  run plonk clone ""
  assert_failure
}

# NOTE: Clone command only accepts remote URLs (https://, git@, git://, user/repo shorthand)
# Local file:// URLs are not supported, so we test with dry-run and remote URL formats

@test "clone from remote repo would work (dry-run)" {
  # Use dry-run to test clone parsing without actually cloning
  run plonk clone --dry-run "testuser/dotfiles"
  assert_success
  assert_output --partial "Dry run"
  assert_output --partial "https://github.com/testuser/dotfiles.git"
}

@test "clone with existing PLONK_DIR shows warning and skips" {
  # Create the PLONK_DIR directory first
  mkdir -p "$PLONK_DIR"
  echo "existing" > "$PLONK_DIR/existing-file"

  # Clone should warn about existing dir (use dry-run to avoid network)
  run plonk clone --dry-run "testuser/dotfiles"
  assert_success
  assert_output --partial "already exists"

  # Original file should still be there
  run cat "$PLONK_DIR/existing-file"
  assert_success
  assert_output "existing"
}

# Dry-run tests

@test "clone --dry-run shows what would happen without making changes" {
  # Remove PLONK_DIR so we test the fresh clone case
  rm -rf "$PLONK_DIR"

  run plonk clone --dry-run "testuser/dotfiles"
  assert_success
  assert_output --partial "Dry run"
  assert_output --partial "would clone"

  # Verify PLONK_DIR was not created by dry-run
  run test -d "$PLONK_DIR"
  assert_failure
}

@test "clone -n is alias for --dry-run" {
  run plonk clone -n "testuser/dotfiles"
  assert_success
  assert_output --partial "Dry run"
}

@test "clone --dry-run with existing PLONK_DIR shows skip message" {
  # Create the PLONK_DIR directory first
  mkdir -p "$PLONK_DIR"

  run plonk clone --dry-run "testuser/dotfiles"
  assert_success
  assert_output --partial "Dry run"
  assert_output --partial "already exists"
  assert_output --partial "would skip clone"
}

# URL format tests

@test "clone parses user/repo GitHub shorthand" {
  # This will fail because it's not a real repo, but should parse the URL correctly
  run plonk clone --dry-run "testuser/testrepo"
  assert_success
  assert_output --partial "https://github.com/testuser/testrepo.git"
}

@test "clone accepts HTTPS URLs" {
  run plonk clone --dry-run "https://github.com/example/dotfiles"
  assert_success
  assert_output --partial "https://github.com/example/dotfiles.git"
}

@test "clone accepts SSH URLs" {
  run plonk clone --dry-run "git@github.com:example/dotfiles.git"
  assert_success
  assert_output --partial "git@github.com:example/dotfiles.git"
}

# Error handling tests

@test "clone with non-existent repo format shows error" {
  # Local paths are not supported - shows URL format error
  run plonk clone "/nonexistent/path/to/repo.git"
  assert_failure
  assert_output --partial "unsupported git URL format"
}

@test "clone with malformed URL shows specific error" {
  run plonk clone ":::invalid:::"
  assert_failure
  assert_output --partial "unsupported git URL format"
}

# NOTE: Package manager detection tests require actually cloning a repo
# which needs network access. These would be integration tests.
# The dry-run tests above verify the clone command parsing works correctly.
