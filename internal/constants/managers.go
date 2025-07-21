// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package constants

// SupportedManagers contains all package managers supported by plonk
var SupportedManagers = []string{
	"apt",
	"cargo",
	"gem",
	"go",
	"homebrew",
	"npm",
	"pip",
}

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "homebrew"
