package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"plonk/pkg/config"
)

var applyCmd = &cobra.Command{
	Use:   "apply [package]",
	Short: "Apply configuration files",
	Long: `Deploy configuration files from the plonk directory to their target locations.

With no arguments, applies all dotfiles and package configurations.
With a package name, applies only that package's configuration files.

Examples:
  plonk apply                                     # Apply all configurations
  plonk apply neovim                              # Apply only neovim configuration`,
	RunE: applyCmdRun,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func applyCmdRun(cmd *cobra.Command, args []string) error {
	return runApply(args)
}

func runApply(args []string) error {
	plonkDir := getPlonkDir()
	
	// Load configuration
	config, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	if len(args) == 0 {
		// Apply all configurations
		return applyAllConfigurations(plonkDir, config)
	} else {
		// Apply specific package configuration
		packageName := args[0]
		return applyPackageConfiguration(plonkDir, config, packageName)
	}
}

func applyAllConfigurations(plonkDir string, config *config.YAMLConfig) error {
	// Apply global dotfiles
	if err := applyDotfiles(plonkDir, config); err != nil {
		return fmt.Errorf("failed to apply dotfiles: %w", err)
	}
	
	// Apply ZSH configuration
	if err := applyZSHConfiguration(config); err != nil {
		return fmt.Errorf("failed to apply ZSH configuration: %w", err)
	}
	
	// Apply package configurations
	if err := applyPackageConfigurations(plonkDir, config); err != nil {
		return fmt.Errorf("failed to apply package configurations: %w", err)
	}
	
	fmt.Printf("Successfully applied all configurations from %s\n", plonkDir)
	return nil
}

func applyPackageConfiguration(plonkDir string, config *config.YAMLConfig, packageName string) error {
	// Find the package and apply its configuration
	packageConfig := findPackageConfig(config, packageName)
	if packageConfig == "" {
		return fmt.Errorf("package '%s' not found or has no configuration", packageName)
	}
	
	if err := applyConfigPath(plonkDir, packageConfig); err != nil {
		return fmt.Errorf("failed to apply %s configuration: %w", packageName, err)
	}
	
	fmt.Printf("Successfully applied %s configuration\n", packageName)
	return nil
}

func applyDotfiles(plonkDir string, config *config.YAMLConfig) error {
	dotfileTargets := config.GetDotfileTargets()
	
	for source, target := range dotfileTargets {
		sourcePath := filepath.Join(plonkDir, source)
		targetPath := expandHomeDir(target)
		
		if err := copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", source, target, err)
		}
	}
	
	return nil
}

func applyPackageConfigurations(plonkDir string, config *config.YAMLConfig) error {
	// Apply Homebrew package configurations
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Config != "" {
			if err := applyConfigPath(plonkDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}
	
	for _, pkg := range config.Homebrew.Casks {
		if pkg.Config != "" {
			if err := applyConfigPath(plonkDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}
	
	// Apply ASDF package configurations
	for _, tool := range config.ASDF {
		if tool.Config != "" {
			if err := applyConfigPath(plonkDir, tool.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", tool.Name, err)
			}
		}
	}
	
	// Apply NPM package configurations
	for _, pkg := range config.NPM {
		if pkg.Config != "" {
			if err := applyConfigPath(plonkDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}
	
	return nil
}

func findPackageConfig(config *config.YAMLConfig, packageName string) string {
	// Check Homebrew packages
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Name == packageName && pkg.Config != "" {
			return pkg.Config
		}
	}
	
	for _, pkg := range config.Homebrew.Casks {
		if pkg.Name == packageName && pkg.Config != "" {
			return pkg.Config
		}
	}
	
	// Check ASDF tools
	for _, tool := range config.ASDF {
		if tool.Name == packageName && tool.Config != "" {
			return tool.Config
		}
	}
	
	// Check NPM packages
	for _, pkg := range config.NPM {
		if pkg.Name == packageName && pkg.Config != "" {
			return pkg.Config
		}
	}
	
	return ""
}

func applyConfigPath(plonkDir, configPath string) error {
	sourcePath := filepath.Join(plonkDir, configPath)
	targetPath := expandHomeDir("~/." + configPath)
	
	return copyFileOrDir(sourcePath, targetPath)
}

func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Write to destination
	return os.WriteFile(dst, data, 0644)
}

func copyFileOrDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if srcInfo.IsDir() {
		return copyDir(src, dst)
	} else {
		return copyFile(src, dst)
	}
}

func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if err := copyFileOrDir(srcPath, dstPath); err != nil {
			return err
		}
	}
	
	return nil
}

func expandHomeDir(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // fallback to original path
	}
	
	if len(path) == 1 || path[1] == '/' {
		return filepath.Join(homeDir, path[1:])
	}
	
	// Handle ~user syntax (though we don't use it in plonk)
	return path
}

func applyZSHConfiguration(cfg *config.YAMLConfig) error {
	// Skip if no ZSH configuration is defined
	if len(cfg.ZSH.EnvVars) == 0 && len(cfg.ZSH.Aliases) == 0 && 
	   len(cfg.ZSH.Inits) == 0 && len(cfg.ZSH.Completions) == 0 &&
	   len(cfg.ZSH.Functions) == 0 && len(cfg.ZSH.ShellOptions) == 0 {
		return nil
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	// Generate and write .zshrc
	zshrcContent := config.GenerateZshrc(&cfg.ZSH)
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(zshrcPath, []byte(zshrcContent), 0644); err != nil {
		return fmt.Errorf("failed to write .zshrc: %w", err)
	}
	
	// Generate and write .zshenv (only if there are environment variables)
	if len(cfg.ZSH.EnvVars) > 0 {
		zshenvContent := config.GenerateZshenv(&cfg.ZSH)
		zshenvPath := filepath.Join(homeDir, ".zshenv")
		if err := os.WriteFile(zshenvPath, []byte(zshenvContent), 0644); err != nil {
			return fmt.Errorf("failed to write .zshenv: %w", err)
		}
	}
	
	return nil
}