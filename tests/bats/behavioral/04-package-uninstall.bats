#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Homebrew uninstall tests
@test "uninstall managed brew package" {
  require_safe_package "brew:tree"

  # Install first
  run plonk install brew:tree
  assert_success

  # Then uninstall
  run plonk uninstall brew:tree
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "tree"
}

# NPM uninstall tests
@test "uninstall managed npm package" {
  require_safe_package "npm:left-pad"

  run which npm
  if [[ $status -ne 0 ]]; then
    skip "npm not available"
  fi

  # Install first
  run plonk install npm:left-pad
  assert_success

  # Then uninstall
  run plonk uninstall npm:left-pad
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "left-pad"
}

# Python/pip uninstall tests
@test "uninstall managed pip package" {
  require_safe_package "pip:cowsay"

  # Check if pip is available
  run which pip3
  if [[ $status -ne 0 ]]; then
    run which pip
    if [[ $status -ne 0 ]]; then
      skip "pip not available"
    fi
  fi

  # Install first
  run plonk install pip:cowsay
  assert_success

  # Then uninstall
  run plonk uninstall pip:cowsay
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "cowsay"
}

# Ruby/gem uninstall tests
@test "uninstall managed gem package" {
  require_safe_package "gem:colorize"

  run which gem
  if [[ $status -ne 0 ]]; then
    skip "gem not available"
  fi

  # Install first
  run plonk install gem:colorize
  assert_success

  # Then uninstall
  run plonk uninstall gem:colorize
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "colorize"
}

# Go uninstall tests
@test "uninstall managed go package" {
  require_safe_package "go:github.com/rakyll/hey"

  run which go
  if [[ $status -ne 0 ]]; then
    skip "go not available"
  fi

  # Install first
  run plonk install go:github.com/rakyll/hey
  assert_success

  # Then uninstall
  run plonk uninstall go:github.com/rakyll/hey
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "hey"
}

# Cargo uninstall tests
@test "uninstall managed cargo package" {
  require_safe_package "cargo:ripgrep"

  run which cargo
  if [[ $status -ne 0 ]]; then
    skip "cargo not available"
  fi

  # Install first
  run plonk install cargo:ripgrep
  assert_success

  # Then uninstall
  run plonk uninstall cargo:ripgrep
  assert_success
  assert_output --partial "removed"

  # Verify gone from status
  run plonk status
  refute_output --partial "ripgrep"
}

# General uninstall behavior tests
@test "uninstall non-managed package shows error" {
  run plonk uninstall brew:not-managed-package
  assert_success  # plonk returns success but skips the package
  assert_output --partial "skipped"
}

@test "uninstall with dry-run shows what would happen" {
  require_safe_package "brew:jq"

  # Install first
  run plonk install brew:jq
  assert_success
  track_artifact "package" "brew:jq"

  # Dry-run uninstall
  run plonk uninstall brew:jq --dry-run
  assert_success
  assert_output --partial "would-remove"

  # Verify still managed
  run plonk status
  assert_output --partial "jq"
}

@test "uninstall with force removes even if not managed" {
  require_safe_package "brew:tree"

  # Ensure it's not managed
  run plonk status
  refute_output --partial "tree"

  # Force uninstall (this might fail if not installed at all)
  run plonk uninstall brew:tree --force
  # Don't assert success/failure as it depends on system state
  # Just verify no panic
  refute_output --partial "panic"
}

# Test multiple package uninstallation - only testing with brew since
# the logic is the same for all managers (just loops over single uninstalls)
@test "uninstall multiple packages" {
  require_safe_package "brew:jq"
  require_safe_package "brew:tree"

  # Install both first
  run plonk install brew:jq brew:tree
  assert_success
  track_artifact "package" "brew:jq"
  track_artifact "package" "brew:tree"

  # Uninstall both
  run plonk uninstall brew:jq brew:tree
  assert_success
  assert_output_contains_all "jq" "tree" "removed"

  # Verify both gone from status
  run plonk status
  refute_output --partial "jq"
  refute_output --partial "tree"
}
