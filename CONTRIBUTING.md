# Contributing to Plonk

## Quick Start

```bash
git clone https://github.com/richhaase/plonk.git
cd plonk
make dev-setup
go test ./...
make install
```

**Requirements:** Go 1.25.0+, Homebrew, Git, Make

## Project Structure

```
plonk/
├── cmd/plonk/              # Entry point
├── internal/
│   ├── commands/           # CLI commands
│   ├── packages/           # Package manager implementations
│   ├── dotfiles/           # Dotfile management
│   ├── orchestrator/       # Coordination
│   ├── config/             # Configuration
│   ├── lock/               # Lock file handling
│   ├── gitops/             # Git automation (auto-commit, push, pull)
│   ├── clone/              # Repository cloning
│   ├── diagnostics/        # Health checks
│   ├── template/           # Template parsing and secret resolvers
│   └── output/             # Output formatting
├── docs/                   # Documentation
└── tests/bats/             # Integration tests
```

See [docs/internals.md](docs/internals.md) for architecture details.

## Development Tasks

```bash
make build        # Build to bin/plonk
make install      # Install to system
make test         # Run tests
make lint         # Run linters
```

## Testing

### Unit Tests

```bash
go test ./...
go test -v ./internal/packages/...
```

### BATS Integration Tests

BATS tests exercise the real CLI with real package managers.

```bash
bats tests/bats/behavioral/
```

Test packages are defined in `tests/bats/config/safe-packages.list`.

## Adding a Package Manager

Plonk supports 5 package managers: brew, cargo, go, pnpm, uv.

To add a new one:

1. Create `internal/packages/newmanager.go` implementing the `Manager` interface:
   ```go
   type Manager interface {
       IsInstalled(ctx context.Context, name string) (bool, error)
       Install(ctx context.Context, name string) error
   }
   ```

2. Register in `internal/packages/registry.go`

3. Add to `SupportedManagers` in `internal/packages/manager.go`

4. Add BATS tests in `tests/bats/behavioral/03-package-managers.bats`

5. Update docs: README.md and docs/reference.md

## Adding a Command

1. Create `internal/commands/newcmd.go`
2. Register with root command in `init()`
3. Add output format support if displaying data
4. Add tests
5. Update docs/reference.md

## Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt`
- Return structured results with per-item status
- Pass context through all layers
- Support table/JSON/YAML output formats
- Treat template resolver values as sensitive: do not put resolved secrets in errors, logs, command output, test fixtures, command arguments, or environment dumps. Use `internal/template.MockSecretResolver` for tests.

## Pull Request Process

1. Fork and create a feature branch
2. Make changes with tests
3. Run `go test ./...`, `go vet ./...`, and `make lint`
4. Run `make test-coverage` for non-trivial behavior changes
5. Submit PR with clear description

### Commit Messages

```
feat: add support for X
fix: handle edge case in Y
docs: update Z documentation
test: add tests for W
```

## Documentation

When changing functionality, update:
- README.md (if user-facing)
- docs/reference.md (CLI/config changes)
- docs/internals.md (architecture changes)
