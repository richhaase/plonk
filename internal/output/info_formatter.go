// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

// InfoOutput represents the output structure for info command
type InfoOutput struct {
	Package     string      `json:"package" yaml:"package"`
	Status      string      `json:"status" yaml:"status"`
	Message     string      `json:"message" yaml:"message"`
	PackageInfo interface{} `json:"package_info,omitempty" yaml:"package_info,omitempty"`
}

// InfoFormatter formats info output
type InfoFormatter struct {
	Data InfoOutput
}

// NewInfoFormatter creates a new formatter
func NewInfoFormatter(data InfoOutput) InfoFormatter {
	return InfoFormatter{Data: data}
}

// TableOutput generates human-friendly table output for info command
func (f InfoFormatter) TableOutput() string {
	i := f.Data
	builder := NewStandardTableBuilder("")

	// Add package name
	builder.AddRow("Package:", i.Package)

	// Add status based on status - simplified version that doesn't access complex package info
	switch i.Status {
	case "managed":
		builder.AddRow("Status:", "Managed by plonk")
	case "installed":
		builder.AddRow("Status:", "Installed (not managed)")
	case "available":
		builder.AddRow("Status:", "Available")
	case "not-found":
		builder.AddRow("Status:", "Not found")
	case "no-managers":
		builder.AddRow("Status:", "No package managers available")
	case "manager-unavailable":
		builder.AddRow("Status:", "Manager unavailable")
	default:
		builder.AddRow("Status:", "Unknown")
	}

	// Add message if available
	if i.Message != "" {
		builder.AddRow("Message:", i.Message)
	}

	return builder.Build()
}

// StructuredData returns the structured data for serialization
func (f InfoFormatter) StructuredData() any {
	return f.Data
}
