#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Install alias tests

@test "plonk i is alias for install" {
  require_safe_package "brew:cowsay"

  run plonk i brew:cowsay
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "added"

  track_artifact "package" "brew:cowsay"

  # Verify it's actually installed
  run brew list cowsay
  assert_success
}

@test "plonk i with multiple packages works" {
  require_safe_package "brew:figlet"
  require_safe_package "brew:sl"

  run plonk i brew:figlet brew:sl
  assert_success
  assert_output_contains_all "figlet" "sl"

  track_artifact "package" "brew:figlet"
  track_artifact "package" "brew:sl"
}

@test "plonk i --dry-run works" {
  require_safe_package "brew:lolcat"

  run plonk i --dry-run brew:lolcat
  assert_success
  assert_output --partial "would-add"
}

# Uninstall alias tests

@test "plonk u is alias for uninstall" {
  require_safe_package "brew:cowsay"

  # First install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Now uninstall using alias
  run plonk u brew:cowsay
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "removed"

  # Verify it's not in lock file
  run cat "$PLONK_DIR/plonk.lock"
  refute_output --partial "cowsay"
}

@test "plonk u --dry-run works" {
  require_safe_package "brew:cowsay"

  # First install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk u --dry-run brew:cowsay
  assert_success
  assert_output --partial "would-remove"

  # Package should still be installed
  run brew list cowsay
  assert_success
}

# Packages alias tests

@test "plonk p is alias for packages" {
  run plonk p
  assert_success
}

@test "plonk p shows installed packages" {
  require_safe_package "brew:cowsay"

  # Install a package first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # List packages using alias
  run plonk p
  assert_success
  assert_output --partial "cowsay"
}

# Dotfiles alias tests

@test "plonk d is alias for dotfiles" {
  run plonk d
  assert_success
}

@test "plonk d shows managed dotfiles" {
  local testfile=".plonk-test-rc"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  create_test_dotfile "$testfile"

  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # List dotfiles using alias
  run plonk d
  assert_success
  assert_output --partial "$testfile"
}

# Status alias test (already exists in 01-basic-commands.bats but included for completeness)

@test "plonk st is alias for status (verification)" {
  run plonk st
  assert_success
  assert_output --partial "0 managed"
}

# Combined alias usage tests

@test "aliases work in typical workflow: i, p, u" {
  require_safe_package "brew:figlet"

  # Install using alias
  run plonk i brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Check packages using alias
  run plonk p
  assert_success
  assert_output --partial "figlet"

  # Uninstall using alias
  run plonk u brew:figlet
  assert_success

  # Verify uninstalled
  run plonk p
  refute_output --partial "figlet"
}

@test "dotfile workflow with d alias: add, d, rm" {
  local testfile=".plonk-test-profile"
  require_safe_dotfile "$testfile"

  # Create test dotfile
  create_test_dotfile "$testfile" "test content"

  # Add it
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # List using alias
  run plonk d
  assert_success
  assert_output --partial "$testfile"

  # Remove it
  run plonk rm "$HOME/$testfile"
  assert_success

  # Verify removed from list
  run plonk d
  refute_output --partial "$testfile"
}
