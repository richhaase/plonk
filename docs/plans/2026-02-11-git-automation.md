# Git Automation for Plonk

Plonk manages dotfiles and packages in `$PLONK_DIR` (a git repo), but users must manually run `git add/commit/push/pull` after every mutation. This plan adds:

1. **Auto-commit** — every command that mutates `$PLONK_DIR` (`add`, `rm`, `track`, `untrack`, `add -y`, `config edit`) automatically commits with a descriptive message.
2. **`plonk push`** — commits any pending changes and pushes to the remote.
3. **`plonk pull`** — pulls remote changes (auto-committing local changes first), with an `--apply|-a` flag to run `plonk apply` afterward.
4. **Config opt-out** — a `git.auto_commit` setting in `plonk.yaml` to disable git operations.
5. **Graceful degradation** — if `$PLONK_DIR` is not a git repo, warn and mention the config setting.

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Auto-commit scope | Commit only (no push) | User controls when to push; works offline; avoids surprise network calls |
| Pull with dirty state | Auto-commit first, then pull | Avoids stash fragility; commits are cheap and reversible |
| Commit messages | Auto-generated from command + args | e.g., `"plonk: add .zshrc .vimrc"`, `"plonk: track brew:ripgrep"` |
| Git library | `os/exec` calling `git` CLI | Simpler than go-git for push/pull (auth, SSH agent, credential helpers all work); go-git is already a dep but its push/pull auth story is painful |
| Config key | `git.auto_commit: true` (default) | Nested under `git` for future git settings; opt-out via `false` |
| Non-git-repo behavior | Warn once per command, skip git ops | Don't error — plonk should still work without git |

## Work

### 1. Create `internal/gitops` package

This is the core git operations module. All git interaction is centralized here.

**`internal/gitops/gitops.go`**

```go
package gitops

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
)

// Client wraps git CLI operations on a specific directory.
type Client struct {
    dir string // the $PLONK_DIR path
}

// New creates a git client for the given directory.
func New(dir string) *Client {
    return &Client{dir: dir}
}

// IsRepo checks if dir is inside a git work tree.
func (c *Client) IsRepo() bool {
    cmd := exec.Command("git", "-C", c.dir, "rev-parse", "--is-inside-work-tree")
    out, err := cmd.Output()
    return err == nil && strings.TrimSpace(string(out)) == "true"
}

// HasRemote checks if the repo has at least one remote configured.
func (c *Client) HasRemote() bool {
    cmd := exec.Command("git", "-C", c.dir, "remote")
    out, err := cmd.Output()
    return err == nil && strings.TrimSpace(string(out)) != ""
}

// IsDirty returns true if there are uncommitted changes (staged or unstaged).
func (c *Client) IsDirty() (bool, error) {
    cmd := exec.Command("git", "-C", c.dir, "status", "--porcelain")
    out, err := cmd.Output()
    if err != nil {
        return false, fmt.Errorf("git status failed: %w", err)
    }
    return strings.TrimSpace(string(out)) != "", nil
}

// Commit stages all changes and commits with the given message.
// Returns nil if there's nothing to commit.
func (c *Client) Commit(message string) error {
    dirty, err := c.IsDirty()
    if err != nil {
        return err
    }
    if !dirty {
        return nil // nothing to commit
    }

    // git add -A
    addCmd := exec.Command("git", "-C", c.dir, "add", "-A")
    if out, err := addCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git add failed: %w\n%s", err, out)
    }

    // git commit -m <message>
    commitCmd := exec.Command("git", "-C", c.dir, "commit", "-m", message)
    if out, err := commitCmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git commit failed: %w\n%s", err, out)
    }

    return nil
}

// Push pushes to the default remote/branch.
func (c *Client) Push() error {
    cmd := exec.Command("git", "-C", c.dir, "push")
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git push failed: %w\n%s", err, out)
    }
    return nil
}

// Pull pulls from the default remote/branch.
func (c *Client) Pull() error {
    cmd := exec.Command("git", "-C", c.dir, "pull")
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("git pull failed: %w\n%s", err, out)
    }
    return nil
}

// CommitMessage builds a commit message from a plonk command and its arguments.
func CommitMessage(command string, args []string) string {
    if len(args) == 0 {
        return fmt.Sprintf("plonk: %s", command)
    }
    // Truncate long arg lists
    display := args
    suffix := ""
    if len(display) > 5 {
        display = display[:5]
        suffix = fmt.Sprintf(" (+%d more)", len(args)-5)
    }
    return fmt.Sprintf("plonk: %s %s%s", command, strings.Join(display, " "), suffix)
}
```

**`internal/gitops/gitops_test.go`**

Test using a temp directory initialized with `git init`:
- `TestIsRepo` — true for git dir, false for plain dir
- `TestIsDirty` — false for clean repo, true after creating a file
- `TestCommit` — creates file, commits, verifies clean state and commit message in log
- `TestCommitNoop` — no error when nothing to commit
- `TestCommitMessage` — verify formatting for various command/arg combos
- `TestPushPull` — can be skipped or tested with a local bare repo as remote

### 2. Add `auto_commit` to config

**`internal/config/config.go`**

Add a `Git` struct and field to `Config`:

```go
// GitConfig contains git-related configuration
type GitConfig struct {
    AutoCommit *bool `yaml:"auto_commit,omitempty"`
}

// Add to Config struct:
type Config struct {
    // ... existing fields ...
    Git GitConfig `yaml:"git,omitempty"`
}
```

Use `*bool` so we can distinguish "not set" (nil → default true) from explicit `false`. Add a helper:

```go
// AutoCommitEnabled returns whether auto-commit is enabled.
// Defaults to true if not explicitly set.
func (c *Config) AutoCommitEnabled() bool {
    if c.Git.AutoCommit == nil {
        return true // default
    }
    return *c.Git.AutoCommit
}
```

**`internal/config/config.go` — `defaultConfig`**: No change needed. The zero value of `*bool` is `nil`, which the helper treats as `true`.

**`internal/config/user_defined.go` — `GetNonDefaultFields`**: The existing reflect-based logic will handle the nested struct automatically since it compares with `DeepEqual`. If `Git` is zero-value `GitConfig{}`, it matches the default and won't be saved. Verify this with a test.

**`internal/config/config_test.go`**: Add tests:
- `TestAutoCommitEnabledDefault` — nil → true
- `TestAutoCommitEnabledExplicitTrue` — true → true
- `TestAutoCommitEnabledExplicitFalse` — false → false
- `TestLoadConfigWithGitAutoCommit` — YAML round-trip with `git:\n  auto_commit: false`

**Documentation in `plonk.yaml`**:
```yaml
# Git integration
git:
  auto_commit: true  # Set to false to disable automatic git commits
```

### 3. Create `internal/gitops/autocommit.go` — convenience wrapper

This is what commands call after a successful mutation. It loads config, checks if auto-commit is enabled, checks if it's a git repo, and commits.

```go
package gitops

import (
    "github.com/richhaase/plonk/internal/config"
    "github.com/richhaase/plonk/internal/output"
)

// AutoCommit is the standard post-mutation hook.
// It loads config from configDir, checks if auto-commit is enabled,
// verifies git repo, and commits if appropriate.
// Errors are reported as warnings, never fatal.
func AutoCommit(configDir string, command string, args []string) {
    cfg := config.LoadWithDefaults(configDir)
    if !cfg.AutoCommitEnabled() {
        return
    }

    client := New(configDir)

    if !client.IsRepo() {
        output.Printf("Warning: %s is not a git repository; changes not committed. Set git.auto_commit: false in plonk.yaml to silence this warning.\n", configDir)
        return
    }

    msg := CommitMessage(command, args)
    if err := client.Commit(msg); err != nil {
        output.Printf("Warning: auto-commit failed: %v\n", err)
    }
}
```

### 4. Wire auto-commit into mutating commands

Each command that changes files in `$PLONK_DIR` gets a call to `gitops.AutoCommit()` after a successful operation. The call goes after output rendering but before returning.

**`internal/commands/add.go` — `runAdd()`**

After `output.RenderOutput(outputData)` and before the final return, add:

```go
// Auto-commit if any files were actually added/updated (not dry-run, not all-failed)
if !opts.DryRun && validateAddResultsErr(results) == nil {
    gitops.AutoCommit(configDir, "add", args)
}
```

Same pattern for `runSyncDrifted()` — add after output, guard on `!dryRun`.

**`internal/commands/rm.go` — `runRm()`**

After `output.RenderOutput(formatter)`, before final return:

```go
if !flags.DryRun && summary.Removed > 0 {
    gitops.AutoCommit(configDir, "rm", args)
}
```

**`internal/commands/track.go` — `runTrack()`**

After the lock file is written (the `if tracked > 0` block), add:

```go
if tracked > 0 {
    if err := lockSvc.Write(lockFile); err != nil {
        return fmt.Errorf("failed to write lock file: %w", err)
    }
    gitops.AutoCommit(configDir, "track", args)
}
```

**`internal/commands/untrack.go` — `runUntrack()`**

Same pattern as track, after lock file write:

```go
if untracked > 0 {
    if err := lockSvc.Write(lockFile); err != nil {
        return fmt.Errorf("failed to write lock file: %w", err)
    }
    gitops.AutoCommit(configDir, "untrack", args)
}
```

**`internal/commands/config_edit.go` — `editConfigVisudoStyle()`**

After `saveNonDefaultValues()` succeeds (the "Success - save" branch), add:

```go
gitops.AutoCommit(configDir, "config edit", nil)
```

`AutoCommit` reads the just-saved config from disk, so the current state of `auto_commit` governs the commit decision.

### 5. Add `plonk push` command

**`internal/commands/push.go`**

```go
package commands

import (
    "fmt"

    "github.com/richhaase/plonk/internal/config"
    "github.com/richhaase/plonk/internal/gitops"
    "github.com/richhaase/plonk/internal/output"
    "github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
    Use:   "push",
    Short: "Commit pending changes and push to remote",
    Long: `Commit any uncommitted changes in your plonk directory and push to the remote.

This stages all changes, creates a commit, and pushes to the default remote.
If there are no changes, only a push is performed.

Examples:
  plonk push    # Commit and push`,
    RunE:         runPush,
    SilenceUsage: true,
}

func init() {
    rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
    configDir := config.GetDefaultConfigDirectory()
    client := gitops.New(configDir)

    if !client.IsRepo() {
        return fmt.Errorf("%s is not a git repository", configDir)
    }

    if !client.HasRemote() {
        return fmt.Errorf("no remote configured for %s", configDir)
    }

    // Commit any pending changes first
    dirty, err := client.IsDirty()
    if err != nil {
        return err
    }
    if dirty {
        msg := gitops.CommitMessage("push", nil)
        if err := client.Commit(msg); err != nil {
            return fmt.Errorf("failed to commit: %w", err)
        }
        output.Println("Committed pending changes")
    }

    // Push
    output.Println("Pushing to remote...")
    if err := client.Push(); err != nil {
        return err
    }
    output.Println("Push complete")
    return nil
}
```

### 6. Add `plonk pull` command

**`internal/commands/pull.go`**

```go
package commands

import (
    "context"
    "fmt"

    "github.com/richhaase/plonk/internal/config"
    "github.com/richhaase/plonk/internal/gitops"
    "github.com/richhaase/plonk/internal/orchestrator"
    "github.com/richhaase/plonk/internal/output"
    "github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
    Use:   "pull",
    Short: "Pull remote changes into plonk directory",
    Long: `Pull remote changes into your plonk directory.

If there are uncommitted local changes, they are committed first to avoid
conflicts. Use --apply to automatically run 'plonk apply' after pulling.

Examples:
  plonk pull            # Pull remote changes
  plonk pull --apply    # Pull and apply`,
    RunE:         runPull,
    SilenceUsage: true,
}

func init() {
    pullCmd.Flags().BoolP("apply", "a", false, "Run plonk apply after pulling")
    rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
    applyAfter, _ := cmd.Flags().GetBool("apply")
    configDir := config.GetDefaultConfigDirectory()
    client := gitops.New(configDir)

    if !client.IsRepo() {
        return fmt.Errorf("%s is not a git repository", configDir)
    }

    // Auto-commit local changes before pulling
    dirty, err := client.IsDirty()
    if err != nil {
        return err
    }
    if dirty {
        msg := gitops.CommitMessage("pre-pull snapshot", nil)
        if err := client.Commit(msg); err != nil {
            return fmt.Errorf("failed to commit local changes: %w", err)
        }
        output.Println("Committed local changes before pull")
    }

    // Pull
    output.Println("Pulling from remote...")
    if err := client.Pull(); err != nil {
        return err
    }
    output.Println("Pull complete")

    // Optionally apply
    if applyAfter {
        output.Println("Applying configuration...")
        homeDir, err := config.GetHomeDir()
        if err != nil {
            return fmt.Errorf("cannot determine home directory: %w", err)
        }
        cfg := config.LoadWithDefaults(configDir)
        ctx := context.Background()

        orch := orchestrator.New(
            orchestrator.WithConfig(cfg),
            orchestrator.WithConfigDir(configDir),
            orchestrator.WithHomeDir(homeDir),
            orchestrator.WithDryRun(false),
        )

        result, err := orch.Apply(ctx)
        output.RenderOutput(result)
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 7. Tests

#### Unit tests

**`internal/gitops/gitops_test.go`**

Test all `Client` methods using temp directories with `git init`. Key tests:

- `TestIsRepo` / `TestIsRepoFalse`
- `TestIsDirty` / `TestIsDirtyClean`
- `TestCommit` — verify commit appears in `git log`
- `TestCommitNoop` — clean repo, no error
- `TestHasRemote` / `TestHasRemoteFalse`
- `TestCommitMessage` — table-driven: various command/arg combos including truncation at 5+ args

**`internal/config/config_test.go`**

- `TestAutoCommitDefault` — zero-value config → `AutoCommitEnabled()` returns true
- `TestAutoCommitExplicitFalse` — returns false
- `TestGitConfigYAMLRoundTrip` — marshal/unmarshal `git: { auto_commit: false }`

#### BATS behavioral tests

**`tests/bats/behavioral/18-git-ops.bats`**

```bash
#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
    setup_test_env

    # Initialize PLONK_DIR as a git repo
    git -C "$PLONK_DIR" init
    git -C "$PLONK_DIR" config user.email "test@test.com"
    git -C "$PLONK_DIR" config user.name "Test"
    git -C "$PLONK_DIR" commit --allow-empty -m "initial"
}

@test "add auto-commits to git" {
    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success

    # Verify git log has a commit from plonk
    run git -C "$PLONK_DIR" log --oneline -1
    assert_output --partial "plonk: add"
}

@test "track auto-commits to git" {
    # ... similar pattern with a safe package
}

@test "push command works" {
    # Set up a bare remote
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"

    # Create a change and commit
    echo "test" > "$PLONK_DIR/testfile"
    git -C "$PLONK_DIR" add -A
    git -C "$PLONK_DIR" commit -m "test change"

    run plonk push
    assert_success
    assert_output --partial "Push complete"
}

@test "pull command works" {
    # ... similar with bare remote setup
}

@test "pull --apply runs apply after pull" {
    # ... verify apply output appears
}

@test "auto-commit disabled via config" {
    echo 'git:' > "$PLONK_DIR/plonk.yaml"
    echo '  auto_commit: false' >> "$PLONK_DIR/plonk.yaml"
    git -C "$PLONK_DIR" add -A && git -C "$PLONK_DIR" commit -m "add config"

    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success

    # Verify repo is dirty (not auto-committed)
    run git -C "$PLONK_DIR" status --porcelain
    refute_output ""
}

@test "non-git-repo warns" {
    # Remove .git from PLONK_DIR
    rm -rf "$PLONK_DIR/.git"

    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success
    assert_output --partial "not a git repository"
    assert_output --partial "auto_commit"
}
```

### 8. File summary

| Action | Path |
|--------|------|
| Create | `internal/gitops/gitops.go` |
| Create | `internal/gitops/autocommit.go` |
| Create | `internal/gitops/gitops_test.go` |
| Create | `internal/commands/push.go` |
| Create | `internal/commands/pull.go` |
| Create | `tests/bats/behavioral/18-git-ops.bats` |
| Modify | `internal/config/config.go` — add `GitConfig` struct, `Git` field, `AutoCommitEnabled()` method |
| Modify | `internal/commands/add.go` — call `gitops.AutoCommit` after successful add |
| Modify | `internal/commands/rm.go` — call `gitops.AutoCommit` after successful rm |
| Modify | `internal/commands/track.go` — call `gitops.AutoCommit` after successful track |
| Modify | `internal/commands/untrack.go` — call `gitops.AutoCommit` after successful untrack |
| Modify | `internal/commands/config_edit.go` — call `gitops.AutoCommit` after successful save |
| Modify | `internal/config/config_test.go` — add auto-commit config tests |

### Validation

1. **Unit tests**: `cd .worktrees/git-ops && go test ./...` — all pass
2. **BATS tests**: `cd .worktrees/git-ops && bats tests/bats/behavioral/18-git-ops.bats`
3. **Manual smoke test**:
   - `export PLONK_DIR=$(mktemp -d) && cd $PLONK_DIR && git init`
   - `plonk add ~/.zshrc` → verify `git log` shows auto-commit
   - `plonk push` → errors with "no remote" (expected)
   - Set `git: { auto_commit: false }` in plonk.yaml → `plonk add ~/.vimrc` → verify no commit
   - Remove `.git` → `plonk add ~/.bashrc` → verify warning message
4. **Lint**: `golangci-lint run ./...`

## Assumptions

- **`git` is on PATH**: We use `os/exec` to call git. The `plonk doctor` command already checks for git availability. We don't add a new dependency.
- **Merge on pull**: We use `git pull` (merge, not rebase). Dotfile changes across machines are independent learnings, not linear incremental progress — merge preserves that intent.
- **No branch tracking logic**: `plonk push`/`pull` rely on git's default remote tracking branch. If the user hasn't set upstream, git will error and we surface that message.
- **Auto-commit is best-effort**: Failures in auto-commit are warnings, not errors. The mutation (add/rm/track/untrack) itself already succeeded — we don't roll it back if git fails.
- **`AutoCommit` reads config from disk**: Callers just pass `configDir`. Whatever `auto_commit` says on disk at call time is what happens.
