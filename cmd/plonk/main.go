// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package main

import (
	"os"

	"plonk/internal/commands"
)

func main() {
	exitCode := commands.ExecuteWithExitCode()
	os.Exit(exitCode)
}
