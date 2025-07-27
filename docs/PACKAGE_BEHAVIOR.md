# Package Install/Uninstall Behavior Specification

This document defines the expected behavior for plonk's install and uninstall commands.

## Install Command

The install command wraps underlying package managers and maintains plonk's lock file.

### Expected Behavior:

1. **Manager Selection**
   - If package has a prefix (e.g., `brew:postgresql`), use the specified manager
   - If no prefix, use the default manager from config (or system default)

2. **Installation & Lock File**
   - Attempt to install the package using the selected manager
   - If installation succeeds: add package info to plonk.lock
   - If package was already installed: still add package info to plonk.lock (idempotent)
   - This ensures plonk can manage pre-existing packages

3. **Error Handling**
   - Only show errors if the install command actually fails
   - Already-installed packages should not be treated as errors

### Example Flow:
```bash
plonk install postgresql  # Uses default manager (e.g., cargo)
plonk install brew:postgresql  # Explicitly uses brew

# If postgresql is already installed via brew:
plonk install brew:postgresql  # Should succeed and add to lock file
```

## Uninstall Command

The uninstall command removes packages from plonk management and optionally from the system.

### Expected Behavior:

1. **Lock File Check**
   - First, check if package exists in plonk.lock
   - If found: use the manager recorded in plonk.lock
   - If not found: use default manager as pass-through

2. **Removal Process**
   - If in plonk.lock:
     - Remove entry from plonk.lock
     - Attempt uninstall with the lock file's manager
     - Report success even if uninstall fails (partial success - removed from plonk management)
   - If not in plonk.lock:
     - Attempt uninstall with default manager
     - Report success/failure based on uninstall result

3. **Success Criteria**
   - Removing from plonk.lock = success (even if system uninstall fails)
   - Pass-through uninstall = success only if uninstall succeeds

### Example Flow:
```bash
# Package in plonk.lock as brew:
plonk uninstall postgresql  # Uses brew (from lock), removes from lock

# Package not in plonk.lock:
plonk uninstall postgresql  # Uses default manager as pass-through
```

## Key Principles

1. **Plonk as a Wrapper**: Plonk wraps existing package managers, adding lock file management
2. **Idempotent Operations**: Installing already-installed packages should succeed
3. **Graceful Degradation**: Removing from plonk management is a success even if system uninstall fails
4. **Lock File Authority**: For managed packages, the lock file determines the correct manager
