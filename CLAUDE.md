# Plonk Development Context

## Current Phase: Integration Testing Implementation (2025-08-04)

### Status
- **Unit test coverage**: 46.0% (represents all safely testable code)
- **Integration testing**: POC infrastructure working with testcontainers-go
- **JSON/YAML output**: Fixed - progress/status to stderr, structured data to stdout

### Recent Achievements
1. **Stderr output fix** (2025-08-04):
   - All progress messages now go to stderr
   - JSON/YAML output is clean on stdout
   - Follows industry standard (kubectl, docker, gh)
   - No breaking changes to existing functionality

2. **Integration test POC** (2025-08-04):
   - Docker containerized testing with testcontainers-go
   - First test (TestInstallPackage) working
   - Safe execution - no impact on developer machines
   - JSON output validation working correctly

### Integration Testing Strategy
- **Docker-only on dev machines** for safety
- **testcontainers-go** for container orchestration
- **JSON validation** for reliable assertions
- **Gradual expansion** from single test POC

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
**Overall Coverage**: 46.0% (up from 32.7% initially, 45.1% after first round, 45.9% after second round)

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
| commands | 17.6% | ✅ Improved via function extraction |
| cmd/plonk | 0% | Cannot test main() |

#### Test Philosophy
- **Safety First**: NO tests may modify system state
- **Business Logic**: All pure functions and utilities tested
- **System Operations**: Documented as intentionally untested
- **Coverage Target**: 50% was aspirational; 45.9% represents all safely testable code with simple extractions

### Build & Release
- **CI/CD**: GitHub Actions with Go 1.23/1.24 matrix testing
- **Release**: GoReleaser with macOS signing/notarization
- **Distribution**: Homebrew via richhaase/homebrew-tap
- **Coverage**: Codecov integration for tracking

### Known Limitations
- No native Windows support (use WSL)
- No package update command (use uninstall/install)
- Basic error messages (post-v1.0 enhancement)
