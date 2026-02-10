# Template Dotfiles

Re-introduce template dotfile support, removed during the dotfiles module simplification (48634e2). This time: minimal, env-var-only substitution with no Go `text/template` dependency.

## How It Works

A file like `$PLONK_DIR/gitconfig.tmpl` is a template dotfile. During `plonk apply`, placeholders are substituted with environment variables, then the rendered result is written to `$HOME/.gitconfig` (dot prefix added, `.tmpl` extension stripped).

```
$PLONK_DIR/gitconfig.tmpl  →  render  →  $HOME/.gitconfig
$PLONK_DIR/config/git/config.tmpl  →  render  →  $HOME/.config/git/config
```

### Syntax

`{{VAR_NAME}}` — literal double braces around an environment variable name.

Example `gitconfig.tmpl`:

```ini
[user]
    email = {{EMAIL}}
    name = {{GIT_USER_NAME}}
```

### Rules

- `.tmpl` files are always rendered, never copied verbatim
- If any referenced variable is not set in the environment, `apply` fails with an error listing the missing variables
- A plain file and a `.tmpl` file must not target the same destination (error if both `gitconfig` and `gitconfig.tmpl` exist)

## Integration With Existing Operations

- **`plonk add`** — no change. Templates are hand-authored in `$PLONK_DIR`.
- **`plonk apply`** — render `.tmpl` files via env var substitution before writing.
- **`plonk status`** — compare rendered output (not raw template) against deployed file.
- **`plonk diff`** — diff rendered output vs deployed file.
- **`plonk rm`** — recognize `.tmpl` source files map to non-`.tmpl` targets.
- **`plonk doctor`** — validate all `.tmpl` files can render (env vars set) and no plain/template conflicts.

## Implementation

Changes touch the existing dotfiles module — no new files.

### `dotfiles.go`

- Add a constant for `.tmpl` extension
- In path-mapping logic, strip `.tmpl` from target paths
- Add `renderTemplate(content []byte) ([]byte, error)` — scan for `{{VAR}}` patterns, look up each via `os.Getenv`, error if any are missing
- In `apply`: route `.tmpl` files through `renderTemplate` before writing
- In status/diff: route `.tmpl` files through `renderTemplate` before comparing

### `reconcile.go`

- Add conflict check: error if both `foo` and `foo.tmpl` exist in the source dir

### `dotfiles_test.go`

- Template rendering: valid substitution, missing var error, multiple vars
- Status/diff with templates: rendered comparison works
- Conflict detection: plain + `.tmpl` targeting same destination

## Not In Scope

- No changes to `add` or `rm`
- No YAML variable files
- No Go `text/template` — simple regex-based `{{VAR}}` substitution only
- No nested templates, conditionals, loops, or other template features
