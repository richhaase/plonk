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

# UV tests
@test "install uv package" {
  require_safe_package "uv:cowsay"

  # Check if uv is available
  run which uv
  if [[ $status -ne 0 ]]; then
    skip "uv not available"
  fi

  run plonk install uv:cowsay
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "added"

  track_artifact "package" "uv:cowsay"

  # Verify it's actually installed by uv
  run uv tool list
  assert_success
  assert_output --partial "cowsay"

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "cowsay"

  # Verify in status
  run plonk status
  assert_output --partial "cowsay"
}

# Pipx tests
@test "install pipx package" {
  require_safe_package "pipx:ruff"

  # Check if pipx is available
  run which pipx
  if [[ $status -ne 0 ]]; then
    skip "pipx not available"
  fi

  run plonk install pipx:ruff
  assert_success
  assert_output --partial "ruff"
  assert_output --partial "added"

  track_artifact "package" "pipx:ruff"

  # Verify it's actually installed by pipx
  run pipx list --short
  assert_success
  assert_output --partial "ruff"

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "ruff"

  # Verify in status
  run plonk status
  assert_output --partial "ruff"
}

# Pnpm tests
@test "install pnpm package" {
  require_safe_package "pnpm:prettier"

  # Check if pnpm is available
  run which pnpm
  if [[ $status -ne 0 ]]; then
    skip "pnpm not available"
  fi

  run plonk install pnpm:prettier
  assert_success
  assert_output --partial "prettier"
  assert_output --partial "added"

  track_artifact "package" "pnpm:prettier"

  # Verify it's actually installed by pnpm
  run pnpm list -g prettier
  assert_success

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "prettier"

  # Verify in status
  run plonk status
  assert_output --partial "prettier"
}

# Conda tests
@test "install conda package" {
  require_safe_package "conda:jq"

  run which conda
  if [[ $status -ne 0 ]]; then
    skip "conda not available"
  fi

  run plonk install conda:jq
  assert_success
  assert_output --partial "jq"
  assert_output --partial "added"

  track_artifact "package" "conda:jq"

  run plonk status
  assert_success
  assert_output --partial "jq"
}

# Dry-run tests

@test "install --dry-run shows what would happen" {
  require_safe_package "brew:lolcat"

  run plonk install --dry-run brew:lolcat
  assert_success
  assert_output --partial "lolcat"
  assert_output --partial "would-add"
}

@test "install --dry-run does not modify lock file" {
  require_safe_package "brew:lolcat"

  # Get initial lock file state
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    initial_content=$(cat "$PLONK_DIR/plonk.lock")
  else
    initial_content=""
  fi

  run plonk install --dry-run brew:lolcat
  assert_success

  # Lock file should not have changed
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    final_content=$(cat "$PLONK_DIR/plonk.lock")
  else
    final_content=""
  fi

  assert [ "$initial_content" = "$final_content" ]
}

@test "install --dry-run does not install package" {
  require_safe_package "brew:fortune"

  # Make sure package is not installed
  run brew uninstall fortune 2>/dev/null || true

  run plonk install --dry-run brew:fortune
  assert_success

  # Package should not be installed
  run brew list fortune
  assert_failure
}

@test "install -n is alias for --dry-run" {
  require_safe_package "brew:lolcat"

  run plonk install -n brew:lolcat
  assert_success
  assert_output --partial "would-add"
}

@test "install --dry-run with multiple packages shows all" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  run plonk install --dry-run brew:figlet brew:sl
  assert_success
  assert_output --partial "figlet"
  assert_output --partial "sl"
}
