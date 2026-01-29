# Why Plonk?

**tl;dr**: There are a ton of great dotfile managers out there, and many great package managers. None of them did it all for me, and made it easy. The goal of plonk is to make setting up a new computer/VM as my ideal development environment in minutes. And, since I'm a shell environment tinkerer, I need to be able to change what plonk knows about.

## The Problem

Setting up a development environment shouldn't be a multi-day project. Yet every time I get a new machine or spin up a VM, I find myself:

1. Installing Homebrew (or another package manager)
2. Remembering which packages I need
3. Hunting down my dotfiles
4. Figuring out symlinks or copying files around
5. Repeating this dance for each language-specific package manager
6. Inevitably forgetting something crucial until I need it at 2 AM

Existing solutions handle parts of this well, but not the whole thing:
- **Dotfile managers** are great at symlinks but don't handle packages
- **Package managers** install software but don't manage configurations
- **Configuration management tools** are overkill for personal machines
- **Shell scripts** work but become unmaintainable spaghetti over time

## My Journey to Plonk

I've tried many approaches to managing my development environment over the years:

### Bash Scripts
Started with the classic approach - bash scripts to deploy dotfiles from a git repo. It worked, until it didn't. As someone who constantly experiments with new tools (switching from neovim to helix, choosing between broot and yazi, adding zoxide to my workflow), maintaining those scripts became a nightmare. Every tool change meant debugging bash.

### Symlink Farms
Next came the symlink approach - keep everything in a git repo and symlink to `$HOME`. But some programs don't play nice with symlinks, and it felt ridiculous to symlink a 1KB config file. Plus, I wanted to update my dotfiles in git without immediately affecting my running system.

### Dotter
[Dotter](https://github.com/SuperCuber/dotter) was a breath of fresh air. I loved its minimalism and simplicity. The global vs local configurations seemed promising, but I never really used them - just meant storing multiple sets of files.

### Chezmoi
[Chezmoi](https://www.chezmoi.io/) is powerful and has excellent setup features. But I wanted something simpler and more opinionated. The templating system, while powerful, added complexity I didn't need. And crucially - it still didn't handle packages.

### The Missing Piece
Every solution I tried solved dotfiles well, but none addressed the other half of my setup: packages. In today's world of amazing CLI/TUI tools, I'm constantly trying new things. I need my package management to be as fluid as my dotfile management.

## Enter Plonk

Plonk is the unified solution I wanted but couldn't find. It manages both packages AND dotfiles in one simple tool:

```bash
# Clone your environment to a new machine
plonk clone yourusername/dotfiles
# Done. Seriously.
```

## Core Philosophy

### Zero Configuration
Plonk works out of the box with sensible defaults. No YAML manifestos or Ruby DSLs required. Your dotfiles ARE the configuration.

### Filesystem as Truth
For dotfiles, the contents of `$PLONK_DIR` define what's managed. No separate tracking files to get out of sync. Add a file, it's managed. Remove it, it's not.

### State, Not Commands
Plonk tracks what should exist, not what commands were run. This makes operations idempotent and shareable between machines.

### Just Enough, Not More
Every feature in plonk has to earn its place. Complex workflows and edge cases are explicitly not goals. Do the common things exceptionally well.

## The Package Manager Manager

Plonk's vision for package management is to be the "package manager manager" - one interface for all the package management operations developers need across the 5 package managers we typically juggle:

- **Homebrew** (brew) - macOS/Linux system packages and tools
- **Cargo** (cargo) - Rust packages
- **Go** (go) - Go binaries
- **PNPM** (pnpm) - Fast Node.js packages
- **UV** (uv) - Python tool management

Plonk follows a simple track/apply model: you install packages with your preferred manager, then use `plonk track` to record them. On a new machine, `plonk apply` installs everything that's missing.

## Why I Built Plonk

Beyond scratching my own itch, plonk exists for three reasons:

### 1. Learning AI-Assisted Development
I wanted to see if I could build something genuinely useful primarily through AI coding agents. The answer is a resounding yes - with important caveats about context engineering, clear specifications, and knowing what you want technically.

### 2. Solving My Own Problem
As someone who's constantly experimenting with new tools, I needed something that could keep up with my tinkering without becoming a maintenance burden itself.

### 3. Returning to Go
After 7-8 years away from Go, I wanted to refresh my knowledge. Go's opinionated design, strict formatting, excellent tooling, and single-binary output seemed perfect for both AI-assisted development and a tool like plonk.

## Goals

* **One command dev environment setup** - From fresh OS to fully configured in minutes
* **Unified management** - Packages and dotfiles together, as they should be
* **Cross-platform** - Same tool whether you're on macOS, Linux, or WSL
* **Git-native** - Your config directory is just a git repo, use all your normal workflows
* **Idiomatic Go** - Simple, maintainable code that's easy to contribute to
* **Fluent UI** - Commands that make sense and output that's actually helpful
* **Future-ready** - Designed to grow with new possibilities

## Non-Goals

* **Enterprise features** - This is for developers, not ops teams
* **Every edge case** - 80/20 rule applies strongly here
* **Backwards compatibility forever** - We'll evolve thoughtfully but not be held back
* **Plugin system** - Core functionality over infinite extensibility

## Lessons from AI-Assisted Development

Building plonk with AI coding agents taught me valuable lessons about the importance of:
- Clear specifications and expected behaviors upfront
- Documenting and planning before executing
- Providing rich context to get good results
- Maintaining architectural discipline to avoid complexity creep

The codebase itself reflects these learnings - it's designed to be AI-friendly with clear interfaces, minimal magic, and straightforward patterns.

## Who Is This For?

Plonk is for developers who:
- Set up new development environments regularly
- Want their tools to get out of the way
- Prefer convention over configuration
- Value simplicity and maintainability
- Are tired of remembering which dotfile manager syntax to use
- Constantly experiment with new CLI/TUI tools
- Need unified management across multiple package managers

## The Name

"Plonk" means to put something down firmly and decisively. That's what it does with your development environment - plonks it right where it should be.
