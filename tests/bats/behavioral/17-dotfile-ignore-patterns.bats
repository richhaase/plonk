#!/usr/bin/env bats

# Tests for automatic ignore of dot-prefixed files in $PLONK_DIR
# Related: plonk-ms6, PR #67

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env

  # Create a minimal plonk config
  cat > "$PLONK_DIR/plonk.yaml" <<EOF
default_manager: brew
EOF
}

@test "dot-prefixed directories in PLONK_DIR are ignored" {
  # Create dot-prefixed directories that should be auto-ignored
  mkdir -p "$PLONK_DIR/.git"
  mkdir -p "$PLONK_DIR/.beads"
  mkdir -p "$PLONK_DIR/.claude"

  # Create files inside them
  echo "ref: refs/heads/main" > "$PLONK_DIR/.git/HEAD"
  echo '{"issues": []}' > "$PLONK_DIR/.beads/issues.jsonl"
  echo '{}' > "$PLONK_DIR/.claude/settings.json"

  # Also create a normal dotfile that SHOULD be tracked
  echo "# normal config" > "$PLONK_DIR/zshrc"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # Should NOT contain dot-prefixed paths
  refute_output --partial ".git"
  refute_output --partial ".beads"
  refute_output --partial ".claude"

  # Should contain normal dotfile
  assert_output --partial "zshrc"
}

@test "dot-prefixed files in PLONK_DIR are ignored" {
  # Create dot-prefixed files that should be auto-ignored
  echo "*.log" > "$PLONK_DIR/.gitignore"
  echo "*.pyc" > "$PLONK_DIR/.gitattributes"

  # Also create a normal dotfile that SHOULD be tracked
  echo "# editor config" > "$PLONK_DIR/editorconfig"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # Should NOT contain dot-prefixed files
  refute_output --partial ".gitignore"
  refute_output --partial ".gitattributes"

  # Should contain normal dotfile
  assert_output --partial "editorconfig"
}

@test "no double-dot paths in dotfiles output" {
  # Create dot-prefixed items that would become ~/.. paths
  mkdir -p "$PLONK_DIR/.testdir"
  echo "test" > "$PLONK_DIR/.testdir/file.txt"
  echo "test" > "$PLONK_DIR/.testfile"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # Should never contain ~/.. pattern (invalid path)
  refute_output --partial "~/.."
}

@test "no double-dot paths in status output" {
  # Create dot-prefixed items
  mkdir -p "$PLONK_DIR/.internal"
  echo "data" > "$PLONK_DIR/.internal/data.txt"

  # Also create normal dotfile
  echo "# config" > "$PLONK_DIR/gitconfig"

  # Run status command
  run plonk status
  assert_success

  # Should never contain ~/.. pattern
  refute_output --partial "~/.."

  # Should contain normal dotfile mapping
  assert_output --partial "gitconfig"
}

@test "nested directories under dot-prefixed paths are also ignored" {
  # Create deeply nested structure under dot-prefixed dir
  mkdir -p "$PLONK_DIR/.git/objects/pack"
  mkdir -p "$PLONK_DIR/.git/refs/heads"
  echo "pack data" > "$PLONK_DIR/.git/objects/pack/pack-abc123.pack"
  echo "ref" > "$PLONK_DIR/.git/refs/heads/main"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # None of the nested paths should appear
  refute_output --partial ".git"
  refute_output --partial "objects"
  refute_output --partial "refs"
  refute_output --partial "pack"
}

@test "normal dotfiles in config subdirectories still work" {
  # Create normal config subdirectory structure (no leading dot in PLONK_DIR)
  mkdir -p "$PLONK_DIR/config/nvim"
  mkdir -p "$PLONK_DIR/config/app"
  echo "set number" > "$PLONK_DIR/config/nvim/init.vim"
  echo "theme: dark" > "$PLONK_DIR/config/app/settings.yaml"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # Should contain the config paths (mapped to ~/.config/...)
  assert_output --partial "config"
  assert_output --partial "init.vim" || assert_output --partial "nvim"
}

@test "plonk.yaml and plonk.lock are still ignored" {
  # These should still be ignored (existing behavior)
  echo "default_manager: brew" > "$PLONK_DIR/plonk.yaml"
  echo "version: 2" > "$PLONK_DIR/plonk.lock"

  # Create a normal dotfile
  echo "# zsh config" > "$PLONK_DIR/zshrc"

  # Run dotfiles command
  run plonk dotfiles
  assert_success

  # Should NOT contain plonk config files
  refute_output --partial "plonk.yaml"
  refute_output --partial "plonk.lock"

  # Should contain normal dotfile
  assert_output --partial "zshrc"
}
