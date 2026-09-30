#!/usr/bin/env bats

load '../lib/test_helper'

setup() {
  setup_test_env
  mkdir -p "$BATS_TEST_TMPDIR/bin"
  export FAKE_PACKAGE_STATE="$BATS_TEST_TMPDIR/installed"
  export FAKE_PACKAGE_LOG="$BATS_TEST_TMPDIR/operations"
  cat > "$BATS_TEST_TMPDIR/bin/brew" <<'SH'
#!/bin/sh
case "$1" in
 list) [ "$2" = "--formula" ] && cat "$FAKE_PACKAGE_STATE" 2>/dev/null; exit 0 ;;
 install) echo "install $3" >> "$FAKE_PACKAGE_LOG"
  [ "$3" = "broken" ] && exit 1
  echo "$3" >> "$FAKE_PACKAGE_STATE" ;;
 uninstall) echo "uninstall $3" >> "$FAKE_PACKAGE_LOG"
  [ "$3" = "broken" ] && exit 1
  sed "/^$3$/d" "$FAKE_PACKAGE_STATE" > "$FAKE_PACKAGE_STATE.new"
  mv "$FAKE_PACKAGE_STATE.new" "$FAKE_PACKAGE_STATE" ;;
 *) exit 1 ;;
esac
SH
  chmod +x "$BATS_TEST_TMPDIR/bin/brew"
  export PATH="$BATS_TEST_TMPDIR/bin:$PATH"
  printf 'git:\n  auto_commit: false\n' > "$PLONK_DIR/plonk.yaml"
}

@test "old package commands are completely removed" {
  run plonk track brew:demo
  assert_failure
  assert_output --partial 'unknown command'
  run plonk untrack brew:demo
  assert_failure
  assert_output --partial 'unknown command'
  run plonk --help
  refute_output --partial '  track '
  refute_output --partial '  untrack '
}

@test "add installs a missing package and tracks only successful installations" {
  run plonk add brew:demo brew:broken
  assert_failure
  assert_output --partial 'installed'
  assert_output --partial 'failed'
  run cat "$PLONK_DIR/plonk.lock"
  assert_output --partial demo
  refute_output --partial broken
  [ "$(cat "$FAKE_PACKAGE_STATE")" = demo ]
}

@test "add tracks installed packages and remains idempotent" {
  echo demo > "$FAKE_PACKAGE_STATE"
  run plonk add brew:demo
  assert_success
  assert_output --partial tracked
  [ ! -f "$FAKE_PACKAGE_LOG" ]
  run plonk add brew:demo
  assert_success
  assert_output --partial 'already installed and tracked'
  [ ! -f "$FAKE_PACKAGE_LOG" ]
  rm "$FAKE_PACKAGE_STATE"
  run plonk add brew:demo
  assert_success
  [ "$(cat "$FAKE_PACKAGE_STATE")" = demo ]
}

@test "rm keeps installed packages unless force is selected" {
  plonk add brew:demo
  run plonk rm brew:demo
  assert_success
  [ "$(cat "$FAKE_PACKAGE_STATE")" = demo ]
  ! grep -q demo "$PLONK_DIR/plonk.lock"
  plonk add brew:demo
  run plonk rm -f brew:demo
  assert_success
  ! grep -q demo "$FAKE_PACKAGE_STATE"
  ! grep -q demo "$PLONK_DIR/plonk.lock"
}

@test "failed uninstall stays tracked and unknown packages stay untouched" {
  echo broken > "$FAKE_PACKAGE_STATE"
  plonk add brew:broken
  run plonk rm -f brew:broken
  assert_failure
  grep -q broken "$PLONK_DIR/plonk.lock"
  run plonk rm -f brew:other
  assert_success
  assert_output --partial 'not tracked'
  ! grep -q other "$FAKE_PACKAGE_LOG"
}

@test "dry runs change neither installations nor lock files" {
  run plonk add -n brew:demo
  assert_success
  assert_output --partial 'would install'
  [ ! -f "$PLONK_DIR/plonk.lock" ]
  [ ! -f "$FAKE_PACKAGE_STATE" ]
  plonk add brew:demo
  cp "$PLONK_DIR/plonk.lock" "$BATS_TEST_TMPDIR/before.lock"
  run plonk rm -nf brew:demo
  assert_success
  cmp "$PLONK_DIR/plonk.lock" "$BATS_TEST_TMPDIR/before.lock"
  [ "$(cat "$FAKE_PACKAGE_STATE")" = demo ]
}

@test "mixed file and package arguments preserve partial successes" {
  local file="$HOME/.plonk-mixed-$BATS_TEST_NUMBER"
  echo test > "$file"
  track_artifact dotfile "${file##*/}"
  run plonk add "$file" brew:demo npm:invalid
  assert_failure
  [ -f "$PLONK_DIR/${file##*/.}" ]
  grep -q demo "$PLONK_DIR/plonk.lock"
  run plonk rm -f "$file" brew:demo
  assert_success
  [ ! -f "$file" ]
  [ ! -f "$PLONK_DIR/${file##*/.}" ]
  ! grep -q demo "$PLONK_DIR/plonk.lock"
}

@test "legacy manager entries can be removed without uninstalling" {
  printf 'version: 3\npackages:\n  npm: [legacy]\n' > "$PLONK_DIR/plonk.lock"
  run plonk rm -f npm:legacy
  assert_failure
  grep -q legacy "$PLONK_DIR/plonk.lock"
  run plonk rm npm:legacy
  assert_success
  ! grep -q legacy "$PLONK_DIR/plonk.lock"
}

@test "empty or option-like package specs are rejected" {
  for spec in brew: brew:-bad :demo unknown:demo; do
    run plonk add "$spec"
    assert_failure
  done
}

@test "force removes template source and deployed target while dry run preserves both" {
  local name="plonk-force-template-$BATS_TEST_NUMBER"
  local target="$HOME/.$name"
  echo '{{env:USER}}' > "$PLONK_DIR/$name.tmpl"
  echo deployed > "$target"
  track_artifact dotfile ".$name"
  run plonk rm -nf "$target"
  assert_success
  assert_output --partial 'would remove source and deployed file'
  [ -f "$target" ]
  [ -f "$PLONK_DIR/$name.tmpl" ]
  run plonk rm -f "$target"
  assert_success
  [ ! -e "$target" ]
  [ ! -e "$PLONK_DIR/$name.tmpl" ]
}

@test "force rejects a directory target without dropping management" {
  local name="plonk-force-directory-$BATS_TEST_NUMBER"
  echo source > "$PLONK_DIR/$name"
  mkdir "$HOME/.$name"
  track_artifact dotfile ".$name"
  run plonk rm -f "$HOME/.$name"
  assert_failure
  assert_output --partial 'directory'
  [ -f "$PLONK_DIR/$name" ]
  [ -d "$HOME/.$name" ]
}

@test "force removes management when the deployed file is already absent" {
  local name="plonk-force-absent-$BATS_TEST_NUMBER"
  echo source > "$PLONK_DIR/$name"
  run plonk rm -f "$HOME/.$name"
  assert_success
  [ ! -e "$PLONK_DIR/$name" ]
}
