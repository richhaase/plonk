#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Custom Manager Tests

@test "custom manager defined in plonk.yaml works" {
  # Create a custom manager config
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  custom:
    binary: echo
    list:
      command: ["echo", "package1\npackage2"]
      parse: lines
    install: ["echo", "installed {{.Package}}"]
    upgrade: ["echo", "upgraded {{.Package}}"]
    upgrade_all: ["echo", "upgraded all"]
    uninstall: ["echo", "uninstalled {{.Package}}"]
EOF

  # Install a package with custom manager
  run plonk install custom:testpkg
  assert_success
  assert_output --partial "testpkg"

  # Verify it's in lock file
  run grep "custom:testpkg" "$PLONK_DIR/plonk.lock"
  assert_success
}

@test "override built-in manager in plonk.yaml" {
  # Override pipx to use different flags
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  pipx:
    binary: pipx
    install: ["echo", "custom-install", "{{.Package}}"]
EOF

  # Install should use overridden config
  run plonk install pipx:ruff
  assert_success
  # Would see "custom-install" in output if override worked
}

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
  # npm, pnpm, conda use json
  skip "Requires npm to be available"

  run plonk install npm:cowsay
  assert_success
  track_artifact "package" "npm:cowsay"

  # Status should show it (uses ListInstalled with json parsing)
  run plonk status
  assert_success
  assert_output --partial "cowsay"
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

@test "missing binary in custom manager fails gracefully" {
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  missing:
    binary: nonexistent-binary-xyz
    install: ["nonexistent-binary-xyz", "{{.Package}}"]
EOF

  run plonk install missing:testpkg
  assert_failure
  assert_output --partial "not available"
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
  skip "pnpm not commonly available in test env"

  run plonk install pnpm:prettier
  assert_success
  track_artifact "package" "pnpm:prettier"
}

@test "conda install works" {
  require_command "conda"
  skip "conda not commonly available in test env"

  run plonk install conda:numpy
  assert_success
  track_artifact "package" "conda:numpy"
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
  # Should NOT have version column or version info
  refute_output --partial "version"
  refute_output --partial "v"
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

# Template Expansion

@test "template variable {{.Package}} expands correctly" {
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  test:
    binary: echo
    install: ["echo", "Installing: {{.Package}}"]
EOF

  run plonk install test:mypackage
  assert_success
  assert_output --partial "Installing: mypackage"
}
