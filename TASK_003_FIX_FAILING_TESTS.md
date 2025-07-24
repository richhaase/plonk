# Task 003: Fix Failing Parser Tests

## Objective
Fix the failing unit tests in the managers package that were identified after the executor removal.

## Quick Context
- These are parsing tests that have incorrect expectations
- The tests themselves are good - they just need their expected values corrected
- No functional code changes needed, only test fixes

## Failing Tests to Fix

1. **cargo_test.go**
   - `TestCargoManager_parseInfoOutput/empty_output` - expects non-nil PackageInfo

2. **gem_test.go**
   - `TestGemManager_parseInfoOutput/gem_with_description_field` - Description field mismatch

3. **goinstall_test.go**
   - `TestParseModulePath` - Version should be "latest" not empty string
   - `TestGoInstallManager_Configuration` - UninstallArgs should be nil

4. **homebrew_test.go**
   - `TestHomebrewManager_parseSearchOutput/output_with_info_messages` - Extra item in results

5. **npm_test.go**
   - `TestNpmManager_parseListOutput/standard_npm_list_output` - Wrong order expected

6. **pip_test.go**
   - `TestPipManager_parseListOutput/malformed_JSON_falls_back_to_plain_text` - Wrong expectation
   - `TestPipManager_parseListOutputPlainText` - Multiple cases with wrong parsing expectations
   - `TestPipManager_parseInfoOutput/package_with_URL_field` - Homepage field mismatch

## Instructions

Fix each test by updating the expected values to match what the parser actually returns. Do NOT change the parser implementations - only fix the test expectations.

## Completion Report

Create `TASK_003_COMPLETION_REPORT.md` with:
- List of tests fixed
- Confirmation that all tests pass
- Any insights about why the tests were failing

## Success Criteria
- `go test ./internal/managers/...` passes with no failures
- No parser implementations were changed
