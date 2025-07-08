// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"
	"plonk/internal/managers"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display overall plonk status",
	Long: `Display the complete status of your plonk-managed environment.

Shows:
- Configuration file status
- Package management state (managed/missing/untracked)
- Dotfile management state (managed/missing/untracked)
- List of all managed items

Examples:
  plonk status           # Show overall status
  plonk status -o json   # Show as JSON
  plonk status -o yaml   # Show as YAML`,
	RunE: runStatus,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Initialize output structure
	outputData := StatusOutput{
		ConfigPath: getConfigPath(configDir),
	}

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// Handle missing config gracefully
		if strings.Contains(err.Error(), "config file not found") {
			outputData.ConfigStatus = "missing"
			outputData.ConfigMessage = "Configuration file not found. Run 'plonk config init' to create one."
			return RenderOutput(outputData, format)
		}
		
		// Handle validation errors
		if strings.Contains(err.Error(), "validation failed") {
			outputData.ConfigStatus = "invalid"
			outputData.ConfigMessage = err.Error()
			return RenderOutput(outputData, format)
		}
		
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	outputData.ConfigStatus = "valid"

	// Get package reconciliation state
	pkgManaged, pkgMissing, pkgUntracked := getPackageState(configDir)
	outputData.PackageState = StateCount{
		Managed:   pkgManaged,
		Missing:   pkgMissing,
		Untracked: pkgUntracked,
	}

	// Get dotfile reconciliation state
	dotManaged, dotMissing, dotUntracked, err := reconcileDotfiles(homeDir, configDir)
	if err == nil {
		outputData.DotfileState = StateCount{
			Managed:   len(dotManaged),
			Missing:   len(dotMissing),
			Untracked: len(dotUntracked),
		}
	}

	// Build managed items lists
	var homebrewPackages []string
	for _, brew := range cfg.Homebrew.Brews {
		homebrewPackages = append(homebrewPackages, brew.Name)
	}
	for _, cask := range cfg.Homebrew.Casks {
		homebrewPackages = append(homebrewPackages, cask.Name)
	}

	var asdfTools []string
	for _, tool := range cfg.ASDF {
		asdfTools = append(asdfTools, fmt.Sprintf("%s@%s", tool.Name, tool.Version))
	}

	var npmPackages []string
	for _, pkg := range cfg.NPM {
		npmPackages = append(npmPackages, pkg.Name)
	}

	outputData.ManagedItems = ManagedItems{
		HomebrewPackages: homebrewPackages,
		ASDFTools:        asdfTools,
		NPMPackages:      npmPackages,
		Dotfiles:         cfg.Dotfiles,
		DefaultManager:   cfg.Settings.DefaultManager,
	}

	return RenderOutput(outputData, format)
}

// getPackageState performs package reconciliation across all managers
func getPackageState(configDir string) (int, int, int) {
	managedCount := 0
	missingCount := 0
	untrackedCount := 0

	// Initialize reconciler components
	configLoader := managers.NewPlonkConfigLoader(configDir)
	
	packageManagers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"asdf":     managers.NewAsdfManager(),
		"npm":      managers.NewNpmManager(),
	}

	versionCheckers := map[string]managers.VersionChecker{
		"homebrew": &managers.HomebrewVersionChecker{},
		"asdf":     &managers.AsdfVersionChecker{},
		"npm":      &managers.NpmVersionChecker{},
	}

	// Reconcile each manager
	for managerName, mgr := range packageManagers {
		if !mgr.IsAvailable() {
			continue
		}

		managerMap := map[string]managers.PackageManager{
			managerName: mgr,
		}
		
		reconciler := managers.NewStateReconciler(configLoader, managerMap, versionCheckers)
		result, err := reconciler.ReconcileManager(managerName)
		if err != nil {
			continue // Skip this manager on error
		}
		
		managedCount += len(result.Managed)
		missingCount += len(result.Missing)
		untrackedCount += len(result.Untracked)
	}

	return managedCount, missingCount, untrackedCount
}

// getConfigPath finds the actual config file path
func getConfigPath(configDir string) string {
	mainPath := filepath.Join(configDir, "plonk.yaml")
	if _, err := os.Stat(mainPath); err == nil {
		return mainPath
	}
	
	repoPath := filepath.Join(configDir, "repo", "plonk.yaml")
	if _, err := os.Stat(repoPath); err == nil {
		return repoPath
	}
	
	return mainPath // Default to main path
}

// StatusOutput represents the complete status output
type StatusOutput struct {
	ConfigPath    string       `json:"config_path" yaml:"config_path"`
	ConfigStatus  string       `json:"config_status" yaml:"config_status"`
	ConfigMessage string       `json:"config_message,omitempty" yaml:"config_message,omitempty"`
	PackageState  StateCount   `json:"package_state" yaml:"package_state"`
	DotfileState  StateCount   `json:"dotfile_state" yaml:"dotfile_state"`
	ManagedItems  ManagedItems `json:"managed_items" yaml:"managed_items"`
}

// StateCount represents state counts for packages or dotfiles
type StateCount struct {
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked" yaml:"untracked"`
}

// ManagedItems represents what is being managed by plonk configuration
type ManagedItems struct {
	HomebrewPackages []string `json:"homebrew_packages" yaml:"homebrew_packages"`
	ASDFTools        []string `json:"asdf_tools" yaml:"asdf_tools"`
	NPMPackages      []string `json:"npm_packages" yaml:"npm_packages"`
	Dotfiles         []string `json:"dotfiles" yaml:"dotfiles"`
	DefaultManager   string   `json:"default_manager" yaml:"default_manager"`
}

// TableOutput generates human-friendly table output for overall status
func (s StatusOutput) TableOutput() string {
	output := "Plonk Status\n============\n\n"
	
	// Config status
	output += fmt.Sprintf("📁 Config: %s ", s.ConfigPath)
	switch s.ConfigStatus {
	case "valid":
		output += "(✅ valid)\n\n"
	case "invalid":
		output += "(❌ invalid)\n"
		output += fmt.Sprintf("   %s\n\n", s.ConfigMessage)
	case "missing":
		output += "(📋 missing)\n"
		output += fmt.Sprintf("   %s\n\n", s.ConfigMessage)
		return output // No further info if config missing
	}

	// Package state
	output += "Package State:\n"
	if s.PackageState.Managed > 0 {
		output += fmt.Sprintf("  ✅ %d managed packages\n", s.PackageState.Managed)
	}
	if s.PackageState.Missing > 0 {
		output += fmt.Sprintf("  ❌ %d missing packages (not installed)\n", s.PackageState.Missing)
	}
	if s.PackageState.Untracked > 0 {
		output += fmt.Sprintf("  🔍 %d untracked packages (installed but not in config)\n", s.PackageState.Untracked)
	}
	if s.PackageState.Managed == 0 && s.PackageState.Missing == 0 && s.PackageState.Untracked == 0 {
		output += "  📦 No packages found\n"
	}

	// Dotfile state
	output += "\nDotfile State:\n"
	if s.DotfileState.Managed > 0 {
		output += fmt.Sprintf("  ✅ %d managed dotfiles\n", s.DotfileState.Managed)
	}
	if s.DotfileState.Missing > 0 {
		output += fmt.Sprintf("  ❌ %d missing dotfiles\n", s.DotfileState.Missing)
	}
	if s.DotfileState.Untracked > 0 {
		output += fmt.Sprintf("  🔍 %d untracked dotfiles\n", s.DotfileState.Untracked)
	}
	if s.DotfileState.Managed == 0 && s.DotfileState.Missing == 0 && s.DotfileState.Untracked == 0 {
		output += "  📄 No dotfiles found\n"
	}

	// Currently managing
	output += "\nCurrently Managing:\n"
	
	hasManaged := false
	if len(s.ManagedItems.HomebrewPackages) > 0 {
		output += fmt.Sprintf("  📦 Homebrew: %s\n", strings.Join(s.ManagedItems.HomebrewPackages, ", "))
		hasManaged = true
	}
	
	if len(s.ManagedItems.ASDFTools) > 0 {
		output += fmt.Sprintf("  🔧 ASDF: %s\n", strings.Join(s.ManagedItems.ASDFTools, ", "))
		hasManaged = true
	}
	
	if len(s.ManagedItems.NPMPackages) > 0 {
		output += fmt.Sprintf("  📦 NPM: %s\n", strings.Join(s.ManagedItems.NPMPackages, ", "))
		hasManaged = true
	}
	
	if len(s.ManagedItems.Dotfiles) > 0 {
		output += fmt.Sprintf("  📄 Dotfiles: %s\n", strings.Join(s.ManagedItems.Dotfiles, ", "))
		hasManaged = true
	}

	if !hasManaged {
		output += "  (No items configured)\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (s StatusOutput) StructuredData() any {
	return s
}