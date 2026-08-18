---
contract_version: "1.1"
id: "plonk-secret-template-resolver"
title: "Secret-Aware Template Resolution with macOS Keychain Backend (Stage A)"
state: completed
created_at: "2026-08-17T23:30:00Z"
updated_at: "2026-08-18T00:20:00Z"
approved_at: "2026-08-17T23:55:00Z"
approved_by: "richhaase"
implemented_at: "2026-08-18T00:00:00Z"
implemented_by: "PR #130, merge commit 7eae869"
---

# plonk-secret-template-resolver — Stage A

## Intent

Enable Plonk templates (`.tmpl`) to resolve sensitive credentials directly from secure system secret stores (beginning with macOS Keychain) at deployment time without placing plaintext secrets into shell environment variables, command arguments, Git repositories, log files, external diff tools, or AI tool observation streams.

---

## In Scope

- **AC1: Centralized, Namespaced Template Directive Grammar**
  - Consolidate regex and parsing logic from `dotfiles.go` and `health.go` into a unified `internal/template` package.
  - Support namespaced directives: `{{keychain:service/account}}` and `{{keychain:service}}` (defaulting account to current user `$USER`).
  - Maintain 100% backwards compatibility for legacy `{{VAR_NAME}}` environment variable syntax and explicit `{{env:VAR_NAME}}`.
  - Directive parsing uses two-phase tokenization (`{{` ... `}}` closure matching followed by provider prefix split) allowing slashes, dashes, dots, and underscores in locator paths.

- **AC2: Pluggable SecretResolver Interface & Structured Error Taxonomy**
  - A modular resolver abstraction (`SecretResolver`) separating template parsing from secret storage backends:
    ```go
    type SecretResolver interface {
        Scheme() string
        Resolve(ctx context.Context, locator string) (string, error)
        RemediationHint(locator string) string
    }
    ```
  - Standardized error classification:
    - `ErrSecretNotFound` (locator does not exist in store)
    - `ErrProviderUnavailable` (running on unsupported OS or missing provider CLI)
    - `ErrKeychainLocked` / `ErrAccessDenied` (OS store locked or authorization failed)
    - `ErrInvalidDirectiveSyntax` (malformed template expression)
  - Core providers implemented in Stage A:
    1. `EnvResolver` (resolves `env:VAR` and legacy `VAR`)
    2. `MacOSKeychainResolver` (resolves `keychain:service[/account]` via `/usr/bin/security`)
    3. `MockSecretResolver` for deterministic, cross-platform unit and integration testing without requiring host keychain access.

- **AC3: Hardened Native macOS Keychain Execution**
  - Query macOS Keychain via direct `exec.CommandContext` to literal `/usr/bin/security find-generic-password -s <service> [-a <account>] -w`.
  - Constant binary path, isolated environment, and argument-separated flags (strictly no shell interpolation or `PATH` lookup).
  - Explicit timeout and non-interactive detection to prevent headless hangs (e.g. SSH / CI) when Keychain is locked.
  - On non-macOS platforms, `keychain:...` directives return `ErrProviderUnavailable` with a clear message.

- **AC4: Strict Redaction and Authorized Write Surface**
  - **Authorized Write Surface**: Resolved secret values are *only* permitted in memory during rendering and written to the authorized target dotfile under `$HOME` (with restrictive `0600` permissions per `dotfiles.rules`).
  - **Forbidden Surfaces**: Secrets must never appear in stdout/stderr, logs, error messages, JSON outputs, CLI flags, temporary diff files, or stack traces.
  - Error messages and `plonk doctor` reports must print *only* the secret locator (service/account), never secret values.
  - `plonk doctor` surfaces provider-owned remediation hints (e.g. `security add-generic-password -s <service> -a <account> -w`) with a note to enter secrets securely without saving to shell history.

- **AC5: Secret-Safe Diff and Drift Inspection**
  - `plonk status` checks drift using in-memory rendered comparison without writing plaintext secret files to disk outside target `$HOME`.
  - `plonk diff` detects secret-bearing directives and redacts secret values with `[REDACTED_SECRET]` before piping content to external diff tools or temporary diff files.
  - An optional `--show-secrets` flag may be added in the future, but safe masked diffing is the strict default.

- **AC6: Verification & Test Suite Integrity**
  - 100% deterministic unit tests across all packages (`internal/template`, `internal/dotfiles`, `internal/diagnostics`, `internal/config`) testing happy path, missing key, locked keychain, invalid syntax, provider errors, and redaction.
  - Existing unit and behavioral test suites remain green on both macOS and Linux (CI).

---

## Out of Scope

- Linux Secret Service / DBus / `libsecret` backend implementation (deferred to Stage A.2).
- Windows Credential Manager backend implementation.
- Writing/storing secrets into Keychain from Plonk CLI (storing remains interactive via `security add-generic-password` or dedicated secret manager; Plonk is strictly a consumer/resolver).
- 1Password / Bitwarden / Vault CLI integrations (reserved for future provider plugins).

---

## Invariants to Preserve

- **Atomic File Deployment**: Restrictive temp file (`0600`) written in target directory, atomically renamed, and permissions set per `dotfiles.rules`.
- **Zero Shell Exposure**: No secret values exported to `os.Environ` or passed through shell command lines.
- **Confinement**: Path confinement within `$HOME` and template conflict rules remain strictly enforced.
- **Backwards Compatibility**: All existing environment-variable templates continue to function identically without config changes.

---

## Completion Assessment

**Verdict: pass.** Implemented in PR #130 (merged as `7eae869`) and released in v0.31.0. Unit tests, lint, security scan, and BATS CI passed. A disposable macOS Keychain item was used to verify resolution, mode `0600`, doctor readiness, secret-masked drift/diff behavior, and cleanup without touching a user credential.

## Validation / Review

1. **Unit Tests (Cross-Platform)**:
   - `go test ./...` passes on Linux and macOS using `MockSecretResolver`.
   - Test invalid syntax (e.g. `{{keychain:}}`, `{{unknown:foo}}`) fails configuration validation.
   - Test missing keychain item error message contains only service name, no keys.
   - Test `plonk diff` masks secret values when comparing drifted secret templates.
2. **Integration Verification (macOS)**:
   - Add test credential via `security add-generic-password -U -s plonk-test-svc -a testuser -w "supersecretval"`.
   - Render template `test.tmpl` containing `key={{keychain:plonk-test-svc/testuser}}`.
   - Verify target file receives rendered secret with configured mode (`0600`).
   - Verify `plonk doctor` reports 0 missing variables.
   - Clean up test credential with `security delete-generic-password -s plonk-test-svc -a testuser`.
3. **CI / Linter Gates**:
   - `go tool golangci-lint run --timeout=10m` passes with 0 issues.
   - `make test-coverage` maintains normalized coverage standards.
