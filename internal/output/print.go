// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import "fmt"

// Printf writes formatted output to stderr for progress/status messages
// This keeps stdout clean for structured output (JSON/YAML)
func Printf(format string, args ...interface{}) {
	progressWriter.Printf(format, args...)
}

// Println writes output with newline to stderr for progress/status messages
func Println(args ...interface{}) {
	progressWriter.Printf("%s\n", fmt.Sprint(args...))
}
