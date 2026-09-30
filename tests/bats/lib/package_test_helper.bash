#!/usr/bin/env bash

# Package test helpers for consolidated package manager testing
# Helpers for package add and forced removal tests

# =============================================================================
# Manager-specific package verification
# =============================================================================

# Verify a package is installed via its package manager
# Usage: verify_package_installed <manager> <package>
# Returns 0 if installed, 1 if not
verify_package_installed() {
  local manager="$1"
  local package="$2"

  case "$manager" in
    brew|homebrew)
      run brew list "$package"
      assert_success
      ;;
    npm)
      run npm list -g "$package"
      assert_success
      ;;
    gem)
      run gem list "$package"
      assert_success
      assert_output --partial "$package"
      ;;
    cargo)
      run cargo install --list
      assert_success
      assert_output --partial "$package"
      ;;
    uv)
      run uv tool list
      assert_success
      assert_output --partial "$package"
      ;;
    pnpm)
      run pnpm list -g "$package"
      assert_success
      ;;
    bun)
      run bun pm ls -g
      assert_success
      assert_output --partial "$package"
      ;;
    *)
      fail "Unknown package manager: $manager"
      ;;
  esac
}

# Verify a package is NOT installed via its package manager
# Usage: verify_package_not_installed <manager> <package>
verify_package_not_installed() {
  local manager="$1"
  local package="$2"

  case "$manager" in
    brew|homebrew)
      run brew list "$package"
      assert_failure
      ;;
    npm)
      run npm list -g "$package"
      assert_failure
      ;;
    gem)
      # gem list returns 0 even if not found
      run gem list "$package"
      assert_success
      refute_output --partial "$package"
      ;;
    cargo)
      run cargo install --list
      assert_success
      refute_output --partial "$package"
      ;;
    uv)
      run uv tool list
      assert_success
      refute_output --partial "$package"
      ;;
    pnpm)
      # pnpm list -g returns 0 even if package not found
      run pnpm list -g "$package"
      refute_output --partial "$package"
      ;;
    bun)
      run bun pm ls -g
      # bun may return success with empty output
      refute_output --partial "$package"
      ;;
    *)
      fail "Unknown package manager: $manager"
      ;;
  esac
}

# =============================================================================
# Lock file and status verification
# =============================================================================

# Verify package is in lock file
verify_in_lock_file() {
  local package="$1"
  run cat "$PLONK_DIR/plonk.lock"
  assert_success
  assert_output --partial "$package"
}

# Verify package is NOT in lock file
verify_not_in_lock_file() {
  local package="$1"
  if [[ -f "$PLONK_DIR/plonk.lock" ]]; then
    run cat "$PLONK_DIR/plonk.lock"
    refute_output --partial "$package"
  fi
}

# Verify package is in plonk status
verify_in_status() {
  local package="$1"
  run plonk status --all
  assert_output --partial "$package"
}

# Verify package is NOT in plonk status
verify_not_in_status() {
  local package="$1"
  run plonk status --all
  refute_output --partial "$package"
}

# =============================================================================
# Complete test helpers for common patterns
# =============================================================================

# Test installing a single package
# Usage: test_install_single <manager> <package>
test_install_single() {
  local manager="$1"
  local package="$2"
  local full_spec="${manager}:${package}"

  require_safe_package "$full_spec"

  run plonk add "$full_spec"
  assert_success
  assert_output --partial "$package"

  track_artifact "package" "$full_spec"

  verify_package_installed "$manager" "$package"
  verify_in_lock_file "$package"
  verify_in_status "$package"
}

# Test uninstalling a managed package
# Usage: test_uninstall_managed <manager> <package>
test_uninstall_managed() {
  local manager="$1"
  local package="$2"
  local full_spec="${manager}:${package}"

  require_safe_package "$full_spec"

  # Install first
  run plonk add "$full_spec"
  assert_success
  track_artifact "package" "$full_spec"

  # Verify it's installed
  verify_package_installed "$manager" "$package"

  # Then uninstall
  run plonk rm -f "$full_spec"
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled
  verify_package_not_installed "$manager" "$package"
  verify_not_in_lock_file "$package"
  verify_not_in_status "$package"
}
