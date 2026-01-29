// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "fmt"

// GetManager returns a Manager by name
func GetManager(name string) (Manager, error) {
	switch name {
	case "brew":
		return NewBrewSimple(), nil
	case "cargo":
		return NewCargoSimple(), nil
	case "go":
		return NewGoSimple(), nil
	case "pnpm":
		return NewPNPMSimple(), nil
	case "uv":
		return NewUVSimple(), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s (supported: %v)", name, SupportedManagers)
	}
}
