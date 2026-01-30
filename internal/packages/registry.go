// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sync"
)

var (
	managerCache = make(map[string]Manager)
	managerMu    sync.Mutex
)

// GetManager returns a Manager by name, caching instances for reuse
func GetManager(name string) (Manager, error) {
	managerMu.Lock()
	defer managerMu.Unlock()

	// Return cached manager if available
	if mgr, ok := managerCache[name]; ok {
		return mgr, nil
	}

	// Create new manager
	var mgr Manager
	switch name {
	case "brew":
		mgr = NewBrewSimple()
	case "cargo":
		mgr = NewCargoSimple()
	case "go":
		mgr = NewGoSimple()
	case "pnpm":
		mgr = NewPNPMSimple()
	case "uv":
		mgr = NewUVSimple()
	default:
		return nil, fmt.Errorf("unsupported package manager: %s (supported: %v)", name, SupportedManagers)
	}

	// Cache and return
	managerCache[name] = mgr
	return mgr, nil
}
