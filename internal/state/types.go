// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package state provides unified state management capabilities for plonk.
// This package implements the core state reconciliation patterns that are used
// across both package management and dotfile management domains.
//
// Types in this package are now aliases to the unified interfaces package.
package state

import (
	"github.com/richhaase/plonk/internal/interfaces"
	"github.com/richhaase/plonk/internal/types"
)

// ItemState is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use interfaces.ItemState directly.
type ItemState = interfaces.ItemState

// Backward compatibility constants
const (
	StateManaged   = interfaces.StateManaged
	StateMissing   = interfaces.StateMissing
	StateUntracked = interfaces.StateUntracked
)

// Item is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use interfaces.Item directly.
type Item = interfaces.Item

// Result is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use types.Result directly.
type Result = types.Result

// Summary is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use types.Summary directly.
type Summary = types.Summary

// ConfigItem is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use interfaces.ConfigItem directly.
type ConfigItem = interfaces.ConfigItem

// ActualItem is an alias for the unified type to maintain backward compatibility.
// Deprecated: Use interfaces.ActualItem directly.
type ActualItem = interfaces.ActualItem
