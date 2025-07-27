#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/cleanup'
load '../lib/assertions'

@test "check for test artifacts before cleanup" {
  setup_test_env

  run check_for_test_artifacts
  # Exit code is number of artifacts found
  echo "# Found $status test artifacts" >&3

  if [[ $status -gt 0 ]]; then
    echo "# Test artifacts detected - cleanup may be needed" >&3
  else
    echo "# No test artifacts found - system is clean" >&3
  fi
}

@test "cleanup all test dotfiles" {
  setup_test_env

  if [[ "$PLONK_TEST_CLEANUP_DOTFILES" == "0" ]]; then
    skip "Dotfile cleanup disabled"
  fi

  run cleanup_all_test_dotfiles
  assert_success
}

@test "cleanup all test packages" {
  setup_test_env

  if [[ "$PLONK_TEST_CLEANUP_PACKAGES" != "1" ]]; then
    skip "Package cleanup not requested (set PLONK_TEST_CLEANUP_PACKAGES=1)"
  fi

  run cleanup_all_test_packages
  assert_success
}

@test "verify cleanup completed" {
  setup_test_env

  run check_for_test_artifacts
  if [[ $status -eq 0 ]]; then
    echo "# All test artifacts cleaned up successfully" >&3
  else
    echo "# Warning: $status test artifacts remain" >&3
    echo "# Run with PLONK_TEST_CLEANUP_PACKAGES=1 to remove packages" >&3
  fi
}
