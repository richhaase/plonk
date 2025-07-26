# IDEA: Config Command Simplification

## Current State

Three subcommands under `config`:
- `plonk config show` - Display current config
- `plonk config edit` - Edit config file
- `plonk config validate` - Validate config syntax

User note: "These could be simplified, and validate is the only truly required functionality, but I don't think edit or show add much overhead."

## Issues to Consider

1. **Subcommand Depth**: Requires typing `plonk config <action>`
2. **Discoverability**: Users might not realize these exist
3. **Validate Importance**: Most critical but least obvious command
4. **Edit Redundancy**: Could just tell users to use their editor

## Potential Solutions

### Option A: Keep Current Structure (Status Quo)
```bash
plonk config show
plonk config edit
plonk config validate
```

**Pros**:
- Clear grouping
- Familiar pattern (like `git config`)
- No breaking changes

**Cons**:
- More typing
- Three commands for simple operations

### Option B: Flags on Single Command
```bash
plonk config              # default: show
plonk config --edit       # or -e
plonk config --validate   # or -v
```

**Pros**:
- Simpler command structure
- Default behavior (show) is most common

**Cons**:
- Flags feel less discoverable than subcommands
- `--validate` conflicts with common `--verbose`

### Option C: Promote Validate, Simplify Others
```bash
plonk validate           # promoted to top-level
plonk config            # shows config
plonk config --edit     # edits config
```

**Pros**:
- Important command more visible
- Simpler common case
- Validate useful for CI/CD

**Cons**:
- Inconsistent patterns
- Breaking change

### Option D: Auto-Validation Everywhere
- Validate on every config read automatically
- Remove explicit validate command
- Show clear errors when config is invalid

**Pros**:
- Users can't forget to validate
- One less command

**Cons**:
- No explicit way to check config without side effects
- Can't validate without running another command

### Option E: Config File Management Pattern
```bash
plonk config                    # shows path and status
plonk config path              # prints path only
plonk config check             # validate (clearer name?)
plonk edit-config              # top-level for importance
```

**Pros**:
- More functionality exposed
- Clearer naming

**Cons**:
- More commands
- `check` vs `validate` naming debate

## Questions for Discussion

1. How often do users need to validate configs explicitly?
2. Is `edit` just convenience or actively useful?
3. Should validation happen automatically on every command?
4. What about config migration/upgrade commands for the future?

## Related Considerations

- Config file location: Should `plonk config` show the path?
- Multiple configs: Might we need `--config-file` flag globally?
- Config init: Should `plonk init` just create a default config?
- Dry run: Should validate show what would happen?

## Integration with Future Features

With hooks and AI Lab configs coming:
- Will config complexity increase?
- Need for config sections: `plonk config show hooks`?
- Config generation: `plonk config generate ai-lab`?

## Recommendation Placeholder

_To be filled after discussion_
