# Phase 3.5: Comprehensive Code Reduction Analysis

## Objective
Perform a detailed analysis to identify concrete opportunities for reducing the codebase from ~14,300 LOC to closer to our 8,000 LOC target. This analysis will inform whether further reduction is feasible without compromising functionality.

## Context for Agent
You are analyzing a CLI tool called "plonk" that manages packages and dotfiles. The codebase has undergone significant refactoring:
- Phase 1: Reorganized into resource-focused architecture
- Phase 2: Introduced Resource abstraction for future extensibility
- Phase 3: Removed some abstractions but only achieved ~500 LOC reduction

We expected to reduce from ~14,800 to ~8,000 LOC but are still at ~14,300. Your task is to find where the remaining ~6,000 lines could be eliminated or consolidated.

## Current Architecture
```
internal/
├── commands/      (~4,000 LOC) - CLI command handlers
├── config/        (~650 LOC)   - Configuration management
├── lock/          (~300 LOC)   - Lock file handling
├── orchestrator/  (~1,000 LOC) - Coordination logic
├── output/        (~500 LOC)   - Output formatting
└── resources/     (~7,500 LOC)
    ├── packages/  (~5,000 LOC) - Package manager implementations
    └── dotfiles/  (~2,500 LOC) - Dotfile operations
```

## Analysis Tasks

### Task 1: Commands Package Analysis (2 hours)
**Analyze `internal/commands/` for reduction opportunities:**

1. **Identify duplicate code patterns:**
   ```bash
   # Find similar code blocks
   grep -r "if err != nil" internal/commands/ | wc -l
   grep -r "return fmt.Errorf" internal/commands/ | wc -l
   ```

2. **Analyze each command file and document:**
   - Lines of boilerplate vs business logic
   - Duplicate error handling patterns
   - Similar flag parsing code
   - Output formatting duplication

3. **Look for consolidation opportunities:**
   - Can similar commands share more code?
   - Is there repeated validation logic?
   - Are there similar patterns in runX functions?

4. **Create a report section:**
   ```markdown
   ## Commands Package Analysis
   - Current: X LOC
   - Potential reduction: Y LOC
   - How: [specific strategies]
   ```

### Task 2: Package Managers Deep Dive (3 hours)
**Analyze `internal/resources/packages/` for commonalities:**

1. **Compare all 6 package managers side-by-side:**
   - List() implementations - how similar are they?
   - Install() implementations - common patterns?
   - Search() implementations - duplicate parsing?
   - Error handling - repeated patterns?

2. **Identify exactly what's unique per manager:**
   ```bash
   # For each manager, what's truly unique?
   diff internal/resources/packages/homebrew.go internal/resources/packages/npm.go
   ```

3. **Look for parsing duplication:**
   - Version parsing
   - Package name extraction
   - Output formatting
   - Command building

4. **Measure potential savings:**
   - If we extracted common patterns, how many lines saved?
   - Could we use a data-driven approach (config per manager)?
   - Is there a way to generate some code?

5. **Create detailed findings:**
   ```markdown
   ## Package Managers Analysis
   - Common patterns across managers: [list]
   - Unique per manager: [list]
   - Potential extraction: X LOC
   - Risks: [what functionality might break]
   ```

### Task 3: Dotfiles Package Investigation (2 hours)
**Examine `internal/resources/dotfiles/` for simplification:**

1. **Analyze the 2,500 LOC:**
   - What operations take the most code?
   - Is there defensive programming we don't need?
   - Are there features we could remove?

2. **Look for over-engineering:**
   - Complex path handling that could be simpler?
   - Validation that Go's type system handles?
   - Error cases that might never happen?

3. **Document findings:**
   ```markdown
   ## Dotfiles Analysis
   - Core functionality: X LOC
   - Safety/validation: Y LOC
   - Could be simplified: Z LOC
   ```

### Task 4: Feature Usage Analysis (1 hour)
**Identify features that could be removed:**

1. **Review all commands and ask:**
   - Which commands are essential vs nice-to-have?
   - Are there flags/options that complicate the code significantly?
   - What features were added but might not be used?

2. **Analyze output formats:**
   - Do we need table, JSON, AND YAML?
   - Could we standardize on fewer formats?

3. **Look at edge cases:**
   - How much code handles rare scenarios?
   - What validation could be removed?

### Task 5: Cross-Package Duplication (2 hours)
**Find duplication across package boundaries:**

1. **Use tools to find duplicate code:**
   ```bash
   # Install and run duplication detector
   go install github.com/mibk/dupl@latest
   dupl -threshold 15 ./internal/
   ```

2. **Manual pattern search:**
   - Error wrapping patterns
   - Context handling
   - String manipulation
   - Path handling

3. **Document all duplications found:**
   ```markdown
   ## Cross-Package Duplication
   - Pattern: [description]
     - Found in: [files]
     - Lines that could be saved: X
   ```

### Task 6: Architecture-Level Opportunities (2 hours)
**Think bigger - could the architecture change?**

1. **Question current structure:**
   - Do we need separate packages and dotfiles?
   - Could commands be more data-driven?
   - Is the orchestrator necessary?

2. **Consider alternative approaches:**
   - Could managers be configured rather than coded?
   - Would a plugin system be simpler?
   - Could we generate code from specifications?

3. **Document radical options:**
   ```markdown
   ## Architecture Alternatives
   - Option 1: [description, LOC savings, risks]
   - Option 2: [description, LOC savings, risks]
   ```

## Deliverable Format

Create `PHASE_3_5_ANALYSIS.md` with:

```markdown
# Phase 3.5: Code Reduction Analysis Report

## Executive Summary
- Current LOC: ~14,300
- Target LOC: ~8,000
- Gap: ~6,300 LOC
- Achievable reduction: [your estimate] LOC
- Recommended approach: [summary]

## Detailed Findings

### 1. Commands Package (4,000 LOC)
[Specific findings with line counts]

### 2. Package Managers (5,000 LOC)
[Detailed analysis with examples]

### 3. Dotfiles Package (2,500 LOC)
[Concrete reduction opportunities]

### 4. Other Opportunities
[Cross-cutting concerns, architecture changes]

## Prioritized Recommendations

1. **High Impact, Low Risk** (X LOC reduction)
   - [Specific action]
   - [Specific action]

2. **Medium Impact, Medium Risk** (Y LOC reduction)
   - [Specific action]
   - [Specific action]

3. **High Impact, High Risk** (Z LOC reduction)
   - [Specific action]
   - [Specific action]

## Conclusion
[Is 8,000 LOC achievable? If not, what's realistic?]
```

## Success Criteria
- [ ] Identified at least 3,000 LOC of potential reduction
- [ ] Provided specific, actionable recommendations
- [ ] Assessed risk vs reward for each recommendation
- [ ] Determined if 8,000 LOC target is realistic
- [ ] Created clear path forward for Phase 4

## Notes for Agent
- Be specific - vague suggestions aren't helpful
- Count actual lines that could be removed
- Consider maintenance burden of suggestions
- Think about code readability, not just line count
- Some duplication is OK if it makes code clearer
- Focus on biggest wins first
