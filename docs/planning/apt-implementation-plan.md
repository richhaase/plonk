# APT Package Manager Implementation Plan

**Status**: 📝 Planning (2025-07-30)

## Overview

Add APT (Advanced Package Tool) support to plonk, enabling management of system packages on Debian-based Linux distributions (Ubuntu, Debian, Pop!_OS, etc.). This is the final major feature required for v1.0.

## Goals

1. **Feature Parity**: APT should work identically to other package managers
2. **Cross-Platform**: Gracefully handle non-Linux systems
3. **Security**: Handle sudo requirements safely
4. **User Experience**: Clear feedback about system packages
5. **Testing**: Comprehensive tests without requiring actual package installation

## Key Challenges

### 1. Privilege Escalation
APT requires sudo for most operations:
- `apt-get install` - requires sudo
- `apt-get remove` - requires sudo
- `apt-cache search` - does NOT require sudo
- `apt-cache show` - does NOT require sudo

### 2. Package Name Differences
APT packages often have different names than other managers:
- `node` → `nodejs`
- `docker` → `docker.io` or `docker-ce`
- Development headers: `libssl-dev` vs `openssl`

### 3. System Package Concerns
- System packages affect the entire OS
- Removing wrong package could break the system
- Updates might require system restart
- Dependencies can be complex

### 4. Cross-Platform Handling
- APT only exists on Debian-based systems
- Need to detect and skip on macOS/other Linux distros
- Doctor command should handle this gracefully

## Design Decisions

### 1. Sudo Handling

**Option A: Fail and Instruct** ✅
- Run commands without sudo
- When they fail, provide clear instructions
- User must run with sudo: `sudo plonk install apt:package`
- Safest approach, explicit user consent

**Option B: Prompt for Password**
- Detect when sudo needed
- Prompt user for password
- More convenient but security concerns

**Option C: Check and Refuse**
- Check if running as root/sudo
- Refuse to run if not elevated
- Very safe but poor UX

**Recommendation**: Option A - Let apt-get fail naturally and provide clear error messages about sudo requirements.

### 2. Package Detection

For `IsInstalled()` check:
- Use `dpkg -l package-name` (fast, local)
- Parse output for installation status
- Alternative: `dpkg-query -W -f='${Status}' package-name`

### 3. Available Packages

For `IsAvailable()` check:
- Use `apt-cache show package-name`
- Check exit code (0 = exists)
- No sudo required

### 4. Search Implementation

For `Search()`:
- Use `apt-cache search keyword`
- Parse output format: `package-name - description`
- Limit results to prevent flooding

### 5. Version Handling

APT versions are complex:
- Format: `1.2.3-ubuntu4.1`
- Include epoch sometimes: `2:1.2.3-ubuntu4.1`
- For now: Store full version string, don't parse

## Implementation Plan

### Phase 1: Basic Structure

1. Create `internal/packagemanager/apt/apt.go`
2. Implement `Manager` interface:
   ```go
   type Manager struct {
       executor exec.Executor
   }
   ```

3. Register in `internal/packagemanager/packagemanager.go`:
   ```go
   func init() {
       RegisterManager("apt", func() PackageManager {
           return apt.New()
       })
   }
   ```

### Phase 2: Core Methods

1. **Name()**: Return "apt"
2. **IsAvailable()**: Check if apt-get and apt-cache exist
3. **IsInstalled()**: Use dpkg to check installation
4. **Install()**: Run apt-get install (will fail without sudo)
5. **Uninstall()**: Run apt-get remove (will fail without sudo)
6. **Search()**: Use apt-cache search
7. **Info()**: Use apt-cache show

### Phase 3: Error Handling

Create clear error messages for permission issues:
```
Error: Installing APT packages requires administrator privileges.
Please run with sudo: sudo plonk install apt:package-name
```

### Phase 4: Testing Strategy

1. **Unit Tests**: Mock exec.Executor
2. **Integration Tests**: Skip on non-Linux
3. **Manual Testing**: Use Docker container

### Phase 5: Doctor Integration

Update doctor command to:
1. Check if system is Debian-based
2. Check if apt-get is available
3. Don't mark as required on non-Debian systems

## User Experience

### Install Example
```bash
$ plonk install apt:htop
Error: Installing APT packages requires administrator privileges.
Please run with sudo: sudo plonk install apt:htop

$ sudo plonk install apt:htop
Installing htop...
✓ htop installed successfully
```

### Search Example
```bash
$ plonk search apt:monitoring
Searching apt for "monitoring"...
- htop - interactive process viewer
- glances - CLI curses-based monitoring tool
- nmon - performance monitoring tool
```

### Platform Detection
```bash
# On macOS
$ plonk install apt:htop
Error: APT package manager is not available on this system.
APT is only available on Debian-based Linux distributions.
```

## Security Considerations

1. **Never store sudo password**
2. **Never automatically escalate privileges**
3. **Clear messages about what requires sudo**
4. **Let user decide when to use sudo**
5. **Log all system package operations**

## Configuration

No special configuration needed. APT works with standard plonk config:
- Default timeout applies
- No special settings required
- Package names in lock file: `apt:package-name`

## Edge Cases

1. **Package not found**: Clear error message
2. **Already installed**: Add to plonk management
3. **Network issues**: Timeout handling
4. **Broken packages**: Report apt-get errors clearly
5. **Different distros**: Ubuntu vs Debian package names

## Success Criteria

1. ✅ Can install/uninstall APT packages (with sudo)
2. ✅ Can search APT packages (without sudo)
3. ✅ Clear error messages for permission issues
4. ✅ Graceful handling on non-Linux systems
5. ✅ Doctor command reports APT status correctly
6. ✅ Lock file correctly tracks APT packages
7. ✅ Tests pass without requiring real installation

## Future Enhancements (Post-v1.0)

1. Package group support (build-essential, etc.)
2. Repository management (add-apt-repository)
3. Update detection (apt-get update)
4. Upgrade command support
5. Better version parsing/comparison
6. Snap package integration

## Decisions Made

1. **Auto-update**: Never run `apt-get update` automatically
   - User has explicit control
   - Keeps operations fast and predictable

2. **Remove vs Purge**: Always use `remove` (not `purge`)
   - Safer default, preserves configuration files
   - User can manually purge if needed

3. **Recommends/Suggests**: Always use `--no-install-recommends`
   - Minimal installations only
   - Keeps system lean and predictable

4. **Virtual packages**: Pass through as-is
   - Let apt-get handle virtual packages naturally
   - User sees APT's prompts and can choose

5. **Architecture**: Pass through as-is
   - Treat `package:arch` as complete package name
   - Store full name in lock file
   - Let APT validate architecture

## Implementation Order

1. Create basic structure and registration
2. Implement IsAvailable() and Search() (no sudo needed)
3. Implement Install/Uninstall with error handling
4. Add IsInstalled() with dpkg
5. Create comprehensive tests
6. Update doctor command
7. Manual testing in Docker
8. Documentation updates
