// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

// SupportedManagers contains all package packages supported by plonk
var SupportedManagers = []string{
	"cargo",
	"gem",
	"go",
	"homebrew",
	"npm",
	"pip",
}

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "homebrew"
