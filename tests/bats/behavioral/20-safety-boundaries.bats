#!/usr/bin/env bats

load '../lib/test_helper'

setup() {
  setup_test_env
  export HOME="$BATS_TEST_TMPDIR/home"
  mkdir -p "$HOME"
  export NO_COLOR=1
  printf 'git:\n  auto_commit: false\n' > "$PLONK_DIR/plonk.yaml"
}

@test "recursive add refuses rendered template targets without copying other files" {
  mkdir -p "$HOME/.demo" "$PLONK_DIR/demo"
  echo ordinary > "$HOME/.demo/a-ordinary"
  echo DUMMY_VALUE > "$HOME/.demo/z-auth"
  echo '{{keychain:svc/acct}}' > "$PLONK_DIR/demo/z-auth.tmpl"
  run plonk add --dry-run "$HOME/.demo"
  assert_failure
  run plonk add "$HOME/.demo"
  assert_failure
  assert_output --partial 'managed as a template'
  [ ! -e "$PLONK_DIR/demo/a-ordinary" ]
  [ ! -e "$PLONK_DIR/demo/z-auth" ]
}

@test "adding config does not track the Plonk repository itself" {
  export PLONK_DIR="$HOME/.config/plonk"
  mkdir -p "$PLONK_DIR" "$HOME/.config/app"
  printf 'git:\n  auto_commit: false\n' > "$PLONK_DIR/plonk.yaml"
  echo ordinary > "$HOME/.config/app/settings"
  ln -s ../plonk/plonk.yaml "$HOME/.config/app/control-alias"
  run plonk add "$HOME/.config"
  assert_success
  [ -f "$PLONK_DIR/config/app/settings" ]
  [ ! -e "$PLONK_DIR/config/app/control-alias" ]
  [ ! -e "$PLONK_DIR/config/plonk" ]
}

@test "forced explicit relative removal deletes only the current directory file" {
  mkdir -p "$HOME/.config/nvim" "$PLONK_DIR/config/nvim"
  echo desired > "$HOME/.config/nvim/init.lua"
  echo other > "$HOME/.init.lua"
  cp "$HOME/.config/nvim/init.lua" "$PLONK_DIR/config/nvim/init.lua"
  cp "$HOME/.init.lua" "$PLONK_DIR/init.lua"
  cd "$HOME/.config/nvim"
  run plonk rm -f ./init.lua
  assert_success
  [ ! -e "$HOME/.config/nvim/init.lua" ]
  [ -f "$HOME/.init.lua" ]
}

@test "invalid configuration blocks apply rather than dropping permissions" {
  cat > "$PLONK_DIR/plonk.yaml" <<'YAML'
dotfiles:
  rules:
    - name: private.tmpl
      mode: "0600"
operation_timeout: invalid
YAML
  echo ordinary > "$PLONK_DIR/private.tmpl"
  run plonk apply --dotfiles
  assert_failure
  assert_output --partial 'failed to load configuration'
  [ ! -e "$HOME/.private" ]
}

@test "mode drift is visible and read only apply preserves legacy lock" {
  cat > "$PLONK_DIR/plonk.yaml" <<'YAML'
git:
  auto_commit: false
dotfiles:
  rules:
    - name: test
      mode: "0600"
YAML
  echo same > "$PLONK_DIR/test"
  cp "$PLONK_DIR/test" "$HOME/.test"
  chmod 0644 "$HOME/.test"
  printf 'version: 2\nresources: []\n' > "$PLONK_DIR/plonk.lock"
  cp "$PLONK_DIR/plonk.lock" "$BATS_TEST_TMPDIR/original.lock"
  run plonk status
  assert_success
  assert_output --partial 'drifted'
  run plonk diff
  assert_success
  assert_output --partial 'permissions 0644 -> 0600'
  run plonk apply --dry-run
  assert_success
  cmp "$BATS_TEST_TMPDIR/original.lock" "$PLONK_DIR/plonk.lock"
}
