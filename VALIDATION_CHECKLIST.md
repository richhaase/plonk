# Plonk Manual Validation Checklist

Let's validate the core functionality together. Check off each item as we test it.

## Build & Basic Setup
- [x] `just build` creates binary in `build/` directory (not repo root)
- [x] `build/plonk --help` shows help message
- [x] `git status` doesn't show `build/plonk` as untracked
- [x] `just install` rebuilds and installs to GOBIN

## Core Commands
- [x] `build/plonk status` - shows package manager status and drift
- [x] `build/plonk check` - alias works (same as status)
- [x] `build/plonk import --dry-run` - shows discovery preview
- [x] `build/plonk install --dry-run` - shows install preview  
- [x] `build/plonk apply --dry-run` - shows apply preview

## Package Commands
- [x] `build/plonk pkg list` - lists packages
- [x] `build/plonk ls` - alias works (same as pkg list)
- [x] `build/plonk pkg search git` - searches for packages

## Global Dry-Run Flag
- [x] `build/plonk --dry-run import` - global flag works
- [x] `build/plonk --dry-run install` - global flag works
- [x] `build/plonk --dry-run apply` - global flag works

## Error Handling
- [x] `build/plonk invalid-command` - shows clear error
- [x] `build/plonk --verbose` - rejects unknown flag
- [x] `build/plonk install nonexistent` - handles missing package

## Current Session Fixes
- [x] Build artifacts go to `build/` not repo root ✅
- [x] Install rebuilds and installs to GOBIN ✅
- [x] No verbose/quiet flags (removed) ✅
- [x] Tests still pass: `go test ./...` ✅

Let's go through these one by one!