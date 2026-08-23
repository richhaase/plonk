// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build unix

package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

// mutationLockName is the advisory lock file used to serialize lock-file
// mutations (read-modify-write cycles) across concurrent plonk processes.
// It is distinct from plonk.lock (the managed state file itself).
const mutationLockName = ".plonk.mutlock"

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
func WithMutationLock(configDir string, fn func() error) error {
	//nolint:gosec // G301: configDir comes from plonk's own config directory resolution
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	lockPath := filepath.Join(configDir, mutationLockName)

	muAny, _ := processMutexes.LoadOrStore(lockPath, &sync.Mutex{})
	mu := muAny.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	//nolint:gosec // G304: lockPath is derived from plonk's own config directory, not user input
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open mutation lock %s: %w", lockPath, err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire mutation lock %s: %w", lockPath, err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:errcheck // best-effort unlock; close() also releases

	return fn()
}
