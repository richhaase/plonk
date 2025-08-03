// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

// Writer abstracts output operations for testing
type Writer interface {
	Printf(format string, args ...interface{})
	IsTerminal() bool
}

// StdoutWriter implements Writer for real stdout
type StdoutWriter struct{}

func (s *StdoutWriter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (s *StdoutWriter) IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// Package-level writer instance
var writer Writer = &StdoutWriter{}

// SetWriter allows tests to override the writer
func SetWriter(w Writer) {
	writer = w
}
