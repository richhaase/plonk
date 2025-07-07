# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Plonk** is a CLI tool for managing shell environments across multiple machines using Homebrew, ASDF, and NPM package managers. Written in Go with strict Test-Driven Development (TDD) practices.

## Documentation File Organization

Each documentation file serves a specific purpose. Understanding these roles prevents duplication and ensures information is found where expected:

### 📋 **README.md** - User Quick Start
**Purpose:** First impression for new users and contributors
**Contains:** Project overview, quick installation, basic usage examples, configuration format
**Audience:** End users, potential contributors
**Avoid:** Detailed development workflows, architectural deep-dives, session management

### 🏗️ **ARCHITECTURE.md** - Design & Decisions  
**Purpose:** System architecture, design philosophy, critical design decisions
**Contains:** Component relationships, design rationale, extension points, security considerations
**Audience:** Contributors, maintainers, system designers
**Avoid:** Feature lists, installation steps, development workflows

### 🤝 **CONTRIBUTING.md** - Developer Workflow
**Purpose:** How to contribute code and follow development practices
**Contains:** TDD workflow, development setup, coding standards, commit guidelines
**Audience:** Contributors (human and AI)
**Avoid:** Architecture details, user instructions, project roadmap

### 🗺️ **CODEBASE_MAP.md** - Navigation Guide
**Purpose:** Find specific code, understand file organization, locate implementation details
**Contains:** Directory structure, file purposes, common patterns, navigation tips
**Audience:** Contributors working with the code
**Avoid:** Design rationale, user documentation, development process

### 📈 **ROADMAP.md** - Strategic Planning
**Purpose:** Strategic direction, upcoming phases, feature planning
**Contains:** Current phases, time estimates, ideas for future consideration
**Audience:** Project stakeholders, potential contributors
**Avoid:** Tactical session details, completed features (see CHANGELOG)

### ✅ **TODO.md** - Session Management
**Purpose:** Tactical session-level work tracking for AI agents
**Contains:** Current session focus, active work items, session notes, context for next session
**Audience:** AI agents, session continuity
**Avoid:** Strategic planning, completed work history, architecture details

### 📖 **CHANGELOG.md** - Historical Record
**Purpose:** Canonical record of what was accomplished when
**Contains:** Version history, feature additions, changes, fixes
**Audience:** Users tracking changes, maintainers
**Avoid:** Future plans, development process, architecture

### 🤖 **CLAUDE.md** - AI Agent Instructions
**Purpose:** Guidance for AI agents working on the codebase
**Contains:** Workflow requirements, session management, coding guidelines, tool usage
**Audience:** AI coding agents
**Avoid:** User-facing documentation, general contribution guidelines

**Cross-Reference Rule:** When information exists in multiple files, one file should be canonical and others should reference it. Example: CONTRIBUTING.md contains the detailed TDD workflow, other files reference it.

## Development Environment

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and workflow commands.

## Architecture Overview

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed system design and [CODEBASE_MAP.md](CODEBASE_MAP.md) for navigation.

## Required TDD Workflow

Follow the Test-Driven Development workflow described in [CONTRIBUTING.md](CONTRIBUTING.md).

**AI Agent Specific Requirements**:
- Sync with TODO.md at session start and maintain it throughout
- Reference ARCHITECTURE.md for system design decisions
- Use CODEBASE_MAP.md for navigation
- Follow established patterns found in existing code

## Session Management

### Session Initialization
1. **Read TODO.md** - Check for active work items and session context
2. **If no active todos** - Check ROADMAP.md for possible work items to suggest
3. **Sync internal todo list** - Match TODO.md content with internal state
4. **Review session notes** - Check TODO.md Notes section for context
5. **Understand current phase** - Documentation refactor, feature development, etc.

### During Session
- **Update TODO.md** when changing task status (pending → in_progress → completed)
- **Commit frequently** with descriptive messages following project patterns
- **Run pre-commit hooks normally** - only use --no-verify when specifically needed
- **Ask questions** when user requests are unclear or conflict with established patterns

### Session End
- **Update TODO.md** to reflect current state for next session
- **Commit any pending work** with clear status in commit message
- **Update related documentation** as needed:
  - Update CHANGELOG.md for user-facing changes
  - Update ARCHITECTURE.md for design changes
  - Update README.md for user workflow changes
  - Update CONTRIBUTING.md for development process changes
- **Leave clear notes** in TODO.md for session continuity

## Tool Usage

### Preferred CLI Tools (via Bash)
- **ripgrep (rg)** instead of grep
- **fd** instead of find  
- **exa** instead of ls
- **sd** instead of sed

## Code Generation Guidelines

### AI-Specific Rules
- **No comments** unless explicitly requested
- **Follow existing patterns** found in the codebase
- **Read before editing** - always use Read tool before Edit/Write
- **Prefer editing** existing files over creating new ones
- **Use established error handling** patterns from error_handling.go
- **Mock external dependencies** in tests using MockCommandExecutor

### Testing Patterns
- **MockCommandExecutor** for unit tests (avoid actual command execution)
- **setupTestEnv(t)** helper for test isolation
- **Table-driven tests** for comprehensive coverage
- **Interface compliance** tests for new package managers

## Standardized Patterns

### Error Handling
```go
// internal/commands/error_handling.go
WrapConfigError(err)                    // Configuration loading errors
WrapInstallError(packageName, err)      // Package installation errors
WrapPackageManagerError("homebrew", err) // Package manager availability errors
```

### Command Structure
All CLI commands follow consistent patterns:
- Cobra command structure in `internal/commands/`
- Argument validation using `ValidateNoArgs()`, `ValidateExactArgs()`
- Error wrapping for consistent user experience
- Comprehensive test coverage with mocks

### File Operations
- Use `internal/directories.Default` for all path operations
- Backup functionality available via `internal/commands/backup.go`

## CLI Commands

For complete command reference, see README.md. Key commands:
- `status` - Package manager availability and drift detection
- `install` - Install packages from config  
- `apply` - Apply configurations (supports --backup, --dry-run)
- `clone`/`pull` - Git repository operations

## Critical Files

- **`internal/commands/root.go`** - CLI structure and command registration
- **`pkg/managers/common.go`** - Core interfaces and patterns
- **`pkg/config/yaml_config.go`** - Configuration structure and parsing
- **`internal/commands/error_handling.go`** - Standardized error patterns
- **`internal/commands/test_helpers.go`** - Test utilities and patterns