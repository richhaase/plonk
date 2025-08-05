// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

// progressWriter is used for all progress/status messages
// Progress messages go to stderr to keep stdout clean for structured output
var progressWriter Writer = &StderrWriter{}

// StderrWriter implements Writer interface for stderr output
type StderrWriter struct{}

// Printf writes formatted output to stderr
func (s *StderrWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// IsTerminal returns true if stderr is connected to a terminal
func (s *StderrWriter) IsTerminal() bool {
	return isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
}
