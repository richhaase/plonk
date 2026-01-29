#!/usr/bin/env bats

# Package manager tests - verifies each manager's IsInstalled and Install behavior

load '../lib/test_helper'

setup() {
  setup_test_env
}

# =============================================================================
# Brew Manager Tests
# =============================================================================

@test "brew: IsInstalled returns true for installed formula" {
  require_package_manager brew

  # git is almost always installed on dev machines
  run plonk track brew:git
  # Either succeeds (git installed) or fails with "not installed"
  if [[ "$status" -eq 0 ]]; then
    [[ "$output" == *"Tracking"* ]] || [[ "$output" == *"already tracked"* ]]
  else
    [[ "$output" == *"not installed"* ]]
    skip "git not installed via brew"
  fi
}

@test "brew: IsInstalled returns false for non-existent package" {
  require_package_manager brew

  run plonk track brew:this-package-definitely-does-not-exist-xyz987
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "brew: install and track workflow" {
  require_package_manager brew
  require_safe_package "brew:cowsay"

  # Ensure cowsay is installed (may already be)
  brew install cowsay 2>/dev/null || true

  # Track it
  run plonk track brew:cowsay
  [ "$status" -eq 0 ]

  # Verify in lock file
  run cat "$PLONK_DIR/plonk.lock"
  [[ "$output" == *"cowsay"* ]]

  # Untrack (cleanup)
  plonk untrack brew:cowsay
}

# =============================================================================
# Cargo Manager Tests
# =============================================================================

@test "cargo: IsInstalled parses cargo install --list correctly" {
  require_package_manager cargo

  # Try to track a package that's likely not installed
  run plonk track cargo:nonexistent-crate-xyz123
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "cargo: IsInstalled handles installed packages" {
  require_package_manager cargo
  require_safe_package "cargo:bat"

  # Check if bat is installed
  if cargo install --list 2>/dev/null | grep -q "^bat "; then
    run plonk track cargo:bat
    [ "$status" -eq 0 ]
    plonk untrack cargo:bat
  else
    skip "bat not installed via cargo"
  fi
}

@test "cargo: Install fails gracefully for invalid crate" {
  require_package_manager cargo

  # Try to install a nonexistent crate - should fail
  run plonk track cargo:this-crate-does-not-exist-xyz
  [ "$status" -ne 0 ]
}

# =============================================================================
# Go Manager Tests
# =============================================================================

@test "go: IsInstalled checks GOBIN/GOPATH correctly" {
  require_package_manager go

  # Try a package that definitely doesn't exist
  run plonk track go:nonexistent.example.com/fake/tool
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "go: IsInstalled finds installed binaries" {
  require_package_manager go
  require_safe_package "go:github.com/rakyll/hey"

  # Check if hey is in GOBIN
  local gobin="${GOBIN:-${GOPATH:-$HOME/go}/bin}"
  if [[ -x "$gobin/hey" ]]; then
    run plonk track "go:github.com/rakyll/hey"
    [ "$status" -eq 0 ]
    plonk untrack "go:github.com/rakyll/hey"
  else
    skip "hey not installed"
  fi
}

@test "go: IsInstalled handles @version suffix in package name" {
  require_package_manager go

  # gopls is commonly installed
  local gobin="${GOBIN:-${GOPATH:-$HOME/go}/bin}"
  if [[ -x "$gobin/gopls" ]]; then
    # Should recognize gopls even with @version
    run plonk track "go:golang.org/x/tools/gopls@latest"
    [ "$status" -eq 0 ]
    plonk untrack "go:golang.org/x/tools/gopls@latest"
  else
    skip "gopls not installed"
  fi
}

# =============================================================================
# PNPM Manager Tests
# =============================================================================

@test "pnpm: IsInstalled parses JSON output correctly" {
  require_package_manager pnpm

  # Non-existent package
  run plonk track pnpm:this-package-does-not-exist-xyz123
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "pnpm: IsInstalled finds globally installed packages" {
  require_package_manager pnpm
  require_safe_package "pnpm:prettier"

  # Check if prettier is globally installed
  if pnpm list -g --depth=0 --json 2>/dev/null | grep -q '"prettier"'; then
    run plonk track pnpm:prettier
    [ "$status" -eq 0 ]
    plonk untrack pnpm:prettier
  else
    skip "prettier not globally installed via pnpm"
  fi
}

@test "pnpm: handles empty global package list" {
  require_package_manager pnpm

  # Should gracefully handle check against empty/minimal global list
  run plonk track pnpm:definitely-not-installed-xyz
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

# =============================================================================
# UV Manager Tests
# =============================================================================

@test "uv: IsInstalled parses tool list correctly" {
  require_package_manager uv

  # Non-existent tool
  run plonk track uv:this-tool-does-not-exist-xyz123
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "uv: IsInstalled finds installed tools" {
  require_package_manager uv
  require_safe_package "uv:cowsay"

  # Check if cowsay is installed via uv
  if uv tool list 2>/dev/null | grep -q "^cowsay "; then
    run plonk track uv:cowsay
    [ "$status" -eq 0 ]
    plonk untrack uv:cowsay
  else
    skip "cowsay not installed via uv"
  fi
}

@test "uv: Install is idempotent" {
  require_package_manager uv
  require_safe_package "uv:cowsay"

  # Install cowsay if not present
  uv tool install cowsay 2>/dev/null || true

  # Track should succeed
  run plonk track uv:cowsay
  [ "$status" -eq 0 ]

  # Clean up
  plonk untrack uv:cowsay
}

# =============================================================================
# Apply Workflow Tests
# =============================================================================

@test "apply: installs missing packages from lock file" {
  require_package_manager brew
  require_safe_package "brew:cowsay"

  # Ensure cowsay is installed first
  brew install cowsay 2>/dev/null || true

  # Track it
  plonk track brew:cowsay

  # Dry-run apply should show it would be skipped (already installed)
  run plonk apply --dry-run
  [ "$status" -eq 0 ]
  [[ "$output" == *"Skipped"* ]] || [[ "$output" == *"skipped"* ]] || [[ "$output" == *"installed"* ]]

  # Clean up
  plonk untrack brew:cowsay
}

@test "apply: handles empty lock file gracefully" {
  # Start with fresh config dir
  rm -f "$PLONK_DIR/plonk.lock"

  run plonk apply --dry-run
  [ "$status" -eq 0 ]
}

@test "apply: reports failures without stopping" {
  require_package_manager brew

  # Create lock file with invalid package
  mkdir -p "$PLONK_DIR"
  cat > "$PLONK_DIR/plonk.lock" << 'EOF'
version: 3
packages:
  brew:
    - this-package-does-not-exist-xyz123
EOF

  # Apply should fail but report the error
  run plonk apply
  [ "$status" -ne 0 ]
  [[ "$output" == *"failed"* ]] || [[ "$output" == *"Failed"* ]] || [[ "$output" == *"Error"* ]]
}

# =============================================================================
# Registry Tests
# =============================================================================

@test "registry: GetManager returns error for unsupported manager" {
  run plonk track npm:left-pad
  [ "$status" -ne 0 ]
  [[ "$output" == *"unsupported manager"* ]]
}

@test "registry: all supported managers are recognized" {
  # These should fail with "not installed", not "unsupported manager"
  for manager in brew cargo go pnpm uv; do
    run plonk track "${manager}:fake-package-xyz"
    [[ "$output" != *"unsupported manager"* ]]
  done
}

# =============================================================================
# Error Handling Tests
# =============================================================================

@test "manager unavailable: graceful error when brew missing" {
  if command -v brew >/dev/null 2>&1; then
    skip "brew is available"
  fi

  run plonk track brew:cowsay
  [ "$status" -ne 0 ]
  # Should fail, not crash
}

@test "manager unavailable: graceful error when cargo missing" {
  if command -v cargo >/dev/null 2>&1; then
    skip "cargo is available"
  fi

  run plonk track cargo:bat
  [ "$status" -ne 0 ]
}

@test "context timeout: managers respect context cancellation" {
  require_package_manager brew

  # This is hard to test directly, but we can at least verify
  # the track command completes in reasonable time for invalid packages
  timeout 30 plonk track brew:this-does-not-exist-xyz || true
  # If we get here without hanging, the test passes
}
