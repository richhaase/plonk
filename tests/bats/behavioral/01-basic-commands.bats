#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

@test "plonk with no args shows help" {
  run plonk
  assert_success
  assert_output --partial "Usage:"
  assert_output --partial "Available Commands:"
}

@test "plonk status works with empty config" {
  run plonk status
  assert_success
  assert_output --partial "0 managed"
}

@test "plonk st alias works" {
  run plonk st
  assert_success
  assert_output --partial "0 managed"
}

@test "help for specific command works" {
  run plonk help install
  assert_success
  assert_output --partial "Install packages"
  assert_output --partial "Examples:"
}

@test "unknown command shows error and suggestions" {
  run plonk unknowncommand
  assert_failure
  assert_output --partial "unknown command"
}

@test "typo in command name shows helpful error" {
  run plonk instal
  assert_failure
  # Should show unknown command error
  assert_output --partial "unknown"
}

@test "plonk --version shows version information" {
  run plonk --version
  assert_success
  # Should show version string
  assert_output --partial "plonk"
}
