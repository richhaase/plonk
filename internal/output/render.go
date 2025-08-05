// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the available output formats
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
)

// RenderOutput renders data in the specified format
func RenderOutput(data OutputData, format OutputFormat) error {
	switch format {
	case OutputTable:
		fmt.Print(data.TableOutput())
		return nil
	case OutputJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data.StructuredData())
	case OutputYAML:
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(data.StructuredData())
	default:
		return fmt.Errorf("unsupported output format: %s (use: table, json, or yaml)", format)
	}
}

// ParseOutputFormat converts string to OutputFormat
func ParseOutputFormat(format string) (OutputFormat, error) {
	switch format {
	case "table":
		return OutputTable, nil
	case "json":
		return OutputJSON, nil
	case "yaml":
		return OutputYAML, nil
	default:
		return OutputTable, fmt.Errorf("unsupported format '%s' (use: table, json, or yaml)", format)
	}
}
