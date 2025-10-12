#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Basic upgrade syntax tests
@test "upgrade command shows help when no arguments given to empty environment" {
  run plonk upgrade
  assert_success
  assert_output --partial "No packages to upgrade"
}

@test "upgrade command rejects trailing colon syntax" {
  require_safe_package "brew:cowsay"

  # First install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Now try invalid trailing colon syntax
  run plonk upgrade brew:
  assert_failure
  assert_output --partial "invalid syntax 'brew:'"
  assert_output --partial "use 'brew' to upgrade"
}

@test "upgrade command rejects unknown package" {
  run plonk upgrade nonexistent-package-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "upgrade command rejects unknown manager:package" {
  run plonk upgrade brew:nonexistent-package-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

# Homebrew upgrade tests
@test "upgrade single brew package" {
  require_safe_package "brew:cowsay"

  # First install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Verify it's installed
  run brew list cowsay
  assert_success

  # Upgrade the specific package
  run plonk upgrade brew:cowsay
  assert_success
  assert_output --partial "cowsay"

  # Should still be installed
  run brew list cowsay
  assert_success
}

@test "upgrade all brew packages" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  # Install multiple packages
  run plonk install brew:figlet brew:sl
  assert_success
  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"

  # Upgrade all homebrew packages
  run plonk upgrade brew
  assert_success
  assert_output --partial "figlet"
  assert_output --partial "sl"

  # Should still be installed
  run brew list figlet
  assert_success
  run brew list sl
  assert_success
}

# NPM upgrade tests
@test "upgrade single npm package" {
  require_package_manager "npm"
  require_safe_package "npm:is-odd"

  # Install package
  run plonk install npm:is-odd
  assert_success
  track_artifact "package" "npm:is-odd"

  # Verify it's installed
  run npm list -g is-odd
  assert_success

  # Upgrade the specific package
  run plonk upgrade npm:is-odd
  assert_success
  assert_output --partial "is-odd"

  # Should still be installed
  run npm list -g is-odd
  assert_success
}

@test "upgrade all npm packages" {
  require_package_manager "npm"
  require_safe_package "npm:is-odd"
  require_safe_package "npm:left-pad"

  # Install multiple packages
  run plonk install npm:is-odd npm:left-pad
  assert_success
  track_artifact "package" "npm:is-odd"
  track_artifact "package" "npm:left-pad"

  # Upgrade all npm packages
  run plonk upgrade npm
  assert_success
  assert_output --partial "is-odd"
  assert_output --partial "left-pad"
}

# UV upgrade tests
@test "upgrade single uv package" {
  require_package_manager "uv"
  require_safe_package "uv:cowsay"

  # Install package
  run plonk install uv:cowsay
  if [[ $status -ne 0 ]]; then
    skip "Failed to install uv:cowsay"
  fi
  track_artifact "package" "uv:cowsay"

  # Upgrade the specific package
  run plonk upgrade uv:cowsay
  assert_success
  assert_output --partial "cowsay"
}

# Gem upgrade tests
@test "upgrade single gem package" {
  require_package_manager "gem"
  require_safe_package "gem:colorize"

  # Install package
  run plonk install gem:colorize
  assert_success
  track_artifact "package" "gem:colorize"

  # Upgrade the specific package
  run plonk upgrade gem:colorize
  assert_success
  assert_output --partial "colorize"
}

# Cargo upgrade tests
@test "upgrade single cargo package" {
  require_package_manager "cargo"
  require_safe_package "cargo:ripgrep"

  # Install package
  run plonk install cargo:ripgrep
  assert_success
  track_artifact "package" "cargo:ripgrep"

  # Upgrade the specific package
  run plonk upgrade cargo:ripgrep
  assert_success
  assert_output --partial "ripgrep"
}

# Pipx upgrade tests
@test "upgrade single pipx package" {
  require_package_manager "pipx"
  require_safe_package "pipx:black"

  # Install package
  run plonk install pipx:black
  if [[ $status -ne 0 ]]; then
    skip "Failed to install pipx:black"
  fi
  track_artifact "package" "pipx:black"

  # Upgrade the specific package
  run plonk upgrade pipx:black
  assert_success
  assert_output --partial "black"
}

# Conda upgrade tests
@test "upgrade single conda package" {
  require_safe_package "conda:jq"

  run which conda
  if [[ $status -ne 0 ]]; then
    skip "conda not available"
  fi

  # Install package
  run plonk install conda:jq
  assert_success
  track_artifact "package" "conda:jq"

  # Upgrade the specific package
  run plonk upgrade conda:jq
  assert_success
  assert_output --partial "jq"
}

# Cross-manager upgrade tests
@test "upgrade package across managers" {
  require_safe_package "brew:cowsay"
  require_package_manager "uv"
  require_safe_package "uv:cowsay"

  # Install cowsay via both managers
  run plonk install brew:cowsay uv:cowsay
  if [[ $status -ne 0 ]]; then
    skip "Failed to install cowsay packages"
  fi
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "uv:cowsay"

  # Upgrade cowsay (should upgrade both)
  run plonk upgrade cowsay
  assert_success
  assert_output --partial "cowsay"
  # Should show both managers (only if both were installed)
  assert_output --partial "brew"
  assert_output --partial "uv"
}

# All packages upgrade test
@test "upgrade all packages" {
  require_safe_package "brew:figlet"
  require_package_manager "npm"
  require_safe_package "npm:is-odd"

  # Install packages from different managers
  run plonk install brew:figlet npm:is-odd
  assert_success
  track_artifact "package" "brew:figlet"
  track_artifact "package" "npm:is-odd"

  # Upgrade all packages
  run plonk upgrade
  assert_success
  assert_output --partial "figlet"
  assert_output --partial "is-odd"
  assert_output --partial "Summary:"
}

# Output format tests
@test "upgrade output in JSON format" {
  require_safe_package "brew:cowsay"

  # Install package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Upgrade with JSON output
  run plonk upgrade brew:cowsay --output json
  assert_success
  assert_output --partial '"command": "upgrade"'
  assert_output --partial '"manager": "brew"'
  assert_output --partial '"package": "cowsay"'
}

@test "upgrade output in YAML format" {
  require_safe_package "brew:sl"

  # Install package
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Upgrade with YAML output
  run plonk upgrade brew:sl --output yaml
  assert_success
  assert_output --partial "command: upgrade"
  assert_output --partial "manager: brew"
  assert_output --partial "package: sl"
}

# Error handling tests
@test "upgrade handles unavailable package manager gracefully" {
  # Try to upgrade with a package manager that doesn't exist
  run plonk upgrade nonexistent-manager:some-package
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "upgrade continues on individual package failures" {
  require_safe_package "brew:figlet"

  # Install one real package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Try to upgrade the real package alongside a fake one via cross-manager syntax
  # This should upgrade figlet successfully but report that fake-package isn't managed
  run plonk upgrade figlet fake-package-xyz
  # This will fail overall, but figlet should still be processed
  assert_failure
  assert_output --partial "not managed by plonk"
}

# Spinner tests for upgrade command
@test "upgrade shows spinner during operation" {
  require_safe_package "brew:cowsay"

  # Install first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Upgrade and check for spinner output
  run plonk upgrade brew:cowsay
  assert_success
  assert_output --partial "Upgrading"
  assert_output --partial "✓"
}

@test "upgrade shows progress indicators for multiple packages" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Upgrade both and check for progress indicators
  run plonk upgrade brew:cowsay brew:figlet
  assert_success
  assert_output --partial "[1/2]"
  assert_output --partial "[2/2]"
  assert_output --partial "Upgrading"
}
