# Doctor Command

Checks system health and configuration for plonk.

## Synopsis

```bash
plonk doctor [options]
```

## Description

The doctor command performs comprehensive health checks on your plonk installation and system configuration. It verifies system requirements, package manager availability, configuration validity, and PATH setup. Doctor provides detailed suggestions for fixing any issues found, making it an essential tool for troubleshooting.

The command runs checks across six categories and reports findings with status levels (PASS, WARN, FAIL, INFO), providing an overall health assessment of your plonk setup.

## Options

- `-o, --output` - Output format (table/json/yaml)

## Behavior

### Health Check Categories

1. **System** - System requirements (Go version, OS, architecture)
2. **Environment** - Environment variables (HOME, PLONK_DIR, PATH)
3. **Permissions** - File permissions on configuration directory
4. **Configuration** - Configuration and lock file existence and validity
5. **Package Managers** - Availability of each package manager
6. **Installation** - Plonk executable accessibility and PATH configuration

### Status Levels

Individual checks report one of four statuses:
- **PASS** (green) - Everything working correctly
- **WARN** (yellow) - Possible degraded behavior or issues needing attention
- **FAIL** (red) - Critical issues preventing plonk from functioning
- **INFO** (blue) - Informational messages (not affecting overall health)

Overall system health:
- **healthy** - All checks pass or only have warnings/info
- **warning** - Some checks have warnings but no failures
- **unhealthy** - One or more checks have failures

### Package Manager Checks

- Homebrew is required (prerequisite) - will show FAIL if missing
- Language package managers are optional - will show WARN if missing
- Only verifies availability, not full functionality
- Use `plonk clone` to automatically install package managers needed by your managed packages

### PATH Configuration Analysis

Doctor provides detailed PATH analysis:
- Shows which package directories are in PATH (marked as available)
- Warns about directories that exist but aren't in PATH
- Notes directories that don't exist yet (informational)
- Provides shell configuration suggestions for missing paths

### Output Formats

**Table Format** (default):
- Hierarchical display with categories and checks
- Color-coded status text
- Detailed messages, issues, and suggestions

**JSON/YAML Format**:
- Structured with overall status and checks array
- Each check includes: name, category, status, message, details
- Optional fields: issues, suggestions

### Error Handling

All health checks run within a 30-second timeout. The command continues even if individual checks fail, providing a complete assessment.

## Examples

```bash
# Run health checks
plonk doctor

# Output as JSON for scripting
plonk doctor -o json

# Output as YAML
plonk doctor -o yaml
```

## Integration

- Use before `plonk clone` to verify prerequisites
- Run when plonk commands fail unexpectedly
- Check after system updates or configuration changes
- Note: `plonk clone` automatically installs package managers for packages in your plonk.lock

## Notes

- There is no `--fix` flag by design - use `plonk clone` for automatic setup
- Missing plonk.lock is not an error (valid for dotfiles-only usage)
- Configuration file issues result in fallback to defaults
- Respects NO_COLOR environment variable for accessibility
