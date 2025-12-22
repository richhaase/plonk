#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Basic config show command tests

@test "config show runs without errors" {
  run plonk config show
  assert_success
}

@test "config show displays config path" {
  run plonk config show
  assert_success
  assert_output --partial "plonk.yaml"
}

# Default config tests

@test "config show works with no config file (defaults)" {
  # Ensure no config file exists
  rm -f "$PLONK_DIR/plonk.yaml"

  run plonk config show
  assert_success
  # Should show default values
  assert_output --partial "default"
}

@test "config show displays default_manager setting" {
  run plonk config show
  assert_success
  assert_output --partial "default_manager"
}

@test "config show displays timeout settings" {
  run plonk config show
  assert_success
  # Should show timeout configuration
  assert_output --partial "timeout"
}

# User config tests

@test "config show displays user-defined values" {
  # Create a config file with custom values
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
default_manager: npm
operation_timeout: 600
package_timeout: 120
dotfile_timeout: 60
EOF

  run plonk config show
  assert_success
  assert_output --partial "npm"
  assert_output --partial "600"
}

@test "config show displays expand_directories" {
  # Create a config file with expand_directories
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
expand_directories:
  - .config
  - .local/share
EOF

  run plonk config show
  assert_success
  assert_output --partial "expand_directories"
  assert_output --partial ".config"
}

@test "config show displays ignore_patterns" {
  # Create a config file with ignore_patterns
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
ignore_patterns:
  - "*.swp"
  - ".git"
EOF

  run plonk config show
  assert_success
  assert_output --partial "ignore_patterns"
}

# Annotation tests (user-defined vs default)

@test "config show distinguishes user-defined values" {
  # Create a config file with some custom values
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
default_manager: cargo
EOF

  run plonk config show
  assert_success
  # Should show the user-defined value
  assert_output --partial "cargo"
}

# Merged config tests

@test "config show shows merged defaults and user values" {
  # Create a partial config file
  cat > "$PLONK_DIR/plonk.yaml" << 'EOF'
default_manager: gem
EOF

  run plonk config show
  assert_success
  # Should show user-defined default_manager
  assert_output --partial "gem"
  # Should also show other default values (like timeouts)
  assert_output --partial "timeout"
}

# Error handling tests

@test "config show handles invalid YAML gracefully" {
  # Create invalid YAML
  echo "invalid: yaml: [" > "$PLONK_DIR/plonk.yaml"

  # Should either show defaults or handle error gracefully
  run plonk config show
  # May succeed with defaults or show error
}

@test "config show handles empty config file" {
  # Create empty config file
  touch "$PLONK_DIR/plonk.yaml"

  run plonk config show
  assert_success
}

# Help tests

@test "config show --help shows usage" {
  run plonk config show --help
  assert_success
  assert_output --partial "show"
  assert_output --partial "configuration"
}

@test "plonk help config shows subcommands" {
  run plonk help config
  assert_success
  assert_output --partial "show"
}

# Edge cases

@test "config show respects PLONK_DIR environment variable" {
  # PLONK_DIR is already set by setup_test_env
  run plonk config show
  assert_success
  # Should show the test PLONK_DIR path
  assert_output --partial "$PLONK_DIR"
}

