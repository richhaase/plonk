---
metadata:
  type: strategic-plan
  topic: ai-development-workflow
  created: 2025-08-12
  status: active

phases:
  - id: 1
    name: Foundation Commands
    status: ready_to_implement
    implementation_plan: docs/phase-1-ai-development-workflow-implementation-plan.md
    dependencies: []
  - id: 2
    name: Advanced State Management
    status: planned
    implementation_plan: null
    dependencies: [1]
  - id: 3
    name: Multi-Expert Review System
    status: planned
    implementation_plan: null
    dependencies: [1, 2]
  - id: 4
    name: Workflow Orchestration
    status: planned
    implementation_plan: null
    dependencies: [1, 2, 3]
  - id: 5
    name: Project Portability
    status: planned
    implementation_plan: null
    dependencies: [1, 2, 3, 4]
---

# Strategic Plan: AI-Assisted Development Workflow

**Date**: August 12, 2025
**Source**: docs/brainstorm-ai-development-workflow.md
**Architect**: Strategic Planning Agent
**Status**: Strategic Planning Complete

## Executive Summary

This plan outlines the implementation of a flexible, AI-agent-orchestrated development methodology via Claude Code slash commands. The system will provide reusable workflow commands that can adapt to different project types while maintaining intelligent guidance throughout the development process.

## Architectural Approach

### Core Design Philosophy
- **Reusable across projects**: Commands work in any codebase, not just plonk
- **State-aware workflow**: Documents track progress and handoffs between agents
- **Flexible entry points**: Workflow adapts to work size (brainstorm→architect→engineer→build→review OR direct engineer→build→review)
- **Expert-level review**: Trust-building through comprehensive assessment capabilities

### Integration with Existing Infrastructure
The workflow system leverages Claude Code's slash command infrastructure rather than building a separate system:
- Commands implemented as `.claude/commands/*.md` files
- Document state management through structured markdown files
- Context loading through file reading and codebase analysis
- Validation through prerequisite checking

### System Architecture

```
Claude Code Slash Commands (.claude/commands/)
├── brainstorm.md     → Pure idea exploration
├── architect.md      → Strategic planning (IMPLEMENTED)
├── engineer.md       → Detailed implementation planning
├── build.md          → Code generation and implementation
├── review.md         → Expert assessment and validation
└── workflow-status.md → Orchestration and state tracking
```

## Implementation Phases

### Phase 1: Foundation Commands - Status: partially_implemented
**Scope**: Implement core workflow commands with basic functionality
**Dependencies**: None
**Timeline**: 1-2 weeks

**Deliverables**:
- `/brainstorm` command (IMPLEMENTED)
- `/architect` command (IMPLEMENTED)
- `/engineer` command with phase detection (PARTIALLY IMPLEMENTED - missing actual planning logic)
- `/build` command with scoped context (PARTIALLY IMPLEMENTED - missing actual build logic)
- `/review` command with multi-expert modes (SHELL ONLY - no review logic implemented)

**Current Issues**:
- Commands have good validation logic but incomplete LLM instructions
- Missing context loading (files aren't read via `@filename` syntax)
- Weak prompt engineering - commands tell users what will happen, not Claude what to do
- No specific expert-mode instructions despite having mode detection

**Immediate Actions Needed**:
1. Add proper LLM instructions to `/review` command after validation
2. Include file loading using `@$REVIEW_FILE` syntax
3. Provide mode-specific expert instructions for programmer/architect/plan modes
4. Add structured output format requirements
5. Include relevant context files for comprehensive analysis

**Success Criteria** (Updated):
- Commands provide clear, actionable instructions to Claude
- Target files are properly loaded and analyzed
- Expert modes deliver specialized analysis as promised
- Output follows consistent, useful format structure

### Phase 2: Advanced State Management - Status: planned
**Scope**: Implement sophisticated document state tracking and agent coordination
**Dependencies**: Phase 1 complete
**Timeline**: 1 week

**Deliverables**:
- Document update permissions system
- Phase status tracking (planned, ready to implement, in progress, completed)
- Auto-detection of next available phases
- Validation of prerequisites and dependencies

**Success Criteria**:
- Agents properly update document states
- Phase progression works automatically
- State conflicts are prevented
- Users can resume workflows across sessions

### Phase 3: Multi-Expert Review System - Status: planned
**Scope**: Implement the "badass" review agent with multiple expert modes
**Dependencies**: Phase 1-2 complete
**Timeline**: 1-2 weeks

**Deliverables**:
- Expert Programmer mode (code quality, patterns, best practices)
- Chief Architect mode (strategic alignment, system integration)
- Plan Reviewer mode (completeness, feasibility, risk assessment)
- Context-aware review based on deliverable type
- Trust-building feedback mechanisms

**Success Criteria**:
- Review agent provides expert-level assessments
- Identifies real problems and improvement opportunities
- Builds user trust through consistent quality
- Prevents agent overstep and scope creep

### Phase 4: Workflow Orchestration - Status: planned
**Scope**: Add meta-commands for workflow management and status tracking
**Dependencies**: Phase 1-3 complete
**Timeline**: 1 week

**Deliverables**:
- `/workflow-status` command showing all documents and phases
- `/context` command for loading agent-specific context
- `/handoff` command for explicit agent transitions
- Workflow visualization and progress tracking

**Success Criteria**:
- Users can easily understand workflow state
- Context loading is efficient and comprehensive
- Handoffs between agents are clean and documented
- Workflow visualization aids decision-making

### Phase 5: Project Portability - Status: planned
**Scope**: Ensure commands work seamlessly across different project types
**Dependencies**: Phase 1-4 complete
**Timeline**: 2 weeks

**Deliverables**:
- Auto-detection of project structure and conventions
- Language-specific architectural guidance
- Configurable templates and output formats
- Integration with existing development tools

**Success Criteria**:
- Commands adapt to Go, Python, JavaScript, etc. projects
- Architecture advice is language and framework appropriate
- Templates can be customized per project
- Integration with common tools (git, CI/CD, etc.)

## Risk Analysis and Mitigation

### Current High-Risk Items

**CRITICAL RISK**: Incomplete LLM prompt engineering in commands
- **Impact**: Commands provide validation but lack clear instructions for Claude
- **Current Status**: HIGH - `/review` validates files but doesn't instruct Claude what to do
- **Mitigation**: Add proper LLM instructions and context loading to all commands
- **Contingency**: Simplify commands to basic prompts if complex context loading fails

**Risk**: Missing context loading in slash commands
- **Impact**: Commands don't load the files they're supposed to analyze
- **Current Status**: HIGH - commands validate file existence but don't use `@filename` syntax
- **Mitigation**: Add file loading and relevant context to each command
- **Contingency**: Use simpler file reading approaches if advanced context loading fails

**Risk**: Document state management complexity
- **Impact**: Workflow becomes unreliable or confusing
- **Current Status**: MEDIUM - not yet implemented
- **Mitigation**: Start with simple state tracking, iterate based on usage
- **Contingency**: Fall back to manual state management if automated system fails

**Risk**: Agent scope creep
- **Impact**: Agents make decisions outside their intended domain
- **Current Status**: LOW - commands currently do nothing
- **Mitigation**: Strict context scoping, clear agent boundaries
- **Contingency**: Add explicit scope validation and user confirmation steps

### Medium-Risk Items

**Risk**: File naming convention conflicts
- **Impact**: Documents overwrite each other or become hard to find
- **Mitigation**: Clear naming conventions, validation of file paths
- **Contingency**: Add versioning or backup mechanisms

**Risk**: Context loading performance
- **Impact**: Commands become slow with large codebases
- **Mitigation**: Incremental context loading, caching strategies
- **Contingency**: Reduce context scope if performance becomes problematic

## Success Metrics

### Technical Metrics
- All commands execute without errors in different project types
- Document state transitions work reliably
- Context loading completes within reasonable time limits
- File naming conflicts are prevented

### User Experience Metrics
- Users can complete full workflows without manual intervention
- Workflow state is always clear and understandable
- Agent handoffs feel natural and helpful
- Review feedback leads to measurable improvements

### Quality Metrics
- Review agent catches real issues and improvement opportunities
- Generated plans are actionable and comprehensive
- Architecture decisions align with best practices
- Code quality improves through workflow usage

## Technical Implementation Details

### Document Schema Design
```yaml
# Strategic Plan Document
metadata:
  type: strategic-plan
  topic: <topic-name>
  created: <timestamp>
  status: active

phases:
  - id: 1
    name: <phase-name>
    status: planned|ready_to_implement|in_progress|completed
    implementation_plan: docs/phase-1-<topic>-implementation-plan.md
    dependencies: []

# Implementation Plan Document
metadata:
  type: implementation-plan
  phase: 1
  topic: <topic-name>
  status: planned|ready_to_implement|in_progress|completed

build_results:
  files_changed: []
  findings: []
  deviations: []
```

### Command Integration Strategy
- Leverage existing plonk patterns for CLI structure
- Use plonk's output formatting system for consistent UX
- Follow plonk's validation and error handling approaches
- Extend plonk's configuration system for workflow settings

### Context Loading Architecture
```
Agent Context Loading:
├── Project Structure Analysis
│   ├── Language/framework detection
│   ├── Key file identification
│   └── Pattern recognition
├── Workflow Document Loading
│   ├── Strategic plans
│   ├── Implementation plans
│   └── Previous build results
└── Codebase Analysis
    ├── Architecture patterns
    ├── Existing interfaces
    └── Recent changes
```

## Immediate Engineering Priorities

### Critical Issues to Address:

1. **Fix `/review` Command Prompt Engineering**
   - Add LLM instructions after the bash validation section
   - Include file loading using `@$REVIEW_FILE` syntax
   - Provide mode-specific expert instructions for programmer/architect/plan modes
   - Define structured output format requirements

2. **Enhance Other Command Prompts**
   - Add proper LLM instructions to `/engineer`, `/build`, and `/brainstorm`
   - Include relevant context loading in each command
   - Provide clear role definitions and output expectations

3. **Add Document State Management via Frontmatter**
   - Commands should read and update YAML frontmatter in workflow documents
   - Track phase progression and status in document metadata
   - Use Claude's file editing capabilities to update documents

4. **Create Command Workflow Integration**
   - Commands should reference and hand off to related documents
   - Include cross-references to strategic plans and implementation plans
   - Validate prerequisites through document state checking

### Development Approach:
1. **Start with `/review` prompt engineering** - most critical for user trust
2. **Add proper LLM instructions to other commands**
3. **Implement document state management via frontmatter**
4. **Test workflow end-to-end with plonk project**
5. **Iterate based on real usage feedback**

## Engineering Handoff Context

The engineer should understand:
- This system extends Claude Code slash commands, not plonk specifically
- Commands should be project-agnostic but can use plonk as test case
- State management is critical but should start simple
- Review agent is the most complex component and should be built last
- User experience and trust-building are primary objectives

**URGENT: Phase 1 needs immediate prompt engineering completion. Current implementation has good validation but lacks LLM instructions. Priority should be completing the `/review` command prompt, then enhancing other command prompts.**

## Updated Implementation Strategy

### Minimum Viable Workflow (Week 1 Priority)
1. **Complete `/review` command prompt** with proper LLM instructions
2. **Basic document state tracking** in YAML frontmatter
3. **Working command integration** through document references
4. **Clear value delivery** through well-engineered prompts

### Success Metrics (Revised)
- `/review` command provides comprehensive, expert-level analysis
- Commands properly load context and provide clear instructions to Claude
- Document state is tracked and updated through frontmatter
- Workflow commands integrate smoothly with proper handoffs
