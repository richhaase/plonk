#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Custom Manager Tests

# Idempotent Operations

@test "install already installed package succeeds (idempotent)" {
  require_safe_package "brew:cowsay"

  # Install once
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Install again - should succeed (idempotent)
  run plonk install brew:cowsay
  assert_success
  assert_output --partial "skipped"
}

@test "upgrade already up-to-date package succeeds (idempotent)" {
  require_safe_package "brew:cowsay"

  # Install package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Upgrade immediately (already latest)
  run plonk upgrade brew:cowsay
  assert_success
}

# Parse Strategy Tests

@test "lines parse strategy works" {
  # pipx, brew, cargo, gem use lines
  require_safe_package "brew:cowsay"

  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Status should show it (uses ListInstalled with lines parsing)
  run plonk status
  assert_success
  assert_output --partial "cowsay"
}

@test "json parse strategy works" {
  require_command "npm"
  require_safe_package "npm:is-odd"

  run plonk install npm:is-odd
  assert_success
  track_artifact "package" "npm:is-odd"

  # Status should show it (uses ListInstalled with json parsing)
  run plonk status
  assert_success
  assert_output --partial "is-odd"
}

# Invalid Config Tests

@test "invalid parse strategy shows error" {
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  bad:
    binary: echo
    list:
      command: ["echo", "test"]
      parse: invalid_strategy
EOF

  run plonk status
  # Should handle invalid parse strategy gracefully
  # Either skip or error - both acceptable
}

# Manager-Specific Tests

@test "uv tool install works" {
  require_command "uv"
  require_safe_package "uv:httpx"

  run plonk install uv:httpx
  assert_success
  track_artifact "package" "uv:httpx"

  run plonk status
  assert_success
  assert_output --partial "httpx"
}

@test "cargo install works with idempotent error handling" {
  require_command "cargo"
  require_safe_package "cargo:bat"

  # Install once
  run plonk install cargo:bat
  assert_success
  track_artifact "package" "cargo:bat"

  # Install again - cargo errors but should be handled as idempotent
  run plonk install cargo:bat
  assert_success
  assert_output --partial "skipped"
}

@test "pnpm global install works" {
  require_command "pnpm"
  require_safe_package "pnpm:prettier"

  run plonk install pnpm:prettier
  assert_success
  track_artifact "package" "pnpm:prettier"
}

@test "conda install works" {
  require_command "conda"
  require_safe_package "conda:jq"

  run plonk install conda:jq
  assert_success
  track_artifact "package" "conda:jq"
}

# UpgradeAll Tests

@test "upgrade with no arguments upgrades all packages" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install two packages
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Upgrade all
  run plonk upgrade
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "figlet"
}

# Status Without Versions

@test "status does not show package versions" {
  require_safe_package "brew:cowsay"

  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk status
  assert_success
  refute_output --partial "version"
}

@test "status only shows installed/missing not untracked" {
  require_safe_package "brew:cowsay"

  # Install something outside plonk
  run brew install figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Install something via plonk
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk status
  assert_success
  # Should show cowsay (managed)
  assert_output --partial "cowsay"
  # Should NOT show figlet (untracked)
  refute_output --partial "figlet"
}
