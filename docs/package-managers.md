# Package Manager Research & Strategy

This document outlines plonk's package manager ecosystem strategy based on competitive research and emerging language trends as of August 2025.

## Current Package Manager Support

Plonk currently supports 11 package managers across multiple language ecosystems:

- **Homebrew (brew)** - System tools and applications (macOS/Linux)
- **NPM (npm)** - JavaScript global packages
- **PNPM (pnpm)** - Fast, disk efficient JavaScript packages (global)
- **Cargo (cargo)** - Rust packages
- **Pipx (pipx)** - Python applications in isolated environments
- **Gem (gem)** - Ruby gems
- **Go install (go)** - Go modules
- **UV (uv)** - Python tool management
- **Pixi (pixi)** - Conda-forge packages
- **Composer (composer)** - PHP global packages and CLI tools
- **.NET Global Tools (dotnet)** - .NET CLI tools and utilities

## Competitive Landscape Analysis

### Direct Competition: Minimal

Plonk occupies a **nearly uncontested market space** for unified package and dotfile management:

- **Nix + Home Manager**: Only true direct competitor, but serves advanced users with steep learning curves
- **Chezmoi + Manual Integration**: Popular dotfiles tool that could integrate packages but requires significant manual setup
- **Custom Solutions**: Most developers use separate tools (brew + dotfiles manager + scripts)

**Key Insight**: The unified package + dotfile management workflow addresses a genuine market gap.

### Adjacent Competition

- **Homebrew Bundle**: Excellent but limited to Homebrew ecosystem
- **Meta Package Manager (MPM)**: Wrapper approach with limited functionality
- **Proto**: Growing adoption for version management across languages

Plonk's multi-package-manager approach is more comprehensive than existing solutions.

## Priority Package Managers for Implementation

### Current Implementation Status

#### 1. **pnpm Global Support** - ✅ **IMPLEMENTED**
- **Status**: Complete and fully functional
- **Performance**: 70% less disk space, significantly faster than npm
- **Adoption**: Increasingly recommended for new projects in 2024-2025
- **Commands**: `pnpm add -g`, `pnpm remove -g`, `pnpm list -g`
- **Value**: Direct npm alternative with better performance
- **Implementation**: Full PackageManager interface support with single-method installation

#### 2. **Bun/BunX Global Support** - ❌ **NOT READY**
- **Status**: Blocked by upstream limitations in Bun itself
- **Performance**: Up to 30x faster package manager than npm
- **Blocking Issues**:
  - ❌ No global package listing (`bun pm ls --global` doesn't exist)
  - ❌ Unreliable global package removal (known issues as of 2023)
  - ❌ No package information/version querying commands
  - ❌ No search capabilities
- **Timeline**: Monitor Bun development; implementation possible Q3-Q4 2025 if upstream issues resolved
- **Alternative**: Consider after pnpm implementation and when Bun matures global package management

### Future Candidates

- **Zig Package Manager**: Monitor for 1.0 release (expected 2025+) and ecosystem maturation
- **Proto Version Manager**: Consider as complement rather than competition for tool version management
- **Mojo/AI Tools**: Watch AI-focused development tool emergence

## Language Ecosystem Trends (2024-2025)

### Key Growth Patterns
- **Performance Focus**: Bun, pnpm, Zig all emphasize speed and efficiency
- **AI/ML Growth**: Python's 7 percentage point increase driving pipx/uv usage
- **TypeScript Surge**: 35% adoption (vs 12% in 2017) driving JS tooling needs
- **Developer Experience**: Tools prioritizing easy setup and consistency

### Fastest Growing Languages
1. **Python**: AI/ML driving massive growth
2. **TypeScript**: Rising across all developer rankings
3. **Rust**: 9th consecutive year as most admired language
4. **Go**: 190% popularity growth, 301% hiring demand growth

## Strategic Positioning

### Market Position
- **"Simple Alternative to Nix"**: Same unified benefits, much lower complexity
- **"Beyond Homebrew Bundle"**: Multi-package manager support
- **"One Command Dev Setup"**: Local-first approach vs cloud development environments

### Key Differentiators
- Unified workflow (packages + dotfiles)
- Multi-package manager support (vs single-ecosystem tools)
- Developer-friendly complexity (vs enterprise IaC tools)
- Cross-platform consistency

## Implementation Roadmap

**Current Phase**: Implement **pnpm** global support
- ✅ Complete PackageManager interface compatibility confirmed
- ✅ Commands available: `pnpm add -g`, `pnpm remove -g`, `pnpm list -g`
- ✅ Address JavaScript ecosystem performance demands
- 🚧 Implementation in progress

**Next Phase**: Monitor **Bun** development and evaluate emerging ecosystems
- Wait for Bun to resolve global package management limitations
- Monitor Zig package manager post-1.0 release
- Evaluate AI/ML focused tooling trends
- Consider new performance-oriented package managers

**Future Considerations**: Additional language ecosystems
- Evaluate demand for additional specialized package managers
- Consider version management integration (Proto, etc.)
- Monitor enterprise and web development needs

---

*Last updated: August 2025*
*Research covers: Stack Overflow Developer Survey 2024-2025, GitHub Octoverse 2024, JetBrains Developer Ecosystem 2024*
