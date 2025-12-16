// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import "fmt"

// RenderOutput renders data in table format
func RenderOutput(data OutputData) {
	fmt.Print(data.TableOutput())
}
