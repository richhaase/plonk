#!/usr/bin/env bats

# Package tracking tests for the simplified track/untrack commands

load '../test_helper'

@test "track requires manager:package format" {
  run plonk track ripgrep
  [ "$status" -ne 0 ]
  [[ "$output" == *"invalid format"* ]]
}

@test "track rejects unsupported manager" {
  run plonk track npm:typescript
  [ "$status" -ne 0 ]
  [[ "$output" == *"unsupported manager"* ]]
}

@test "track fails for uninstalled package" {
  run plonk track brew:this-package-does-not-exist-xyz123
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "untrack requires manager:package format" {
  run plonk untrack ripgrep
  [ "$status" -ne 0 ]
  [[ "$output" == *"invalid format"* ]]
}

@test "untrack skips packages not being tracked" {
  run plonk untrack brew:not-tracked-xyz123
  [ "$status" -eq 0 ]
  [[ "$output" == *"not tracked"* ]]
}
