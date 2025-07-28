# Improve Filtering for ~/.config Directory

## Problem Summary

When running `plonk status --dotfiles --unmanaged`, the output for files under `~/.config` contains too many non-configuration files that make it difficult to identify actual configuration files users might want to manage.

## Current Behavior

The scanner shows all files in expanded directories (like `.config`), which includes:
- Cache files and directories
- Lock files
- Session data
- Log files
- Temporary files
- Application state/data files

## Examples of Noise

Common patterns that are likely not user configuration:
- `*cache*` directories and files
- `*.lock` files
- `logs/` directories
- Session files
- Database files (`.db`, `.sqlite`)
- Compiled or generated files

## Desired Behavior

Show only files that are likely to be user-editable configuration files, such as:
- Files with config extensions: `.conf`, `.toml`, `.yaml`, `.yml`, `.json`, `.ini`, `.cfg`
- Known configuration files for popular tools
- Exclude obvious non-config patterns

## Potential Solutions

1. **Enhanced ignore patterns**: Add more default patterns for common non-config files
2. **Extension-based filtering**: Only show files with known config extensions
3. **Hybrid approach**: Combine ignore patterns with preferred extensions
4. **Configurable filtering**: Allow users to customize what's shown via config

This improvement would significantly enhance the usability of the `--unmanaged` flag by reducing noise and focusing on files users actually care about managing.
