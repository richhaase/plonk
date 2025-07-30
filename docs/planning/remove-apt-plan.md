# Plan: Remove APT Support

## Rationale
APT doesn't align with plonk's philosophy:
- Requires system-wide changes (sudo)
- Not portable across developer environments
- Breaks the "dotfiles repository" model
- Complicates the user experience

## What to Remove

### 1. Code Files
- [ ] `internal/resources/packages/apt.go`
- [ ] `internal/resources/packages/apt_test.go`
- [ ] `internal/resources/packages/platform.go` (simplify or remove)
- [ ] `internal/resources/packages/platform_test.go`

### 2. Integration Tests
- [ ] `test/integration/apt_test.go`
- [ ] Platform-specific logic in `test/integration/crossplatform_test.go`
- [ ] CI workflow APT-specific steps in `.github/workflows/ci.yml`

### 3. Documentation Updates
- [ ] Remove APT from README.md
- [ ] Remove APT from docs/architecture.md
- [ ] Remove APT from docs/cli.md
- [ ] Remove APT from docs/cmds/package_management.md
- [ ] Remove APT from docs/cmds/doctor.md
- [ ] Remove APT from docs/installation.md
- [ ] Remove APT from docs/CONFIGURATION.md
- [ ] Remove APT mentions from CLAUDE.md

### 4. Simplifications
- [ ] Remove platform detection complexity
- [ ] Simplify package manager availability checks
- [ ] Update doctor command output

## What to Keep/Enhance

### 1. Homebrew on Linux
- Already works via linuxbrew
- No code changes needed
- Just documentation updates

### 2. Language Package Managers
- npm, cargo, pip, gem, go
- All work identically on Linux and macOS
- User-space, no sudo required

## Implementation Steps

### Phase 1: Remove APT Code
1. Delete apt.go and apt_test.go
2. Remove APT from registry
3. Simplify platform detection (no need for Linux distro detection)
4. Update tests

### Phase 2: Update Documentation
1. Remove all APT references
2. Add Linux setup guide focusing on Homebrew
3. Update v1.0 readiness docs

### Phase 3: Update Linux Testing Plan
1. Focus on Homebrew installation on Linux
2. Test language package managers
3. Verify consistent behavior with macOS

## Benefits
- Simpler codebase (remove ~1000 lines)
- No sudo complexity
- True cross-platform dotfiles
- Consistent user experience
- Easier to maintain

## Linux Developer Workflow
```bash
# Install Homebrew on Linux
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Use plonk exactly like macOS
plonk install ripgrep fd bat
plonk install npm:prettier cargo:tokei
```

This is the right approach for plonk's philosophy!
