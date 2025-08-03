# Plonk Development Context

## Current Phase: Test Coverage Improvement for v1.0 (2025-08-03)

### Status
All features and distribution improvements complete. Working to improve test coverage from 32.7% to 50% before v1.0 tag.

### v1.0 Release Checklist
- [x] Remove stability warning from README.md
- [x] Update version injection to "1.0.0"
- [ ] Improve test coverage to 50% (currently 32.7%)
- [ ] Create and push v1.0.0 tag
- [x] Verify release workflow (tested with v0.9.5)
- [x] Homebrew distribution with code signing/notarization

### Test Coverage Plan
- **Current**: 34.6% overall (up from 32.7%)
- **Target**: 50% for v1.0 (need +15.4%)
- **Strategy**: Focus on high-impact, low-risk improvements
- **Details**: See [Consolidated Test Improvement Plan](docs/planning/consolidated-test-improvement-plan.md)

### Recent Progress
- Created internal/testutil package with BufferWriter for testing
- Added Writer interface to output package for testability
- Achieved 80% coverage for output package (was 0%)
- All tests passing with no external command calls

### Key Coverage Findings
- Config package has 38.4% coverage (not 0% as initially thought)
- Dotfiles package has 50.3% coverage (not 0% as initially thought)
- No coverage reporting bugs exist
- Commands package (9.2%) is highest priority for improvement

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
- Unit tests for business logic only, no mocks for CLIs
- Integration tests in CI only to protect developer systems
- Existing CommandExecutor interface pattern for mocking

## Technical Details

### System Requirements
- **Go**: 1.23+ (required by tool dependencies)
- **Platforms**: macOS, Linux (including WSL)
- **Prerequisites**: Homebrew, Git

### Test Coverage
| Package | Current | Target | Priority | Status |
|---------|---------|--------|----------|---------|
| commands | 9.2% | 40% | **HIGH** | Pending |
| output | 80.0% | 80% | **HIGH** | ✅ Complete |
| clone | 0% | 30% | **MEDIUM** | Pending |
| diagnostics | 13.7% | 40% | **MEDIUM** | Pending |
| orchestrator | 0.7% | 40% | **LOW** | Pending |
| testutil | 100% | - | - | ✅ Complete |
| **Overall** | **34.6%** | **50%** | | In Progress |

### Build & Release
- **CI/CD**: GitHub Actions with Go 1.23/1.24 matrix testing
- **Release**: GoReleaser with macOS signing/notarization
- **Distribution**: Homebrew via richhaase/homebrew-tap
- **Coverage**: Codecov integration for tracking

### Known Limitations
- No native Windows support (use WSL)
- No package update command (use uninstall/install)
- Basic error messages (post-v1.0 enhancement)
