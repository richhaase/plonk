// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"github.com/fatih/color"
)

// InitColors should be called early in command execution to set up color support
func InitColors() {
	// The fatih/color package automatically handles:
	// 1. NO_COLOR environment variable
	// 2. Terminal capability detection
	// 3. Windows console support

	// Check both stdout and stderr for terminal status
	stdoutIsTerminal := writer.IsTerminal()
	stderrIsTerminal := progressWriter.IsTerminal()

	// Disable colors if neither stdout nor stderr is a terminal
	if !stdoutIsTerminal && !stderrIsTerminal {
		color.NoColor = true
	}
}

// colorize applies color to text only if colors are enabled
func colorize(text string, attrs ...color.Attribute) string {
	// color.NoColor is checked internally by the color package
	return color.New(attrs...).Sprint(text)
}

// Common status words with appropriate coloring - sorted by color then alphabetically

// Green (Success) status words
func Available() string { return colorize("available", color.FgGreen) }
func Deployed() string  { return colorize("deployed", color.FgGreen) }
func Managed() string   { return colorize("managed", color.FgGreen) }
func Success() string   { return colorize("success", color.FgGreen) }
func Valid() string     { return colorize("Valid", color.FgGreen) }

// Red (Error) status words
func Invalid() string      { return colorize("Invalid", color.FgRed) }
func Missing() string      { return colorize("missing", color.FgRed) }
func NotAvailable() string { return colorize("not available", color.FgRed) }

// Yellow (Warning) status words
func Drifted() string   { return colorize("drifted", color.FgYellow) }
func Unmanaged() string { return colorize("unmanaged", color.FgYellow) }

// Dim (Skip) status words

// Plain text status words (no color)

// Additional color functions for specific use cases
func ColorError(text string) string { return colorize(text, color.FgRed) }
func ColorInfo(text string) string  { return colorize(text, color.FgBlue) }
func ColorAdded(text string) string { return colorize(text, color.FgGreen) }
func ColorRemoved(text string) string {
	return colorize(text, color.FgRed)
}
