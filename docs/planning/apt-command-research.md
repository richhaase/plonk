# APT Command Research

## Command Analysis

### 1. Check if package is installed

**Option A: dpkg -l**
```bash
dpkg -l htop
```
Output when installed:
```
ii  htop  3.2.1-1  amd64  interactive process viewer
```
Output when not installed:
```
dpkg-query: no packages found matching htop
```

**Option B: dpkg-query**
```bash
dpkg-query -W -f='${Status}\n' htop
```
Output when installed:
```
install ok installed
```
Output when not installed:
```
dpkg-query: no packages found matching htop
```

**Recommendation**: Use dpkg-query for cleaner parsing

### 2. Search for packages

**Command**: `apt-cache search <keyword>`
```bash
apt-cache search monitoring
```
Output format:
```
htop - interactive process viewer
glances - CLI curses-based monitoring tool
nagios-nrpe-server - Nagios Remote Plugin Executor Server
```

**Considerations**:
- Can return many results
- Each line is: `package-name - description`
- No sudo required
- Should limit results in code

### 3. Get package info

**Command**: `apt-cache show <package>`
```bash
apt-cache show htop
```
Output includes:
```
Package: htop
Version: 3.2.1-1
Priority: optional
Section: utils
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Installed-Size: 353
Depends: libc6 (>= 2.34), libncursesw6 (>= 6), libtinfo6 (>= 6)
Description: interactive process viewer
 ...
```

**Key fields**:
- Version: Full version string
- Depends: Dependencies
- Description: Package description
- Exit code 0 if package exists

### 4. Install package

**Command**: `apt-get install -y <package>`
```bash
sudo apt-get install -y htop
```

Without sudo:
```
E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)
E: Unable to acquire the dpkg frontend lock (/var/lib/dpkg/lock-frontend), are you root?
```

**Flags to consider**:
- `-y` or `--yes`: Automatic yes to prompts
- `--no-install-recommends`: Don't install recommended packages
- `-q` or `--quiet`: Less output

### 5. Remove package

**Command**: `apt-get remove -y <package>`
```bash
sudo apt-get remove -y htop
```

**Remove vs Purge**:
- `remove`: Removes package but keeps config files
- `purge`: Removes package and config files
- Recommendation: Use `remove` for safety

### 6. Check if APT is available

Check for commands:
```bash
which apt-get
which apt-cache
```

Check if Debian-based:
```bash
test -f /etc/debian_version
```

Or check for apt directories:
```bash
test -d /var/lib/apt
```

## Exit Codes

- `0`: Success
- `100`: apt-get errors (like package not found)
- `1`: General errors
- `13`: Permission denied (common without sudo)

## Package Name Edge Cases

### Virtual Packages
Some packages are virtual (provided by others):
```bash
apt-cache show mail-transport-agent
```
Returns multiple packages that provide this

### Architecture Suffixes
Some packages have architecture:
```bash
apt-get install libc6:i386
```

### Development Headers
Common pattern:
- Runtime: `libssl` or `openssl`
- Development: `libssl-dev`

### Different Names
Common differences from other package managers:
- `node` → `nodejs`
- `docker` → `docker.io` or `docker-ce`
- `postgres` → `postgresql`

## Platform Detection

### Check if Debian-based

**Option 1**: Check for debian_version
```go
if _, err := os.Stat("/etc/debian_version"); err == nil {
    // Debian-based system
}
```

**Option 2**: Check os-release
```bash
grep -i debian /etc/os-release
grep -i ubuntu /etc/os-release
```

**Option 3**: Check for apt directories
```go
if _, err := os.Stat("/var/lib/apt"); err == nil {
    // Has APT
}
```

## Sudo Handling Patterns

### Pattern 1: Check if root
```go
if os.Geteuid() != 0 {
    return fmt.Errorf("this command requires root privileges")
}
```

### Pattern 2: Parse apt-get error
```go
if strings.Contains(err.Error(), "Permission denied") ||
   strings.Contains(err.Error(), "are you root?") {
    return fmt.Errorf("requires sudo: sudo plonk install apt:%s", pkg)
}
```

## Testing Considerations

### Docker Testing
```dockerfile
FROM ubuntu:22.04
RUN apt-get update
COPY plonk /usr/local/bin/
# Test commands
```

### Mock Responses
Need to mock:
1. dpkg-query output for installed check
2. apt-cache search results
3. apt-cache show output
4. apt-get install errors
5. Platform detection files

## APT Lock Files

APT uses lock files that can cause issues:
- `/var/lib/dpkg/lock-frontend`
- `/var/lib/apt/lists/lock`

If another apt process is running:
```
E: Could not get lock /var/lib/dpkg/lock-frontend - open (11: Resource temporarily unavailable)
E: Unable to acquire the dpkg frontend lock (/var/lib/dpkg/lock-frontend), is another process using it?
```

## Performance Considerations

1. **apt update**: Potentially slow, should not run automatically
2. **apt-cache search**: Can return thousands of results
3. **dpkg -l**: Fast, uses local database
4. **apt-cache show**: Fast, uses local cache

## Recommendations Summary

1. Use `dpkg-query` for installation checks
2. Use `apt-cache` for search and info (no sudo)
3. Let `apt-get` fail naturally for permission errors
4. Provide clear, actionable error messages
5. Use `/etc/debian_version` for platform detection
6. Add reasonable timeouts (30s for install, 5s for search)
7. Use `--no-install-recommends` by default
8. Limit search results to prevent flooding
