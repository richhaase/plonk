// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build unix

package lock

import (
	"os"
	"os/exec"
	"path/filepath"
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
				err := WithMutationLock(dir, func() error {
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
	if err := WithMutationLock(dir, func() error { return nil }); err != nil {
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
		err := WithMutationLock(os.Getenv("PLONK_MUTLOCK_DIR"), func() error {
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
	if err := WithMutationLock(dir, func() error { return nil }); err != nil {
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
