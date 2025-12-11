// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

// StageUpdate prints a stage update for multi-stage operations
func StageUpdate(stage string) {
	progressWriter.Printf("%s\n", stage)
}
