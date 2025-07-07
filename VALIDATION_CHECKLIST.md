# Plonk Manual Validation Checklist

Let's validate the core functionality together. Check off each item as we test it.

## Build & Basic Setup
- [x] `just build` creates binary in `build/` directory (not repo root)
- [x] `build/plonk --help` shows help message
- [x] `git status` doesn't show `build/plonk` as untracked
- [x] `just install` rebuilds and installs to GOBIN

## Core Commands
- [ ] `build/plonk status` - shows package manager status and drift
- [ ] `build/plonk check` - alias works (same as status)
- [ ] `build/plonk import --dry-run` - shows discovery preview
- [ ] `build/plonk install --dry-run` - shows install preview  
- [ ] `build/plonk apply --dry-run` - shows apply preview

## Package Commands
- [ ] `build/plonk pkg list` - lists packages
- [ ] `build/plonk ls` - alias works (same as pkg list)
- [ ] `build/plonk pkg search git` - searches for packages

## Global Dry-Run Flag
- [ ] `build/plonk --dry-run import` - global flag works
- [ ] `build/plonk --dry-run install` - global flag works
- [ ] `build/plonk --dry-run apply` - global flag works

## Error Handling
- [ ] `build/plonk invalid-command` - shows clear error
- [ ] `build/plonk --verbose` - rejects unknown flag
- [ ] `build/plonk install nonexistent` - handles missing package

## Current Session Fixes
- [ ] Build artifacts go to `build/` not repo root ✅
- [ ] Install copies from `build/` instead of rebuilding ✅
- [ ] No verbose/quiet flags (removed) ✅
- [ ] Tests still pass: `go test ./...`

Let's go through these one by one!