#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Homebrew tests
@test "install single brew package" {
  require_safe_package "brew:cowsay"

  run plonk install brew:cowsay
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "added"

  track_artifact "package" "brew:cowsay"

  # Verify it's actually installed by brew
  run brew list cowsay
  assert_success

  # Verify it's in plonk lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "cowsay"

  # Verify it's in status
  run plonk status
  assert_output --partial "cowsay"
}

# Test multiple package installation - only testing with brew since
# the logic is the same for all managers (just loops over single installs)
@test "install multiple brew packages" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  run plonk install brew:figlet brew:sl
  assert_success
  assert_output_contains_all "figlet" "sl"

  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"

  # Verify they're actually installed by brew
  run brew list figlet
  assert_success
  run brew list sl
  assert_success

  # Verify both in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output_contains_all "figlet" "sl"

  # Verify both in status
  run plonk status
  assert_output_contains_all "figlet" "sl"
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

  # Verify it's actually installed by npm
  run npm list -g is-odd
  assert_success

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "is-odd"

  # Verify in status
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

  # Verify it's actually installed by pip
  run pip show six
  assert_success

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "six"

  # Verify in status
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

  # Verify it's actually installed by gem
  run gem list colorize
  assert_success
  assert_output --partial "colorize"

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "colorize"

  # Verify in status
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

  # Verify it's actually installed by go - check binary exists
  # Go installs to GOBIN if set, otherwise GOPATH/bin
  local gobin="$(go env GOBIN)"
  if [[ -z "$gobin" ]]; then
    gobin="$(go env GOPATH)/bin"
  fi
  run test -f "$gobin/hey"
  assert_success

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "hey"

  # Verify in status
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

  # Verify it's actually installed by cargo
  run cargo install --list
  assert_success
  assert_output --partial "ripgrep"

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "ripgrep"

  # Verify in status
  run plonk status
  assert_output --partial "ripgrep"
}

# General installation behavior tests
@test "install with dry-run doesn't actually install" {
  require_safe_package "brew:fortune"

  run plonk install brew:fortune --dry-run
  assert_success
  assert_output --partial "would"

  # Verify not actually installed by brew
  run brew list fortune
  assert_failure

  # Verify not in lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "fortune"
  fi

  # Verify not in status
  run plonk status
  refute_output --partial "fortune"
}

@test "install already managed package shows appropriate message" {
  require_safe_package "brew:cowsay"

  # First install
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Verify it's actually installed
  run brew list cowsay
  assert_success

  # Try to install again - should skip but still succeed
  run plonk install brew:cowsay
  assert_success
  assert_output --partial "skipped"
}

@test "install already-installed but unmanaged package adds to lock file" {
  require_safe_package "brew:fortune"

  # Ensure it's not already managed by removing from lock if present
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    grep -v "name: fortune" "$PLONK_DIR/plonk.lock" > "$PLONK_DIR/plonk.lock.tmp" || true
    mv "$PLONK_DIR/plonk.lock.tmp" "$PLONK_DIR/plonk.lock"
  fi

  # Install directly with brew (not via plonk)
  run brew install fortune

  # Verify it's not in plonk's management
  run plonk status
  refute_output --partial "fortune"

  # Install via plonk - should succeed and add to lock
  run plonk install brew:fortune
  assert_success
  assert_output --partial "added"

  # Verify it's now managed
  run plonk status
  assert_output --partial "fortune"

  track_artifact "package" "brew:fortune"
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
