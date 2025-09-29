# Agent Guide for Plonk

## Commands
- **Build**: `just build` or `go build -o bin/plonk ./cmd/plonk`
- **Test all**: `go test ./...`
- **Test single**: `go test ./internal/resources/packages -run TestHomebrew -v`
- **Lint**: `just lint` or `go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m`
- **Coverage**: `just test-coverage` (excludes internal/testutil, tools, cmd/* from totals)

## Architecture
- **Entry point**: `cmd/plonk/` - thin CLI wrapper
- **Commands**: `internal/commands/` - CLI command implementations (install, add, clone, apply, status, etc.)
- **Orchestrator**: `internal/orchestrator/` - coordinates package & dotfile reconciliation across resource types
- **Packages**: `internal/resources/packages/` - 12 package manager implementations (brew, npm, pnpm, cargo, pipx, conda, gem, go, uv, pixi, composer, dotnet)
- **Dotfiles**: `internal/resources/dotfiles/` - dotfile scanning, deployment, atomic operations
- **Config**: `internal/config/` - user configuration (plonk.yaml) with sensible defaults
- **Lock**: `internal/lock/` - package state tracking (plonk.lock in YAML)
- **State reconciliation**: Compare desired (lock) vs actual (system) state, apply differences

## Code Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html) and use `gofmt` (automatic)
- Return structured results with per-item status instead of panicking
- Pass `context.Context` through all layers for cancellation/timeout
- Support table, JSON, and YAML output formats for all commands
- Use interfaces for extensibility (PackageManager, Resource abstractions)
- Table-driven tests preferred; mock external dependencies; test success and error paths
- No inline comments unless code is complex; explanations belong in commit messages/docs
