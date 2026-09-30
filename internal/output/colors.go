// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"os"

	"github.com/fatih/color"
)

// InitColors should be called early in command execution to set up color support
func InitColors() {
	// Output color follows stdout independently of progress on stderr.
	color.NoColor = !writer.IsTerminal() || colorsDisabled()
}

// colorize applies color to text only if colors are enabled
func colorize(text string, attrs ...color.Attribute) string {
	// color.NoColor is checked internally by the color package
	c := color.New(attrs...)
	if color.NoColor {
		c.DisableColor()
	} else {
		c.EnableColor()
	}
	return c.Sprint(text)
}

// Status word with coloring
func Success() string { return colorize("success", color.FgGreen) }

// Color functions for specific use cases
func ColorError(text string) string { return colorize(text, color.FgRed) }
func ColorInfo(text string) string  { return colorize(text, color.FgBlue) }
func ColorAdded(text string) string { return colorize(text, color.FgGreen) }
func ColorRemoved(text string) string {
	return colorize(text, color.FgRed)
}

func colorsDisabled() bool { return os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" }

// ColorState uses color only to reinforce a state that is also written in text.
func ColorState(state string) string { return colorState(state, state, nil) }

func colorState(text, state string, target Writer) string {
	var attr color.Attribute
	switch state {
	case "error", "failed", "fail", "unhealthy":
		attr = color.FgRed
	case "missing", "drifted", "warn", "warning":
		attr = color.FgYellow
	case "info", "would add", "would update", "would remove", "would install", "would clone", "would apply", "would create", "would check":
		attr = color.FgBlue
	case "skipped", "untracked", "unknown":
		return text
	default:
		attr = color.FgGreen
	}
	if target == nil {
		return colorize(text, attr)
	}
	c := color.New(attr)
	if target.IsTerminal() && !colorsDisabled() {
		c.EnableColor()
	} else {
		c.DisableColor()
	}
	return c.Sprint(text)
}
