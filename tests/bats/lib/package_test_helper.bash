#!/usr/bin/env bash

# Package test helpers for consolidated package manager testing
# Reduces duplication across install, uninstall, and upgrade tests

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
  run plonk status
  assert_output --partial "$package"
}

# Verify package is NOT in plonk status
verify_not_in_status() {
  local package="$1"
  run plonk status
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

  run plonk install "$full_spec"
  assert_success
  assert_output --partial "$package"
  assert_output --partial "added"

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
  run plonk install "$full_spec"
  assert_success
  track_artifact "package" "$full_spec"

  # Verify it's installed
  verify_package_installed "$manager" "$package"

  # Then uninstall
  run plonk uninstall "$full_spec"
  assert_success
  assert_output --partial "removed"

  # Verify actually uninstalled
  verify_package_not_installed "$manager" "$package"
  verify_not_in_lock_file "$package"
  verify_not_in_status "$package"
}

# Test upgrading a single package
# Usage: test_upgrade_single <manager> <package>
test_upgrade_single() {
  local manager="$1"
  local package="$2"
  local full_spec="${manager}:${package}"

  require_safe_package "$full_spec"

  # Install first
  run plonk install "$full_spec"
  if [[ $status -ne 0 ]]; then
    skip "Failed to install $full_spec"
  fi
  track_artifact "package" "$full_spec"

  # Upgrade the specific package
  run plonk upgrade "$full_spec"
  assert_success
  assert_output --partial "$package"

  # Should still be installed after upgrade
  verify_package_installed "$manager" "$package"
}

# Test upgrading all packages for a manager
# Usage: test_upgrade_all_manager <manager> <package1> [package2...]
test_upgrade_all_manager() {
  local manager="$1"
  shift
  local packages=("$@")

  # Install all packages
  local full_specs=()
  for pkg in "${packages[@]}"; do
    local full_spec="${manager}:${pkg}"
    require_safe_package "$full_spec"
    full_specs+=("$full_spec")
  done

  run plonk install "${full_specs[@]}"
  assert_success
  for full_spec in "${full_specs[@]}"; do
    track_artifact "package" "$full_spec"
  done

  # Upgrade all for this manager
  run plonk upgrade "$manager"
  assert_success
  for pkg in "${packages[@]}"; do
    assert_output --partial "$pkg"
  done

  # All should still be installed
  for pkg in "${packages[@]}"; do
    verify_package_installed "$manager" "$pkg"
  done
}
