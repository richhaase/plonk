// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package main

import (
	"os"

	"plonk/internal/commands"
)

// Version information, injected at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	exitCode := commands.ExecuteWithExitCode(version, commit, date)
	os.Exit(exitCode)
}
