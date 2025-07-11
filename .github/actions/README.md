# Plonk Composite Actions

This directory contains reusable composite actions for the Plonk project's CI/CD workflows.

## Available Actions

### 🔧 setup-go-env
**Purpose**: Setup Go environment with dependencies and optional development tools

**Inputs**:
- `go-version` (optional): Go version to install (default: from go.mod)
- `install-homebrew` (optional): Install Homebrew on Linux (default: false)
- `install-just` (optional): Install Just task runner (default: false)
- `cache-key-suffix` (optional): Additional cache key suffix

**Usage**:
```yaml
- uses: ./.github/actions/setup-go-env
  with:
    install-homebrew: 'true'
    install-just: 'true'
```

### 🧪 run-tests
**Purpose**: Run Go tests with optional coverage and mocking

**Inputs**:
- `coverage` (optional): Enable coverage reporting (default: false)
- `coverage-format` (optional): Coverage format - 'ci' or 'html' (default: ci)
- `generate-mocks` (optional): Generate mocks before testing (default: false)
- `upload-codecov` (optional): Upload coverage to Codecov (default: false)

**Usage**:
```yaml
- uses: ./.github/actions/run-tests
  with:
    coverage: 'true'
    generate-mocks: 'true'
    upload-codecov: 'true'
```

### ✅ quality-checks
**Purpose**: Run linting, formatting, and security checks

**Inputs**:
- `run-linter` (optional): Run golangci-lint (default: true)
- `run-formatter` (optional): Run goimports formatter (default: true) 
- `run-security` (optional): Run security checks (default: false)
- `test-build` (optional): Test that the project builds (default: true)

**Usage**:
```yaml
- uses: ./.github/actions/quality-checks
  with:
    run-linter: 'true'
    run-security: 'true'
```

## Benefits

1. **Eliminates Duplication**: Common setup steps defined once
2. **Consistency**: Same environment setup across all workflows
3. **Maintainability**: Update logic in one place
4. **Flexibility**: Configurable inputs for different use cases
5. **Caching**: Optimized caching strategies built-in

## Development

When modifying these actions:

1. Test changes in a feature branch
2. Update this README if inputs/outputs change
3. Version the action if breaking changes are made
4. Consider backwards compatibility for existing workflows