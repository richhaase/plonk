// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !unix

package lock

import "context"

// WithMutationLock is a no-op on platforms without flock support.
// Mutations are serialized within a process only.
func WithMutationLock(ctx context.Context, configDir string, fn func() error) error {
	return fn()
}
