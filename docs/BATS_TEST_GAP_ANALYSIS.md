# BATS Test Suite Gap Analysis

## Executive Summary

This document analyzes the current BATS test suite coverage against all Plonk CLI commands and features. The goal is to identify gaps that could allow broken user flows to reach production.

**Current State:**

- 15 test files with ~131 test cases
- 8 package managers tested (brew, npm, gem, cargo, uv, pipx, pnpm, conda)
- Good coverage for basic operations, weaker coverage for edge cases and advanced features

**Key Findings:**

- Several commands have no tests at all (clone, doctor, config show)
- Command aliases are undertested (only `st` is tested)
- Install and upgrade dry-run flags are not tested
- Custom package manager configuration (V2 extensibility) is not tested

---

## Coverage Matrix

### Commands Overview

| Command             | Tested  | Coverage Level | Priority Gaps                       |
| ------------------- | ------- | -------------- | ----------------------------------- |
| `plonk` (root)      | Partial | Basic          | Version flag format validation      |
| `plonk install`     | Yes     | Good           | Dry-run, alias `i`, default manager |
| `plonk uninstall`   | Yes     | Good           | Alias `u`                           |
| `plonk upgrade`     | Yes     | Good           | Dry-run, pnpm manager               |
| `plonk packages`    | Partial | Basic          | Alias `p`                           |
| `plonk add`         | Yes     | Good           | Symlink handling                    |
| `plonk rm`          | Yes     | Moderate       | Directory removal (has bug comment) |
| `plonk apply`       | Yes     | Good           | Selective file apply edge cases     |
| `plonk dotfiles`    | Partial | Basic          | Alias `d`                           |
| `plonk diff`        | Partial | Basic          | Specific file argument              |
| `plonk status`      | Yes     | Good           | (alias `st` tested)                 |
| `plonk clone`       | **No**  | None           | Entire command untested             |
| `plonk config show` | **No**  | None           | Entire command untested             |
| `plonk config edit` | Skip    | N/A            | Interactive - skip per discussion   |
| `plonk doctor`      | **No**  | None           | Entire command untested             |

### Package Manager Coverage

| Manager     | Install | Uninstall | Upgrade | Status  | Notes                      |
| ----------- | ------- | --------- | ------- | ------- | -------------------------- |
| brew        | ✅      | ✅        | ✅      | ✅      | Complete                   |
| npm         | ✅      | ✅        | ✅      | ✅      | Complete                   |
| gem         | ✅      | ✅        | ✅      | ✅      | Complete                   |
| cargo       | ✅      | ✅        | ✅      | ✅      | Complete                   |
| uv          | ✅      | ✅        | ✅      | ✅      | Complete                   |
| pipx        | ✅      | ✅        | ✅      | ✅      | Complete                   |
| pnpm        | ✅      | ✅        | ❌      | Partial | Missing upgrade test       |
| conda       | ✅      | ✅        | ✅      | ✅      | Complete                   |
| custom (go) | ❌      | ❌        | ❌      | ❌      | Test user-defined managers |

---

## Detailed Gap Analysis

### 1. CRITICAL: Untested Commands

#### 1.1 Clone Command (`plonk clone`)

**Priority: High**

The clone command is essential for new user onboarding. It:

- Clones a git repository
- Reads plonk.lock to detect required managers
- Runs apply to set up the system

**Missing Tests:**

- Basic clone from local git repo
- Clone with various URL formats (user/repo, https://, git@)
- Clone with existing PLONK_DIR (should fail/warn)
- Clone dry-run mode
- Clone with invalid/non-existent repo
- Clone with lock file containing packages

**Implementation Note:** Use local bare git repo for testing (no network dependency).

#### 1.2 Doctor Command (`plonk doctor`)

**Priority: High**

The doctor command is the primary diagnostic tool. It:

- Checks system info
- Verifies package manager availability
- Validates config
- Shows install hints

**Missing Tests:**

- Basic doctor output with all managers available
- Doctor output with some managers missing
- Doctor with invalid config file
- Doctor shows correct install hints
- Doctor shows correct manager versions

#### 1.3 Config Show Command (`plonk config show`)

**Priority: Medium**

**Missing Tests:**

- Show default config (no plonk.yaml)
- Show merged config (defaults + overrides)
- Show annotations for user-defined values
- Show custom manager definitions

---

### 2. HIGH: Missing Flag/Option Coverage

#### 2.1 Install Dry-Run (`plonk install --dry-run`)

**Priority: High**

Currently tested for uninstall, apply, add, rm but NOT for install.

**Missing Tests:**

- `plonk install --dry-run brew:cowsay` shows what would happen
- Dry-run doesn't modify lock file
- Dry-run doesn't install package

#### 2.2 Upgrade Dry-Run (`plonk upgrade --dry-run`)

**Priority: High**

Not tested at all.

**Missing Tests:**

- `plonk upgrade --dry-run` shows what would happen
- Dry-run doesn't modify packages

#### 2.3 Install Alias (`plonk i`)

**Priority: Medium**

The `i` alias for install is not tested.

**Missing Tests:**

- `plonk i brew:cowsay` works like `plonk install`

#### 2.4 Other Aliases

**Priority: Medium**

| Command   | Alias | Tested |
| --------- | ----- | ------ |
| install   | i     | ❌     |
| uninstall | u     | ❌     |
| packages  | p     | ❌     |
| dotfiles  | d     | ❌     |
| status    | st    | ✅     |

---

### 3. MEDIUM: Default Manager Behavior

#### 3.1 Current State

All install tests use explicit `manager:package` syntax.

#### 3.2 Missing Tests

- Install with default manager (no prefix): `plonk install cowsay`
- Default manager from config
- Invalid default manager handling

---

### 4. MEDIUM: Custom Package Manager (V2 Config)

#### 4.1 Current State

No tests for user-defined package managers via plonk.yaml config.

#### 4.2 Missing Tests

Using `go` as the test case for custom manager:

- Define custom manager in plonk.yaml
- Install package with custom manager
- List packages with custom manager
- Upgrade package with custom manager
- Uninstall package with custom manager

**Example Config:**

```yaml
managers:
  go:
    binary: go
    list:
      command: [go, list, -m, -json, all]
      parse: json
    install:
      command: [go, install, "{{.Package}}@latest"]
    uninstall:
      command: [go, clean, -i, "{{.Package}}"]
```

---

### 5. LOW: Edge Cases and Error Handling

#### 5.1 Already Covered

- Invalid package specs
- Non-existent packages
- Missing managers
- Empty lock file
- Invalid config YAML

#### 5.2 Missing Tests

- Concurrent plonk operations (lock file contention)
- Very long package names
- Special characters in package names
- Unicode in dotfile names
- Large dotfiles
- Binary dotfiles
- Dotfile permission preservation
- Symlink handling in dotfiles

---

### 6. LOW: Version and Help

#### 6.1 Current State

- Basic `plonk --version` tested
- Basic `plonk help <command>` tested

#### 6.2 Missing Tests

- Version format validation (dev vs release)
- Help for all commands
- Unknown command handling

---

## Recommended Test Additions

### Priority 1: Critical (Must Have)

| Test File                 | Tests to Add                                        | Est. Count |
| ------------------------- | --------------------------------------------------- | ---------- |
| `13-clone-command.bats`   | Clone from local repo, URL formats, dry-run, errors | 8          |
| `14-doctor-command.bats`  | Basic output, manager availability, install hints   | 6          |
| `15-config-show.bats`     | Default config, merged config, annotations          | 4          |
| `03-package-install.bats` | Add dry-run tests                                   | 3          |
| `08-package-upgrade.bats` | Add dry-run tests, pnpm upgrade                     | 3          |

### Priority 2: High (Should Have)

| Test File                 | Tests to Add                         | Est. Count |
| ------------------------- | ------------------------------------ | ---------- |
| `16-command-aliases.bats` | All command aliases (i, u, p, d)     | 4          |
| `17-custom-managers.bats` | User-defined manager via config (go) | 6          |
| `01-basic-commands.bats`  | Unknown command handling             | 2          |
| Various                   | Default manager behavior             | 3          |

### Priority 3: Medium (Nice to Have)

| Test File                 | Tests to Add              | Est. Count |
| ------------------------- | ------------------------- | ---------- |
| `18-version-formats.bats` | Version format validation | 2          |
| Various                   | Help for all commands     | 8          |

### Priority 4: Low (Future Consideration)

| Category    | Tests                               | Est. Count |
| ----------- | ----------------------------------- | ---------- |
| Edge cases  | Special chars, unicode, large files | 6          |
| Permissions | Dotfile permission preservation     | 2          |
| Symlinks    | Symlink handling in dotfiles        | 3          |

---

## Test Infrastructure Enhancements

### Required Changes

1. **Add local git repo helper** - For clone command testing
2. **Add pnpm to require_package_manager()** - Currently missing
3. **Go already available** - Docker container includes Go for custom manager testing

### Future Consideration: Parallelization

Tests could be parallelized by:

- Grouping tests by package manager (run brew tests parallel with npm tests)
- Separating read-only tests (status, packages, dotfiles, doctor)
- Using BATS `--jobs N` flag

This would require:

- Unique artifact names per test (avoid collisions)
- Per-manager safe package isolation
- Lock file for shared resources

**Recommendation:** Defer parallelization until coverage is complete.

---

## Summary Statistics

| Category         | Before       | After          | Improvement |
| ---------------- | ------------ | -------------- | ----------- |
| Commands Tested  | 10/15        | 14/15          | +4          |
| Test Cases       | ~131         | 223            | +92         |
| Manager Coverage | 8/8          | 9/9 (+ custom) | +1          |
| Dry-run Coverage | 4/7 commands | 7/7 commands   | +3          |
| Alias Coverage   | 1/5          | 5/5            | +4          |

---

## Implementation Status

All Priority 1 and Priority 2 items have been implemented:

### New Test Files Created

| File                        | Tests | Description                        |
| --------------------------- | ----- | ---------------------------------- |
| `13-clone-command.bats`     | 16    | Clone from local repo, dry-run    |
| `14-doctor-command.bats`    | 21    | Basic output, manager availability |
| `15-config-show.bats`       | 17    | Default config, merged config      |
| `16-command-aliases.bats`   | 12    | All aliases (i, u, p, d)           |
| `17-custom-managers.bats`   | 15    | Go as user-defined manager         |

### Existing Files Updated

| File                        | Added Tests | Changes                            |
| --------------------------- | ----------- | ---------------------------------- |
| `01-basic-commands.bats`    | 4           | Unknown command, version tests     |
| `03-package-install.bats`   | 5           | Install dry-run tests              |
| `08-package-upgrade.bats`   | 9           | Upgrade dry-run, pnpm tests        |

### Infrastructure Updates

- Added `pnpm` and `conda` to `require_package_manager()` helper
- Added safe dotfiles for clone testing

---

## Appendix: Safe List Updates

### Packages Added

```
# Already in safe list
go:github.com/rakyll/hey
```

### Dotfiles Added

```
# For clone testing
.plonk-test-clone-rc
.config/plonk-test-clone/config.yaml
```

---

_Generated: 2025-12-16_
_Updated: 2025-12-16 (implementation complete)_
_Branch: feat/enhance-bats-test-suite_
