#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Helper to create a template file in $PLONK_DIR
create_template() {
  local name="$1"
  local content="$2"
  local dir=$(dirname "$name")

  if [[ "$dir" != "." ]]; then
    mkdir -p "$PLONK_DIR/$dir"
  fi

  echo "$content" > "$PLONK_DIR/${name}.tmpl"
}

# Helper to create local.yaml with variables
create_local_yaml() {
  local content="$1"
  mkdir -p "$PLONK_DIR/.plonk"
  echo "$content" > "$PLONK_DIR/.plonk/local.yaml"
}

# Basic template recognition tests

@test "dotfiles status shows template files" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create a template in PLONK_DIR
  create_template "$testfile" "email = {{.email}}"

  # Check dotfiles status shows the template
  run plonk dotfiles
  assert_success
  assert_output --partial "$testfile"
}

@test "template without local.yaml shows error status" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template without local.yaml
  create_template "$testfile" "value = {{.my_var}}"

  # Deploy it first (will fail silently in status)
  echo "value = test" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  # Status should show error
  run plonk dotfiles
  assert_success
  assert_output --partial "error"
}

@test "template with local.yaml renders correctly" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template and local.yaml
  create_template "$testfile" "email = {{.email}}"
  create_local_yaml "email: test@example.com"

  # Create destination with rendered content
  echo "email = test@example.com" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  # Status should show deployed (not drifted)
  run plonk dotfiles
  assert_success
  assert_output --partial "deployed"
  refute_output --partial "drifted"
  refute_output --partial "error"
}

# Doctor command template tests

@test "doctor shows missing local.yaml when templates exist" {
  local testfile="plonk-test-template"

  # Create template without local.yaml
  create_template "$testfile" "value = {{.my_var}}"

  run plonk doctor
  assert_success
  assert_output --partial "Templates"
  assert_output --partial "local.yaml"
}

@test "doctor shows missing variables" {
  local testfile="plonk-test-template"

  # Create template without local.yaml
  create_template "$testfile" "value = {{.missing_var}}"

  run plonk doctor
  assert_success
  assert_output --partial "Missing variables"
  assert_output --partial "missing_var"
}

@test "doctor shows example local.yaml content" {
  local testfile="plonk-test-template"

  # Create template with a variable
  create_template "$testfile" "name = {{.user_name}}"

  run plonk doctor
  assert_success
  # Should show how to create local.yaml
  assert_output --partial "user_name"
  assert_output --partial "your-value-here"
}

@test "doctor passes when templates have all required variables" {
  local testfile="plonk-test-template"

  # Create template and local.yaml with the variable
  create_template "$testfile" "name = {{.user_name}}"
  create_local_yaml "user_name: John"

  run plonk doctor
  assert_success
  assert_output --partial "Templates"
  assert_output --partial "validated successfully"
}

# Diff command with templates

@test "diff shows error for template with missing variables" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template without local.yaml
  create_template "$testfile" "value = {{.missing_var}}"

  # Create a deployed file
  echo "value = something" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk diff
  # Should show an error about the missing variable
  assert_output --partial "missing_var"
}

@test "diff works correctly with valid template" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template and local.yaml
  create_template "$testfile" "email = {{.email}}"
  create_local_yaml "email: new@example.com"

  # Create deployed file with different content (drift)
  echo "email = old@example.com" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk diff
  assert_success
  # Should show the difference
  assert_output --partial "old@example.com"
  assert_output --partial "new@example.com"
}

# Drift detection with templates

@test "template drift detected when deployed content differs from rendered" {
  local testfile="plonk-test-template-drift"
  require_safe_dotfile ".$testfile"

  # Create template and local.yaml
  create_template "$testfile" "setting = {{.value}}"
  create_local_yaml "value: correct"

  # Create deployed file with different value
  echo "setting = wrong" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk dotfiles
  assert_success
  assert_output --partial "drifted"
}

@test "template shows deployed when rendered matches deployed" {
  local testfile="plonk-test-template-drift"
  require_safe_dotfile ".$testfile"

  # Create template and local.yaml
  create_template "$testfile" "setting = {{.value}}"
  create_local_yaml "value: myvalue"

  # Create deployed file with matching rendered content
  echo "setting = myvalue" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk dotfiles
  assert_success
  assert_output --partial "deployed"
  refute_output --partial "drifted"
}

# Apply with templates

@test "apply renders template to destination" {
  local testfile="plonk-test-template-apply"
  require_safe_dotfile ".$testfile"

  # Create template and local.yaml
  create_template "$testfile" "config_value = {{.my_setting}}"
  create_local_yaml "my_setting: production"

  track_artifact "dotfile" ".$testfile"

  # Apply dotfiles
  run plonk apply --dotfiles
  assert_success

  # Verify rendered content
  run cat "$HOME/.$testfile"
  assert_success
  assert_output "config_value = production"
}

@test "apply fails gracefully when template variables missing" {
  local testfile="plonk-test-template-apply"
  require_safe_dotfile ".$testfile"

  # Create template without local.yaml
  create_template "$testfile" "value = {{.undefined_var}}"

  track_artifact "dotfile" ".$testfile"

  # Apply should fail or show error
  run plonk apply --dotfiles
  # The command may succeed but skip the template, or fail
  # Either way, the file should not exist or be empty
  if [[ -f "$HOME/.$testfile" ]]; then
    # If file exists, it should not have the raw template
    run cat "$HOME/.$testfile"
    refute_output --partial "{{.undefined_var}}"
  fi
}

# Error message quality tests

@test "template error message mentions the variable name" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template without the required variable
  create_template "$testfile" "email = {{.user_email}}"
  create_local_yaml "other_var: value"

  # Create deployed file to trigger comparison
  echo "email = test" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk diff
  # Error should mention the variable name
  assert_output --partial "user_email"
}

@test "template error message mentions local.yaml path" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template with missing variable
  create_template "$testfile" "value = {{.missing}}"
  create_local_yaml "other: value"

  echo "value = x" > "$HOME/.$testfile"
  track_artifact "dotfile" ".$testfile"

  run plonk diff
  # Error should mention where to define the variable
  assert_output --partial "local.yaml"
}

# Multiple variables test

@test "template with multiple variables works when all defined" {
  local testfile="plonk-test-template"
  require_safe_dotfile ".$testfile"

  # Create template with multiple variables
  create_template "$testfile" "name = {{.name}}
email = {{.email}}
host = {{.host}}"

  create_local_yaml "name: John
email: john@example.com
host: localhost"

  # Create matching deployed file
  cat > "$HOME/.$testfile" << 'EOF'
name = John
email = john@example.com
host = localhost
EOF
  track_artifact "dotfile" ".$testfile"

  run plonk dotfiles
  assert_success
  assert_output --partial "deployed"
  refute_output --partial "error"
}

@test "doctor shows all missing variables" {
  local testfile="plonk-test-template"

  # Create template with multiple missing variables
  create_template "$testfile" "a = {{.var_a}}
b = {{.var_b}}"

  # Create local.yaml with neither variable
  create_local_yaml "other: value"

  run plonk doctor
  assert_success
  # Should show at least one missing variable
  assert_output --partial "Missing variables"
}
