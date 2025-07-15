# Package Managers

This document tracks the package managers currently supported by plonk and those under consideration for future support.

## Currently Supported Package Managers

### 1. Homebrew
- **Type**: System package manager (macOS/Linux)
- **Scope**: Global packages only
- **Primary Use**: System utilities, development tools, applications
- **Commands**: `brew install`, `brew uninstall`, `brew list`
- **Platform**: macOS, Linux

### 2. npm (Node Package Manager)
- **Type**: Language package manager
- **Scope**: Global packages only (`npm -g`)
- **Primary Use**: Node.js CLI tools and global utilities
- **Commands**: `npm install -g`, `npm uninstall -g`, `npm list -g`
- **Platform**: Cross-platform

### 3. Cargo (Rust)
- **Type**: Language package manager
- **Scope**: Global installations (`cargo install`)
- **Primary Use**: Rust CLI tools and utilities
- **Commands**: `cargo install`, `cargo uninstall`, `cargo install --list`
- **Platform**: Cross-platform

### 4. pip (Python)
- **Type**: Language package manager
- **Scope**: User installations (`pip install --user`)
- **Primary Use**: Python CLI tools and utilities (black, flake8, pytest, etc.)
- **Commands**: `pip install --user`, `pip uninstall`, `pip list --user`
- **Platform**: Cross-platform
- **Special Behavior**: Works with any Python version in PATH (system, pyenv, conda, etc.)

### 5. gem (Ruby)
- **Type**: Language package manager
- **Scope**: User/global installations (prefers `--user-install`)
- **Primary Use**: Ruby CLI tools (bundler, rails, rubocop, pry, etc.)
- **Commands**: `gem install`, `gem uninstall`, `gem list`
- **Platform**: Cross-platform
- **Special Behavior**: Only tracks gems that provide executables

## Package Managers Under Consideration

### System Package Managers

#### High Priority
1. **apt (Advanced Package Tool)**
   - Debian/Ubuntu systems
   - Most common Linux package manager
   - Would enable Linux support

2. **dnf/yum**
   - Fedora/RHEL/CentOS systems
   - Major enterprise Linux distributions
   - Similar interface to apt

3. **pacman**
   - Arch Linux
   - Popular among developers
   - AUR support could be interesting

#### Medium Priority
4. **MacPorts**
   - Alternative to Homebrew on macOS
   - Some prefer it for certain packages

5. **Chocolatey**
   - Windows package manager
   - Would enable Windows support

6. **Scoop**
   - Windows package manager
   - Developer-focused alternative to Chocolatey

7. **winget**
   - Microsoft's official Windows package manager
   - Growing adoption

### Language/Development Package Managers

#### High Priority
1. **go install**
   - Go modules and tools
   - Global tools: gopls, golangci-lint
   - Simpler than most - single binary installations

#### Medium Priority
2. **composer (PHP)**
   - PHP package manager
   - Global tools: phpcs, phpstan
   - Less commonly used globally

3. **pipx**
   - Python application installer
   - Specifically designed for global CLI tools
   - Isolates each tool in its own virtual environment
   - Better alternative to pip for CLI tools

4. **yarn**
   - Alternative to npm
   - Some developers prefer it
   - Global package support similar to npm

5. **pnpm**
   - Another npm alternative
   - Growing popularity
   - Global package support

### Specialized Package Managers

#### Low Priority
1. **nix**
   - Functional package manager
   - Cross-platform
   - Complex but powerful

2. **asdf**
   - Version manager for multiple languages
   - Different paradigm than other package managers
   - Might conflict with plonk's approach

3. **conda/mamba**
   - Scientific Python ecosystem
   - Manages environments and packages
   - Complex interaction with system Python

4. **snap**
   - Ubuntu's universal package system
   - Containerized applications
   - Different paradigm

5. **flatpak**
   - Desktop application distribution
   - Sandboxed applications
   - GUI-focused

## Implementation Considerations

### For System Package Managers
- Need sudo/admin privileges for installation
- Platform-specific implementations
- Version detection varies by system
- Package naming differences across distributions

### For Language Package Managers
- Version manager interactions (pyenv, rbenv, nvm)
- Global vs local installation detection
- Path and environment considerations
- Build tool requirements

### Design Principles
1. Focus on globally installed packages only
2. Gracefully handle missing package managers
3. Work with whatever version is currently active (via PATH)
4. Don't try to manage environments or versions
5. Simple state reconciliation: installed or not installed

## Recommendations

### Next Package Managers to Implement
1. **apt** - Enable Linux support
2. **go install** - Simple implementation, growing ecosystem
3. **dnf/yum** - Enterprise Linux support
4. **pipx** - Better Python tool isolation

### Future Considerations
- Windows support (Chocolatey or winget)
- Alternative npm clients (yarn, pnpm)
- Specialized tools (pipx for Python isolation)
