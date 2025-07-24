# Task 013: Standardize Output Formats (YAML + Human)

## Objective
Remove JSON output support and standardize on YAML (machine-readable) and human-readable table formats, aligning with the YAML-based configuration system.

## Quick Context
- **Current**: 3 formats supported (table, json, yaml) in output.go
- **User Request**: Drop JSON, keep YAML + human-readable for consistency with YAML config
- **Impact**: ~16 files use ParseOutputFormat/RenderOutput functions
- **Rationale**: YAML config → YAML machine output makes more sense than JSON

## Work Required

### Phase 1: Update Output Format Constants
1. **Remove OutputJSON constant** from output.go:20
2. **Update ParseOutputFormat()** to reject "json" format (output.go:49-61)
3. **Update RenderOutput()** to remove JSON case (output.go:31-47)
4. **Update error messages** to reference only "table" and "yaml"

### Phase 2: Update All Command Help Text
1. **Find all --output flag descriptions** that mention JSON
2. **Update help text** to show only "table|yaml" options
3. **Update command long descriptions** that reference JSON output
4. **Verify flag validation** uses new format restrictions

### Phase 3: Update Documentation References
1. **Search for JSON output examples** in comments
2. **Update any CLI usage examples** in code comments
3. **Update config command descriptions** that mention JSON
4. **Ensure consistency** in format naming (table/yaml only)

### Phase 4: Test Output Format Changes
1. **Verify format validation** rejects "json" with clear error
2. **Test YAML output** still works correctly
3. **Test table output** (default) still works
4. **Verify automation use cases** work with YAML

## Files to Update (16 total)
**Commands with output formats**:
- config_show.go, config_validate.go, search.go, ls.go, env.go
- dotfiles.go, status.go, doctor.go, info.go, sync.go
- rm.go, add.go, uninstall.go, install.go
- shared.go, output.go

## Rationale
- **Consistency**: YAML config files → YAML machine output
- **Simplicity**: 2 formats instead of 3 reduces complexity
- **User Experience**: YAML is more human-readable than JSON anyway
- **Maintenance**: Less code to maintain, fewer test cases

## Preservation Requirements
- ✅ **YAML output functionality** fully preserved
- ✅ **Human-readable table output** (default) unchanged
- ✅ **All automation use cases** continue working with YAML
- ✅ **CLI interface compatibility** (just removes json option)

## Expected Benefits
- **Simplified codebase**: Remove JSON encoding/handling code
- **Consistent output philosophy**: YAML everywhere for structured data
- **Reduced maintenance**: Fewer output format branches to test
- **Clearer user experience**: Less choice paralysis (table vs yaml)

## Breaking Changes Assessment
- **Minor breaking change**: `--output json` will now error
- **Migration path**: Users should use `--output yaml` instead
- **Impact**: Low (YAML and JSON are easily convertible)

## Success Criteria
1. ✅ **JSON format completely removed** from ParseOutputFormat
2. ✅ **All help text updated** to show only table/yaml options
3. ✅ **Error messages updated** for invalid format attempts
4. ✅ **All 16 command files verified** for format consistency
5. ✅ **YAML output still works** for all automation use cases
6. ✅ **Clear error on json attempts** with migration guidance

## Dependencies
- **No conflicts with other tasks**: Output format is independent
- **Can proceed immediately**: Simple, focused change

## Completion Report
Create `TASK_013_COMPLETION_REPORT.md` with:
- **List of all files updated** with specific changes
- **Before/after help text examples** for key commands
- **Verification of YAML output** for automation scenarios
- **Error message examples** for invalid format attempts
- **Breaking change documentation** for users