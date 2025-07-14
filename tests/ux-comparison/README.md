# Plonk UI/UX Comparison Tool

A simple tool to compare UI/UX output between the installed version of plonk and a compiled version, helping ensure that refactoring doesn't break the CLI and that enhancements work as expected.

## Prerequisites

- Python 3 with PyYAML:
  ```bash
  pip install pyyaml
  ```
- An installed version of `plonk` in your PATH
- The `just` build tool

## Usage

### First Time Setup

Capture a baseline from your currently installed plonk:

```bash
./compare.py --update-baseline
```

This creates a `baseline.json` file with the current UI outputs.

### Testing Your Changes

After making code changes, run:

```bash
./compare.py
```

This will:
1. Build your code (unless `--skip-build` is used)
2. Run the same test scenarios against both versions
3. Show which scenarios have identical output vs differences
4. Save detailed diffs for review

### Options

- `--update-baseline` - Update the baseline from installed plonk
- `-f PATTERN, --filter PATTERN` - Only run scenarios matching the pattern
- `--skip-build` - Skip building plonk (if already built)
- `-v, --verbose` - Show detailed output

### Interpreting Results

The tool simply shows you what changed. You decide if the changes are expected:

- **During refactoring**: Any UI changes likely indicate a bug
- **During enhancement**: Changes should match your new features

Results are saved in timestamped directories under `results/` with:
- Individual diff files for each changed scenario
- A markdown summary report

### Customizing Test Scenarios

Edit `scenarios.yaml` to add or modify test cases. The default scenarios include:
- Basic commands (help, version)
- List operations (all, packages, dotfiles)
- Status commands
- Info and manager queries
- Error cases
- Safe dry-run operations

## Safety

The tool protects your system by:
- Using a temporary copy of your plonk configuration
- Never modifying your actual plonk setup
- Only running read-only commands (except dry-run tests)
- Restoring the original environment after testing

## Example Workflow

1. Before starting a refactor:
   ```bash
   ./compare.py --update-baseline
   ```

2. After making changes:
   ```bash
   ./compare.py
   ```

3. Review any differences to ensure they're expected
4. If adding new features, update the baseline when done:
   ```bash
   ./compare.py --update-baseline
   ```
