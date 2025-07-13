// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// ItemType represents the type of item being processed
type ItemType int

const (
	ItemTypePackage ItemType = iota
	ItemTypeDotfile
	ItemTypeAmbiguous
)

// String returns the string representation of ItemType
func (it ItemType) String() string {
	switch it {
	case ItemTypePackage:
		return "package"
	case ItemTypeDotfile:
		return "dotfile"
	case ItemTypeAmbiguous:
		return "ambiguous"
	default:
		return "unknown"
	}
}

// DetectionRule defines a pattern-based rule for item type detection
type DetectionRule struct {
	Pattern     *regexp.Regexp
	Type        ItemType
	Confidence  float64
	Description string
}

// DetectionRules contains all the patterns for intelligent detection
var DetectionRules = []DetectionRule{
	// High confidence dotfile patterns
	{regexp.MustCompile(`^~/`), ItemTypeDotfile, 1.0, "Tilde path"},
	{regexp.MustCompile(`^\.`), ItemTypeDotfile, 0.95, "Hidden file"},
	{regexp.MustCompile(`/`), ItemTypeDotfile, 0.9, "Contains path separator"},

	// Package patterns with edge cases
	{regexp.MustCompile(`^@[\w-]+/[\w.-]+$`), ItemTypePackage, 0.95, "Scoped npm package"},
	{regexp.MustCompile(`^[\w.-]+\.(js|ts|json|toml|yaml|yml)$`), ItemTypeDotfile, 0.8, "Config file extension"},

	// Ambiguous cases requiring disambiguation
	{regexp.MustCompile(`^(config|settings|preferences)$`), ItemTypeAmbiguous, 0.5, "Common ambiguous name"},
}

// DetectItemType performs simple item type detection using basic heuristics
func DetectItemType(item string) ItemType {
	// Path-like (contains /, ~, starts with .) → dotfile
	if strings.Contains(item, "/") || strings.HasPrefix(item, "~") || strings.HasPrefix(item, ".") {
		return ItemTypeDotfile
	}
	return ItemTypePackage
}

// DetectItemTypeWithConfidence performs enhanced item type detection with confidence scoring
func DetectItemTypeWithConfidence(item string) (ItemType, float64, error) {
	// Check edge cases first
	if itemType, handled := HandleEdgeCases(item); handled {
		if itemType == ItemTypeAmbiguous {
			return itemType, 0.5, fmt.Errorf("ambiguous item type for: %s", item)
		}
		return itemType, 0.9, nil
	}

	// Apply detection rules
	for _, rule := range DetectionRules {
		if rule.Pattern.MatchString(item) {
			if rule.Type == ItemTypeAmbiguous {
				return rule.Type, rule.Confidence, fmt.Errorf("ambiguous item type for: %s", item)
			}
			return rule.Type, rule.Confidence, nil
		}
	}

	// Default: assume package if no path-like characteristics
	if !strings.Contains(item, "/") && !strings.HasPrefix(item, ".") {
		return ItemTypePackage, 0.7, nil
	}

	return ItemTypeAmbiguous, 0.5, fmt.Errorf("ambiguous item type for: %s", item)
}

// HandleEdgeCases handles specific edge cases that detection rules might miss
func HandleEdgeCases(item string) (ItemType, bool) {
	edgeCases := map[string]ItemType{
		// Package names that look like paths
		"node.js":      ItemTypePackage,
		"font-awesome": ItemTypePackage,
		"vue.js":       ItemTypePackage,

		// Common ambiguous items
		"config":   ItemTypeAmbiguous,
		"settings": ItemTypeAmbiguous,
		"bin":      ItemTypeAmbiguous,

		// Files that look like packages
		"package.json": ItemTypeDotfile,
		"Cargo.toml":   ItemTypeDotfile,
		"go.mod":       ItemTypeDotfile,
	}

	if itemType, exists := edgeCases[item]; exists {
		return itemType, true
	}

	return ItemTypeAmbiguous, false
}

// ProcessMixedItems separates a list of items into packages and dotfiles
func ProcessMixedItems(items []string) (packages []string, dotfiles []string) {
	for _, item := range items {
		if DetectItemType(item) == ItemTypeDotfile {
			dotfiles = append(dotfiles, item)
		} else {
			packages = append(packages, item)
		}
	}
	return packages, dotfiles
}

// ProcessMixedItemsWithFlags separates items considering flag overrides
func ProcessMixedItemsWithFlags(items []string, flags *CommandFlags) (packages []string, dotfiles []string, ambiguous []string) {
	for _, item := range items {
		itemType, confidence, err := DetectItemTypeWithConfidence(item)

		// Apply flag overrides
		if flags != nil {
			itemType = ApplyTypeOverride(itemType, flags)
		}

		// Handle ambiguous items
		if err != nil && itemType == ItemTypeAmbiguous && (flags == nil || (!flags.PackageOnly && !flags.DotfileOnly)) {
			ambiguous = append(ambiguous, item)
			continue
		}

		// Categorize based on final determination
		switch itemType {
		case ItemTypeDotfile:
			dotfiles = append(dotfiles, item)
		case ItemTypePackage:
			packages = append(packages, item)
		default:
			// If we can't determine and no flags are set, default to package
			if confidence > 0.6 {
				packages = append(packages, item)
			} else {
				ambiguous = append(ambiguous, item)
			}
		}
	}
	return packages, dotfiles, ambiguous
}

// ResolveAmbiguousItem resolves an ambiguous item using user preference or flags
func ResolveAmbiguousItem(item string, userPreference ItemType, flags map[string]bool) ItemType {
	// Check explicit flags first
	if flags["package"] {
		return ItemTypePackage
	}
	if flags["dotfile"] {
		return ItemTypeDotfile
	}

	// Use user preference or prompt for clarification
	if userPreference != ItemTypeAmbiguous {
		return userPreference
	}

	// Default fallback based on context
	return ItemTypePackage
}

// CommandFlags represents unified flags for the new command structure
type CommandFlags struct {
	Manager     string
	DryRun      bool
	Force       bool
	Verbose     bool
	Output      string
	PackageOnly bool
	DotfileOnly bool
}

// ParseUnifiedFlags parses command flags into a unified structure
func ParseUnifiedFlags(cmd *cobra.Command) (*CommandFlags, error) {
	flags := &CommandFlags{}

	// Parse manager flags with precedence
	if brew, _ := cmd.Flags().GetBool("brew"); brew {
		flags.Manager = "homebrew"
	} else if npm, _ := cmd.Flags().GetBool("npm"); npm {
		flags.Manager = "npm"
	} else if cargo, _ := cmd.Flags().GetBool("cargo"); cargo {
		flags.Manager = "cargo"
	}

	// Parse type override flags
	flags.PackageOnly, _ = cmd.Flags().GetBool("package")
	flags.DotfileOnly, _ = cmd.Flags().GetBool("dotfile")

	// Validate flag combinations
	if flags.PackageOnly && flags.DotfileOnly {
		return nil, fmt.Errorf("cannot specify both --package and --dotfile")
	}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")
	flags.Output, _ = cmd.Flags().GetString("output")

	return flags, nil
}

// ApplyTypeOverride applies flag-based type overrides
func ApplyTypeOverride(itemType ItemType, flags *CommandFlags) ItemType {
	if flags.PackageOnly {
		return ItemTypePackage
	}
	if flags.DotfileOnly {
		return ItemTypeDotfile
	}
	return itemType
}

// MixedOperationError represents errors from mixed operations
type MixedOperationError struct {
	PackageResults []OperationResult `json:"packages"`
	DotfileResults []OperationResult `json:"dotfiles"`
	PartialSuccess bool              `json:"partial_success"`
	Summary        OperationSummary  `json:"summary"`
}

// OperationSummary provides summary information for mixed operations
type OperationSummary struct {
	TotalItems     int `json:"total_items"`
	PackagesAdded  int `json:"packages_added"`
	DotfilesLinked int `json:"dotfiles_linked"`
	Failed         int `json:"failed"`
	Skipped        int `json:"skipped"`
}

// Error implements the error interface
func (e MixedOperationError) Error() string {
	if e.PartialSuccess {
		return fmt.Sprintf("partial success: %d/%d items processed successfully",
			e.Summary.PackagesAdded+e.Summary.DotfilesLinked, e.Summary.TotalItems)
	}
	return fmt.Sprintf("operation failed: %d/%d items failed", e.Summary.Failed, e.Summary.TotalItems)
}

// OperationResult represents the result of a single operation (reused from operations package)
type OperationResult struct {
	Name           string                 `json:"name"`
	Manager        string                 `json:"manager,omitempty"`
	Status         string                 `json:"status"`
	Error          error                  `json:"error,omitempty"`
	AlreadyManaged bool                   `json:"already_managed,omitempty"`
	FilesProcessed int                    `json:"files_processed,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
