# v1.0 Implementation Workflow

## Process Overview

For each piece of work in the v1.0 implementation, we follow this 8-step process:

### 1. Review Work To Be Done
- Check v1-implementation-plan.md for remaining tasks
- Review current phase and priorities
- Identify dependencies or blockers

### 2. Select Next Work Body
- Choose the next task based on:
  - Current phase (Foundation → Core → Polish)
  - Priority (High → Medium → Low)
  - Dependencies (unblocked tasks first)
  - User guidance

### 3. Plan and Document Work
- Create detailed implementation plan including:
  - Specific files to modify
  - Functions/methods to add or change
  - Test cases to add
  - Expected behavior changes
  - Edge cases to handle
- Document in the v1-implementation-plan.md or create specific plan doc

### 4. Receive User Review and Approval
- Present plan to user
- Discuss any concerns or alternatives
- Revise based on feedback
- Get explicit approval to proceed

### 5. Complete Work
- Implement the approved plan
- Write tests
- Handle edge cases
- Follow existing code patterns

### 6. Validate Completion of Plan
- Run tests to ensure functionality works
- Verify all plan items were addressed
- Check for regressions
- Test edge cases

### 7. Update All Relevant Documents
- Update CLAUDE.md progress tracking
- Update command documentation if behavior changed
- Update planning documents to mark complete
- Add any new learnings or decisions

### 8. Commit
- Stage all changes
- Write descriptive commit message
- Include what was implemented and why
- Run pre-commit hooks

### 9. Repeat
- Return to step 1 for next task

## Current Status

**Current Phase**: Phase 1 - Foundation (Week 1)
**Completed**: .plonk/ directory exclusion
**Next Task**: Progress Indicators

## Benefits of This Workflow

1. **Clear Planning**: Each task is well-defined before implementation
2. **User Alignment**: Approval before coding prevents rework
3. **Documentation**: Everything stays up-to-date
4. **Traceability**: Clear record of what was planned vs implemented
5. **Quality**: Validation ensures plan was fully executed

## Example Workflow Execution

### Example: Progress Indicators

1. **Review**: Need to add progress output to long operations
2. **Select**: Progress Indicators (1-2 days, High priority)
3. **Plan**:
   - Add progress to install/uninstall/apply/search
   - Use "X of Y" format
   - Update every item
   - Add to existing output functions
4. **User Review**: User approves plan, requests "2/5: package" format
5. **Complete**: Implement changes to output functions
6. **Validate**: Test all commands with multiple packages
7. **Update**: Mark complete in CLAUDE.md, update command docs
8. **Commit**: "feat: add progress indicators to long operations"
9. **Repeat**: Move to Doctor Code Consolidation

This workflow ensures systematic progress through v1.0 implementation with clear communication and documentation at each step.
