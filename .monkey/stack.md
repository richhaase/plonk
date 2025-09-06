# Stack Profile (generated)

**Generated:** 2025-09-02T21:47:32Z

## Languages / Frameworks

- **Go** (1.23.10) - Primary language, CLI application using Cobra framework
- **BATS** - Bash Automated Testing System for integration/behavioral tests
- **YAML** - Configuration files and CI/CD workflows
- **Bash** - Shell scripting for build automation and test helpers

## Build / Test / Run

- **Build:** `just build` - Builds versioned binary with ldflags injection
- **Test:** `just test` - Runs Go unit tests (`go test ./...`)
- **Integration Test:** `just test-bats` - Runs BATS behavioral tests
- **Coverage:** `just test-coverage` - Generates test coverage reports
- **Lint:** `just lint` - Runs golangci-lint with extended ruleset
- **Format:** `just format` - Auto-formats code with goimports
- **Clean:** `just clean` - Removes build artifacts and clears test cache
- **Dev Setup:** `just dev-setup` - Complete development environment setup
- **Release:** GoReleaser for cross-platform binary distribution

## Entrypoints / Hot paths

- **Main binary:** `cmd/plonk/main.go` - CLI entry point with version handling
- **Commands:** `internal/commands/` - Cobra command implementations
- **Core logic:** `internal/orchestrator/` - Main application orchestration
- **Resources:** `internal/resources/` - Package and dotfile management
- **Configuration:** `internal/config/` - User configuration handling
- **Output:** `internal/output/` - Formatted CLI output and progress display

## External services

- **Package Managers:** Homebrew, npm, pnpm, cargo, gem, pip, pipx, conda, composer, dotnet, go install
- **Version Control:** Git operations for cloning and repository management
- **File System:** Dotfile synchronization and atomic file operations
- **Lock Files:** YAML-based state tracking for installed packages and dotfiles
