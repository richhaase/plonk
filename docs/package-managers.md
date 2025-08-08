# Package Manager Research & Strategy

This document outlines plonk's package manager ecosystem strategy based on competitive research and emerging language trends as of August 2025.

## Current Package Manager Support

Plonk currently supports 9 package managers across multiple language ecosystems:

- **Homebrew (brew)** - System tools and applications (macOS/Linux)
- **NPM (npm)** - JavaScript global packages
- **Cargo (cargo)** - Rust packages
- **Gem (gem)** - Ruby gems
- **Go install (go)** - Go modules
- **UV (uv)** - Python tool management
- **Pixi (pixi)** - Conda-forge packages

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

### Immediate Priority

#### 1. **pnpm Global Support**
- **Status**: 70% less disk space, significantly faster than npm
- **Adoption**: "Increasingly recommended for new projects" in 2024-2025
- **Commands**: `pnpm add -g`, `pnpm remove -g`, `pnpm list -g`
- **Value**: Direct npm alternative with better performance
- **Why Important**: Higher current adoption than alternatives like Bun

#### 2. **Bun/BunX Global Support**
- **Status**: Strong developer interest, early but growing adoption
- **Performance**: Up to 30x faster package manager than npm
- **Commands**: `bunx add --global`, `bunx remove --global`, `bun pm ls --global`
- **Value**: Performance-focused JavaScript ecosystem alternative
- **Why Important**: Positions plonk in emerging high-performance JS tooling

### Next Phase

#### 3. **.NET Global Tools** (`dotnet tool`)
- **Gap**: Major enterprise language ecosystem completely missing from plonk
- **Commands**: `dotnet tool install/list/uninstall -g`
- **Value**: Massive enterprise adoption, rich CLI tool ecosystem
- **Examples**: `dotnet-ef`, `dotnet-outdated`, code generators

#### 4. **PHP Composer Global** (`composer global`)
- **Gap**: Web development ecosystem with huge adoption
- **Commands**: `composer global require/show/remove`
- **Value**: PHP has enormous web development usage
- **Examples**: `phpunit`, `php_codesniffer`, documentation generators

### Future Monitoring

- **Proto Version Manager**: Consider as complement rather than competition for tool version management
- **Zig Package Manager**: Monitor for 1.0 release (expected 2025+) and ecosystem maturation
- **Mojo/AI Tools**: Watch AI-focused development tool emergence

## Language Ecosystem Trends (2024-2025)

### Key Growth Patterns
- **Performance Focus**: Bun, pnpm, Zig all emphasize speed and efficiency
- **AI/ML Growth**: Python's 7 percentage point increase driving pip/uv usage
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

## Implementation Recommendations

**Phase 1**: Implement **pnpm** and **Bun** global support
- Address JavaScript ecosystem performance demands
- Capitalize on growing adoption of npm alternatives

**Phase 2**: Add **.NET Global Tools** and **PHP Composer Global**
- Fill major language ecosystem gaps
- Expand enterprise and web development coverage

**Phase 3**: Monitor and evaluate emerging ecosystems
- Zig package manager post-1.0
- AI/ML focused tooling trends
- New performance-oriented package managers

---

*Last updated: August 2025*
*Research covers: Stack Overflow Developer Survey 2024-2025, GitHub Octoverse 2024, JetBrains Developer Ecosystem 2024*
