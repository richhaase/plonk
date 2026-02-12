// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import "fmt"

// RenderOutput renders data in table format.
// No-op if data is nil.
func RenderOutput(data OutputData) {
	if data == nil {
		return
	}
	fmt.Print(data.TableOutput())
}
