# Doctor Command

The `plonk doctor` command checks system health and configuration.

## Description

The doctor command performs comprehensive health checks on your plonk installation and system configuration. It verifies system requirements, package manager availability, configuration validity, and PATH setup. With the `--fix` flag, doctor can automatically install missing package managers, making it an essential tool for both troubleshooting issues and initial system setup.

## Behavior

### Core Function

Doctor runs a series of health checks across six categories and reports findings with three status levels:
- **PASS** (✅): Everything working correctly
- **WARN** (⚠️): Possible degraded behavior or issues needing attention
- **ERROR** (❌): Critical issues preventing plonk from functioning

Overall status reflects the most severe issue found: ERROR > WARNING > PASS.

### Check Categories

1. **System**
   - System requirements (Go version, OS, architecture)

2. **Environment**
   - Environment variables (HOME, PLONK_DIR, PATH)

3. **Permissions**
   - File permissions on configuration directory

4. **Configuration**
   - Configuration file existence and validity
   - Lock file existence and validity
   - Shows details like package counts, ignore patterns

5. **Package Managers**
   - Availability of each package manager
   - Functionality verification (currently redundant with availability)

6. **Installation**
   - Plonk executable accessibility
   - PATH configuration for package installation directories

### Command Options

- `--fix` - Offer to install missing package managers
- `--yes` - Auto-install without prompts (requires `--fix`)
- `-o, --output` - Output format (table/json/yaml)

### Output Formats

**Table Format** (default):
- Hierarchical display with categories and checks
- Color-coded status indicators
- Detailed messages, issues, and suggestions
- Human-readable with formatting

**JSON Format**:
- Structured with `overall` status and `checks` array
- Each check includes: name, category, status, message, details
- Optional fields: issues, suggestions
- Uses different field naming (details vs Details)

**YAML Format**:
- Same structure as JSON but in YAML syntax
- Preserves all fields including multi-line suggestions

### Fix Behavior

With `--fix` flag:
- Identifies missing package managers
- Prompts user to install each missing manager (unless `--yes`)
- Installs via official methods:
  - Homebrew: Official installer script
  - Cargo: rustup installer
  - Others: Via default_manager

Currently limited to package manager installation only.

### Error Conditions

ERROR status occurs when:
- Default package manager (e.g., brew) is not installed
- PLONK_DIR does not exist
- Critical system requirements not met

### PATH Configuration Check

Doctor provides detailed PATH analysis:
- Shows which package directories are in PATH (✅)
- Warns about directories that exist but aren't in PATH (⚠️)
- Notes directories that don't exist yet (ℹ️)
- Provides shell configuration suggestions for missing paths

### Integration with Setup

The `plonk setup` command uses the same code as `doctor --fix` internally for package manager installation. This ensures consistency between initial setup and later health checks.

### Special Behaviors

- Package manager "availability" and "functionality" are currently equivalent
- PATH suggestions are informational only - not auto-fixable
- Missing plonk.lock is not an error (valid for dotfiles-only usage)
- Configuration file issues result in fallback to defaults

## Implementation Notes

The doctor command provides comprehensive system health checking through a structured diagnostics system:

**Command Structure:**
- Entry point: `internal/commands/doctor.go`
- Health checks: `internal/diagnostics/health.go`
- Fix functionality: delegates to `internal/setup/tools.go`

**Key Implementation Flow:**

1. **Command Processing:**
   - Parses `--fix` and `--yes` flags
   - Runs `diagnostics.RunHealthChecks()` with 30-second timeout
   - Delegates fix behavior to setup package
   - Re-runs health checks after fixes to show updated status

2. **Health Check Categories:**
   - **DISCREPANCY**: Uses different status terminology than documented
   - Implementation uses: `pass`, `warn`, `fail`, `info`
   - Documentation says: `PASS`, `WARN`, `ERROR`
   - Categories: `system`, `environment`, `permissions`, `configuration`, `package-managers`, `installation`

3. **Overall Status Calculation:**
   - **DISCREPANCY**: Uses different overall status values
   - Implementation: `healthy`, `warning`, `unhealthy`
   - Documentation: `PASS`, `WARNING`, `ERROR`
   - Priority: `fail` > `warn` > `pass` (any failure makes overall unhealthy)

4. **Individual Health Checks:**
   - System Requirements: OS/architecture validation
   - Environment Variables: HOME, PLONK_DIR, PATH checks
   - Permissions: Configuration directory access
   - Configuration File: Existence and validity of plonk.yaml
   - Configuration Validity: YAML parsing and validation
   - Lock File: Existence and validity of plonk.lock
   - Lock File Validity: YAML parsing and structure validation
   - Package Manager Availability: Tests each manager via registry
   - **DISCREPANCY**: Package Manager Functionality is identical to availability (not differentiated)
   - Executable Path: Plonk binary accessibility
   - PATH Configuration: Package manager directories in PATH

5. **Fix Behavior:**
   - **DISCREPANCY**: Actually delegates to `setup.CheckAndInstallToolsFromReport()`, not direct `doctor --fix`
   - Uses same setup infrastructure as `plonk setup`
   - Only fixes package manager installation issues
   - `--yes` flag bypasses interactive prompts
   - Re-runs all health checks after fixes

6. **Output Formatting:**
   - Table format uses hierarchical display with fixed category ordering
   - JSON/YAML formats use flat check arrays
   - **DISCREPANCY**: Documentation mentions color coding but implementation uses emoji icons
   - Status icons: ✅ (pass), ⚠️ (warn), ❌ (fail), ℹ️ (info), ❓ (unknown)

**Category Organization:**
- Fixed ordering: system → environment → permissions → configuration → package-managers → installation
- Uses `strings.Title()` for category display formatting
- Groups all checks by category for table output

**Error Conditions:**
- 30-second timeout for all health checks combined
- Missing plonk.lock treated as informational, not error
- Configuration file errors fall back to defaults
- Package manager unavailability results in warnings, not failures

**Integration Details:**
- **DISCREPANCY**: Shares code with setup via `setup.CheckAndInstallToolsFromReport()`
- Does not call `plonk doctor --fix` as a subprocess
- Uses same package manager installation logic as setup command

**Bugs Identified:**
1. Status terminology mismatch between docs (`PASS`/`WARN`/`ERROR`) and implementation (`pass`/`warn`/`fail`)
2. Overall status values differ: docs use `PASS`/`WARNING`/`ERROR`, implementation uses `healthy`/`warning`/`unhealthy`
3. Package Manager Functionality check is identical to Availability check (no differentiation)
4. Documentation says "color-coded" but implementation uses emoji icons
5. Documentation doesn't mention 30-second timeout for all checks

## Improvements

- Extend `--fix` to address all fixable issues, not just package managers
- Review and standardize all health check behaviors
- Revisit check categories for better organization
- Remove redundant "functionality" check or differentiate from "availability"
- Consider having setup directly call doctor instead of duplicating code
- Add auto-fix capabilities for PATH configuration issues
- Provide copy-paste ready PATH export commands for user's specific shell
