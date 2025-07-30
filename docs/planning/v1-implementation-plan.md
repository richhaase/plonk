# v1.0 Implementation Plan

## Current Status (2025-07-30)
- Phase 1 started
- 1 of 7 tasks complete (14%)
- .plonk/ directory exclusion ✅

## Remaining Tasks Overview

### Phase 1: Foundation (Week 1) - IN PROGRESS

#### Task: Progress Indicators (Next Priority)
**Effort**: 1-2 days
**Priority**: High
**Description**: Add periodic status output during long-running operations

**Key Operations Needing Progress**:
1. `plonk install` - Show "Installing package X of Y..."
2. `plonk apply` - Show progress through packages and dotfiles
3. `plonk search` - Show which managers are being searched
4. `plonk clone` - Show clone progress, apply progress

**Implementation Approach**:
- Simple counter-based approach (X of Y)
- Update existing output methods
- No fancy progress bars for v1.0
- Consider: Should we show time elapsed?

**Questions for User**:
1. Format preference: "Installing htop (2 of 5)..." vs "Installing 2/5: htop"?
2. Update frequency: Every item or batch updates?
3. Show time elapsed or ETA?
4. Silent mode flag needed?

#### Task: Doctor Code Consolidation
**Effort**: 1-2 days
**Priority**: Medium
**Description**: Extract shared health check logic from clone command

**Current Duplication**:
- Clone command shells out to `plonk doctor --fix`
- Could use internal functions directly
- Both check package managers, PATH, etc.

**Benefits**:
- Reduce code duplication
- Consistent behavior
- Easier testing
- Better error handling

**Questions for User**:
1. Should clone continue to use `plonk doctor --fix` or internal functions?
2. Any specific consolidation priorities?

### Phase 2: Core Features (Week 2)

#### Task: APT Package Manager Support
**Effort**: 3-5 days
**Priority**: High (Linux requirement)
**Description**: Add apt package manager for Debian/Ubuntu systems

**Key Challenges**:
1. **Sudo handling** - How to handle password prompts?
2. **Package naming** - apt uses different names than brew
3. **Update cache** - When to run `apt update`?
4. **Non-interactive mode** - CI/CD considerations

**Design Questions for User**:
1. Sudo approach:
   - Check sudo upfront and fail if needed?
   - Pass through sudo prompts?
   - Require pre-authenticated sudo?
2. Package name mapping:
   - Simple 1:1 mapping?
   - Config file for mappings?
   - Accept both brew and apt names?
3. Cache updates:
   - Always run apt update first?
   - Only if install fails?
   - User flag to control?

#### Task: Dotfile Drift Detection
**Effort**: 2-3 days
**Priority**: TOP (User's critical gap)
**Description**: Show when deployed dotfiles differ from source

**Requirements**:
- Add "out-of-sync" state to reconciliation
- Show in `plonk status` output
- Clear indication of what `plonk apply` will change

**Implementation Approach**:
1. Compare file contents (not just existence)
2. Show diff summary (not full diff)
3. New state: "out-of-sync" (in addition to managed/missing/unmanaged)

**Design Questions for User**:
1. Comparison approach:
   - Checksum comparison (fast)?
   - Byte-by-byte comparison?
   - Ignore whitespace changes?
2. Status output:
   - New column showing sync state?
   - Separate section for out-of-sync files?
   - Show count in summary?
3. Detail level:
   - Just show "out-of-sync" flag?
   - Show last modified times?
   - Show size differences?

### Phase 3: Polish & Release (Week 3)

#### Task: Linux Platform Testing
**Effort**: 2-3 days
**Priority**: High
**Description**: Test on major Linux distributions

**Test Matrix**:
- Ubuntu 22.04 LTS
- Ubuntu 24.04 LTS
- Fedora (latest)
- Arch Linux
- Debian (stable)

**Test Areas**:
1. Installation methods
2. Package manager detection
3. PATH configuration
4. Dotfile paths
5. Default directories

**Questions for User**:
1. Priority distributions?
2. Minimum versions to support?
3. WSL-specific testing needed?

#### Task: Documentation Updates
**Effort**: 1-2 days
**Priority**: Medium
**Description**: Update all documentation for v1.0

**Required Updates**:
1. Remove all "setup" command references
2. Update installation guide with current binaries
3. Document .plonk/ directory
4. Review/update stability warnings
5. Add Linux-specific sections

**Questions for User**:
1. Keep stability warning until v1.0 release?
2. Separate Linux installation guide?
3. Add troubleshooting guide?

## Execution Strategy

### Week 1 Remaining (Days 2-5)
- Day 2-3: Progress Indicators
- Day 4-5: Doctor Code Consolidation

### Week 2 (Days 6-10)
- Day 6-8: APT Package Manager
- Day 9-10: Dotfile Drift Detection

### Week 3 (Days 11-15)
- Day 11-13: Linux Platform Testing
- Day 14-15: Documentation Updates
- Day 15: Final review and v1.0 prep

## Open Questions Summary

1. **Progress Indicators**: Format, frequency, time display?
2. **Doctor Consolidation**: Internal functions vs shell out?
3. **APT Support**: Sudo handling, package mapping, cache strategy?
4. **Drift Detection**: Comparison method, output format, detail level?
5. **Linux Testing**: Priority distros, minimum versions?
6. **Documentation**: Stability warning, separate guides?

## Next Steps

1. Get user input on open questions
2. Create detailed design for progress indicators
3. Start implementation based on priorities
4. Update this plan as decisions are made
