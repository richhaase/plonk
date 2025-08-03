// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotfileResource_ID(t *testing.T) {
	resource := &DotfileResource{}
	assert.Equal(t, "dotfiles", resource.ID())
}
