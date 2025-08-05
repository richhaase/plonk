// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

// OutputData defines the interface for command output data
type OutputData interface {
	TableOutput() string // Human-friendly table format
	StructuredData() any // Data structure for json/yaml/toml
}
