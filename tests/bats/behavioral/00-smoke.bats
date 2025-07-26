#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

@test "plonk binary can be built and is executable" {
  setup_test_env

  # Verify test binary exists
  run which plonk
  assert_success

  # Verify it's executable
  run test -x "$(which plonk)"
  assert_success

  # Verify it's our test binary, not system plonk
  run which plonk
  assert_output --partial "$BATS_TEST_TMPDIR"
}

@test "test environment setup works correctly" {
  setup_test_env

  # Verify environment variables
  assert [ -n "$PLONK_DIR" ]
  assert [ -d "$PLONK_DIR" ]
  assert [ -n "$CLEANUP_FILE" ]
  assert [ -f "$CLEANUP_FILE" ]
}

@test "safe lists load correctly" {
  setup_test_env

  # Verify safe packages loaded
  assert [ ${#SAFE_PACKAGES[@]} -gt 0 ]

  # Verify safe dotfiles loaded
  assert [ ${#SAFE_DOTFILES[@]} -gt 0 ]


  # Test package checking (run creates subshell, so test directly)
  if is_safe_package "brew:jq"; then
    assert true
  else
    assert false "brew:jq should be in safe list"
  fi

  if is_safe_package "brew:definitely-not-safe"; then
    assert false "brew:definitely-not-safe should NOT be in safe list"
  else
    assert true
  fi
}

@test "cleanup tracking works" {
  setup_test_env

  # Track some artifacts
  track_artifact "dotfile" ".test-file"
  track_artifact "package" "brew:test"

  # Verify they were tracked
  run grep "dotfile:.test-file" "$CLEANUP_FILE"
  assert_success

  run grep "package:brew:test" "$CLEANUP_FILE"
  assert_success
}

@test "plonk help command works" {
  run plonk help
  assert_success
  assert_output --partial "Usage:"
  assert_output --partial "Available Commands:"
}

@test "plonk version shows version info" {
  run plonk --version
  assert_success
  assert_output --partial "plonk"
}

@test "BATS can run tests with proper isolation" {
  setup_test_env

  # Verify we're using isolated config
  run echo "$PLONK_DIR"
  assert_output --partial "/plonk-config"

  # Verify it's not the user's real config
  refute_output --partial "$HOME/.config/plonk"
}
