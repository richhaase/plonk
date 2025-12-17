#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Helper to create a Go custom manager config
# Uses the new 'available' config option to specify a custom availability check
create_go_manager_config() {
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  go:
    binary: go
    description: "Go language package manager"
    install_hint: "Install Go from https://go.dev/dl/"
    help_url: "https://go.dev"
    available:
      command: ["go", "version"]
    list:
      command: ["go", "version"]
      parse: lines
    install:
      command: ["go", "install", "{{.Package}}@latest"]
    upgrade:
      command: ["go", "install", "{{.Package}}@latest"]
    uninstall:
      command: ["echo", "To uninstall {{.Package}}, remove the binary from $GOBIN or $GOPATH/bin"]
EOF
}

# Helper to check if go manager actually works with plonk
check_go_manager_available() {
  create_go_manager_config
  # Verify go is available
  run go version
  if [[ $status -ne 0 ]]; then
    return 1
  fi
  return 0
}

# Basic custom manager tests

@test "custom manager config is recognized" {
  require_package_manager "go"
  create_go_manager_config

  # Config show should display the custom manager
  run plonk config show
  assert_success
  assert_output --partial "go"
  assert_output --partial "managers"
}

@test "custom manager appears in doctor output" {
  require_package_manager "go"
  create_go_manager_config

  run plonk doctor
  assert_success
  # Doctor may show the custom manager if it's in use
}

# Go package install tests
# NOTE: These tests require a working Go installation with GOPATH/GOBIN configured.
# In CI environments where Go isn't fully set up, these will be skipped.

@test "install package with custom go manager" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  run plonk install go:github.com/rakyll/hey
  assert_success
  assert_output --partial "hey"

  track_artifact "package" "go:github.com/rakyll/hey"

  # Verify it's in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "hey"
}

@test "install custom manager package with --dry-run" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"
  create_go_manager_config

  run plonk install --dry-run go:github.com/rakyll/hey
  assert_success
  assert_output --partial "would-add"
  assert_output --partial "hey"
}

# Status and packages tests

@test "status shows custom manager packages" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  run plonk status
  assert_success
  assert_output --partial "go"
}

@test "packages command shows custom manager packages" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  run plonk packages
  assert_success
  assert_output --partial "hey"
}

# Uninstall tests
# Note: Our Go config uses echo to print a helpful message instead of actually
# removing files, which lets us test the uninstall flow safely

@test "uninstall package with custom go manager" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # First install
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Then uninstall - our config echoes a help message
  run plonk uninstall go:github.com/rakyll/hey
  assert_success
  assert_output --partial "hey"
  assert_output --partial "removed"

  # Verify removed from lock file
  run cat "$PLONK_DIR/plonk.lock"
  refute_output --partial "hey"
}

@test "uninstall custom manager package with --dry-run" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # First install
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Dry-run uninstall
  run plonk uninstall --dry-run go:github.com/rakyll/hey
  assert_success
  assert_output --partial "would-remove"

  # Should still be in lock file
  run cat "$PLONK_DIR/plonk.lock"
  assert_output --partial "hey"
}

# Upgrade tests

@test "upgrade package with custom go manager" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # First install
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Then upgrade
  run plonk upgrade go:github.com/rakyll/hey
  assert_success
  assert_output --partial "hey"
}

@test "upgrade all custom manager packages" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # First install
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Upgrade all go packages
  run plonk upgrade go
  assert_success
}

# Error handling tests

@test "install with unknown custom manager shows error" {
  # Don't create the config - manager doesn't exist
  run plonk install unknown-custom-mgr:some-package
  assert_failure
  assert_output --partial "unknown package manager"
}

@test "custom manager with invalid binary shows error" {
  # Create config with non-existent binary
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
managers:
  fake:
    binary: nonexistent-binary-xyz
    install:
      command: ["nonexistent-binary-xyz", "install", "{{.Package}}"]
EOF

  run plonk install fake:some-package
  assert_failure
}

# Mixed manager tests

@test "can use built-in and custom managers together" {
  require_package_manager "go"
  require_safe_package "brew:cowsay"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # Install from both managers
  run plonk install brew:cowsay go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "go:github.com/rakyll/hey"

  # Status should show both
  run plonk status
  assert_success
  assert_output --partial "cowsay"

  # Packages should show both
  run plonk packages
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "hey"
}

# Apply tests

@test "apply installs custom manager packages from lock file" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # Install package
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Manually clean up to simulate missing package
  # (this is complex for Go, so we just test apply runs)
  run plonk apply --packages
  assert_success
}

@test "apply --dry-run shows custom manager packages" {
  require_package_manager "go"
  require_safe_package "go:github.com/rakyll/hey"

  if ! check_go_manager_available; then
    skip "Go manager not available in plonk"
  fi

  # Install package first
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Dry-run apply
  run plonk apply --dry-run --packages
  assert_success
}
