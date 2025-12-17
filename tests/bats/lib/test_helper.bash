#!/usr/bin/env bash

# Check if BATS support libraries are available
if [[ -d "/usr/local/lib/bats-support" ]]; then
  load '/usr/local/lib/bats-support/load'
  load '/usr/local/lib/bats-assert/load'
elif [[ -d "/opt/homebrew/lib/bats-support" ]]; then
  # M1 Mac location
  load '/opt/homebrew/lib/bats-support/load'
  load '/opt/homebrew/lib/bats-assert/load'
else
  # Basic fallback assertions if bats-support not installed
  assert() {
    # Evaluate the entire expression
    if ! eval "$@"; then
      echo "Assertion failed: $*" >&2
      return 1
    fi
  }

  assert_success() {
    if [[ "$status" -ne 0 ]]; then
      echo "Expected success, got status $status" >&2
      echo "Output: $output" >&2
      return 1
    fi
  }

  assert_failure() {
    if [[ "$status" -eq 0 ]]; then
      echo "Expected failure, got success" >&2
      echo "Output: $output" >&2
      return 1
    fi
  }

  assert_output() {
    if [[ "$1" == "--partial" ]]; then
      if [[ "$output" != *"$2"* ]]; then
        echo "Expected output to contain: $2" >&2
        echo "Actual output: $output" >&2
        return 1
      fi
    else
      if [[ "$output" != "$1" ]]; then
        echo "Expected output: $1" >&2
        echo "Actual output: $output" >&2
        return 1
      fi
    fi
  }

  refute_output() {
    if [[ "$1" == "--partial" ]]; then
      if [[ "$output" == *"$2"* ]]; then
        echo "Expected output NOT to contain: $2" >&2
        echo "Actual output: $output" >&2
        return 1
      fi
    fi
  }
fi

# Global variables
export PLONK_TEST_DIR="$BATS_TEST_DIRNAME/.."
export SAFE_PACKAGES_FILE="$PLONK_TEST_DIR/config/safe-packages.list"
export SAFE_DOTFILES_FILE="$PLONK_TEST_DIR/config/safe-dotfiles.list"


# Initialize test environment
setup_test_env() {
  # Create isolated plonk config directory
  export PLONK_DIR="$BATS_TEST_TMPDIR/plonk-config"
  mkdir -p "$PLONK_DIR"

  # Create cleanup tracking file
  export CLEANUP_FILE="$BATS_TEST_TMPDIR/cleanup.list"
  : > "$CLEANUP_FILE"

  # Load safe lists
  load_safe_lists
}

# Load safe lists from files or environment
load_safe_lists() {
  # Load safe packages
  if [[ -n "$PLONK_TEST_SAFE_PACKAGES" ]]; then
    IFS=',' read -ra SAFE_PACKAGES <<< "$PLONK_TEST_SAFE_PACKAGES"
  else
    # More portable than mapfile
    SAFE_PACKAGES=()
    while IFS= read -r line; do
      # Remove comments and trim whitespace
      line=$(echo "$line" | sed 's/#.*//' | sed 's/[[:space:]]*$//')
      [[ -n "$line" ]] && SAFE_PACKAGES+=("$line")
    done < <(grep -v '^#' "$SAFE_PACKAGES_FILE" 2>/dev/null | grep -v '^$')
  fi
  # Export array properly for subshells
  export SAFE_PACKAGES

  # Load safe dotfiles
  if [[ -n "$PLONK_TEST_SAFE_DOTFILES" ]]; then
    IFS=',' read -ra SAFE_DOTFILES <<< "$PLONK_TEST_SAFE_DOTFILES"
  else
    # More portable than mapfile
    SAFE_DOTFILES=()
    while IFS= read -r line; do
      # Remove comments and trim whitespace
      line=$(echo "$line" | sed 's/#.*//' | sed 's/[[:space:]]*$//')
      [[ -n "$line" ]] && SAFE_DOTFILES+=("$line")
    done < <(grep -v '^#' "$SAFE_DOTFILES_FILE" 2>/dev/null | grep -v '^$')
  fi
  # Export array properly for subshells
  export SAFE_DOTFILES
}

# Track an artifact for cleanup
track_artifact() {
  local type="$1"
  local name="$2"
  echo "${type}:${name}" >> "$CLEANUP_FILE"
}

# Check if a package is in the safe list
is_safe_package() {
  local package="$1"
  for safe in "${SAFE_PACKAGES[@]}"; do
    if [[ "$package" == "$safe" ]] || [[ "$package" == *":${safe#*:}" ]]; then
      return 0
    fi
  done
  return 1
}
export -f is_safe_package

# Check if a dotfile is in the safe list
is_safe_dotfile() {
  local dotfile="$1"
  for safe in "${SAFE_DOTFILES[@]}"; do
    if [[ "$dotfile" == "$safe" ]] || [[ "$(basename "$dotfile")" == "$safe" ]]; then
      return 0
    fi
  done
  return 1
}
export -f is_safe_dotfile


# Require a safe package or skip
require_safe_package() {
  local package="$1"
  if ! is_safe_package "$package"; then
    skip "Package $package not in safe list"
  fi
}

# Require a safe dotfile or skip
require_safe_dotfile() {
  local dotfile="$1"
  if ! is_safe_dotfile "$dotfile"; then
    skip "Dotfile $dotfile not in safe list"
  fi
}

# Require a command to be available or skip
require_command() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    skip "$cmd not available"
  fi
}

# Require a package manager to be available or skip
require_package_manager() {
  local manager="$1"
  case "$manager" in
    brew|homebrew)
      if ! command -v brew >/dev/null 2>&1; then
        skip "Homebrew not available"
      fi
      ;;
    npm)
      if ! command -v npm >/dev/null 2>&1; then
        skip "NPM not available"
      fi
      ;;
    gem)
      if ! command -v gem >/dev/null 2>&1; then
        skip "Gem not available"
      fi
      ;;
    go)
      if ! command -v go >/dev/null 2>&1; then
        skip "Go not available"
      fi
      ;;
    cargo)
      if ! command -v cargo >/dev/null 2>&1; then
        skip "Cargo not available"
      fi
      ;;
    uv)
      if ! command -v uv >/dev/null 2>&1; then
        skip "UV not available"
      fi
      ;;
    pixi)
      if ! command -v pixi >/dev/null 2>&1; then
        skip "Pixi not available"
      fi
      ;;
    pipx)
      if ! command -v pipx >/dev/null 2>&1; then
        skip "Pipx not available"
      fi
      ;;
    pnpm)
      if ! command -v pnpm >/dev/null 2>&1; then
        skip "Pnpm not available"
      fi
      ;;
    conda)
      if ! command -v conda >/dev/null 2>&1; then
        skip "Conda not available"
      fi
      ;;
    *)
      skip "Unknown package manager: $manager"
      ;;
  esac
}

# Create a test config file
create_test_config() {
  local content="$1"
  echo "$content" > "$PLONK_DIR/plonk.yaml"
}



# Create a test dotfile
create_test_dotfile() {
  local name="$1"
  local content="${2:-# Test file created by BATS}"

  require_safe_dotfile "$name"

  local dir=$(dirname "$name")
  if [[ "$dir" != "." ]]; then
    mkdir -p "$HOME/$dir"
  fi

  echo "$content" > "$HOME/$name"
  track_artifact "dotfile" "$name"
}

# Cleanup function for teardown
cleanup_test_artifacts() {
  if [[ ! -f "$CLEANUP_FILE" ]]; then
    return
  fi

  while IFS=: read -r type name; do
    case "$type" in
      dotfile)
        rm -rf "$HOME/$name" 2>/dev/null || true
        ;;
      package)
        if [[ "$PLONK_TEST_CLEANUP_PACKAGES" != "0" ]]; then
          plonk uninstall "$name" --force 2>/dev/null || true
        fi
        ;;
    esac
  done < "$CLEANUP_FILE"
}

# Standard teardown
teardown() {
  if [[ "$PLONK_TEST_CLEANUP_DOTFILES" != "0" ]]; then
    cleanup_test_artifacts
  fi
}
