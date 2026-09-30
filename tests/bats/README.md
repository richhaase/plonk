# BATS Tests

BATS exercises the real CLI and package managers. Use Docker: host execution can
install packages and overwrite files under `$HOME`, even though `$PLONK_DIR` is
isolated per test.

## Run

```bash
make docker-test-all            # Build the current checkout and run all tests
make docker-test                # Run the existing image
make docker-test-smoke          # Smoke tests in the existing image
make docker-test-file file=tests/bats/behavioral/19-output-presentation.bats
make docker-shell              # Debug interactively
```

Rebuild after source changes: Compose does not mount the checkout by default.
Tests requiring unavailable managers or fixtures may skip; inspect the BATS
summary and skip reasons.

## Write tests

- Use fixtures from `config/safe-packages.list` and `config/safe-dotfiles.list`.
- Initialize the environment with `lib/test_helper.bash` and track artifacts for
  cleanup using its helpers.
- Assert behavior, exit codes, and relevant output rather than entire transcripts.
- Validate in Docker before submitting changes.

`behavioral/19-output-presentation.bats` covers compact actions, inventories,
healthy status, dry runs, diagnostics, help groups, and terminal color. Its
Python 3 helper (`lib/capture_terminal.py`) captures a pseudo-terminal and checks
`NO_COLOR` and `TERM=dumb`. Captured nonterminal output must be free of ANSI color.

## Environment

| Variable | Default | Purpose |
|---|---|---|
| `PLONK_TEST_CLEANUP_PACKAGES` | `1` | Set `0` to keep installed test packages |
| `PLONK_TEST_CLEANUP_DOTFILES` | `1` | Set `0` to keep created test dotfiles |
| `PLONK_TEST_SAFE_PACKAGES` | `config/safe-packages.list` | Override allowed packages with a comma-separated list |
| `PLONK_TEST_SAFE_DOTFILES` | `config/safe-dotfiles.list` | Override allowed dotfiles with a comma-separated list |

For disposable host environments only, build `plonk` onto `PATH`, install BATS
and Python 3, then run `make test-bats`. Review fixtures first. The cleanup suite
is `tests/bats/cleanup/99-cleanup-all.bats`; cleanup is not a substitute for
isolation or a backup.
