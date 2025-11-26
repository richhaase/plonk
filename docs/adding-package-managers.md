# Adding Custom Package Managers

Plonk can manage any package manager that exposes CLI commands. You do not need to change code—define managers entirely in `plonk.yaml`.

## Quick Start
1) Ensure the manager’s CLI is installed and on `$PATH`.
2) Add a `managers.<name>` block in `~/.config/plonk/plonk.yaml`.
3) Run `plonk doctor` (or `plonk packages`) to confirm the manager is detected and its packages are parsed correctly.
4) Manage packages with `plonk install <name:pkg>` or set `default_manager` to avoid prefixes.

## List Parsing Strategies
Choose one per manager in `list`:
- `lines` (default): splits stdout on newlines, takes the first token per line.
- `json`: expects a JSON array of objects; `json_field` names the string field to extract.
- `json-map`: expects a JSON object (or nested object at `json_field`); collects map keys.
- `jsonpath`: JSONPath selectors drive extraction:
  - `keys_from`: JSONPath selecting object(s); collects their keys.
  - `values_from`: JSONPath selecting string value(s).
  - `normalize`: `lower` or `none` (default).
Behavior: invalid JSON or selector errors fail; if output is non-empty but no names are extracted, parsing fails to avoid silent “all missing.”

## Field Reference (per manager)
```yaml
managers:
  <name>:
    binary: <cli name>            # required
    list:
      command: [<cmd>, ...]       # required
      parse_strategy: lines|json|json-map|jsonpath
      json_field: <field>         # json/json-map only
      keys_from: <jsonpath>       # jsonpath only
      values_from: <jsonpath>     # jsonpath only
      normalize: lower|none       # optional (default none)
    install:
      command: [<cmd>, ...]
      idempotent_errors: ["already installed", ...]
    uninstall:
      command: [<cmd>, ...]
      idempotent_errors: ["not installed", ...]
    upgrade:
      command: [<cmd>, ...]
      idempotent_errors: ["already up-to-date", ...]
    upgrade_all:
      command: [<cmd>, ...]
      idempotent_errors: ["already up-to-date", ...]
```
Commands are arrays (one token per element). `idempotent_errors` make operations treat matching stderr/stdout as success.

## Examples

### pnpm (built-in default, using JSONPath)
```yaml
managers:
  pnpm:
    binary: pnpm
    list:
      command: [pnpm, list, -g, --depth=0, --json]
      parse_strategy: jsonpath
      keys_from: "$[*].dependencies"
    install:
      command: [pnpm, add, -g, "{{.Package}}"]
    uninstall:
      command: [pnpm, remove, -g, "{{.Package}}"]
    upgrade:
      command: [pnpm, update, -g, "{{.Package}}"]
    upgrade_all:
      command: [pnpm, update, -g]
```

### npm (map of dependencies)
```yaml
managers:
  npm:
    binary: npm
    list:
      command: [npm, list, -g, --depth=0, --json]
      parse_strategy: jsonpath
      keys_from: "$.dependencies"
    install:
      command: [npm, install, -g, "{{.Package}}"]
    uninstall:
      command: [npm, uninstall, -g, "{{.Package}}"]
```

## Tips
- Scopes/case: use `normalize: lower` if the manager’s output varies in case.
- Exit codes: plonk relies on command exit codes; mark harmless errors in `idempotent_errors`.
- Debug parsing: run the list command manually, capture its JSON, and test JSONPath selectors with a local tool before updating `plonk.yaml`.
- Shareable: custom managers live in your `plonk.yaml`; others can reuse them by copying that block.
