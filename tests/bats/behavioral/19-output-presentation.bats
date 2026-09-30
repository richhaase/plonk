#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
  export NO_COLOR=1
  cat > "$PLONK_DIR/plonk.yaml" <<'EOF'
git:
  auto_commit: false
EOF
}

make_status_fixture() {
  create_test_dotfile '.plonk-test-rc' 'healthy'
  cp "$HOME/.plonk-test-rc" "$PLONK_DIR/plonk-test-rc"
  create_test_dotfile '.plonk-test-gitconfig' 'local changes'
  echo source > "$PLONK_DIR/plonk-test-gitconfig"
  echo missing > "$PLONK_DIR/plonk-test-profile"
}

@test "compact status keeps actionable items and omits healthy rows" {
  make_status_fixture
  run plonk status
  assert_success
  assert_output --partial 'drifted  ~/.plonk-test-gitconfig'
  assert_output --partial 'missing  ~/.plonk-test-profile'
  refute_output --partial '.plonk-test-rc'
  refute_output --partial 'Summary:'
  refute_output --partial '===='
  refute_output --partial $'\033['
}

@test "full status uses aligned states with complete item identities" {
  make_status_fixture
  run plonk status --all
  assert_success
  assert_output --partial 'deployed       ~/.plonk-test-rc'
  assert_output --partial 'drifted        ~/.plonk-test-gitconfig'
  assert_output --partial 'missing        ~/.plonk-test-profile'
  assert_output --partial 'Summary: 1 managed, 1 missing, 1 drifted'
}

@test "healthy status confirms the check without listing healthy items" {
  create_test_dotfile '.plonk-test-rc' 'healthy'
  cp "$HOME/.plonk-test-rc" "$PLONK_DIR/plonk-test-rc"
  run plonk status
  assert_success
  assert_output 'All managed items are in sync.'
}

@test "package errors retain the manager and explanation" {
  cat > "$PLONK_DIR/plonk.lock" <<'EOF'
version: 3
packages:
  bogomgr:
    - somepkg
EOF
  run plonk packages
  assert_success
  assert_output --partial 'error          bogomgr:somepkg'
  assert_output --partial 'unsupported manager: bogomgr'
}

@test "dry run labels planned actions and preserves the target" {
  create_test_dotfile '.plonk-test-rc' 'local'
  echo source > "$PLONK_DIR/plonk-test-rc"
  run plonk apply --dotfiles --dry-run
  assert_success
  assert_output --partial 'Dry run'
  assert_output --partial 'would update'
  assert_output --partial '1 planned, 0 failed'
  assert_output --partial 'No changes made.'
  run cat "$HOME/.plonk-test-rc"
  assert_output 'local'
}

@test "compact add and remove explain copy direction and retained deployed file" {
  create_test_dotfile '.plonk-test-rc' 'local'
  run plonk add "$HOME/.plonk-test-rc"
  assert_success
  assert_output --partial 'added'
  assert_output --partial 'copied to $PLONK_DIR/plonk-test-rc'
  refute_output --partial 'Original:'
  run plonk rm "$HOME/.plonk-test-rc"
  assert_success
  assert_output --partial 'removed'
  assert_output --partial 'deployed file kept'
  [ -f "$HOME/.plonk-test-rc" ]
  [ ! -f "$PLONK_DIR/plonk-test-rc" ]
}

@test "doctor uses readable groups without literal Markdown" {
  run plonk doctor
  assert_output --partial 'System readiness'
  assert_output --partial 'Configuration'
  refute_output --partial '**'
  refute_output --partial '##'
  refute_output --partial $'\033['
}

@test "help groups commands by purpose and retains command help" {
  run plonk --help
  assert_success
  assert_output_contains_all 'Inspect:' 'Manage:' 'Sync:' 'Configure:'
  run plonk status --help
  assert_success
  assert_output --partial '--all'
}

@test "terminal color is meaningful and respects NO_COLOR and TERM=dumb" {
  require_command python3
  echo missing > "$PLONK_DIR/plonk-test-profile"
  run python3 "$PLONK_TEST_DIR/lib/capture_terminal.py" color plonk status
  assert_success
  assert_output --partial $'\033[33mmissing\033[0m'
  assert_output --partial '  ~/.plonk-test-profile'
  run python3 "$PLONK_TEST_DIR/lib/capture_terminal.py" no-color plonk status
  assert_success
  refute_output --partial $'\033['
  run python3 "$PLONK_TEST_DIR/lib/capture_terminal.py" dumb plonk status
  assert_success
  refute_output --partial $'\033['
  run python3 "$PLONK_TEST_DIR/lib/capture_terminal.py" color plonk unknowncommand
  assert_failure
  assert_output --partial $'\033[31merror\033[0m'
  assert_output --partial 'unknown command'
}
