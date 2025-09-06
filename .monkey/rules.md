# Project Development Rules

Rules below are normative and use RFC 2119 language (MUST, SHOULD, MAY, etc.).
Edit or remove rules with care; each rule is delimited for easy maintenance.

<!-- RULE-BEGIN: 2025-08-28T00:00:00Z no-emojis -->
## No Emojis in Output
**Level:** MUST NOT
**Directive:** Code MUST NOT include any emojis in output, comments, commit messages, or any generated text.
**Scope:** All plonk command output, code files, and communication
**Rationale:** Professional tools maintain clean, text-based output without decorative characters

<!-- RULE-END: no-emojis -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:01Z exact-scope-only -->
## Exact Scope Implementation
**Level:** MUST
**Directive:** Implementations MUST include only the exact features requested with no additional enhancements.
**Scope:** All feature development and bug fixes
**Rationale:** Unrequested features are bugs that waste time, complicate reviews, and violate user trust
**Exceptions:** Safety-critical error handling may be added if the absence would cause system harm

<!-- RULE-END: exact-scope-only -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:02Z no-system-modification-tests -->
## Safe Unit Testing
**Level:** MUST NOT
**Directive:** Unit tests MUST NOT modify the host system through package installations, file system changes outside temp directories, or system command execution.
**Scope:** All unit tests in test files
**Rationale:** Tests that modify developer systems put machines at risk and violate the principle of safe testing

<!-- RULE-END: no-system-modification-tests -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:03Z prefer-edit-over-create -->
## File Creation Restriction
**Level:** SHOULD NOT
**Directive:** New files SHOULD NOT be created unless absolutely necessary for the requested task.
**Scope:** All development work
**Rationale:** Editing existing files maintains consistency and reduces codebase complexity

<!-- RULE-END: prefer-edit-over-create -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:04Z professional-output -->
## Professional Output Standards
**Level:** MUST
**Directive:** Command output MUST be clean and professional like git, docker, or kubectl without conversational language.
**Scope:** All plonk CLI command output
**Rationale:** Professional tools maintain consistent, focused output that users can rely on in scripts and automation

<!-- RULE-END: professional-output -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:05Z no-documentation-files -->
## Documentation File Restriction
**Level:** MUST NOT
**Directive:** Documentation files MUST NOT be created unless explicitly requested by the user.
**Scope:** All *.md and README files
**Rationale:** Proactive documentation creation violates the exact-scope principle and adds unwanted files

<!-- RULE-END: no-documentation-files -->
---

<!-- RULE-BEGIN: 2025-08-28T00:00:06Z integration-tests-bats-only -->
## Integration Test Framework
**Level:** MUST
**Directive:** Integration tests MUST use BATS framework and require express user permission before execution.
**Scope:** All integration tests that interact with system resources
**Rationale:** BATS provides controlled CLI testing while requiring explicit permission protects developer systems

<!-- RULE-END: integration-tests-bats-only -->
---
