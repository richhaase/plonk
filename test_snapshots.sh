#!/bin/bash
# Capture current UI/UX behavior for regression testing

SNAPSHOT_DIR="test_snapshots"
mkdir -p "$SNAPSHOT_DIR"

echo "Capturing current plonk behavior snapshots..."

# Basic commands
plonk --help > "$SNAPSHOT_DIR/help.txt" 2>&1
plonk > "$SNAPSHOT_DIR/status.txt" 2>&1
plonk ls > "$SNAPSHOT_DIR/ls.txt" 2>&1
plonk ls --output json > "$SNAPSHOT_DIR/ls_json.txt" 2>&1
plonk ls --output yaml > "$SNAPSHOT_DIR/ls_yaml.txt" 2>&1
plonk dotfiles > "$SNAPSHOT_DIR/dotfiles.txt" 2>&1
plonk env > "$SNAPSHOT_DIR/env.txt" 2>&1
plonk doctor > "$SNAPSHOT_DIR/doctor.txt" 2>&1

# Config commands
plonk config show > "$SNAPSHOT_DIR/config_show.txt" 2>&1
plonk config show --output json > "$SNAPSHOT_DIR/config_show_json.txt" 2>&1

# Error cases (important for UI/UX)
plonk add nonexistent-package-12345 > "$SNAPSHOT_DIR/error_add.txt" 2>&1
plonk rm nonexistent-package-12345 > "$SNAPSHOT_DIR/error_rm.txt" 2>&1
plonk info nonexistent-package-12345 > "$SNAPSHOT_DIR/error_info.txt" 2>&1

echo "Snapshots captured in $SNAPSHOT_DIR/"
echo "Use these to verify no UI/UX changes during refactoring"
