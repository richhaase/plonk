# IDEA: Output Format Flag Optimization

## Current State

- Every command accepts `-o table|json|yaml` flag
- Default is `table` for human readability
- User note: "We do need the output flag on all commands, to make it easy for plonk to use for machines and humans"

## Issues to Consider

1. **Repetition**: Users must specify `-o json` on every command in scripts
2. **Environment Detection**: Could auto-detect when piped or in CI
3. **Global Config**: No way to set preferred format permanently
4. **Format Completeness**: Do all commands support all formats equally well?

## Potential Solutions

### Option A: Keep Current Design (Status Quo)
```bash
plonk status -o json
plonk ls -o yaml
plonk info ripgrep -o json
```

**Pros**:
- Explicit control
- No surprises
- Simple implementation

**Cons**:
- Repetitive in scripts
- No smart defaults

### Option B: Environment Variable
```bash
export PLONK_OUTPUT=json
plonk status              # uses json
plonk ls                  # uses json
plonk info -o table       # override still works
```

**Pros**:
- Set once for session
- Scripts can set their preference
- Follows common patterns (like EDITOR)

**Cons**:
- Hidden state
- Can cause surprises

### Option C: Auto-Detection
```bash
plonk status              # table when terminal
plonk status | jq .       # json when piped
CI=true plonk status      # json when CI detected
plonk status -o table     # explicit override
```

**Pros**:
- Smart defaults
- Works well in common cases
- No configuration needed

**Cons**:
- Magic behavior
- Can be surprising
- Harder to test

### Option D: Config File Setting
```yaml
# plonk.yaml
defaults:
  output_format: json
```

```bash
plonk status              # uses config default
plonk status -o table     # override
```

**Pros**:
- Persistent preference
- Visible in config
- Per-project settings possible

**Cons**:
- Another config option
- Not helpful for one-off scripts

### Option E: Combined Approach
1. Check explicit flag first (`-o`)
2. Then environment variable (`PLONK_OUTPUT`)
3. Then config file setting
4. Then auto-detect (if enabled)
5. Finally default to table

**Pros**:
- Maximum flexibility
- Covers all use cases
- Predictable precedence

**Cons**:
- Complex implementation
- Many ways to set it

## Questions for Discussion

1. How often do users switch between formats?
2. Is auto-detection too magic?
3. Should some commands have different defaults? (e.g., `plonk config show` as YAML?)
4. What about partial outputs (errors in table, data in JSON)?

## Format-Specific Considerations

### Table Format
- Best for: Human reading, quick status checks
- Issues: Column width, wrapping, Unicode support

### JSON Format
- Best for: Scripting, jq processing, CI/CD
- Issues: Large outputs, pretty vs compact

### YAML Format
- Best for: Config generation, human editing
- Issues: Complex escaping, multiline strings

## Alternative Ideas

1. **Shorthand Flags**: `-j` for JSON, `-y` for YAML
2. **Format in Command**: `plonk status:json` or `plonk-json status`
3. **Separate Commands**: `plonk-cli` vs `plonk-api`
4. **Content Negotiation**: Different formats for different parts (errors always text?)

## Recommendation Placeholder

_To be filled after discussion_
