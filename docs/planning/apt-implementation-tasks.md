# APT Implementation Task List

## Pre-Implementation Research

### 1. Command Analysis
- [ ] Document exact commands for each operation
- [ ] Test behavior with missing packages
- [ ] Test behavior without sudo
- [ ] Understand dpkg output formats
- [ ] Research apt-cache search limits

### 2. Platform Detection
- [ ] How to detect Debian-based systems
- [ ] Check for apt-get vs apt command
- [ ] Handle Ubuntu vs Debian differences
- [ ] Test on WSL (Windows Subsystem for Linux)

## Implementation Tasks

### Phase 1: Basic Structure (2 hours)
- [ ] Create `internal/packagemanager/apt/` directory
- [ ] Create `apt.go` with Manager struct
- [ ] Implement `Name()` method
- [ ] Implement `IsAvailable()` method
- [ ] Register in packagemanager.go
- [ ] Create `apt_test.go` with basic tests

### Phase 2: Read Operations (2 hours)
- [ ] Implement `Search()` method
  - [ ] Use apt-cache search
  - [ ] Parse output format
  - [ ] Handle timeouts
  - [ ] Limit results
- [ ] Implement `Info()` method
  - [ ] Use apt-cache show
  - [ ] Extract version info
  - [ ] Handle missing packages
- [ ] Implement `IsInstalled()` method
  - [ ] Use dpkg -l or dpkg-query
  - [ ] Parse status output
  - [ ] Handle partially installed states
- [ ] Add tests for all read operations

### Phase 3: Write Operations (3 hours)
- [ ] Implement `Install()` method
  - [ ] Build apt-get install command
  - [ ] Handle sudo failure
  - [ ] Create helpful error messages
  - [ ] Add --no-install-recommends flag
- [ ] Implement `Uninstall()` method
  - [ ] Build apt-get remove command
  - [ ] Handle sudo failure
  - [ ] Decide remove vs purge
- [ ] Create common error handler for permission issues
- [ ] Add progress indicators
- [ ] Add tests with mocked executor

### Phase 4: Integration (2 hours)
- [ ] Update doctor command
  - [ ] Check for Debian-based system
  - [ ] Check apt-get availability
  - [ ] Mark as optional on non-Debian
- [ ] Test with other commands
  - [ ] plonk install apt:package
  - [ ] plonk search apt:keyword
  - [ ] plonk info apt:package
  - [ ] plonk status (shows apt packages)
- [ ] Verify lock file handling
- [ ] Test cross-platform behavior

### Phase 5: Testing (3 hours)
- [ ] Unit tests
  - [ ] Mock all exec calls
  - [ ] Test error conditions
  - [ ] Test output parsing
  - [ ] Test timeout handling
- [ ] Integration tests
  - [ ] Skip on non-Linux
  - [ ] Use test containers if available
- [ ] Manual testing checklist
  - [ ] Test on Ubuntu 22.04
  - [ ] Test on Debian 12
  - [ ] Test on macOS (should fail gracefully)
  - [ ] Test permission errors
  - [ ] Test network timeouts

### Phase 6: Documentation (1 hour)
- [ ] Update package_management.md
- [ ] Add APT examples to README
- [ ] Document sudo requirements
- [ ] Add to architecture.md
- [ ] Create troubleshooting section

## Testing Scenarios

### Happy Path
1. Search for available package
2. Check if installed (not installed)
3. Install package (with sudo)
4. Check if installed (installed)
5. Uninstall package (with sudo)

### Error Cases
1. Install without sudo - should fail with helpful message
2. Install non-existent package - should fail with apt error
3. Search with network down - should timeout
4. Run on macOS - should fail with platform error

### Edge Cases
1. Package already installed via apt
2. Package with complex name (lib-dev)
3. Virtual package names
4. Partially configured packages

## Manual Testing Commands

```bash
# Build plonk
go build -o plonk ./cmd/plonk

# Test search (no sudo needed)
./plonk search apt:htop

# Test info (no sudo needed)
./plonk info apt:htop

# Test install without sudo (should fail gracefully)
./plonk install apt:htop

# Test install with sudo
sudo ./plonk install apt:htop

# Check status
./plonk status --packages

# Test uninstall
sudo ./plonk uninstall apt:htop

# Test doctor
./plonk doctor
```

## Code Structure

```
internal/
  packagemanager/
    apt/
      apt.go          # Main implementation
      apt_test.go     # Unit tests
      errors.go       # APT-specific errors
      parser.go       # Output parsing helpers
      parser_test.go  # Parser tests
```

## Definition of Done

- [ ] All unit tests pass
- [ ] Integration tests pass on Linux
- [ ] Graceful failure on non-Linux systems
- [ ] Clear error messages for permission issues
- [ ] Doctor command updated
- [ ] Documentation complete
- [ ] Manual testing on Ubuntu and Debian
- [ ] Code review passed
