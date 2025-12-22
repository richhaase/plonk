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
- **Packages**: `internal/resources/packages/` - 8 hardcoded package managers (brew, npm, pnpm, bun, cargo, gem, go, uv) with self-installation support
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

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
