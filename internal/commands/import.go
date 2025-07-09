// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"plonk/internal/config"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var (
	importInteractive bool
	importPackages    bool
	importDotfiles    bool
	importAll         bool
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import existing packages and dotfiles into plonk management",
	Long: `The import command helps new plonk users discover and selectively import their existing
packages and dotfiles into plonk management. It provides an interactive interface to choose
what to import, making it easy to get started with plonk.

This command will:
- Discover existing packages installed on your system
- Find dotfiles in your home directory
- Present an interactive selection interface
- Safely import selected items into plonk configuration

Examples:
  plonk import                    # Interactive import of packages and dotfiles
  plonk import --packages         # Import packages only
  plonk import --dotfiles         # Import dotfiles only
  plonk import --all              # Import all discovered items without interaction`,
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolVarP(&importInteractive, "interactive", "i", true, "Interactive selection (default)")
	importCmd.Flags().BoolVarP(&importPackages, "packages", "p", false, "Import packages only")
	importCmd.Flags().BoolVarP(&importDotfiles, "dotfiles", "d", false, "Import dotfiles only")
	importCmd.Flags().BoolVarP(&importAll, "all", "a", false, "Import all discovered items")
}

func runImport(cmd *cobra.Command, args []string) error {
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

	// Load or create configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			cfg = &config.Config{
				Settings: config.Settings{
					DefaultManager: "homebrew",
				},
				Dotfiles: []config.DotfileEntry{},
			}
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Discover untracked items
	discoveredPackages, discoveredDotfiles, err := discoverUntrackedItems(cfg, homeDir, configDir)
	if err != nil {
		return fmt.Errorf("failed to discover untracked items: %w", err)
	}

	// Filter items based on flags
	var itemsToImport ImportItems
	if importPackages && !importDotfiles {
		itemsToImport.Packages = discoveredPackages
	} else if importDotfiles && !importPackages {
		itemsToImport.Dotfiles = discoveredDotfiles
	} else {
		// Both or neither specified - include both
		itemsToImport.Packages = discoveredPackages
		itemsToImport.Dotfiles = discoveredDotfiles
	}

	// Apply intelligent filtering
	itemsToImport = applyIntelligentFiltering(itemsToImport)

	// Handle different modes
	if importAll {
		// Import all items without interaction
		return importAllItems(cfg, itemsToImport, configDir, format)
	} else {
		// Interactive selection (default behavior)
		return importInteractiveMode(cfg, itemsToImport, configDir, format)
	}
}

// ImportItems represents items that can be imported
type ImportItems struct {
	Packages []DiscoveredPackage `json:"packages" yaml:"packages"`
	Dotfiles []DiscoveredDotfile `json:"dotfiles" yaml:"dotfiles"`
}

// DiscoveredPackage represents a package that can be imported
type DiscoveredPackage struct {
	Name        string `json:"name" yaml:"name"`
	Manager     string `json:"manager" yaml:"manager"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// DiscoveredDotfile represents a dotfile that can be imported
type DiscoveredDotfile struct {
	Name        string `json:"name" yaml:"name"`
	Path        string `json:"path" yaml:"path"`
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// discoverUntrackedItems discovers packages and dotfiles that aren't managed by plonk
func discoverUntrackedItems(cfg *config.Config, homeDir, configDir string) ([]DiscoveredPackage, []DiscoveredDotfile, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create package provider: %w", err)
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Register dotfile provider
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	// Reconcile to get state
	summary, err := reconciler.ReconcileAll(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to reconcile state: %w", err)
	}

	var packages []DiscoveredPackage
	var dotfiles []DiscoveredDotfile

	// Extract untracked packages
	for _, result := range summary.Results {
		if result.Domain == "package" {
			for _, item := range result.Untracked {
				packages = append(packages, DiscoveredPackage{
					Name:        item.Name,
					Manager:     result.Manager,
					Description: getPackageDescription(item.Name, result.Manager),
				})
			}
		}
	}

	// Extract untracked dotfiles
	for _, result := range summary.Results {
		if result.Domain == "dotfile" {
			for _, item := range result.Untracked {
				source, destination := generateDotfilePaths(item.Name, homeDir)
				dotfiles = append(dotfiles, DiscoveredDotfile{
					Name:        item.Name,
					Path:        item.Path,
					Source:      source,
					Destination: destination,
					Description: getDotfileDescription(item.Name),
				})
			}
		}
	}

	return packages, dotfiles, nil
}

// applyIntelligentFiltering filters out system files and prioritizes commonly imported items
func applyIntelligentFiltering(items ImportItems) ImportItems {
	// Filter dotfiles
	var filteredDotfiles []DiscoveredDotfile
	for _, dotfile := range items.Dotfiles {
		if !isSystemDotfile(dotfile.Name) {
			filteredDotfiles = append(filteredDotfiles, dotfile)
		}
	}

	// Filter packages (less aggressive filtering)
	var filteredPackages []DiscoveredPackage
	for _, pkg := range items.Packages {
		if !isSystemPackage(pkg.Name, pkg.Manager) {
			filteredPackages = append(filteredPackages, pkg)
		}
	}

	// Sort by name
	sort.Slice(filteredPackages, func(i, j int) bool {
		return filteredPackages[i].Name < filteredPackages[j].Name
	})

	sort.Slice(filteredDotfiles, func(i, j int) bool {
		return filteredDotfiles[i].Name < filteredDotfiles[j].Name
	})

	return ImportItems{
		Packages: filteredPackages,
		Dotfiles: filteredDotfiles,
	}
}

// generateDotfilePaths generates source and destination paths for a dotfile
func generateDotfilePaths(name, homeDir string) (string, string) {
	destination := "~/" + name
	source := config.TargetToSource(destination)
	return source, destination
}

// Helper functions for filtering and recommendations
func isSystemDotfile(name string) bool {
	systemFiles := []string{
		".DS_Store", ".Trash", ".CFUserTextEncoding", ".cups",
		".cache", ".npm", ".bundle", ".gem", ".cargo", ".rustup",
		".local", ".lesshst", ".zsh_history", ".bash_history",
		".zcompdump", ".zcompcache", ".viminfo", ".zsh_sessions",
	}
	
	for _, sysFile := range systemFiles {
		if name == sysFile {
			return true
		}
	}
	
	// Filter out certain directory patterns
	if strings.HasPrefix(name, ".config") {
		// Only recommend certain config directories
		return !strings.Contains(name, "/nvim/") && !strings.Contains(name, "/git/") && !strings.Contains(name, "/alacritty/")
	}
	
	return false
}

func isSystemPackage(name, manager string) bool {
	// System packages that users typically don't want to track
	systemPackages := map[string][]string{
		"homebrew": {
			"ca-certificates", "openssl", "sqlite", "zlib", "readline",
			"ncurses", "libidn2", "libunistring", "libxcrypt", "util-linux",
			"oniguruma", "coreutils", "findutils", "gnu-sed", "grep",
		},
		"npm": {
			"npm", "corepack",
		},
	}
	
	if sysPkgs, exists := systemPackages[manager]; exists {
		for _, sysPkg := range sysPkgs {
			if name == sysPkg {
				return true
			}
		}
	}
	
	return false
}


func getPackageDescription(name, manager string) string {
	descriptions := map[string]string{
		"git":        "Version control system",
		"curl":       "Command line tool for transferring data",
		"wget":       "Network downloader",
		"jq":         "Command-line JSON processor",
		"htop":       "Interactive process viewer",
		"tree":       "Directory tree viewer",
		"fzf":        "Command-line fuzzy finder",
		"neovim":     "Hyperextensible Vim-based text editor",
		"tmux":       "Terminal multiplexer",
		"ripgrep":    "Fast text search tool",
		"fd":         "Fast alternative to find",
		"bat":        "Cat clone with syntax highlighting",
		"exa":        "Modern replacement for ls",
		"typescript": "JavaScript with types",
		"prettier":   "Code formatter",
		"eslint":     "JavaScript linter",
		"nodemon":    "Node.js development utility",
	}
	
	if desc, exists := descriptions[name]; exists {
		return desc
	}
	
	return ""
}

func getDotfileDescription(name string) string {
	descriptions := map[string]string{
		".zshrc":        "Zsh shell configuration",
		".bashrc":       "Bash shell configuration",
		".vimrc":        "Vim editor configuration",
		".gitconfig":    "Git configuration",
		".tmux.conf":    "Tmux configuration",
		".editorconfig": "Editor configuration",
		".gitignore_global": "Global Git ignore patterns",
	}
	
	if desc, exists := descriptions[name]; exists {
		return desc
	}
	
	// Pattern-based descriptions
	if strings.HasPrefix(name, ".config/nvim/") {
		return "Neovim configuration"
	} else if strings.HasPrefix(name, ".config/git/") {
		return "Git configuration"
	} else if strings.HasPrefix(name, ".config/alacritty/") {
		return "Alacritty terminal configuration"
	}
	
	return ""
}

// importAllItems imports all discovered items without interaction
func importAllItems(cfg *config.Config, items ImportItems, configDir string, format OutputFormat) error {
	if len(items.Packages) == 0 && len(items.Dotfiles) == 0 {
		fmt.Println("🎉 No untracked items found! Your system is already well-managed by plonk.")
		return nil
	}

	// Show what will be imported
	fmt.Printf("📋 Importing all %d packages and %d dotfiles:\n", len(items.Packages), len(items.Dotfiles))
	for _, pkg := range items.Packages {
		fmt.Printf("  📦 %s (%s)", pkg.Name, pkg.Manager)
		if pkg.Description != "" {
			fmt.Printf(" - %s", pkg.Description)
		}
		fmt.Println()
	}
	for _, dotfile := range items.Dotfiles {
		fmt.Printf("  📄 %s", dotfile.Name)
		if dotfile.Description != "" {
			fmt.Printf(" - %s", dotfile.Description)
		}
		fmt.Println()
	}

	fmt.Println() // Add spacing before import starts

	var imported []string
	
	// Import packages
	for _, pkg := range items.Packages {
		if err := importPackage(cfg, pkg, configDir); err != nil {
			fmt.Printf("Warning: Failed to import package %s: %v\n", pkg.Name, err)
			continue
		}
		imported = append(imported, fmt.Sprintf("📦 %s (%s)", pkg.Name, pkg.Manager))
	}

	// Import dotfiles
	for _, dotfile := range items.Dotfiles {
		if err := importDotfile(cfg, dotfile, configDir); err != nil {
			fmt.Printf("Warning: Failed to import dotfile %s: %v\n", dotfile.Name, err)
			continue
		}
		imported = append(imported, fmt.Sprintf("📄 %s", dotfile.Name))
	}

	// Save configuration
	if err := saveConfig(cfg, configDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show results
	if format == OutputTable {
		fmt.Printf("\n✅ Successfully imported %d items:\n", len(imported))
		for _, item := range imported {
			fmt.Printf("  %s\n", item)
		}
	}

	return nil
}

// importInteractiveMode provides an interactive selection interface
func importInteractiveMode(cfg *config.Config, items ImportItems, configDir string, format OutputFormat) error {
	if len(items.Packages) == 0 && len(items.Dotfiles) == 0 {
		fmt.Println("🎉 No untracked items found! Your system is already well-managed by plonk.")
		return nil
	}

	fmt.Println("🔍 Welcome to plonk import!")
	fmt.Println("I found some packages and dotfiles that aren't managed by plonk yet.")
	fmt.Println("Let's help you import them into plonk management.")
	fmt.Println()

	var selectedPackages []DiscoveredPackage
	var selectedDotfiles []DiscoveredDotfile

	// Interactive package selection
	if len(items.Packages) > 0 {
		fmt.Printf("📦 Found %d packages:\n", len(items.Packages))
		selectedPackages = selectPackagesInteractive(items.Packages)
	}

	// Interactive dotfile selection
	if len(items.Dotfiles) > 0 {
		fmt.Printf("📄 Found %d dotfiles:\n", len(items.Dotfiles))
		selectedDotfiles = selectDotfilesInteractive(items.Dotfiles)
	}

	// Show selection summary
	totalSelected := len(selectedPackages) + len(selectedDotfiles)
	if totalSelected == 0 {
		fmt.Println("No items selected. Exiting.")
		return nil
	}

	fmt.Printf("\n📋 You selected %d items to import:\n", totalSelected)
	for _, pkg := range selectedPackages {
		fmt.Printf("  📦 %s (%s)", pkg.Name, pkg.Manager)
		if pkg.Description != "" {
			fmt.Printf(" - %s", pkg.Description)
		}
		fmt.Println()
	}
	for _, dotfile := range selectedDotfiles {
		fmt.Printf("  📄 %s", dotfile.Name)
		if dotfile.Description != "" {
			fmt.Printf(" - %s", dotfile.Description)
		}
		fmt.Println()
	}

	// Confirmation
	fmt.Print("\nProceed with import? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Import canceled.")
		return nil
	}

	// Import selected items
	var imported []string
	
	for _, pkg := range selectedPackages {
		if err := importPackage(cfg, pkg, configDir); err != nil {
			fmt.Printf("Warning: Failed to import package %s: %v\n", pkg.Name, err)
			continue
		}
		imported = append(imported, fmt.Sprintf("📦 %s (%s)", pkg.Name, pkg.Manager))
	}

	for _, dotfile := range selectedDotfiles {
		if err := importDotfile(cfg, dotfile, configDir); err != nil {
			fmt.Printf("Warning: Failed to import dotfile %s: %v\n", dotfile.Name, err)
			continue
		}
		imported = append(imported, fmt.Sprintf("📄 %s", dotfile.Name))
	}

	// Save configuration
	if err := saveConfig(cfg, configDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show results
	fmt.Printf("\n✅ Successfully imported %d items:\n", len(imported))
	for _, item := range imported {
		fmt.Printf("  %s\n", item)
	}

	fmt.Println("\n🎉 Welcome to plonk! Your packages and dotfiles are now managed.")
	fmt.Println("💡 Tip: Run 'plonk status' to see your managed items.")

	return nil
}

// selectPackagesInteractive provides interactive package selection
func selectPackagesInteractive(packages []DiscoveredPackage) []DiscoveredPackage {
	if len(packages) == 0 {
		return nil
	}

	fmt.Println("\nEnter numbers separated by spaces (e.g., 1 3 5) or 'all' for all:")
	
	// Display packages with numbers
	for i, pkg := range packages {
		fmt.Printf("  %2d. %s (%s)", i+1, pkg.Name, pkg.Manager)
		if pkg.Description != "" {
			fmt.Printf(" - %s", pkg.Description)
		}
		fmt.Println()
	}

	// Get user selection
	fmt.Print("\nSelect packages: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	if input == "all" {
		return packages
	}

	// Parse numbers
	var selected []DiscoveredPackage
	parts := strings.Fields(input)
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil && num > 0 && num <= len(packages) {
			selected = append(selected, packages[num-1])
		}
	}

	return selected
}

// selectDotfilesInteractive provides interactive dotfile selection
func selectDotfilesInteractive(dotfiles []DiscoveredDotfile) []DiscoveredDotfile {
	if len(dotfiles) == 0 {
		return nil
	}

	fmt.Println("\nEnter numbers separated by spaces (e.g., 1 3 5) or 'all' for all:")
	
	// Display dotfiles with numbers
	for i, dotfile := range dotfiles {
		fmt.Printf("  %2d. %s", i+1, dotfile.Name)
		if dotfile.Description != "" {
			fmt.Printf(" - %s", dotfile.Description)
		}
		fmt.Println()
	}

	// Get user selection
	fmt.Print("\nSelect dotfiles: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	if input == "all" {
		return dotfiles
	}

	// Parse numbers
	var selected []DiscoveredDotfile
	parts := strings.Fields(input)
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil && num > 0 && num <= len(dotfiles) {
			selected = append(selected, dotfiles[num-1])
		}
	}

	return selected
}

// importPackage imports a single package
func importPackage(cfg *config.Config, pkg DiscoveredPackage, configDir string) error {
	// Check if already in config
	if isPackageInConfig(cfg, pkg.Name, pkg.Manager) {
		return nil
	}

	// Add to configuration
	return addPackageToConfig(cfg, pkg.Name, pkg.Manager)
}

// importDotfile imports a single dotfile
func importDotfile(cfg *config.Config, dotfile DiscoveredDotfile, configDir string) error {
	// Check if already managed
	for _, entry := range cfg.Dotfiles {
		if entry.Destination == dotfile.Destination {
			return nil
		}
	}

	// Copy dotfile to plonk config directory
	sourcePath := filepath.Join(configDir, dotfile.Source)
	if err := copyDotfile(dotfile.Path, sourcePath); err != nil {
		return fmt.Errorf("failed to copy dotfile: %w", err)
	}

	// Add to configuration
	newEntry := config.DotfileEntry{
		Source:      dotfile.Source,
		Destination: dotfile.Destination,
	}
	cfg.Dotfiles = append(cfg.Dotfiles, newEntry)

	return nil
}

// ImportOutput represents the output structure for import command
type ImportOutput struct {
	Items ImportItems `json:"items" yaml:"items"`
}

// TableOutput generates human-friendly table output
func (a ImportOutput) TableOutput() string {
	return "" // Handled in command logic
}

// StructuredData returns the structured data for serialization
func (a ImportOutput) StructuredData() any {
	return a.Items
}