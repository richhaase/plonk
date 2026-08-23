// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build unix

package lock

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestWithMutationLockSerializes verifies that the advisory lock serializes
// critical sections within a process (flock locks are per open file
// description, so independent acquires contend even in-process).
func TestWithMutationLockSerializes(t *testing.T) {
	dir := t.TempDir()

	const goroutines = 16
	const iterations = 50
	var counter int

	var mu sync.Mutex // guards test error reporting only
	var testErr error

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				err := WithMutationLock(context.Background(), dir, func() error {
					// Unlocked read-modify-write; only safe under the lock
					v := counter
					v++
					counter = v
					return nil
				})
				if err != nil {
					mu.Lock()
					testErr = err
					mu.Unlock()
					return
				}
			}
		}()
	}
	wg.Wait()

	if testErr != nil {
		t.Fatalf("WithMutationLock failed: %v", testErr)
	}
	if counter != goroutines*iterations {
		t.Errorf("lost updates: counter = %d, want %d", counter, goroutines*iterations)
	}
}

// TestWithMutationLockCreatesLockFile verifies the advisory lock file is
// created in the config directory.
func TestWithMutationLockCreatesLockFile(t *testing.T) {
	dir := t.TempDir()
	if err := WithMutationLock(context.Background(), dir, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".plonk.mutlock")); err != nil {
		t.Errorf("expected mutation lock file to exist: %v", err)
	}
}

// TestWithMutationLockCrossProcess verifies mutual exclusion across processes:
// a child process acquires the lock, signals readiness through a file, and
// holds the lock for a fixed duration while the parent attempts to acquire
// it. The parent's acquisition must block until the child releases. Uses the
// standard subprocess re-invocation test pattern with an explicit readiness
// handshake instead of timing-based synchronization.
func TestWithMutationLockCrossProcess(t *testing.T) {
	const childHold = 500 * time.Millisecond

	if os.Getenv("PLONK_MUTLOCK_CHILD") == "1" {
		readyPath := os.Getenv("PLONK_MUTLOCK_READY")
		err := WithMutationLock(context.Background(), os.Getenv("PLONK_MUTLOCK_DIR"), func() error {
			// Signal that the lock is held, then hold it for the test duration
			if err := os.WriteFile(readyPath, []byte("1"), 0644); err != nil {
				os.Exit(2)
			}
			time.Sleep(childHold)
			return nil
		})
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	dir := t.TempDir()
	readyPath := filepath.Join(dir, "ready")
	cmd := exec.Command(os.Args[0], "-test.run=TestWithMutationLockCrossProcess")
	cmd.Env = append(os.Environ(),
		"PLONK_MUTLOCK_CHILD=1",
		"PLONK_MUTLOCK_DIR="+dir,
		"PLONK_MUTLOCK_READY="+readyPath)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()

	// Wait for the explicit readiness signal: the child now holds the lock
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(readyPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			cmd.Process.Kill()
			t.Fatal("timed out waiting for child to acquire the lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// The ready file is written while the child holds the lock; wait out the
	// remaining hold window with margin so the parent's attempt overlaps the hold
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	if err := WithMutationLock(context.Background(), dir, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	// The parent must have blocked until the child released the lock
	if elapsed < childHold-200*time.Millisecond {
		t.Errorf("parent acquired lock after %v; expected to block ~%v (no cross-process exclusion)", elapsed, childHold)
	}
}

// TestWriteConcurrentNoTempCollisions verifies concurrent Write calls (each
// using a unique temp file) leave no stray temp files and a valid lock.
func TestWriteConcurrentNoTempCollisions(t *testing.T) {
	dir := t.TempDir()
	// NewLockV3Service takes the config directory and resolves the lock file
	// path inside it, so temp files are created in dir alongside plonk.lock
	svc := NewLockV3Service(dir)

	const writers = 16
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lock := NewLockV3()
			lock.AddPackage("brew", "pkg"+string(rune('a'+i)))
			if err := svc.Write(lock); err != nil {
				t.Errorf("Write failed: %v", err)
			}
		}(w)
	}
	wg.Wait()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "plonk.lock" {
			t.Errorf("stray file left behind: %s", e.Name())
		}
	}

	lock, err := svc.Read()
	if err != nil {
		t.Fatalf("lock file unreadable after concurrent writes: %v", err)
	}
	if lock.Version != 3 {
		t.Errorf("lock version = %d, want 3", lock.Version)
	}
}

// TestWithMutationLockCanceledWhileContended verifies that acquisition honors
// context cancellation: when another process holds the lock, a canceled
// context must return ctx.Err() promptly instead of blocking until release.
func TestWithMutationLockCanceledWhileContended(t *testing.T) {
	if os.Getenv("PLONK_MUTLOCK_CHILD") == "1" {
		readyPath := os.Getenv("PLONK_MUTLOCK_READY")
		err := WithMutationLock(context.Background(), os.Getenv("PLONK_MUTLOCK_DIR"), func() error {
			if err := os.WriteFile(readyPath, []byte("1"), 0644); err != nil {
				os.Exit(2)
			}
			// Hold the lock long enough for the parent to attempt and cancel
			time.Sleep(10 * time.Second)
			return nil
		})
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	dir := t.TempDir()
	readyPath := filepath.Join(dir, "ready")
	cmd := exec.Command(os.Args[0], "-test.run=TestWithMutationLockCanceledWhileContended")
	cmd.Env = append(os.Environ(),
		"PLONK_MUTLOCK_CHILD=1",
		"PLONK_MUTLOCK_DIR="+dir,
		"PLONK_MUTLOCK_READY="+readyPath)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()

	// Wait until the child holds the lock
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(readyPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for child to acquire the lock")
		}
		time.Sleep(10 * time.Millisecond)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WithMutationLock(ctx, dir, func() error { return nil })
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected ctx.Err() while contended with canceled context, got nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("cancellation took %v; expected prompt return", elapsed)
	}
}

// TestWithMutationLockExcludesFromGit verifies that the internal lock file is
// added to .git/info/exclude so `git add -A` (auto-commit) never stages it.
func TestWithMutationLockExcludesFromGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	if err := WithMutationLock(context.Background(), dir, func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("failed to read .git/info/exclude: %v", err)
	}
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == ".plonk.mutlock" {
			found = true
		}
	}
	if !found {
		t.Errorf(".plonk.mutlock not in .git/info/exclude:\n%s", data)
	}

	// Verify git add -A does not stage the lock file
	if out, err := exec.Command("git", "-C", dir, "add", "-A").CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}
	if strings.Contains(string(out), ".plonk.mutlock") {
		t.Errorf("mutation lock file staged by git add -A:\n%s", out)
	}
}

// TestWithMutationLockPreCanceledContext verifies that an already-canceled
// context prevents the callback from running at all — no state mutation after
// cancellation, even when the lock is uncontended.
func TestWithMutationLockPreCanceledContext(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := WithMutationLock(ctx, dir, func() error {
		called = true
		return nil
	})

	if err == nil {
		t.Fatal("expected error from pre-canceled context, got nil")
	}
	if called {
		t.Error("callback ran despite pre-canceled context")
	}
}

// TestWithMutationLockWorktreeExclude verifies that in a linked git worktree
// (where .git is a file pointing at the main repository's worktrees dir), the
// mutation lock exclusion lands in the correct exclude file and git add -A
// does not stage .plonk.mutlock.
func TestWithMutationLockWorktreeExclude(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	mainDir := t.TempDir()
	if out, err := exec.Command("git", "-C", mainDir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	commitFile(t, mainDir, "README.md")

	wtDir := t.TempDir()
	if out, err := exec.Command("git", "-C", mainDir, "worktree", "add", wtDir, "-b", "wt-test").CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\n%s", err, out)
	}
	t.Cleanup(func() {
		exec.Command("git", "-C", mainDir, "worktree", "remove", "--force", wtDir).Run()
	})

	if err := WithMutationLock(context.Background(), wtDir, func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	// The exclude entry must resolve via git rev-parse (for a linked worktree
	// this is the main repository's exclude file, which covers the worktree)
	resolved, err := exec.Command("git", "-C", wtDir, "rev-parse", "--git-path", "info/exclude").Output()
	if err != nil {
		t.Fatalf("rev-parse failed: %v", err)
	}
	excludePath := strings.TrimSpace(string(resolved))
	if !filepath.IsAbs(excludePath) {
		excludePath = filepath.Join(wtDir, excludePath)
	}
	data, err := os.ReadFile(excludePath)
	if err != nil {
		t.Fatalf("failed to read resolved exclude %s: %v", excludePath, err)
	}
	if !strings.Contains(string(data), ".plonk.mutlock") {
		t.Errorf("exclusion missing from %s:\n%s", excludePath, data)
	}

	if out, err := exec.Command("git", "-C", wtDir, "add", "-A").CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	status, err := exec.Command("git", "-C", wtDir, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}
	if strings.Contains(string(status), ".plonk.mutlock") {
		t.Errorf("mutation lock file staged by git add -A in worktree:\n%s", status)
	}
}

// commitFile writes and commits a file so git worktree add has a HEAD.
func commitFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
}
