#!/usr/bin/env bash
# This file is automatically loaded by BATS and runs once for the entire test suite

setup_suite() {
  # Find project root
  local test_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  local project_root="$(cd "$test_dir/../../.." && pwd)"

  # Create a shared directory for the entire test suite
  export PLONK_SUITE_DIR="$(mktemp -d /tmp/plonk-bats-suite-XXXXXX)"
  export PLONK_TEST_BINARY="$PLONK_SUITE_DIR/plonk"

  # Build plonk once for the entire test suite
  echo "# Building plonk from source for test suite..." >&3
  (cd "$project_root" && go build -o "$PLONK_TEST_BINARY" ./cmd/plonk) || {
    echo "Failed to build plonk binary" >&2
    return 1
  }

  # Verify build succeeded
  if [[ ! -f "$PLONK_TEST_BINARY" ]]; then
    echo "Build succeeded but binary not found at $PLONK_TEST_BINARY" >&2
    return 1
  fi

  # Add to PATH for all tests
  export PATH="$PLONK_SUITE_DIR:$PATH"
}

teardown_suite() {
  # Clean up suite directory
  if [[ -n "$PLONK_SUITE_DIR" ]] && [[ -d "$PLONK_SUITE_DIR" ]]; then
    rm -rf "$PLONK_SUITE_DIR"
  fi
}
