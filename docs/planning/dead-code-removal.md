# Dead Code Removal Plan

**Created:** 2025-12-11
**Source:** `docs/planning/2025-12-09-review.md` (Category A dead code)
**Approach:** Incremental removal with validation between each step

---

## Validation Commands

```bash
# Unit tests (Claude can run)
go test ./...

# Lint check (Claude can run)
golangci-lint run

# Build verification (Claude can run)
go build ./...

# BATS tests (user runs - Docker required)
just test-bats
```

---

## Removal Steps

### Step 1: commands/helpers.go - Deprecated/Duplicate Functions ✅

**Target:** `internal/commands/helpers.go`

**Removed:**
- [x] `ParsePackageSpec` (lines 24-30) - deprecated, use `packages.ParsePackageSpec`
- [x] `IsValidManager` (lines 33-42) - unused
- [x] `GetValidManagers` (lines 45-48) - unused
- [x] `GetMetadataString` (lines 141-149) - unused exported version
- [x] Removed unused `packages` import
- [x] Removed tests for above functions from `helpers_test.go`

**NOT Removed (false positive from deadcode):**
- `getMetadataString` - actually used by `add.go`

**Validation:**
- [x] `go test ./...` - PASS
- [x] `golangci-lint run` - PASS (0 issues)
- [x] `go build ./...` - PASS
- [ ] User: `just test-bats`

---

### Step 2: commands/status.go - Unused Sort Functions ✅

**Target:** `internal/commands/status.go`

**Removed:**
- [x] `sortItems` (lines 209-214)
- [x] `sortItemsByManager` (lines 217-231)
- [x] Removed unused `sort` and `strings` imports
- [x] Deleted `status_test.go` (only contained tests for removed functions)

**Validation:**
- [x] `go test ./...` - PASS
- [x] `golangci-lint run` - PASS (0 issues)
- [x] `go build ./...` - PASS
- [ ] User: `just test-bats`

---

### Step 3: output/colors.go - Unused Color Helpers ✅

**Target:** `internal/output/colors.go`

**Removed:**
- [x] `Available()`, `Deployed()`, `Managed()`, `Valid()` (green status words)
- [x] `Invalid()`, `Missing()`, `NotAvailable()` (red status words)
- [x] `Drifted()`, `Unmanaged()` (yellow status words)
- [x] Updated `colors_test.go` to remove tests for deleted functions

**Kept (in use):**
- `Success()` - used by config_edit.go
- `ColorError()`, `ColorInfo()`, `ColorAdded()`, `ColorRemoved()` - used by config_formatter.go

**Validation:**
- [x] `go test ./...` - PASS
- [x] `golangci-lint run` - PASS (0 issues)
- [x] `go build ./...` - PASS
- [ ] User: `just test-bats`

---

### Step 4: output/dotfile_list_formatter.go - Entire File ✅

**Target:** `internal/output/dotfile_list_formatter.go`

**Removed:**
- [x] Deleted `dotfile_list_formatter.go` (entire formatter unused)
- [x] Deleted `dotfile_list_formatter_more_test.go` (tests for deleted code)

**Validation:**
- [x] `go test ./...` - PASS
- [x] `golangci-lint run` - PASS (0 issues)
- [x] `go build ./...` - PASS
- [ ] User: `just test-bats`

---

### Step 5: output/progress.go - ProgressUpdate

**Target:** `internal/output/progress.go`

**Remove:**
- [ ] `ProgressUpdate` (lines 8-15)

**Note:** Check if `StageUpdate` is also dead after this removal. If so, consider removing entire file.

**Validation:**
- [ ] `go test ./...`
- [ ] `golangci-lint run`
- [ ] `go build ./...`
- [ ] User: `just test-bats`

---

### Step 6: output/spinner.go - Unused Spinner Helpers

**Target:** `internal/output/spinner.go`

**Remove:**
- [ ] `Spinner.UpdateText` (lines 89-93)
- [ ] `WithSpinner` (lines 158-165)
- [ ] `WithSpinnerResult` (lines 168-179)

**Validation:**
- [ ] `go test ./...`
- [ ] `golangci-lint run`
- [ ] `go build ./...`
- [ ] User: `just test-bats`

---

### Step 7: output/utils.go - Unused Utilities

**Target:** `internal/output/utils.go`

**Remove:**
- [ ] `TruncateString` (lines 181-189)
- [ ] `FormatValidationError` (lines 194-196)
- [ ] `FormatNotFoundError` (lines 199-209)

**Validation:**
- [ ] `go test ./...`
- [ ] `golangci-lint run`
- [ ] `go build ./...`
- [ ] User: `just test-bats`

---

### Step 8: resources/types.go - Unused Result Methods

**Target:** `internal/resources/types.go`

**Remove:**
- [ ] `Result.Count` (lines 73-75)
- [ ] `Result.IsEmpty` (lines 78-80)
- [ ] `Result.AddToSummary` (lines 83-88)
- [ ] `CreateDomainSummary` (lines 177-192)

**Validation:**
- [ ] `go test ./...`
- [ ] `golangci-lint run`
- [ ] `go build ./...`
- [ ] User: `just test-bats`

---

### Step 9: dotfiles/manager.go - Unused Internal Helpers

**Target:** `internal/resources/dotfiles/manager.go`

**Remove:**
- [ ] `Manager.computeFileHash` (line 554)
- [ ] `Manager.createCompareFunc` (line 559)

**Validation:**
- [ ] `go test ./...`
- [ ] `golangci-lint run`
- [ ] `go build ./...`
- [ ] User: `just test-bats`

---

## Post-Removal Verification

After all steps complete:

- [ ] Run full deadcode analysis: `go run golang.org/x/tools/cmd/deadcode@latest -test=false ./...`
- [ ] Verify Category A is clear (only Category B/C should remain)
- [ ] Full BATS test suite
- [ ] Manual smoke test of key commands: `plonk status`, `plonk doctor`, `plonk apply --dry-run`

---

## Progress Tracking

| Step | Description | Status | Date |
|------|-------------|--------|------|
| 1 | helpers.go deprecated/duplicate | ✅ Complete | 2025-12-11 |
| 2 | status.go sort functions | ✅ Complete | 2025-12-11 |
| 3 | colors.go unused helpers | ✅ Complete | 2025-12-11 |
| 4 | dotfile_list_formatter.go | ✅ Done (awaiting BATS) | 2025-12-11 |
| 5 | progress.go ProgressUpdate | Pending | |
| 6 | spinner.go unused helpers | Pending | |
| 7 | utils.go unused utilities | Pending | |
| 8 | types.go Result methods | Pending | |
| 9 | manager.go internal helpers | Pending | |

---

## Notes

- Each step should be a separate commit for easy rollback
- If any validation fails, investigate before proceeding
- Category B (test-only wrappers) deferred to separate cleanup effort
