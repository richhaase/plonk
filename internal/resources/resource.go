// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import "context"

// Resource represents any manageable resource (packages, dotfiles, future services)
type Resource interface {
	ID() string
	Desired() []Item // Set by orchestrator from config
	Actual(ctx context.Context) []Item
	Apply(ctx context.Context, item Item) error
}
