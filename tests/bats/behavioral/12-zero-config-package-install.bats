#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "brew install works without plonk.yaml" {
  require_package_manager "brew"
  require_safe_package "brew:cowsay"

  # Ensure zero-config state
  rm -f "$PLONK_DIR/plonk.yaml"
  if [[ -f "$PLONK_DIR/plonk.yaml" ]]; then
    fail "Expected no plonk.yaml before installing"
  fi

  run plonk install brew:cowsay
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "added"
  track_artifact "package" "brew:cowsay"

  if [[ -f "$PLONK_DIR/plonk.yaml" ]]; then
    fail "Zero-config install should not create plonk.yaml automatically"
  fi

  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "brew:cowsay"
  assert_output --partial "manager: brew"
}

@test "npm install records metadata without plonk.yaml" {
  require_package_manager "npm"
  require_safe_package "npm:left-pad"

  rm -f "$PLONK_DIR/plonk.yaml"
  if [[ -f "$PLONK_DIR/plonk.yaml" ]]; then
    fail "Expected no plonk.yaml before installing"
  fi

  run plonk install npm:left-pad
  assert_success
  assert_output --partial "left-pad"
  assert_output --partial "added"
  track_artifact "package" "npm:left-pad"

  if [[ -f "$PLONK_DIR/plonk.yaml" ]]; then
    fail "Zero-config install should not create plonk.yaml automatically"
  fi

  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "npm:left-pad"
  assert_output --partial "full_name: left-pad"
}
