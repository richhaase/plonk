// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"sync"
	"testing"
	"time"
)

// buildLargeLock constructs a lock with n package resources for testing.
func buildLargeLock(n int) *Lock {
	l := &Lock{Version: LockFileVersion, Resources: make([]ResourceEntry, 0, n)}
	for i := 0; i < n; i++ {
		name := "pkg" + itoa(i)
		entry := ResourceEntry{
			Type:        "package",
			ID:          "brew:" + name,
			InstalledAt: time.Now().Format(time.RFC3339),
			Metadata: map[string]interface{}{
				"manager": "brew",
				"name":    name,
				"version": "1.0." + itoa(i%10),
			},
		}
		l.Resources = append(l.Resources, entry)
	}
	return l
}

// minimal int->string to avoid fmt overhead in tight loops
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	bp := len(b)
	n := i
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		bp--
		b[bp] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		bp--
		b[bp] = '-'
	}
	return string(b[bp:])
}

func TestYAMLLockService_LargeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	svc := NewYAMLLockService(dir)

	const N = 1000
	in := buildLargeLock(N)
	if err := svc.Write(in); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	out, err := svc.Read()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if out.Version != LockFileVersion {
		t.Fatalf("unexpected version: %d", out.Version)
	}
	if got := len(out.Resources); got != N {
		t.Fatalf("resource count mismatch: got %d want %d", got, N)
	}

	// spot-check first/last entries metadata
	if out.Resources[0].Type != "package" || out.Resources[0].Metadata["manager"] != "brew" {
		t.Fatalf("unexpected first entry: %+v", out.Resources[0])
	}
	if out.Resources[N-1].Metadata["name"] != "pkg"+itoa(N-1) {
		t.Fatalf("unexpected last entry name: %+v", out.Resources[N-1].Metadata["name"])
	}
}

func TestYAMLLockService_ConcurrentWriteReadAtomic(t *testing.T) {
	dir := t.TempDir()
	svc := NewYAMLLockService(dir)

	const rounds = 20
	var wg sync.WaitGroup

	// writer repeatedly writes growing locks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= rounds; i++ {
			l := buildLargeLock(i * 50)
			if err := svc.Write(l); err != nil {
				t.Errorf("write round %d failed: %v", i, err)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// reader repeatedly reads and validates parse/version
	wg.Add(1)
	go func() {
		defer wg.Done()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			l, err := svc.Read()
			if err != nil {
				t.Errorf("read failed: %v", err)
				return
			}
			if l.Version != LockFileVersion {
				t.Errorf("unexpected version: %d", l.Version)
				return
			}
		}
	}()

	wg.Wait()

	// final read should reflect last write size
	l, err := svc.Read()
	if err != nil {
		t.Fatalf("final read failed: %v", err)
	}
	want := rounds * 50
	if got := len(l.Resources); got != want {
		t.Fatalf("final size mismatch: got %d want %d", got, want)
	}
}
