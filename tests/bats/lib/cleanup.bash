#!/usr/bin/env bash

# Remove all test dotfiles
cleanup_all_test_dotfiles() {
  load_safe_lists

  for dotfile in "${SAFE_DOTFILES[@]}"; do
    if [[ -e "$HOME/$dotfile" ]]; then
      rm -rf "$HOME/$dotfile"
      echo "Removed test dotfile: $dotfile"
    fi
  done
}

# Uninstall all test packages
cleanup_all_test_packages() {
  load_safe_lists

  for package in "${SAFE_PACKAGES[@]}"; do
    # Check if package is managed by plonk
    if plonk status 2>/dev/null | grep -q "${package#*:}"; then
      plonk uninstall "$package" --force 2>/dev/null || true
      echo "Removed test package: $package"
    fi
  done
}

# Full cleanup of all test artifacts
cleanup_all_test_artifacts() {
  echo "Starting full test cleanup..."
  cleanup_all_test_dotfiles
  cleanup_all_test_packages
  echo "Cleanup complete"
}

# Check for test artifacts
check_for_test_artifacts() {
  local found_artifacts=0

  echo "Checking for test artifacts..."

  load_safe_lists

  # Check dotfiles
  for dotfile in "${SAFE_DOTFILES[@]}"; do
    if [[ -e "$HOME/$dotfile" ]]; then
      echo "Found test dotfile: $dotfile"
      ((found_artifacts++))
    fi
  done

  # Check packages
  for package in "${SAFE_PACKAGES[@]}"; do
    if plonk status 2>/dev/null | grep -q "${package#*:}"; then
      echo "Found test package: $package"
      ((found_artifacts++))
    fi
  done

  return $found_artifacts
}
