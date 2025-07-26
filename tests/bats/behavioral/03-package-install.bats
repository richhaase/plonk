#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Homebrew tests
@test "install single brew package" {
  require_safe_package "brew:jq"

  run plonk install brew:jq
  assert_success
  assert_output --partial "jq"
  assert_output --partial "added"

  track_artifact "package" "brew:jq"

  # Verify it's in status
  run plonk status
  assert_output --partial "jq"
}

# Test multiple package installation - only testing with brew since
# the logic is the same for all managers (just loops over single installs)
@test "install multiple brew packages" {
  require_safe_package "brew:jq"
  require_safe_package "brew:tree"

  run plonk install brew:jq brew:tree
  assert_success
  assert_output_contains_all "jq" "tree"

  track_artifact "package" "brew:jq"
  track_artifact "package" "brew:tree"

  # Verify both in status
  run plonk status
  assert_output_contains_all "jq" "tree"
}

# NPM tests
@test "install npm package" {
  require_safe_package "npm:is-odd"

  # Check if npm is available first
  run which npm
  if [[ $status -ne 0 ]]; then
    skip "npm not available"
  fi

  run plonk install npm:is-odd
  assert_success
  assert_output --partial "is-odd"
  assert_output --partial "added"

  track_artifact "package" "npm:is-odd"

  run plonk status
  assert_output --partial "is-odd"
}


# Python/pip tests
@test "install pip package" {
  require_safe_package "pip:six"

  # Check if pip is available
  run which pip3
  if [[ $status -ne 0 ]]; then
    run which pip
    if [[ $status -ne 0 ]]; then
      skip "pip not available"
    fi
  fi

  run plonk install pip:six
  assert_success
  assert_output --partial "six"
  assert_output --partial "added"

  track_artifact "package" "pip:six"

  run plonk status
  assert_output --partial "six"
}

# Ruby/gem tests
@test "install gem package" {
  require_safe_package "gem:colorize"

  # Check if gem is available
  run which gem
  if [[ $status -ne 0 ]]; then
    skip "gem not available"
  fi

  run plonk install gem:colorize
  assert_success
  assert_output --partial "colorize"
  assert_output --partial "added"

  track_artifact "package" "gem:colorize"

  run plonk status
  assert_output --partial "colorize"
}

# Go tests
@test "install go package" {
  require_safe_package "go:github.com/rakyll/hey"

  # Check if go is available
  run which go
  if [[ $status -ne 0 ]]; then
    skip "go not available"
  fi

  run plonk install go:github.com/rakyll/hey
  assert_success
  assert_output --partial "hey"
  assert_output --partial "added"

  track_artifact "package" "go:github.com/rakyll/hey"

  run plonk status
  assert_output --partial "hey"
}

# Cargo tests
@test "install cargo package" {
  require_safe_package "cargo:ripgrep"

  # Check if cargo is available
  run which cargo
  if [[ $status -ne 0 ]]; then
    skip "cargo not available"
  fi

  run plonk install cargo:ripgrep
  assert_success
  assert_output --partial "ripgrep"
  assert_output --partial "added"

  track_artifact "package" "cargo:ripgrep"

  run plonk status
  assert_output --partial "ripgrep"
}

# General installation behavior tests
@test "install with dry-run doesn't actually install" {
  require_safe_package "brew:jq"

  run plonk install brew:jq --dry-run
  assert_success
  assert_output --partial "would"

  # Verify not actually installed
  run plonk status
  refute_output --partial "jq"
}

@test "install already managed package shows appropriate message" {
  require_safe_package "brew:jq"

  # First install
  run plonk install brew:jq
  assert_success
  track_artifact "package" "brew:jq"

  # Try to install again
  run plonk install brew:jq
  assert_success
  assert_output --partial "skipped"
}

@test "install shows error for non-existent package" {
  run plonk install brew:definitely-not-real-xyz123
  assert_failure
  assert_output --partial "failed"
  refute_output --partial "panic"
}

@test "install with invalid manager shows error" {
  run plonk install fake-manager:package
  assert_failure
  assert_output --partial "failed"
  refute_output --partial "panic"
}
