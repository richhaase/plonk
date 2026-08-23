// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build unix

package lock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// mutationLockName is the advisory lock file used to serialize lock-file
// mutations (read-modify-write cycles) across concurrent plonk processes.
// It is distinct from plonk.lock (the managed state file itself).
const mutationLockName = ".plonk.mutlock"

// lockRetryInterval is how often a contended acquisition re-attempts the
// non-blocking flock while waiting, so cancellation is honored promptly.
const lockRetryInterval = 50 * time.Millisecond

// processMutexes provides in-process serialization per lock path, layered
// under the inter-process flock. flock alone is sufficient for correctness
// across processes, but the in-process mutex guarantees memory visibility
// between goroutines (and keeps the race detector happy).
var processMutexes sync.Map // string -> *sync.Mutex

// WithMutationLock serializes lock-file mutations for the repository in
// configDir. It first acquires an in-process mutex for the lock path, then an
// advisory exclusive flock, so concurrent plonk processes (and goroutines)
// cannot interleave read-modify-write cycles and lose updates. The callback
// fn runs while the lock is held.
//
// Acquisition is cancellation-aware: if ctx is canceled while another process
// holds the lock, this returns ctx.Err() instead of blocking indefinitely.
func WithMutationLock(ctx context.Context, configDir string, fn func() error) error {
	// Refuse to start mutating state once cancellation has been requested
	if err := ctx.Err(); err != nil {
		return err
	}

	//nolint:gosec // G301: configDir comes from plonk's own config directory resolution
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	lockPath := filepath.Join(configDir, mutationLockName)

	muAny, _ := processMutexes.LoadOrStore(lockPath, &sync.Mutex{})
	mu := muAny.(*sync.Mutex)

	// In-process mutex acquisition must also honor cancellation
	if err := acquireProcessMutex(ctx, mu); err != nil {
		return err
	}
	defer mu.Unlock()

	// Keep the internal coordination file out of auto-commits (git add -A)
	// without touching the user's tracked .gitignore
	excludeFromGit(configDir, mutationLockName)

	//nolint:gosec // G304: lockPath is derived from plonk's own config directory, not user input
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open mutation lock %s: %w", lockPath, err)
	}
	defer f.Close()

	if err := acquireFlock(ctx, f); err != nil {
		return fmt.Errorf("failed to acquire mutation lock %s: %w", lockPath, err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:errcheck // best-effort unlock; close() also releases

	// Final cancellation check before entering the critical section, so a
	// signal that arrived during (uncontended) acquisition never mutates state
	if err := ctx.Err(); err != nil {
		return err
	}

	return fn()
}

// acquireProcessMutex acquires mu, waiting at most until ctx is canceled.
func acquireProcessMutex(ctx context.Context, mu *sync.Mutex) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	locked := make(chan struct{})
	go func() {
		mu.Lock()
		close(locked)
	}()
	select {
	case <-locked:
		return nil
	case <-ctx.Done():
		// The goroutine still holds/will hold the mutex; unlock it from here
		// to release it for other waiters. This is safe because the channel
		// close happens-after Lock() returns, and if we raced, either we or
		// the goroutine owns it — the goroutine never uses it after closing.
		go func() {
			<-locked
			mu.Unlock()
		}()
		return ctx.Err()
	}
}

// acquireFlock acquires an exclusive advisory lock on f, honoring ctx
// cancellation while contended by retrying a non-blocking acquire.
func acquireFlock(ctx context.Context, f *os.File) error {
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockRetryInterval):
			// retry
		}
	}
}

// excludeFromGit adds entry to the repository's .git/info/exclude so the
// internal coordination file is never staged by `git add -A`. This is a
// no-op when configDir is not a git repository. Unlike .gitignore,
// info/exclude is local-only and never touches the user's tracked files.
func excludeFromGit(configDir, entry string) {
	excludePath := resolveGitExcludePath(configDir)
	if excludePath == "" {
		return // not a git repository; best-effort
	}

	data, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return // no permission / not a repo; best-effort
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return // already excluded
		}
	}

	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // no permission; best-effort
	}
	defer f.Close()

	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, entry)
}

// resolveGitExcludePath returns the absolute path of the repository's local
// exclude file, or "" if configDir is not a git repository. It uses
// `git rev-parse --git-path info/exclude`, which resolves correctly for
// worktrees and submodules (where .git is a file, not a directory) and for
// linked worktrees whose exclude lives under .git/worktrees/<name>/info/.
// Falls back to the plain <configDir>/.git/info/exclude path when git is
// unavailable or fails.
func resolveGitExcludePath(configDir string) string {
	//nolint:gosec // G204: git args are constant strings; configDir is plonk's own config directory
	out, err := exec.Command("git", "-C", configDir, "rev-parse", "--git-path", "info/exclude").Output()
	if err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			if filepath.IsAbs(p) {
				return p
			}
			return filepath.Join(configDir, p)
		}
	}

	// Fallback: plain repository layout
	fallback := filepath.Join(configDir, ".git", "info", "exclude")
	//nolint:gosec // G304: path derived from plonk's own config directory, not user input
	if info, err := os.Stat(filepath.Join(configDir, ".git")); err == nil && !info.IsDir() {
		// .git is a file (worktree/submodule gitlink) and git rev-parse
		// failed; we cannot safely resolve the real exclude path.
		return ""
	}
	return fallback
}
