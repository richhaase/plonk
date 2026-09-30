# Contributing to Plonk

## Quick Start

```bash
git clone https://github.com/richhaase/plonk.git
cd plonk
make build
go test ./...
make install
```

**Requirements:** Go version specified in `go.mod`, Git, and Make. Docker is
required for isolated BATS testing. `make dev-setup` additionally requires
`pre-commit` and installs repository hooks.

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
make install      # Install to GOPATH/bin
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

BATS tests exercise the CLI and real package managers. Run them in Docker:

```bash
make docker-test-all
```

See [tests/bats/README.md](tests/bats/README.md) for focused runs and test fixtures.

## Adding a Package Manager

Plonk supports 5 package managers: brew, cargo, go, pnpm, uv.

To add a new one:

1. Create `internal/packages/newmanager.go` implementing the `Manager` interface:
   ```go
   type Manager interface {
       IsInstalled(ctx context.Context, name string) (bool, error)
       Install(ctx context.Context, name string) error
       Uninstall(ctx context.Context, name string) error
   }
   ```

2. Register in `internal/packages/registry.go`

3. Add to `SupportedManagers` in `internal/packages/manager.go`

4. Add BATS tests in `tests/bats/behavioral/03-package-managers.bats`

5. Update docs: README.md and docs/reference.md

## Adding a Command

1. Create `internal/commands/newcmd.go`
2. Register with root command in `init()`
3. Use `internal/output` formatters if displaying data
4. Add tests
5. Update docs/reference.md

## Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt`
- Return structured results with per-item status
- Pass context through all layers
- Keep command output in `internal/output` formatters
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

### Required Status Checks

CI runs the following checks. Repository branch protection determines which
checks block merging:

| Check | Workflow | What it enforces |
|---|---|---|
| Unit Tests | `CI` | Tests with coverage, lint (including gosec), and formatting |
| Integration Tests (BATS) | `CI` | Behavioral tests in Docker |
| Security Scan | `Security Check` | govulncheck on dependencies; gosec via golangci-lint |

Notes for maintainers:

- Formatting is verified read-only (`gofmt -l -s`); the workflow never rewrites the checkout. If it fails, run `gofmt -w -s .` locally and push.
- Releases (`v*` tags) additionally run the Release Quality Gate (formatting, lint, unit tests, security) before GoReleaser publishes.
- These checks must be configured as *required* in the repository's branch protection settings for `main` (Settings → Branches → Branch protection rule → Require status checks).

## Documentation

When changing functionality, update:
- README.md (if user-facing)
- docs/reference.md (CLI/config changes)
- docs/internals.md (architecture changes)
