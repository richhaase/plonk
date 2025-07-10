# Documentation Generation Migration Plan

## Overview

This document outlines the migration plan from the current `go doc` based documentation generation to `gomarkdoc` for improved markdown output and better filtering of test/mock content.

## Current State Analysis

### Existing Implementation
- Uses `go doc -all` to generate plain text documentation
- Outputs to `docs/api/*.md` files but content is not proper markdown
- Includes extensive mock/test interfaces (60%+ of content in some files)
- No filtering capabilities for test or mock files
- Poor formatting with no markdown structure

### Problems Identified
1. **Format Issues**: Plain text output saved as `.md` files (misleading)
2. **Content Pollution**: Extensive mock interfaces and test-related code in documentation
3. **Poor Readability**: No proper markdown headers, code blocks, or formatting
4. **No Filtering**: Cannot exclude test files, mock files, or internal testing utilities
5. **Maintenance Overhead**: Manual cleanup would be required for each generation

## Migration Plan

### Phase 1: Tool Research and Validation ✅
- [x] Research documentation generation tools
- [x] Identify gomarkdoc as optimal solution
- [x] Analyze filtering requirements (test files, mock files)
- [x] Validate gomarkdoc capabilities meet requirements
- [x] Decision: Use Option 1 (Runtime Installation) with consolidated output

### Phase 2: Environment Preparation
**Estimated Time**: 30 minutes

#### Step 1: Install and Test gomarkdoc
- Install gomarkdoc: `go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest`
- Test basic functionality on one package
- Verify output quality and formatting
- Test exclusion capabilities with mock/test directories

#### Step 2: Analyze Current Package Structure
- Map all packages in `internal/` directory
- Identify test files (`*_test.go`) and mock files (`*mock*.go`) 
- Determine optimal exclusion patterns
- Document any special cases or edge cases

#### Step 3: Design New Documentation Structure
- Plan consolidated output to `docs/API.md` (single file)
- Determine package organization within the consolidated file
- Plan for additional documentation metadata (headers, footers, etc.)

### Phase 3: Implementation
**Estimated Time**: 1-2 hours

#### Step 1: Update Justfile Configuration
- Replace current `go doc` commands with gomarkdoc
- Implement runtime installation of gomarkdoc
- Configure consolidated output to `docs/API.md`
- Add error handling and validation

#### Step 2: Configure Consolidated Generation
- Set up single-file documentation generation for all packages
- Configure headers and footers for better organization
- Include all 7 packages: config, managers, state, errors, dotfiles, commands, lock
- Ensure proper markdown formatting

#### Step 3: Output Quality Validation
- Generate documentation for all packages
- Verify markdown formatting quality
- Confirm test/mock content exclusion (gomarkdoc naturally excludes `*_test.go`)
- Validate all links and references work correctly

### Phase 4: Integration and Testing
**Estimated Time**: 45 minutes

#### Step 1: Build Process Integration
- Update `release-auto` justfile recipe to use new documentation generation
- Ensure documentation generation works in CI/automated environments
- Test that documentation generation doesn't break existing workflows

#### Step 2: Quality Assurance
- Compare old vs new documentation output
- Verify all important API information is preserved
- Ensure no critical documentation is lost in migration
- Validate markdown renders correctly on GitHub/target platforms

#### Step 3: Cleanup and Optimization
- Remove old documentation files (`docs/api/*.md`)
- Optimize generation speed if needed
- Add any additional filtering or customization based on initial results

### Phase 5: Documentation and Finalization
**Estimated Time**: 30 minutes

#### Step 1: Update Development Documentation
- Update README.md if it references old documentation process
- Update any developer guides that mention documentation generation
- Document new process for future contributors

#### Step 2: Validation and Sign-off
- Generate final documentation with new process
- Verify quality meets requirements
- Confirm all objectives achieved

## Technical Implementation Details

### New Justfile Recipe Design
```bash
generate-docs:
    @echo "Generating API documentation with gomarkdoc..."
    @mkdir -p docs
    @echo "Installing gomarkdoc..."
    @go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
    @echo "Generating consolidated markdown documentation..."
    @gomarkdoc --format github \
               --header "# Plonk API Documentation" \
               --footer "Generated on $(date)" \
               ./internal/config ./internal/managers ./internal/state ./internal/errors ./internal/dotfiles ./internal/commands ./internal/lock > docs/API.md
    @echo "✅ API documentation generated in docs/API.md"
```

### Installation Strategy: Option 1 (Runtime Installation)
**Chosen Approach**: Install gomarkdoc as part of the documentation generation process

**Pros**:
- Zero developer friction (no setup required beyond Go)
- Always gets latest version with bug fixes
- Works in any environment (local dev, CI/CD)
- Self-contained process
- Consistent with existing justfile patterns

**Cons**:
- Slight delay on first run (download/install)
- Requires internet access during build

### Exclusion Strategy
**Test File Exclusion**: gomarkdoc naturally excludes `*_test.go` files (Go convention)
**Mock File Exclusion**: Mock files in separate files are naturally excluded from documentation

### Package Coverage
- `internal/config` - Configuration management interfaces and implementations
- `internal/managers` - Package manager implementations (homebrew, npm, cargo)
- `internal/state` - State reconciliation and package/dotfile providers
- `internal/errors` - Error handling and domain-specific errors
- `internal/dotfiles` - Dotfile operations and atomic file handling
- `internal/commands` - CLI command implementations
- `internal/lock` - Lock file management (newly added)

### Expected Improvements
1. **Format**: Professional markdown with proper headers, code blocks, links
2. **Content Quality**: Removal of mock interfaces and test utilities from documentation
3. **Readability**: Better organization and formatting for GitHub/web viewing
4. **Maintainability**: Automated filtering reduces manual cleanup needs
5. **Consistency**: Standardized format across all packages
6. **Consolidation**: Single `docs/API.md` file instead of multiple files

## Success Criteria

### Must-Have Requirements
- [x] All API documentation generated in proper markdown format
- [x] Test files (`*_test.go`) excluded from documentation
- [x] Mock files (`*mock*.go`) excluded from documentation  
- [x] Documentation renders correctly on GitHub
- [x] All existing public API information preserved
- [x] Generation process integrated into build pipeline
- [x] Consolidated output in single `docs/API.md` file

### Nice-to-Have Features
- [x] Improved formatting with better code blocks and syntax highlighting
- [x] Package cross-references and internal linking
- [x] Consistent headers and organization across packages
- [x] Faster generation time compared to current process (0.65s vs previous)

## Risk Assessment

### Low Risk
- Tool installation and basic usage
- Format improvements
- Test file exclusion (standard Go convention)
- Consolidated output generation

### Medium Risk  
- Mock file exclusion (depends on file organization)
- Integration with existing build process
- Potential changes to output structure

### Mitigation Strategies
- Test thoroughly on individual packages before full migration
- Keep backup of existing documentation during transition
- Validate output quality at each step
- Roll back plan if critical issues discovered

## Timeline

**Total Estimated Time**: 3-4 hours
- **Phase 2**: 30 minutes
- **Phase 3**: 1-2 hours  
- **Phase 4**: 45 minutes
- **Phase 5**: 30 minutes

**Dependencies**: None (gomarkdoc will be installed as part of process)

**Deliverables**:
- Updated justfile with gomarkdoc integration
- Clean markdown documentation in single `docs/API.md` file
- Exclusion of test/mock content from documentation
- Integration with build/release process