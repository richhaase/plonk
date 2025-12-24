#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Basic doctor command tests

@test "doctor command runs without errors" {
  run plonk doctor
  assert_success
}

@test "doctor shows system requirements check" {
  run plonk doctor
  assert_success
  assert_output --partial "System Requirements"
}

@test "doctor shows environment variables check" {
  run plonk doctor
  assert_success
  assert_output --partial "Environment Variables"
  assert_output --partial "HOME"
  assert_output --partial "PLONK_DIR"
}

@test "doctor shows permissions check" {
  run plonk doctor
  assert_success
  assert_output --partial "Permissions"
}

@test "doctor shows configuration checks" {
  run plonk doctor
  assert_success
  assert_output --partial "Configuration"
}

@test "doctor shows lock file checks" {
  run plonk doctor
  assert_success
  assert_output --partial "Lock File"
}

@test "doctor shows package managers check" {
  run plonk doctor
  assert_success
  assert_output --partial "Package Managers"
}

@test "doctor shows executable path check" {
  run plonk doctor
  assert_success
  assert_output --partial "Executable Path"
}

# Overall health status tests

@test "doctor shows healthy status when no issues" {
  run plonk doctor
  assert_success
  # Should show healthy or operational when all checks pass
  assert_output --partial "healthy" || assert_output --partial "operational" || assert_output --partial "pass"
}

# Package manager availability tests

@test "doctor shows brew availability when installed" {
  # Install a brew package to make brew appear in required managers
  require_safe_package "brew:cowsay"
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk doctor
  assert_success
  assert_output --partial "brew"
  assert_output --partial "available"
}

@test "doctor shows npm availability when installed" {
  require_safe_package "npm:is-odd"
  run plonk install npm:is-odd
  assert_success
  track_artifact "package" "npm:is-odd"

  run plonk doctor
  assert_success
  assert_output --partial "npm"
}

# Configuration checks

@test "doctor handles missing config file gracefully" {
  # Ensure no config file exists
  rm -f "$PLONK_DIR/plonk.yaml"

  run plonk doctor
  assert_success
  # Should indicate using defaults
  assert_output --partial "using defaults" || assert_output --partial "does not exist"
}

@test "doctor handles valid config file" {
  # Create a valid config file
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
default_manager: brew
operation_timeout: 300
EOF

  run plonk doctor
  assert_success
  assert_output --partial "Configuration"
}

@test "doctor detects invalid YAML config" {
  # Create invalid YAML
  echo "invalid: yaml: [" > "$PLONK_DIR/plonk.yaml"

  run plonk doctor
  # Doctor should still run but may show warning/error for config
  assert_success
}

# Lock file checks

@test "doctor handles missing lock file gracefully" {
  # Ensure no lock file exists
  rm -f "$PLONK_DIR/plonk.lock"

  run plonk doctor
  assert_success
  # Should indicate lock file doesn't exist
  assert_output --partial "Lock File"
}

@test "doctor handles valid lock file" {
  # Create a valid lock file
  cat > "$PLONK_DIR/plonk.lock" << 'EOF'
version: 1
resources: []
EOF

  run plonk doctor
  assert_success
  assert_output --partial "Lock File"
}

@test "doctor shows package counts from lock file" {
  require_safe_package "brew:cowsay"

  # Install a package first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  run plonk doctor
  assert_success
  # Should show package count in lock file details
  assert_output --partial "brew"
}

# Help output test

@test "doctor --help shows usage information" {
  run plonk doctor --help
  assert_success
  assert_output --partial "doctor"
  assert_output --partial "health"
}

@test "doctor help via plonk help doctor" {
  run plonk help doctor
  assert_success
  assert_output --partial "doctor"
}

# OS-specific tests

@test "doctor shows correct OS information" {
  run plonk doctor
  assert_success

  # Should show either darwin or linux
  if [[ "$OSTYPE" == "darwin"* ]]; then
    assert_output --partial "darwin"
  else
    assert_output --partial "linux"
  fi
}

# Integration test

@test "doctor after installing packages shows correct state" {
  require_safe_package "brew:figlet"

  # Start fresh
  run plonk doctor
  assert_success

  # Install a package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Doctor should now show that package
  run plonk doctor
  assert_success
  assert_output --partial "brew"
}
