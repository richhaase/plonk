// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateYAML checks if the provided YAML content has valid syntax.
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

// ValidatePackageName checks if a package name is valid.
func ValidatePackageName(name string) error {
	// Trim and check for empty
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	// Check for invalid characters
	for _, char := range trimmed {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			return fmt.Errorf("invalid package name - cannot contain whitespace: %q", name)
		}
		// Allow letters, numbers, hyphens, underscores, dots, @, and /
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' || char == '@' || char == '/') {
			return fmt.Errorf("invalid package name - contains invalid character %q: %q", char, name)
		}
	}

	// Check for invalid patterns
	if strings.HasPrefix(trimmed, "-") || strings.HasSuffix(trimmed, "-") {
		return fmt.Errorf("package name cannot start or end with hyphen: %q", name)
	}

	return nil
}

// ValidateFilePath checks if a file path is valid for plonk configuration.
func ValidateFilePath(path string) error {
	// Trim and check for empty
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check for absolute paths
	if filepath.IsAbs(trimmed) {
		return fmt.Errorf("file path cannot be absolute: %q", path)
	}

	// Check for invalid characters
	for _, char := range trimmed {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			return fmt.Errorf("invalid file path - cannot contain whitespace: %q", path)
		}
		// Disallow special characters that could be problematic
		if char == '!' || char == '@' || char == '#' || char == '$' || char == '%' ||
			char == '^' || char == '&' || char == '*' || char == '(' || char == ')' ||
			char == '=' || char == '+' || char == '[' || char == ']' || char == '{' ||
			char == '}' || char == '|' || char == '\\' || char == ':' || char == ';' ||
			char == '"' || char == '\'' || char == '<' || char == '>' || char == '?' {
			return fmt.Errorf("invalid file path - contains invalid character %q: %q", char, path)
		}
	}

	return nil
}

// ValidateConfigContent validates the content of a plonk configuration.
func ValidateConfigContent(content []byte) error {
	// First validate YAML syntax
	if err := ValidateYAML(content); err != nil {
		return err
	}

	// Parse the YAML into a generic structure
	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate package names in homebrew section
	if homebrew, ok := config["homebrew"]; ok {
		if err := validateHomebrewPackages(homebrew); err != nil {
			return err
		}
	}

	// Validate package names in npm section
	if npm, ok := config["npm"]; ok {
		if err := validateNPMPackages(npm); err != nil {
			return err
		}
	}

	// Validate package names in asdf section
	if asdf, ok := config["asdf"]; ok {
		if err := validateASDF(asdf); err != nil {
			return err
		}
	}

	// Validate dotfiles section
	if dotfiles, ok := config["dotfiles"]; ok {
		if err := validateDotfiles(dotfiles); err != nil {
			return err
		}
	}

	return nil
}

// validateHomebrewPackages validates package names in homebrew section
func validateHomebrewPackages(homebrew interface{}) error {
	homebrewMap, ok := homebrew.(map[string]interface{})
	if !ok {
		return nil // Skip if not a map
	}

	// Check brews
	if brews, ok := homebrewMap["brews"]; ok {
		if err := validatePackageList(brews, "homebrew brews"); err != nil {
			return err
		}
	}

	// Check casks
	if casks, ok := homebrewMap["casks"]; ok {
		if err := validatePackageList(casks, "homebrew casks"); err != nil {
			return err
		}
	}

	return nil
}

// validateNPMPackages validates package names in npm section
func validateNPMPackages(npm interface{}) error {
	return validatePackageList(npm, "npm")
}

// validateASDF validates package names in asdf section
func validateASDF(asdf interface{}) error {
	return validatePackageList(asdf, "asdf")
}

// validateDotfiles validates file paths in dotfiles section
func validateDotfiles(dotfiles interface{}) error {
	dotfilesList, ok := dotfiles.([]interface{})
	if !ok {
		return nil // Skip if not a list
	}

	for i, dotfile := range dotfilesList {
		if dotfilePath, ok := dotfile.(string); ok {
			if err := ValidateFilePath(dotfilePath); err != nil {
				return fmt.Errorf("invalid file path in dotfiles[%d]: %w", i, err)
			}
		}
	}

	return nil
}

// validatePackageList validates a list of packages (can be strings or objects with name field)
func validatePackageList(packages interface{}, section string) error {
	packageList, ok := packages.([]interface{})
	if !ok {
		return nil // Skip if not a list
	}

	for i, pkg := range packageList {
		var packageName string
		var configPath string
		var hasConfig bool

		// Handle string packages
		if name, ok := pkg.(string); ok {
			packageName = name
		} else if pkgMap, ok := pkg.(map[string]interface{}); ok {
			// Handle object packages with name field
			if name, ok := pkgMap["name"]; ok {
				if nameStr, ok := name.(string); ok {
					packageName = nameStr
				}
			} else if section == "npm" {
				// For npm, check if there's a "package" field instead
				if pkgName, ok := pkgMap["package"]; ok {
					if pkgStr, ok := pkgName.(string); ok {
						packageName = pkgStr
					}
				}
			}

			// Check for config field
			if config, ok := pkgMap["config"]; ok {
				hasConfig = true
				if configStr, ok := config.(string); ok {
					configPath = configStr
				}
			}
		}

		// Always validate package names, even if empty (to catch empty strings)
		if err := ValidatePackageName(packageName); err != nil {
			return fmt.Errorf("invalid package name in %s[%d]: %w", section, i, err)
		}

		// Validate config path if config field was present (including empty strings)
		if hasConfig {
			if err := ValidateFilePath(configPath); err != nil {
				return fmt.Errorf("invalid file path in %s[%d]: %w", section, i, err)
			}
		}
	}

	return nil
}
