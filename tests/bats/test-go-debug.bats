#\!/usr/bin/env bats

load 'lib/test_helper'
load 'lib/assertions'

@test "debug go install" {
  setup_test_env

  echo "# PLONK_DIR: $PLONK_DIR" >&3
  echo "# PATH: $PATH" >&3
  echo "# Which plonk: $(which plonk)" >&3

  # Check if hey is already installed
  echo "# Checking if hey exists before install..." >&3
  if test -f "$(go env GOPATH)/bin/hey"; then
    echo "# WARNING: hey already exists at $(go env GOPATH)/bin/hey" >&3
  fi

  echo "# Running: plonk install go:github.com/rakyll/hey" >&3
  run plonk install go:github.com/rakyll/hey
  echo "# Exit status: $status" >&3
  echo "# Output: $output" >&3

  # Check lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    echo "# Lock file contents:" >&3
    cat "$PLONK_DIR/plonk.lock" >&3
  fi
}
