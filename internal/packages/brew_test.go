// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBrewCachedInventoryHonorsCancellation(t *testing.T) {
	b := NewBrewSimple()
	b.installed = map[string]bool{"demo": true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	installed, err := b.IsInstalled(ctx, "demo")
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, installed)
}
