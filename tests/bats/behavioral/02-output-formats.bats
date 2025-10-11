#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "status supports table format (default)" {
  run plonk status
  assert_success
  # Look for table-like formatting
  assert_output --partial "Plonk Status"
}
