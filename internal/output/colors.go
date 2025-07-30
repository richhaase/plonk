// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// InitColors should be called early in command execution to set up color support
func InitColors() {
	// The fatih/color package automatically handles:
	// 1. NO_COLOR environment variable
	// 2. Terminal capability detection
	// 3. Windows console support

	// We only need to check if output is going to a terminal
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	// Disable colors if output is not to a terminal (piped/redirected)
	if !isTerminal {
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
func Done() string      { return colorize("done", color.FgGreen) }
func Installed() string { return colorize("installed", color.FgGreen) }
func Managed() string   { return colorize("managed", color.FgGreen) }
func Pass() string      { return colorize("PASS", color.FgGreen) }
func Removed() string   { return colorize("removed", color.FgGreen) }
func Success() string   { return colorize("success", color.FgGreen) }
func Valid() string     { return colorize("Valid", color.FgGreen) }

// Red (Error) status words
func Error() string        { return colorize("error", color.FgRed) }
func Fail() string         { return colorize("FAIL", color.FgRed) }
func Failed() string       { return colorize("failed", color.FgRed) }
func Invalid() string      { return colorize("Invalid", color.FgRed) }
func Missing() string      { return colorize("missing", color.FgRed) }
func NotAvailable() string { return colorize("not available", color.FgRed) }

// Yellow (Warning) status words
func Unmanaged() string { return colorize("unmanaged", color.FgYellow) }
func Warn() string      { return colorize("WARN", color.FgYellow) }
func Warning() string   { return colorize("warning", color.FgYellow) }

// Dim (Skip) status words
func Skipped() string { return colorize("skipped", color.Faint) }

// Plain text status words (no color)
func Info() string { return "INFO" }

// Additional color functions for specific use cases
func ColorSuccess(text string) string { return colorize(text, color.FgGreen) }
func ColorError(text string) string   { return colorize(text, color.FgRed) }
func ColorWarning(text string) string { return colorize(text, color.FgYellow) }
func ColorDim(text string) string     { return colorize(text, color.Faint) }
