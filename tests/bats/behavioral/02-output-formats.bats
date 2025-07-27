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

@test "info supports table format" {
  run plonk info jq
  assert_success
  # Table format should show package details
  assert_output --partial "Package:"
  assert_output --partial "jq"
}

@test "search supports table format" {
  run plonk search jq
  assert_success
  # Table format shows search results
  assert_output --partial "jq"
}
