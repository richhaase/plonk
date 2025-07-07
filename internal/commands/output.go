package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Output prints normal output unless quiet mode is enabled
func Output(cmd *cobra.Command, format string, args ...interface{}) {
	if !IsQuiet(cmd) {
		fmt.Printf(format, args...)
	}
}

// Outputln prints normal output with newline unless quiet mode is enabled
func Outputln(cmd *cobra.Command, format string, args ...interface{}) {
	if !IsQuiet(cmd) {
		fmt.Printf(format+"\n", args...)
	}
}

// Verbose prints output only in verbose mode
func Verbose(cmd *cobra.Command, format string, args ...interface{}) {
	if IsVerbose(cmd) && !IsQuiet(cmd) {
		fmt.Printf("[verbose] "+format, args...)
	}
}

// Verboseln prints verbose output with newline
func Verboseln(cmd *cobra.Command, format string, args ...interface{}) {
	if IsVerbose(cmd) && !IsQuiet(cmd) {
		fmt.Printf("[verbose] "+format+"\n", args...)
	}
}

// Warning always prints warnings (even in quiet mode)
func Warning(format string, args ...interface{}) {
	fmt.Printf("Warning: "+format, args...)
}

// Warningln always prints warnings with newline
func Warningln(format string, args ...interface{}) {
	fmt.Printf("Warning: "+format+"\n", args...)
}

// Error always prints errors (even in quiet mode)
func Error(format string, args ...interface{}) {
	fmt.Printf("Error: "+format, args...)
}

// Errorln always prints errors with newline
func Errorln(format string, args ...interface{}) {
	fmt.Printf("Error: "+format+"\n", args...)
}
