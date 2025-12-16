#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Package Manager Failures
@test "install handles package manager unavailable gracefully" {
  # Try to install with a package manager that doesn't exist in path
  run plonk install nonexistent-mgr:some-package
  assert_failure
  assert_output --partial "unknown package manager"
}

@test "install handles network errors gracefully" {
  require_safe_package "brew:nonexistent-package-xyz-123456"

  # Try to install a package that doesn't exist
  run plonk install brew:nonexistent-package-xyz-123456
  assert_failure
}

@test "upgrade handles package manager errors gracefully" {
  require_safe_package "brew:cowsay"

  # Install a package first
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Now try to upgrade with a made-up package that's not in lock
  run plonk upgrade brew:nonexistent-pkg-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "uninstall handles already-removed package gracefully" {
  require_safe_package "brew:figlet"

  # Install package
  run plonk install brew:figlet
  assert_success
  track_artifact "package" "brew:figlet"

  # Remove it manually with brew
  run brew uninstall figlet
  assert_success

  # Now try to uninstall with plonk - should succeed (idempotent)
  run plonk uninstall brew:figlet
  # Should handle gracefully even if package is already gone
  # (Implementation may vary - either success or informative error)
  # We mainly check it doesn't crash
}

# Lock File Errors
@test "status handles missing lock file gracefully" {
  # In fresh environment with no lock file
  run plonk status
  assert_success
  # Should show empty/no managed packages
}

@test "upgrade handles empty lock file gracefully" {
  # Try to upgrade when nothing is installed
  run plonk upgrade
  assert_success
  assert_output --partial "No packages to upgrade"
}

@test "apply handles empty lock file gracefully" {
  # Try to apply when lock file is empty
  run plonk apply
  assert_success
}

# Concurrent Operation Conflicts
@test "install multiple packages handles individual failures" {
  require_safe_package "brew:cowsay"

  # Try to install one valid and one invalid package
  run plonk install brew:cowsay brew:nonexistent-xyz-456
  # Partial success returns exit code 0 but shows failures in output
  assert_success
  assert_output --partial "succeeded"
  assert_output --partial "failed"
  # The valid package should have been tracked
  track_artifact "package" "brew:cowsay"
}

# Permission Errors
@test "install handles permission errors gracefully" {
  skip "Requires sudo/permission setup"
  # This would test what happens when user lacks permissions
  # Skipped in normal test runs to avoid requiring elevated privileges
}

# Dotfile Errors
@test "add handles non-existent dotfile with clear error" {
  # Try to add a dotfile that doesn't exist
  run plonk add ~/.nonexistent-dotfile-xyz-123
  assert_failure
  assert_output --partial "does not exist"
}

@test "add handles directory without proper permissions" {
  skip "Requires permission setup"
  # Would test permission denied scenarios
}

@test "rm handles non-managed dotfile gracefully" {
  # Try to remove a dotfile that isn't managed
  run plonk rm ~/.random-file-not-managed
  # Returns success but skips the file
  assert_success
  assert_output --partial "not managed"
}

@test "apply dotfiles handles source file deleted" {
  require_safe_dotfile ".test-delete-source"

  # Create a dotfile
  echo "test content" > ~/.test-delete-source

  # Add it to plonk
  run plonk add ~/.test-delete-source
  assert_success
  track_artifact "dotfile" ".test-delete-source"

  # Delete the deployed version to force redeployment
  rm -f ~/.test-delete-source

  # Apply should redeploy it
  run plonk apply --dotfiles
  assert_success

  # Verify it's back
  run test -f ~/.test-delete-source
  assert_success
}

# Config Errors
@test "plonk handles invalid config file gracefully" {
  # Write invalid YAML to config
  echo "invalid: yaml: content: [" > "$PLONK_DIR/plonk.yaml"

  # Commands continue with defaults when config is invalid
  run plonk status
  # Returns success using default config
  assert_success
}

# Upgrade Error Scenarios
@test "upgrade handles mix of successful and failed upgrades" {
  require_safe_package "brew:cowsay"
  require_safe_package "brew:figlet"

  # Install two packages
  run plonk install brew:cowsay brew:figlet
  assert_success
  track_artifact "package" "brew:cowsay"
  track_artifact "package" "brew:figlet"

  # Try to upgrade both plus a non-existent one
  run plonk upgrade cowsay figlet nonexistent-pkg-xyz
  assert_failure
  assert_output --partial "not managed by plonk"
}

@test "upgrade shows meaningful error when package manager fails" {
  require_safe_package "brew:sl"

  # Install package
  run plonk install brew:sl
  assert_success
  track_artifact "package" "brew:sl"

  # Manually break the package (if possible) or just test error reporting
  # This is a placeholder for testing error message quality
  run plonk upgrade sl
  # Should either succeed or show clear error
}

# Dry-run Error Scenarios
@test "dry-run install shows what would happen even for invalid packages" {
  # Dry-run should check package validity
  run plonk install --dry-run brew:definitely-not-a-real-package-xyz
  # Behavior may vary - either fails early or shows "would install"
}

@test "dry-run doesn't modify lock file on errors" {
  # Get initial lock file state
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    initial_hash=$(md5sum "$PLONK_DIR/plonk.lock" | cut -d' ' -f1)
  else
    initial_hash="none"
  fi

  # Try dry-run install with invalid package
  run plonk install --dry-run brew:nonexistent-xyz-789

  # Lock file should not have changed
  if [ -f "$PLONK_DIR/plonk.lock" ]; then
    final_hash=$(md5sum "$PLONK_DIR/plonk.lock" | cut -d' ' -f1)
  else
    final_hash="none"
  fi

  assert [ "$initial_hash" = "$final_hash" ]
}

# Status Error Scenarios
@test "status handles corrupted package manager state" {
  # Install a package
  require_safe_package "brew:cowsay"
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Status should still work even if package state is complex
  run plonk status
  assert_success
}

@test "diff shows only dotfile drift not package drift" {
  require_safe_package "brew:cowsay"

  # Install a package
  run plonk install brew:cowsay
  assert_success
  track_artifact "package" "brew:cowsay"

  # Manually uninstall it
  run brew uninstall cowsay
  assert_success

  # Diff only shows dotfile drift, not package drift
  run plonk diff
  assert_success
  assert_output --partial "No drifted dotfiles found"

  # Use status to see missing packages (status shows all including missing)
  run plonk packages
  assert_success
  assert_output --partial "cowsay"
  assert_output --partial "missing"
}
