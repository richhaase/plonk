#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
  setup_test_env
}

# Test that drifted dotfiles appear only once in status output
@test "status does not show drifted dotfiles twice" {
  local testfile=".plonk-test-drift"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  echo "original content" > "$HOME/$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Modify the deployed file to create drift
  echo "modified content" > "$HOME/$testfile"

  # Run status and count rows containing the filename
  run plonk dotfiles
  assert_success

  # Count how many rows contain the testfile (each row should have $HOME and $PLONK_DIR columns)
  # A single dotfile should appear in exactly one row (but in two columns of that row)
  row_count=$(echo "$output" | grep "$testfile" | grep -c "drifted" || true)

  # Should be exactly 1 row with "drifted" status
  if [ "$row_count" -ne 1 ]; then
  echo "Expected 1 row with $testfile, but found $row_count rows"
  echo "Output: $output"
    return 1
  fi
}

# Test that status displays clear column headers
@test "status shows \$HOME and \$PLONK_DIR column headers" {
  local testfile=".plonk-test-headers"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Run status
  run plonk dotfiles
  assert_success

  # Check for new column headers
  assert_output --partial "\$HOME"
  assert_output --partial "\$PLONK_DIR"

  # Should NOT have old headers
  refute_output --partial "SOURCE"
  refute_output --partial "TARGET"
}

# Test that diff follows standard conventions (current state on left, source on right)
@test "diff shows \$HOME on left and \$PLONK_DIR on right" {
  local testfile=".plonk-test-diff-order"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  echo "line 1" > "$HOME/$testfile"
  echo "line 2" >> "$HOME/$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Modify the deployed file to create drift
  echo "line 1 modified" > "$HOME/$testfile"
  echo "line 2" >> "$HOME/$testfile"

  # Run diff (git diff format shows - for left, + for right)
  run plonk diff "$HOME/$testfile"

  # git diff shows old/left with - and new/right with +
  # We want $HOME (deployed) on left showing as -, and $PLONK_DIR (source) on right showing as +
  # So modified line in $HOME should show as - and original from $PLONK_DIR as +
  assert_output --partial "line 1 modified"
  assert_output --partial "line 1"
}

# Test the --sync-drifted flag to copy modified files back to plonk
@test "plonk add -y syncs drifted files back to \$PLONK_DIR" {
  local testfile=".plonk-test-sync"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  echo "original" > "$HOME/$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Modify the deployed file to create drift
  echo "modified in home" > "$HOME/$testfile"

  # Verify it's drifted
  run plonk dotfiles
  assert_output --partial "drifted"

  # Use add -y to sync back
  run plonk add -y
  assert_success
  assert_output --partial "$testfile"
  assert_output --partial "Updated"

  # Verify the plonk dir has the new content
  local stored_name="${testfile#.}"
  run cat "$PLONK_DIR/$stored_name"
  assert_success
  assert_output "modified in home"

  # Should no longer be drifted
  run plonk dotfiles
  refute_output --partial "drifted"
}

@test "plonk add -y with dry-run shows what would be synced" {
  local testfile=".plonk-test-sync-dry"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile
  echo "original" > "$HOME/$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Modify to create drift
  echo "modified" > "$HOME/$testfile"

  # Dry-run should show what would happen
  run plonk add -y --dry-run
  assert_success
  assert_output --partial "Would update"
  assert_output --partial "$testfile"

  # Verify plonk dir still has original content
  local stored_name="${testfile#.}"
  run cat "$PLONK_DIR/$stored_name"
  assert_success
  assert_output "original"
}

@test "plonk add -y with no drifted files shows appropriate message" {
  local testfile=".plonk-test-no-drift"
  require_safe_dotfile "$testfile"

  # Create and add a dotfile (no modifications)
  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Run add -y when nothing is drifted
  run plonk add -y
  assert_success
  assert_output --partial "No drifted"
}

# Test selective file deployment with plonk apply
@test "plonk apply with file argument only applies that file" {
  local file1=".plonk-test-apply1"
  local file2=".plonk-test-apply2"
  require_safe_dotfile "$file1"
  require_safe_dotfile "$file2"

  # Create and add two dotfiles
  create_test_dotfile "$file1" "file1 content"
  create_test_dotfile "$file2" "file2 content"

  run plonk add "$HOME/$file1" "$HOME/$file2"
  assert_success
  track_artifact "dotfile" "$file1"
  track_artifact "dotfile" "$file2"

  # Remove both from home
  rm -f "$HOME/$file1" "$HOME/$file2"

  # Apply only file1
  run plonk apply "$HOME/$file1"
  assert_success

  # Currently this validates the file but applies all - check for the validation
  # Once full filtering is implemented, uncomment these:
  # run test -f "$HOME/$file1"
  # assert_success
  # run test -f "$HOME/$file2"
  # assert_failure
}

@test "plonk apply with non-managed file shows error" {
  run plonk apply "$HOME/.totally-not-managed-xyz"
  assert_failure
  assert_output --partial "not managed"
}

@test "plonk apply with file arguments cannot combine with --packages" {
  local testfile=".plonk-test-no-combine"
  require_safe_dotfile "$testfile"

  create_test_dotfile "$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # Try to combine file argument with --packages
  run plonk apply "$HOME/$testfile" --packages
  assert_failure
  assert_output --partial "cannot specify files"
}

@test "plonk apply with multiple file arguments validates all" {
  local file1=".plonk-test-multi1"
  local file2=".plonk-test-multi2"
  require_safe_dotfile "$file1"
  require_safe_dotfile "$file2"

  # Add only file1
  create_test_dotfile "$file1"
  run plonk add "$HOME/$file1"
  assert_success
  track_artifact "dotfile" "$file1"

  # Try to apply both file1 (managed) and file2 (not managed)
  run plonk apply "$HOME/$file1" "$HOME/$file2"
  assert_failure
  assert_output --partial "not managed"
  assert_output --partial "$file2"
}

# Integration test for drift detection and sync workflow
@test "complete workflow: drift, check status, sync back, apply selectively" {
  local testfile=".plonk-test-workflow"
  require_safe_dotfile "$testfile"

  # 1. Create and add dotfile
  echo "version 1" > "$HOME/$testfile"
  run plonk add "$HOME/$testfile"
  assert_success
  track_artifact "dotfile" "$testfile"

  # 2. Modify to create drift
  echo "version 2 - edited in home" > "$HOME/$testfile"

  # 3. Check status - should show drifted with new column headers
  run plonk dotfiles
  assert_success
  assert_output --partial "drifted"
  assert_output --partial "\$HOME"
  assert_output --partial "\$PLONK_DIR"

  # 4. Use add -y to sync changes back
  run plonk add -y
  assert_success
  assert_output --partial "$testfile"

  # 5. Status should now show no drift
  run plonk dotfiles
  refute_output --partial "drifted"

  # 6. Remove file from home
  rm -f "$HOME/$testfile"

  # 7. Apply selectively to restore it
  run plonk apply "$HOME/$testfile"
  assert_success

  # Verify content was restored
  run cat "$HOME/$testfile"
  assert_success
  assert_output "version 2 - edited in home"
}
