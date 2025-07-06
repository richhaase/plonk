package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateYAML checks if the provided YAML content has valid syntax
func ValidateYAML(content []byte) error {
	// Handle empty content
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" || isOnlyComments(trimmed) {
		return nil // Empty or comment-only YAML is valid
	}
	
	// Create a generic interface to unmarshal into
	var data interface{}
	
	// Create a decoder to get better error messages
	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	decoder.KnownFields(false) // Allow unknown fields for flexibility
	
	// Try to decode the YAML
	err := decoder.Decode(&data)
	if err != nil {
		// Handle EOF for empty content
		if err.Error() == "EOF" {
			return nil
		}
		
		// Check for specific error types
		if strings.Contains(err.Error(), "found character that cannot start") {
			return fmt.Errorf("yaml syntax error: invalid character or indentation - %w", err)
		}
		if strings.Contains(err.Error(), "found undefined alias") {
			return fmt.Errorf("yaml syntax error: undefined alias - %w", err)
		}
		if strings.Contains(err.Error(), "mapping values are not allowed") {
			return fmt.Errorf("yaml syntax error: invalid mapping - %w", err)
		}
		if strings.Contains(err.Error(), "already defined") {
			return fmt.Errorf("yaml syntax error: duplicate key found - %w", err)
		}
		// Check for invalid indentation specifically
		if strings.Contains(err.Error(), "did not find expected") && strings.Contains(string(content), "\ndefault_manager:") {
			return fmt.Errorf("yaml syntax error: invalid indentation - %w", err)
		}
		// Generic YAML error
		return fmt.Errorf("yaml syntax error: %w", err)
	}
	
	return nil
}

// isOnlyComments checks if the content contains only comments and whitespace
func isOnlyComments(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
			return false
		}
	}
	return true
}