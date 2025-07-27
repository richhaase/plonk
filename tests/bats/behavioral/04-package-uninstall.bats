#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Homebrew uninstall tests
@test "uninstall managed brew package" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Verify it's installed
  run brew list sl
  assert_success

  # Then uninstall
  run plonk uninstall brew:sl
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by brew
  run brew list sl
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "sl"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "sl"
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
  track_artifact "package" "npm:left-pad"

  # Verify it's installed
  run npm list -g left-pad
  assert_success

  # Then uninstall
  run plonk uninstall npm:left-pad
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by npm
  run npm list -g left-pad
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "left-pad"
  fi

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
  track_artifact "package" "pip:cowsay"

  # Verify it's installed
  run pip show cowsay
  assert_success

  # Then uninstall
  run plonk uninstall pip:cowsay
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by pip
  run pip show cowsay
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "cowsay"
  fi

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
  track_artifact "package" "gem:colorize"

  # Verify it's installed
  run gem list colorize
  assert_success
  assert_output --partial "colorize"

  # Then uninstall
  run plonk uninstall gem:colorize
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by gem
  run gem list colorize
  assert_success  # gem list returns 0 even if not found
  refute_output --partial "colorize"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "colorize"
  fi

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
  track_artifact "package" "go:github.com/rakyll/hey"

  # Verify it's installed by go - check binary exists
  # Go installs to GOBIN if set, otherwise GOPATH/bin
  local gobin="$(go env GOBIN)"
  if [[ -z "$gobin" ]]; then
    gobin="$(go env GOPATH)/bin"
  fi
  run test -f "$gobin/hey"
  assert_success

  # Then uninstall
  run plonk uninstall go:github.com/rakyll/hey
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by go - binary should be gone
  # Go installs to GOBIN if set, otherwise GOPATH/bin
  local gobin="$(go env GOBIN)"
  if [[ -z "$gobin" ]]; then
    gobin="$(go env GOPATH)/bin"
  fi
  run test -f "$gobin/hey"
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "hey"
  fi

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
  track_artifact "package" "cargo:ripgrep"

  # Verify it's installed by cargo
  run cargo install --list
  assert_success
  assert_output --partial "ripgrep"

  # Then uninstall
  run plonk uninstall cargo:ripgrep
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by cargo
  run cargo install --list
  assert_success
  refute_output --partial "ripgrep"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "ripgrep"
  fi

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
  require_safe_package "brew:figlet"

  # Install first
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Verify it's installed
  run brew list figlet
  assert_success

  # Dry-run uninstall
  run plonk uninstall brew:figlet --dry-run
  assert_success
  assert_output --partial "would-remove"

  # Verify still installed by brew
  run brew list figlet
  assert_success

  # Verify still in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_output --partial "figlet"

  # Verify still managed
  run plonk status
  assert_output --partial "figlet"
}

@test "uninstall with force removes even if not managed" {
  require_safe_package "brew:fortune"

  # Ensure it's not managed by plonk
  run plonk status
  refute_output --partial "fortune"

  # Check if installed by brew (might or might not be)
  run brew list fortune
  local was_installed=$status

  # Force uninstall
  run plonk uninstall brew:fortune --force
  # Should succeed even if not managed by plonk
  assert_success
  refute_output --partial "panic"

  # If it was installed, verify it's now gone
  if [[ $was_installed -eq 0 ]]; then
    run brew list fortune
    assert_failure
  fi
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
