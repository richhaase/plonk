# Package Manager Metadata & Parsing Pipeline

Status: Draft design for v2.1 package manager refactor.
Scope: Describes the configuration schema and processing stages for parsing
package manager output and deriving metadata in a manager-agnostic way.

## Goals

- Move all manager-specific parsing and metadata logic into configuration.
- Support both simple line-based output and richer JSON structures (arrays or maps).
- Provide a generic mechanism for:
  - Normalizing package names (e.g., stripping npm scopes if desired).
  - Extracting additional metadata (scope, full name, version, etc.).
- Keep the core engine (`GenericManager`, `ManagerRegistry`) manager-agnostic.

## Configuration Schema

The schema is defined in `internal/config/managers.go`:

```go
type ManagerConfig struct {
    Binary            string                           `yaml:"binary,omitempty"`
    List              ListConfig                       `yaml:"list,omitempty"`
    Install           CommandConfig                    `yaml:"install,omitempty"`
    Upgrade           CommandConfig                    `yaml:"upgrade,omitempty"`
    UpgradeAll        CommandConfig                    `yaml:"upgrade_all,omitempty"`
    Uninstall         CommandConfig                    `yaml:"uninstall,omitempty"`
    NameTransform     *NameTransformConfig             `yaml:"name_transform,omitempty"`
    MetadataExtractors map[string]MetadataExtractorConfig `yaml:"metadata_extractors,omitempty"`
}

type ListConfig struct {
    Command       []string `yaml:"command,omitempty"`
    Parse         string   `yaml:"parse,omitempty"`
    ParseStrategy string   `yaml:"parse_strategy,omitempty"` // alias for Parse
    JSONField     string   `yaml:"json_field,omitempty"`
}

type NameTransformConfig struct {
    Type        string `yaml:"type,omitempty"`        // e.g. "regex"
    Pattern     string `yaml:"pattern,omitempty"`     // regex pattern
    Replacement string `yaml:"replacement,omitempty"` // replacement template
}

type MetadataExtractorConfig struct {
    Pattern string `yaml:"pattern,omitempty"` // optional regex
    Group   int    `yaml:"group,omitempty"`   // capturing group to use
    Source  string `yaml:"source,omitempty"`  // e.g. "json_field", "name"
    Field   string `yaml:"field,omitempty"`   // JSON field name when Source is "json_field"
}
```

### Parse Strategy

- `ListConfig.Parse` (and alias `parse_strategy`) controls how `GenericManager`
  interprets the list command output:
  - `"lines"`: one package per line; core takes the first token.
  - `"json"`: JSON array of objects; core extracts the string field named by `json_field`.
- Future extensions may add additional parse modes (e.g., JSON maps or nested trees).

### Name Transform

`NameTransform` lets a manager normalize package names after parsing:

- Example: strip npm scopes from names for display, or normalize aliases.
- Implemented as a pluggable transform; initial implementation will support
  a regex-based transform:
  - `pattern` is a regular expression.
  - `replacement` follows Go regexp replacement semantics.

### Metadata Extractors

`MetadataExtractors` is a map from metadata key to extraction rule:

- Example for npm scoped packages:

```yaml
metadata_extractors:
  scope:
    pattern: "^@([^/]+)/.*$"
    group: 1
  full_name:
    source: "name"
```

- Example for JSON-derived version fields:

```yaml
metadata_extractors:
  version:
    source: "json_field"
    field: "version"
```

- The engine will evaluate extractors after parsing names and/or JSON objects
  and populate `metadata` for lock entries and operation results.

## Processing Pipeline

The intended end-to-end pipeline for listing packages is:

1. **Execute list command**
   - Use `ListConfig.Command` against the configured binary.

2. **Parse output**
   - Use `ListConfig.Parse` / `parse_strategy` to choose parser:
     - `"lines"` → split on newlines, extract package names.
     - `"json"` → unmarshal JSON array, use `json_field` as the name.
   - Future: support JSON maps (e.g., npm dependencies) via additional parse modes.

3. **Apply name transform (optional)**
   - If `NameTransform` is configured, apply it to each parsed package name.

4. **Extract metadata (optional)**
   - For each package, run configured `MetadataExtractors` to derive values like
     `scope`, `version`, or `full_name`.
   - These metadata values are stored alongside packages in the lock file and
     can be used by higher-level operations (e.g., `upgrade` matching).

5. **Reconciliation**
   - Parsed names and metadata feed into the generic reconciliation engine
     (no manager-specific logic).

## Backward Compatibility

- Existing managers (brew, gem, cargo, conda, uv, pipx, npm, pnpm) continue
  to work with their configuration:
  - `parse` / `parse_strategy` and `json_field` remain the primary selectors.
  - npm and pnpm have been migrated to JSON-based list output using the new
    strategies (`json` / `json-map`) without changing lock-file semantics.
- New fields (`name_transform`, `metadata_extractors`) are optional and are
  currently used for npm only; other managers remain unchanged until explicitly
  configured.

## Next Steps

1. Implement support for additional JSON parse modes (e.g., nested maps beyond
   simple `dependencies` keys) driven by `ListConfig`.
2. Wire `NameTransform` into the package listing pipeline.
3. Extend `MetadataExtractors` usage beyond install-time lock writes to
   upgrade matching and other operations where richer metadata is useful.
4. Add validation for regex patterns and extractor configuration in
   `internal/config/validators.go`.
