# Plonk Development Context

## Current Phase: Test Coverage Improvement for v1.0 (2025-08-03)

### Status
All features and distribution improvements complete. Working to improve test coverage from 32.7% to 50% before v1.0 tag.

### v1.0 Release Checklist
- [x] Remove stability warning from README.md
- [x] Update version injection to "1.0.0"
- [ ] Improve test coverage to 50% (currently 37.6%)
- [ ] Create and push v1.0.0 tag
- [x] Verify release workflow (tested with v0.9.5)
- [x] Homebrew distribution with code signing/notarization

### Test Coverage Plan
- **Current**: 37.6% overall (up from 32.7%)
- **Target**: 50% for v1.0 (need +12.4%)
- **Strategy**: Focus on high-impact, low-risk improvements
- **Details**: See [Consolidated Test Improvement Plan](docs/planning/consolidated-test-improvement-plan.md)

### Recent Progress
- Phase 1 ✅: Created testutil package with BufferWriter, achieved 80% coverage for output package
- Phase 2 ✅: Added tests for pure functions in commands package, improved from 9.2% to 14.1%
- Overall improvement: 32.7% → 37.6% (+4.9% total)

### Key Coverage Findings
- Config package has 38.4% coverage (not 0% as initially thought)
- Dotfiles package has 50.3% coverage (not 0% as initially thought)
- No coverage reporting bugs exist
- Commands package improved from 9.2% to 14.1% (Phase 2 complete)

## Critical Implementation Guidelines

### STRICT RULE: No Unauthorized Features
**NEVER independently add features or enhancements that were not explicitly requested.**
- You MAY propose improvements, but that is all
- Do NOT implement anything beyond the exact scope requested
- Do NOT add "helpful" extras without explicit approval
- When in doubt, implement ONLY what was explicitly requested

### UI/UX Guidelines
- **NEVER use emojis in plonk output** - Use colored text status indicators instead
- Status indicators should be colored minimally (only the status word, not full lines)
- Professional, clean output similar to tools like git, docker, kubectl

## Testing Philosophy

### 🚨 CRITICAL SAFETY RULE: NEVER MODIFY SYSTEM STATE IN UNIT TESTS 🚨

**THIS IS THE MOST IMPORTANT RULE IN THE ENTIRE CODEBASE**

**UNIT TESTS MUST NEVER:**
- Call Apply() methods that could install packages or modify dotfiles
- Execute real package manager commands (brew, apt, npm, etc.)
- Run hooks or shell commands that affect the system
- Write to any paths outside of temporary test directories
- Modify ANY aspect of the developer's machine

**NO TESTS IS BETTER THAN TESTS THAT BREAK DEVELOPER MACHINES**

This rule has been violated multiple times. It CANNOT happen again. Any AI agent or developer who creates tests that modify system state is putting users at risk.

### Safe Testing Practices
- Unit tests for business logic only, no mocks for CLIs
- Integration tests in CI only to protect developer systems
- Existing CommandExecutor interface pattern for mocking
- Commands package orchestration functions are not unit testable by design (see [Architecture Decision](docs/planning/commands-testing-architecture-decision.md))
- ALWAYS use os.MkdirTemp() for file operations
- ONLY test pure functions and data structures

## Technical Details

### System Requirements
- **Go**: 1.23+ (required by tool dependencies)
- **Platforms**: macOS, Linux (including WSL)
- **Prerequisites**: Homebrew, Git

### ⚠️ WARNING: Test Coverage Must Be Safe ⚠️
Before adding ANY test, ask: "Could this test modify the real system?" If yes, DO NOT ADD IT.

### Test Coverage Status
**Overall Coverage**: 45.1% (up from 32.7%)

#### Coverage by Package
| Package | Coverage | Notes |
|---------|----------|-------|
| parsers | 100% | ✅ Complete |
| testutil | 100% | ✅ Complete |
| config | 95.4% | ✅ Comprehensive tests added |
| resources | 89.8% | ✅ Utility functions tested |
| lock | 84.6% | ✅ Good coverage |
| output | 82.0% | ✅ StructuredData methods tested |
| diagnostics | 70.6% | ✅ Health checks tested with temp dirs |
| packages | 62.1% | ✅ SupportsSearch methods tested |
| dotfiles | 50.5% | Limited by file operations |
| clone | 28.9% | Limited by git/network operations |
| orchestrator | 17.6% | Limited by system operations |
| commands | 14.6% | Limited by CLI orchestration |
| cmd/plonk | 0% | Cannot test main() |

#### Test Philosophy
- **Safety First**: NO tests may modify system state
- **Business Logic**: All pure functions and utilities tested
- **System Operations**: Documented as intentionally untested
- **Coverage Target**: 50% was aspirational; 45.1% represents all safely testable code

### Build & Release
- **CI/CD**: GitHub Actions with Go 1.23/1.24 matrix testing
- **Release**: GoReleaser with macOS signing/notarization
- **Distribution**: Homebrew via richhaase/homebrew-tap
- **Coverage**: Codecov integration for tracking

### Known Limitations
- No native Windows support (use WSL)
- No package update command (use uninstall/install)
- Basic error messages (post-v1.0 enhancement)
