#!/bin/bash
# Verify UI/UX hasn't changed by comparing snapshots

SNAPSHOT_DIR="test_snapshots"
TEMP_DIR="test_snapshots_temp"
mkdir -p "$TEMP_DIR"

echo "Verifying UI/UX hasn't changed..."

# Capture current state
plonk --help > "$TEMP_DIR/help.txt" 2>&1
plonk > "$TEMP_DIR/status.txt" 2>&1
plonk ls > "$TEMP_DIR/ls.txt" 2>&1
plonk ls --output json > "$TEMP_DIR/ls_json.txt" 2>&1
plonk ls --output yaml > "$TEMP_DIR/ls_yaml.txt" 2>&1
plonk dotfiles > "$TEMP_DIR/dotfiles.txt" 2>&1
plonk config show > "$TEMP_DIR/config_show.txt" 2>&1
plonk config show --output json > "$TEMP_DIR/config_show_json.txt" 2>&1

# Compare snapshots
FAILED=0
for file in "$SNAPSHOT_DIR"/*.txt; do
    basename=$(basename "$file")
    if [ -f "$TEMP_DIR/$basename" ]; then
        if ! diff -q "$file" "$TEMP_DIR/$basename" > /dev/null; then
            echo "❌ CHANGED: $basename"
            diff "$file" "$TEMP_DIR/$basename"
            FAILED=1
        else
            echo "✅ OK: $basename"
        fi
    fi
done

# Cleanup
rm -rf "$TEMP_DIR"

if [ $FAILED -eq 0 ]; then
    echo "✅ All UI/UX tests passed!"
else
    echo "❌ UI/UX changes detected!"
    exit 1
fi
