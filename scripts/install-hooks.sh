#!/bin/bash
# Install git hooks for plonk development

set -e

HOOK_DIR=".git/hooks"
PRE_COMMIT_HOOK="$HOOK_DIR/pre-commit"

# Create hooks directory if it doesn't exist
mkdir -p "$HOOK_DIR"

# Create pre-commit hook
cat > "$PRE_COMMIT_HOOK" << 'EOF'
#!/bin/sh
# Pre-commit hook for Plonk
# Runs pre-commit checks using just

set -e

echo "Running pre-commit checks..."

# Run the unified pre-commit checks
just precommit

# Add only Go files that may have been formatted by goimports
git add *.go **/*.go 2>/dev/null || true

echo "✅ Pre-commit checks passed!"
EOF

# Make hook executable
chmod +x "$PRE_COMMIT_HOOK"

echo "✅ Git hooks installed successfully!"
echo "The pre-commit hook will now run 'just precommit' before each commit."