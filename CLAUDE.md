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

Use `justfile` for development tasks. Run `just --list` to see available commands.

## Architecture Overview

See [CODEBASE_MAP.md](CODEBASE_MAP.md) for navigation and current implementation status.

## Required TDD Workflow

Follow Test-Driven Development practices:
1. Write failing tests first
2. Write minimal code to pass tests
3. Refactor while keeping tests green
4. Commit frequently with descriptive messages

**AI Agent Requirements**:
- Use CODEBASE_MAP.md for navigation
- Follow established patterns found in existing code
- Focus on currently implemented features (see CODEBASE_MAP.md development status)

## Session Management

### Session Start
1. **Read CODEBASE_MAP.md** - Understand current implementation status
2. **Check existing code** - Review patterns and structure before making changes
3. **Run tests** - Ensure existing functionality works (`just test`)

### During Session
- **Follow TDD workflow** - Write tests first, then implement
- **Commit frequently** with descriptive messages
- **Run quality checks** - Use `just precommit` before committing
- **Ask questions** when requirements are unclear

### Session End
- **Run final tests** - Ensure all tests pass
- **Update documentation** if needed (CODEBASE_MAP.md for structural changes)
- **Commit final changes** with clear messages

## Development Tools

### Just Task Runner
- **`just --list`** - Show available tasks
- **`just build`** - Build the binary
- **`just test`** - Run tests
- **`just precommit`** - Run quality checks

### Preferred CLI Tools
- **ripgrep (rg)** for searching
- **fd** for finding files

## Code Guidelines

### AI-Specific Rules
- **No comments** unless explicitly requested
- **Follow existing patterns** found in the codebase
- **Read before editing** - always use Read tool before Edit/Write
- **Prefer editing** existing files over creating new ones
- **Use standard Go error handling** patterns

### Testing Patterns
- **Standard Go testing** with table-driven tests
- **Mock external dependencies** in tests
- **Test coverage** for all major components
- **Interface compliance** tests for new package managers

## Code Patterns

### Error Handling
```go
// Standard Go error patterns
return fmt.Errorf("config error: %w", err)
return fmt.Errorf("package manager error: %w", err)
```

### Command Structure
All CLI commands follow Cobra patterns:
- Commands in `internal/commands/`
- Subcommands organized by functionality
- Consistent error handling
- Comprehensive test coverage

## Current CLI Commands

**Currently Implemented:**
- `status` - Package manager availability and counts
- `pkg list` - List packages by manager
- `config show` - Show configuration
- `dot list` - List dotfiles

**Planned:** Installation, configuration application, repository management

## Key Files

- **`internal/commands/root.go`** - CLI structure and command registration
- **`internal/managers/common.go`** - Core interfaces and patterns
- **`internal/config/yaml_config.go`** - Configuration structure and parsing
- **`internal/commands/output.go`** - Output formatting utilities
- **`internal/managers/reconciler.go`** - State reconciliation logic