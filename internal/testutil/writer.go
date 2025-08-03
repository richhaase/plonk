// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package testutil provides common test utilities for the plonk project
package testutil

import (
	"bytes"
	"fmt"
)

// BufferWriter captures output for testing
type BufferWriter struct {
	Buffer   bytes.Buffer
	terminal bool
}

// NewBufferWriter creates a test writer with configurable terminal detection
func NewBufferWriter(terminal bool) *BufferWriter {
	return &BufferWriter{terminal: terminal}
}

// Printf writes formatted output to the buffer
func (b *BufferWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&b.Buffer, format, args...)
}

// IsTerminal returns the configured terminal status
func (b *BufferWriter) IsTerminal() bool {
	return b.terminal
}

// String returns the captured output as a string
func (b *BufferWriter) String() string {
	return b.Buffer.String()
}

// Reset clears the buffer
func (b *BufferWriter) Reset() {
	b.Buffer.Reset()
}
