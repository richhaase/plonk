#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Homebrew uninstall tests
@test "uninstall managed brew package" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Verify it's installed
  run brew list sl
  assert_success

  # Then uninstall
  run plonk uninstall brew:sl
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by brew
  run brew list sl
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "sl"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "sl"
}

# NPM uninstall tests
@test "uninstall managed npm package" {
  require_safe_package "npm:left-pad"

  run which npm
  if [[ $status -ne 0 ]]; then
    skip "npm not available"
  fi

  # Install first
  run plonk install npm:left-pad
  assert_success
  track_artifact "package" "npm:left-pad"

  # Verify it's installed
  run npm list -g left-pad
  assert_success

  # Then uninstall
  run plonk uninstall npm:left-pad
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by npm
  run npm list -g left-pad
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "left-pad"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "left-pad"
}


# Ruby/gem uninstall tests
@test "uninstall managed gem package" {
  require_safe_package "gem:colorize"

  run which gem
  if [[ $status -ne 0 ]]; then
    skip "gem not available"
  fi

  # Install first
  run plonk install gem:colorize
  assert_success
  track_artifact "package" "gem:colorize"

  # Verify it's installed
  run gem list colorize
  assert_success
  assert_output --partial "colorize"

  # Then uninstall
  run plonk uninstall gem:colorize
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by gem
  run gem list colorize
  assert_success  # gem list returns 0 even if not found
  refute_output --partial "colorize"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "colorize"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "colorize"
}

# Go uninstall tests
@test "uninstall managed go package" {
  require_safe_package "go:github.com/rakyll/hey"

  run which go
  if [[ $status -ne 0 ]]; then
    skip "go not available"
  fi

  # Install first
  run plonk install go:github.com/rakyll/hey
  assert_success
  track_artifact "package" "go:github.com/rakyll/hey"

  # Verify it's installed by go - check binary exists
  # Go installs to GOBIN if set, otherwise GOPATH/bin
  local gobin="$(go env GOBIN)"
  if [[ -z "$gobin" ]]; then
    gobin="$(go env GOPATH)/bin"
  fi
  run test -f "$gobin/hey"
  assert_success

  # Then uninstall
  run plonk uninstall go:github.com/rakyll/hey
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by go - binary should be gone
  # Go installs to GOBIN if set, otherwise GOPATH/bin
  local gobin="$(go env GOBIN)"
  if [[ -z "$gobin" ]]; then
    gobin="$(go env GOPATH)/bin"
  fi
  run test -f "$gobin/hey"
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "hey"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "hey"
}

# Cargo uninstall tests
@test "uninstall managed cargo package" {
  require_safe_package "cargo:ripgrep"

  run which cargo
  if [[ $status -ne 0 ]]; then
    skip "cargo not available"
  fi

  # Install first
  run plonk install cargo:ripgrep
  assert_success
  track_artifact "package" "cargo:ripgrep"

  # Verify it's installed by cargo
  run cargo install --list
  assert_success
  assert_output --partial "ripgrep"

  # Then uninstall
  run plonk uninstall cargo:ripgrep
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by cargo
  run cargo install --list
  assert_success
  refute_output --partial "ripgrep"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "ripgrep"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "ripgrep"
}

# UV uninstall tests
@test "uninstall managed uv package" {
  require_safe_package "uv:rich-cli"

  run which uv
  if [[ $status -ne 0 ]]; then
    skip "uv not available"
  fi

  # Install first
  run plonk install uv:rich-cli
  assert_success
  track_artifact "package" "uv:rich-cli"

  # Verify it's installed by uv
  run uv tool list
  assert_success
  assert_output --partial "rich-cli"

  # Then uninstall
  run plonk uninstall uv:rich-cli
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by uv
  run uv tool list
  assert_success
  refute_output --partial "rich-cli"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "rich-cli"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "rich-cli"
}

# Pixi uninstall tests
@test "uninstall managed pixi package" {
  require_safe_package "pixi:gcal"

  run which pixi
  if [[ $status -ne 0 ]]; then
    skip "pixi not available"
  fi

  # Install first
  run plonk install pixi:gcal
  assert_success
  track_artifact "package" "pixi:gcal"

  # Verify it's installed by pixi
  run pixi global list
  assert_success
  assert_output --partial "gcal"

  # Then uninstall
  run plonk uninstall pixi:gcal
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by pixi
  run pixi global list
  assert_success
  refute_output --partial "gcal"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "gcal"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "gcal"
}

# Composer uninstall tests
@test "uninstall managed composer package" {
  require_safe_package "composer:splitbrain/php-cli"

  run which composer
  if [[ $status -ne 0 ]]; then
    skip "composer not available"
  fi

  # Install first
  run plonk install composer:splitbrain/php-cli
  assert_success
  track_artifact "package" "composer:splitbrain/php-cli"

  # Verify it's installed
  run composer global show splitbrain/php-cli
  assert_success

  # Then uninstall
  run plonk uninstall composer:splitbrain/php-cli
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by composer
  run composer global show splitbrain/php-cli
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "splitbrain/php-cli"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "splitbrain/php-cli"
}

@test "uninstall second managed composer package" {
  require_safe_package "composer:minicli/minicli"

  run which composer
  if [[ $status -ne 0 ]]; then
    skip "composer not available"
  fi

  # Install first
  run plonk install composer:minicli/minicli
  assert_success
  track_artifact "package" "composer:minicli/minicli"

  # Verify it's installed
  run composer global show minicli/minicli
  assert_success

  # Then uninstall
  run plonk uninstall composer:minicli/minicli
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by composer
  run composer global show minicli/minicli
  assert_failure

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "minicli/minicli"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "minicli/minicli"
}

# Pipx uninstall tests
@test "uninstall managed pipx package" {
  require_safe_package "pipx:black"

  run which pipx
  if [[ $status -ne 0 ]]; then
    skip "pipx not available"
  fi

  # Install first
  run plonk install pipx:black
  assert_success
  track_artifact "package" "pipx:black"

  # Verify it's installed by pipx
  run pipx list --short
  assert_success
  assert_output --partial "black"

  # Then uninstall
  run plonk uninstall pipx:black
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by pipx
  run pipx list --short
  assert_success
  refute_output --partial "black"

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "black"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "black"
}

# Conda uninstall tests
@test "uninstall managed conda package" {
  require_safe_package "conda:tree"

  # Check if conda/mamba is available
  run which mamba
  local has_mamba=$status
  run which conda
  local has_conda=$status

  if [[ $has_mamba -ne 0 && $has_conda -ne 0 ]]; then
    skip "conda/mamba not available"
  fi

  # Install first
  run plonk install conda:tree
  assert_success
  track_artifact "package" "conda:tree"

  # Verify it's installed by conda/mamba
  if [[ $has_mamba -eq 0 ]]; then
    run mamba list -n base tree
    assert_success
    assert_output --partial "tree"
  else
    run conda list -n base tree
    assert_success
    assert_output --partial "tree"
  fi

  # Then uninstall
  run plonk uninstall conda:tree
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled by conda/mamba
  if [[ $has_mamba -eq 0 ]]; then
    run mamba list -n base tree
    assert_success
    refute_output --partial "tree"
  else
    run conda list -n base tree
    assert_success
    refute_output --partial "tree"
  fi

  # Verify gone from lock file
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "tree"
  fi

  # Verify gone from status
  run plonk status
  refute_output --partial "tree"
}

# General uninstall behavior tests
@test "uninstall non-managed package acts as pass-through" {
  require_safe_package "brew:fortune"

  # Install directly with brew (not via plonk)
  run brew install fortune
  assert_success

  # Verify it's not managed by plonk
  run plonk status
  refute_output --partial "fortune"

  # Uninstall via plonk - should pass through to brew
  run plonk uninstall brew:fortune
  assert_success
  assert_output --partial "removed"

  # Verify it was actually uninstalled
  run brew list fortune
  assert_failure
}

# Test multiple package uninstallation - only testing with brew since
# the logic is the same for all managers (just loops over single uninstalls)
@test "uninstall multiple packages" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Uninstall both
  run plonk uninstall brew:cowsay brew:figlet
  assert_success
  assert_output_contains_all "cowsay" "figlet" "removed"

  # Verify both gone from status
  run plonk status
  refute_output --partial "cowsay"
  refute_output --partial "figlet"
}

@test "uninstall without prefix uses manager from lock file" {
  require_safe_package "brew:sl"

  # Install with brew
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Set default manager to something else
  cat > "$PLONK_DIR/plonk.yaml" <<EOF
default_manager: npm
EOF

  # Uninstall without prefix - should use brew from lock file
  run plonk uninstall sl
  assert_success
  assert_output --partial "removed"

  # Verify it was uninstalled with brew (not npm)
  run brew list sl
  assert_failure
}

@test "uninstall succeeds when package removed from lock even if system uninstall fails" {
  require_safe_package "brew:figlet"

  # Install package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Manually uninstall from system
  run brew uninstall figlet --force

  # Now plonk uninstall should still succeed (removes from lock)
  run plonk uninstall brew:figlet
  assert_success
  assert_output --partial "removed"

  # Verify it's gone from lock file
  run grep "name: figlet" "$PLONK_DIR/plonk.lock"
  assert_failure
}

# Spinner tests for uninstall command
@test "uninstall shows spinner during operation" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Uninstall and check for spinner output
  run plonk uninstall brew:sl
  assert_success
  assert_output --partial "Uninstalling"
  assert_output --partial "✓"
  assert_output --partial "removed"
}

@test "uninstall shows completion message after spinner" {
  require_safe_package "brew:figlet"

  # Install first
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Uninstall and verify completion message
  run plonk uninstall brew:figlet
  assert_success
  assert_output --partial "Uninstalling: brew:figlet"
  assert_output --partial "✓"
}

@test "uninstall shows error message when removal fails" {
  # Try to uninstall a non-existent package
  run plonk uninstall brew:nonexistentpackage123456
  assert_failure
  assert_output --partial "Failed to uninstall"
}

@test "uninstall in dry-run mode shows spinner without actual removal" {
  require_safe_package "brew:sl"

  # Install first
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Dry-run uninstall
  run plonk uninstall --dry-run brew:sl
  assert_success
  assert_output --partial "Uninstalling"
  assert_output --partial "would-remove"

  # Verify package is still installed
  run brew list sl
  assert_success
}

@test "uninstall multiple packages shows progress indicators" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install both first
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Uninstall both and check for progress indicators
  run plonk uninstall brew:cowsay brew:figlet
  assert_success
  assert_output --partial "[1/2]"
  assert_output --partial "[2/2]"
  assert_output --partial "Uninstalling"
}
