#!/usr/bin/env bash
#
# Docker entrypoint script for plonk BATS tests
#
# This script sets up the environment and runs tests with proper configuration.
#
# Usage:
#   ./docker-entrypoint.sh              # Run all behavioral tests
#   ./docker-entrypoint.sh smoke        # Run smoke tests only
#   ./docker-entrypoint.sh <file.bats>  # Run specific test file
#   ./docker-entrypoint.sh bash         # Start interactive shell
#   ./docker-entrypoint.sh verify       # Verify all package managers

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
info() { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Verify a command exists
check_command() {
    local cmd="$1"
    local name="${2:-$cmd}"
    if command -v "$cmd" &>/dev/null; then
        local version
        case "$cmd" in
            brew) version=$(brew --version | head -1) ;;
            npm) version=$(npm --version) ;;
            pnpm) version=$(pnpm --version) ;;
            cargo) version=$(cargo --version) ;;
            pipx) version=$(pipx --version 2>&1) ;;
            conda) version=$(conda --version) ;;
            gem) version=$(gem --version) ;;
            uv) version=$(uv --version) ;;
            go) version=$(go version | awk '{print $3}') ;;
            bats) version=$(bats --version) ;;
            plonk) version=$(plonk --version 2>&1 | head -1) ;;
            *) version="installed" ;;
        esac
        success "$name: $version"
        return 0
    else
        error "$name: not found"
        return 1
    fi
}

# Verify all package managers are available
verify_package_managers() {
    info "Verifying package managers..."
    echo ""

    local failed=0

    echo "Core tools:"
    check_command go "Go" || ((failed++))
    check_command bats "BATS" || ((failed++))
    check_command plonk "Plonk" || ((failed++))
    echo ""

    echo "Package managers:"
    check_command brew "Homebrew" || ((failed++))
    check_command npm "npm" || ((failed++))
    check_command pnpm "pnpm" || ((failed++))
    check_command cargo "Cargo" || ((failed++))
    check_command pipx "pipx" || ((failed++))
    check_command conda "Conda" || ((failed++))
    check_command gem "Gem" || ((failed++))
    check_command uv "uv" || ((failed++))
    echo ""

    if [[ $failed -eq 0 ]]; then
        success "All package managers verified!"
        return 0
    else
        error "$failed package manager(s) not available"
        return 1
    fi
}

# Run tests
run_tests() {
    local test_path="$1"

    info "Running BATS tests: $test_path"
    info "Cleanup packages: ${PLONK_TEST_CLEANUP_PACKAGES:-1}"
    info "Cleanup dotfiles: ${PLONK_TEST_CLEANUP_DOTFILES:-1}"
    echo ""

    cd /home/plonk/plonk

    exec bats "$test_path"
}

# Main entry point
main() {
    # Source environment for interactive shells
    if [[ -f ~/.bashrc ]]; then
        # shellcheck disable=SC1091
        source ~/.bashrc 2>/dev/null || true
    fi

    # Parse command
    local cmd="${1:-all}"
    shift || true

    case "$cmd" in
        all|"")
            verify_package_managers
            echo ""
            run_tests "tests/bats/behavioral/"
            ;;
        smoke)
            run_tests "tests/bats/behavioral/00-smoke.bats"
            ;;
        verify|check)
            verify_package_managers
            ;;
        bash|shell|sh)
            info "Starting interactive shell..."
            exec bash
            ;;
        help|--help|-h)
            echo "Plonk BATS Test Container"
            echo ""
            echo "Usage: docker-entrypoint.sh [command]"
            echo ""
            echo "Commands:"
            echo "  all, (default)   Run all behavioral tests"
            echo "  smoke            Run smoke tests only"
            echo "  verify, check    Verify all package managers are available"
            echo "  bash, shell      Start interactive bash shell"
            echo "  help             Show this help message"
            echo "  <file.bats>      Run specific test file"
            echo ""
            echo "Examples:"
            echo "  docker-entrypoint.sh"
            echo "  docker-entrypoint.sh smoke"
            echo "  docker-entrypoint.sh tests/bats/behavioral/01-basic-commands.bats"
            echo "  docker-entrypoint.sh bash"
            ;;
        *.bats)
            run_tests "$cmd"
            ;;
        bats)
            # Pass through to bats directly
            exec bats "$@"
            ;;
        *)
            # Pass through any other command
            exec "$cmd" "$@"
            ;;
    esac
}

main "$@"
