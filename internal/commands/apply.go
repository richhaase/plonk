package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"plonk/internal/directories"
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
  plonk apply neovim                              # Apply only neovim configuration
  plonk apply --backup                            # Apply all configurations with backup`,
	RunE: applyCmdRun,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().Bool("backup", false, "Create backups of existing configuration files before applying")
	applyCmd.Flags().Bool("dry-run", false, "Show what would be applied without making any changes")
}

func applyCmdRun(cmd *cobra.Command, args []string) error {
	backup, _ := cmd.Flags().GetBool("backup")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return runApplyWithAllOptions(args, backup, dryRun)
}

func runApply(args []string) error {
	return runApplyWithBackup(args, false)
}

func runApplyWithBackup(args []string, backup bool) error {
	return runApplyWithAllOptions(args, backup, false)
}

func runApplyWithAllOptions(args []string, backup bool, dryRun bool) error {
	plonkDir := directories.Default.PlonkDir()

	// Load configuration
	cfg, err := config.LoadConfig(plonkDir)
	if err != nil {
		return WrapConfigError(err)
	}

	// In dry-run mode, show what would be applied
	if dryRun {
		if len(args) == 0 {
			return previewAllConfigurations(plonkDir, cfg)
		} else {
			packageName := args[0]
			return previewPackageConfiguration(plonkDir, cfg, packageName)
		}
	}

	// If backup is requested, create backups before applying
	if backup {
		if err := createBackupsBeforeApply(cfg, args); err != nil {
			return fmt.Errorf("failed to create backups: %w", err)
		}
	}

	if len(args) == 0 {
		// Apply all configurations
		return applyAllConfigurations(plonkDir, cfg)
	} else {
		// Apply specific package configuration
		packageName := args[0]
		return applyPackageConfiguration(plonkDir, cfg, packageName)
	}
}

func applyAllConfigurations(plonkDir string, config *config.Config) error {
	// Apply global dotfiles
	if err := applyDotfiles(plonkDir, config); err != nil {
		return fmt.Errorf("failed to apply dotfiles: %w", err)
	}

	// Apply ZSH configuration
	if err := applyZSHConfiguration(config); err != nil {
		return fmt.Errorf("failed to apply ZSH configuration: %w", err)
	}

	// Apply Git configuration
	if err := applyGitConfiguration(config); err != nil {
		return fmt.Errorf("failed to apply Git configuration: %w", err)
	}

	// Apply package configurations
	if err := applyPackageConfigurations(plonkDir, config); err != nil {
		return fmt.Errorf("failed to apply package configurations: %w", err)
	}

	fmt.Printf("Successfully applied all configurations from %s\n", plonkDir)
	return nil
}

func applyPackageConfiguration(plonkDir string, config *config.Config, packageName string) error {
	// Find the package and apply its configuration
	packageConfig := findPackageConfig(config, packageName)
	if packageConfig == "" {
		return fmt.Errorf("package '%s' not found or has no configuration", packageName)
	}

	// Determine the correct source directory
	sourceDir := getSourceDirectory(plonkDir)

	if err := applyConfigPath(sourceDir, packageConfig); err != nil {
		return fmt.Errorf("failed to apply %s configuration: %w", packageName, err)
	}

	fmt.Printf("Successfully applied %s configuration\n", packageName)
	return nil
}

func applyDotfiles(plonkDir string, config *config.Config) error {
	dotfileTargets := config.GetDotfileTargets()

	// Determine the correct source directory (check repo subdirectory first)
	sourceDir := getSourceDirectory(plonkDir)

	for source, target := range dotfileTargets {
		sourcePath := filepath.Join(sourceDir, source)
		targetPath := directories.Default.ExpandHomeDir(target)

		if err := copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", source, target, err)
		}
	}

	return nil
}

func applyPackageConfigurations(plonkDir string, config *config.Config) error {
	// Determine the correct source directory
	sourceDir := getSourceDirectory(plonkDir)

	// Apply Homebrew package configurations
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Config != "" {
			if err := applyConfigPath(sourceDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}

	for _, pkg := range config.Homebrew.Casks {
		if pkg.Config != "" {
			if err := applyConfigPath(sourceDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}

	// Apply ASDF package configurations
	for _, tool := range config.ASDF {
		if tool.Config != "" {
			if err := applyConfigPath(sourceDir, tool.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", tool.Name, err)
			}
		}
	}

	// Apply NPM package configurations
	for _, pkg := range config.NPM {
		if pkg.Config != "" {
			if err := applyConfigPath(sourceDir, pkg.Config); err != nil {
				return fmt.Errorf("failed to apply %s config: %w", pkg.Name, err)
			}
		}
	}

	return nil
}

// getSourceDirectory returns the directory where source files are located
// Checks repo subdirectory first, then falls back to main plonk directory
func getSourceDirectory(plonkDir string) string {
	repoDir := filepath.Join(plonkDir, "repo")

	// Check if repo directory exists and has content
	if entries, err := os.ReadDir(repoDir); err == nil && len(entries) > 0 {
		return repoDir
	}

	// Fall back to main plonk directory
	return plonkDir
}

func findPackageConfig(config *config.Config, packageName string) string {
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
	targetPath := directories.Default.ExpandHomeDir("~/." + configPath)

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

func applyZSHConfiguration(cfg *config.Config) error {
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

func applyGitConfiguration(cfg *config.Config) error {
	// Skip if no Git configuration is defined
	if len(cfg.Git.User) == 0 && len(cfg.Git.Core) == 0 && len(cfg.Git.Aliases) == 0 &&
		len(cfg.Git.Color) == 0 && len(cfg.Git.Delta) == 0 && len(cfg.Git.Fetch) == 0 &&
		len(cfg.Git.Pull) == 0 && len(cfg.Git.Push) == 0 && len(cfg.Git.Status) == 0 &&
		len(cfg.Git.Diff) == 0 && len(cfg.Git.Log) == 0 && len(cfg.Git.Init) == 0 &&
		len(cfg.Git.Rerere) == 0 && len(cfg.Git.Branch) == 0 && len(cfg.Git.Rebase) == 0 &&
		len(cfg.Git.Merge) == 0 && len(cfg.Git.Filter) == 0 {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Generate and write .gitconfig
	gitconfigContent := config.GenerateGitconfig(&cfg.Git)
	gitconfigPath := filepath.Join(homeDir, ".gitconfig")
	if err := os.WriteFile(gitconfigPath, []byte(gitconfigContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitconfig: %w", err)
	}

	return nil
}

// createBackupsBeforeApply determines which files will be overwritten and creates backups
func createBackupsBeforeApply(cfg *config.Config, args []string) error {
	var filesToBackup []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(args) == 0 {
		// Backing up for full apply - check all files that will be written

		// Add dotfiles
		dotfileTargets := cfg.GetDotfileTargets()
		for _, target := range dotfileTargets {
			targetPath := directories.Default.ExpandHomeDir(target)
			filesToBackup = append(filesToBackup, targetPath)
		}

		// Add ZSH configuration files if ZSH config exists
		if len(cfg.ZSH.EnvVars) > 0 || len(cfg.ZSH.Aliases) > 0 ||
			len(cfg.ZSH.Inits) > 0 || len(cfg.ZSH.Completions) > 0 ||
			len(cfg.ZSH.Functions) > 0 || len(cfg.ZSH.ShellOptions) > 0 {
			filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".zshrc"))

			// Only backup .zshenv if there are environment variables
			if len(cfg.ZSH.EnvVars) > 0 {
				filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".zshenv"))
			}
		}

		// Add Git configuration file if Git config exists
		if len(cfg.Git.User) > 0 || len(cfg.Git.Core) > 0 || len(cfg.Git.Aliases) > 0 ||
			len(cfg.Git.Color) > 0 || len(cfg.Git.Delta) > 0 || len(cfg.Git.Fetch) > 0 ||
			len(cfg.Git.Pull) > 0 || len(cfg.Git.Push) > 0 || len(cfg.Git.Status) > 0 ||
			len(cfg.Git.Diff) > 0 || len(cfg.Git.Log) > 0 || len(cfg.Git.Init) > 0 ||
			len(cfg.Git.Rerere) > 0 || len(cfg.Git.Branch) > 0 || len(cfg.Git.Rebase) > 0 ||
			len(cfg.Git.Merge) > 0 || len(cfg.Git.Filter) > 0 {
			filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".gitconfig"))
		}

		// Add package configuration files
		filesToBackup = append(filesToBackup, getPackageConfigFilesToBackup(cfg)...)
	} else {
		// Backing up for specific package apply - only backup that package's config
		packageName := args[0]
		packageConfig := findPackageConfig(cfg, packageName)
		if packageConfig != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + packageConfig)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	// Create backups using existing backup functionality
	return BackupConfigurationFiles(filesToBackup)
}

// getPackageConfigFilesToBackup returns all package configuration file paths
func getPackageConfigFilesToBackup(cfg *config.Config) []string {
	var filesToBackup []string

	// Homebrew package configurations
	for _, pkg := range cfg.Homebrew.Brews {
		if pkg.Config != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + pkg.Config)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	for _, pkg := range cfg.Homebrew.Casks {
		if pkg.Config != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + pkg.Config)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	// ASDF package configurations
	for _, tool := range cfg.ASDF {
		if tool.Config != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + tool.Config)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	// NPM package configurations
	for _, pkg := range cfg.NPM {
		if pkg.Config != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + pkg.Config)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	return filesToBackup
}

// previewAllConfigurations shows what would be applied for all configurations
func previewAllConfigurations(plonkDir string, config *config.Config) error {
	fmt.Printf("Dry-run mode: Showing what would be applied from %s\n\n", plonkDir)

	// Preview global dotfiles
	if err := previewDotfiles(plonkDir, config); err != nil {
		return fmt.Errorf("failed to preview dotfiles: %w", err)
	}

	// Preview ZSH configuration
	if err := previewZSHConfiguration(config); err != nil {
		return fmt.Errorf("failed to preview ZSH configuration: %w", err)
	}

	// Preview Git configuration
	if err := previewGitConfiguration(config); err != nil {
		return fmt.Errorf("failed to preview Git configuration: %w", err)
	}

	// Preview package configurations
	if err := previewPackageConfigurations(plonkDir, config); err != nil {
		return fmt.Errorf("failed to preview package configurations: %w", err)
	}

	fmt.Printf("\nDry-run complete. No files were modified.\n")
	return nil
}

// previewPackageConfiguration shows what would be applied for a specific package
func previewPackageConfiguration(plonkDir string, config *config.Config, packageName string) error {
	fmt.Printf("Dry-run mode: Showing what would be applied for package '%s'\n\n", packageName)

	// Find the package and preview its configuration
	packageConfig := findPackageConfig(config, packageName)
	if packageConfig == "" {
		return fmt.Errorf("package '%s' not found or has no configuration", packageName)
	}

	// Determine the correct source directory
	sourceDir := getSourceDirectory(plonkDir)

	if err := previewConfigPath(sourceDir, packageConfig); err != nil {
		return fmt.Errorf("failed to preview %s configuration: %w", packageName, err)
	}

	fmt.Printf("\nDry-run complete. No files were modified.\n")
	return nil
}

// previewDotfiles shows what dotfiles would be applied
func previewDotfiles(plonkDir string, config *config.Config) error {
	dotfileTargets := config.GetDotfileTargets()

	if len(dotfileTargets) == 0 {
		return nil
	}

	fmt.Printf("Dotfiles that would be applied:\n")

	// Determine the correct source directory (check repo subdirectory first)
	sourceDir := getSourceDirectory(plonkDir)

	for source, target := range dotfileTargets {
		sourcePath := filepath.Join(sourceDir, source)
		targetPath := directories.Default.ExpandHomeDir(target)

		// Check if source exists
		if _, err := os.Stat(sourcePath); err != nil {
			fmt.Printf("  ⚠️  %s -> %s (source not found)\n", source, target)
		} else {
			// Check if target exists
			if _, err := os.Stat(targetPath); err == nil {
				fmt.Printf("  📝 %s -> %s (would overwrite existing file)\n", source, target)
			} else {
				fmt.Printf("  ✨ %s -> %s (would create new file)\n", source, target)
			}
		}
	}

	fmt.Println()
	return nil
}

// previewZSHConfiguration shows what ZSH configuration would be applied
func previewZSHConfiguration(cfg *config.Config) error {
	// Skip if no ZSH configuration is defined
	if len(cfg.ZSH.EnvVars) == 0 && len(cfg.ZSH.Aliases) == 0 &&
		len(cfg.ZSH.Inits) == 0 && len(cfg.ZSH.Completions) == 0 &&
		len(cfg.ZSH.Functions) == 0 && len(cfg.ZSH.ShellOptions) == 0 {
		return nil
	}

	fmt.Printf("ZSH configuration that would be generated:\n")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check .zshrc
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if _, err := os.Stat(zshrcPath); err == nil {
		fmt.Printf("  📝 .zshrc (would overwrite existing file)\n")
	} else {
		fmt.Printf("  ✨ .zshrc (would create new file)\n")
	}

	// Check .zshenv (only if there are environment variables)
	if len(cfg.ZSH.EnvVars) > 0 {
		zshenvPath := filepath.Join(homeDir, ".zshenv")
		if _, err := os.Stat(zshenvPath); err == nil {
			fmt.Printf("  📝 .zshenv (would overwrite existing file)\n")
		} else {
			fmt.Printf("  ✨ .zshenv (would create new file)\n")
		}
	}

	fmt.Println()
	return nil
}

// previewGitConfiguration shows what Git configuration would be applied
func previewGitConfiguration(cfg *config.Config) error {
	// Skip if no Git configuration is defined
	if len(cfg.Git.User) == 0 && len(cfg.Git.Core) == 0 && len(cfg.Git.Aliases) == 0 &&
		len(cfg.Git.Color) == 0 && len(cfg.Git.Delta) == 0 && len(cfg.Git.Fetch) == 0 &&
		len(cfg.Git.Pull) == 0 && len(cfg.Git.Push) == 0 && len(cfg.Git.Status) == 0 &&
		len(cfg.Git.Diff) == 0 && len(cfg.Git.Log) == 0 && len(cfg.Git.Init) == 0 &&
		len(cfg.Git.Rerere) == 0 && len(cfg.Git.Branch) == 0 && len(cfg.Git.Rebase) == 0 &&
		len(cfg.Git.Merge) == 0 && len(cfg.Git.Filter) == 0 {
		return nil
	}

	fmt.Printf("Git configuration that would be generated:\n")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check .gitconfig
	gitconfigPath := filepath.Join(homeDir, ".gitconfig")
	if _, err := os.Stat(gitconfigPath); err == nil {
		fmt.Printf("  📝 .gitconfig (would overwrite existing file)\n")
	} else {
		fmt.Printf("  ✨ .gitconfig (would create new file)\n")
	}

	fmt.Println()
	return nil
}

// previewPackageConfigurations shows what package configurations would be applied
func previewPackageConfigurations(plonkDir string, config *config.Config) error {
	hasAnyPackageConfig := false

	// Check Homebrew package configurations
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Config != "" {
			hasAnyPackageConfig = true
			break
		}
	}

	if !hasAnyPackageConfig {
		for _, pkg := range config.Homebrew.Casks {
			if pkg.Config != "" {
				hasAnyPackageConfig = true
				break
			}
		}
	}

	// Check ASDF package configurations
	if !hasAnyPackageConfig {
		for _, tool := range config.ASDF {
			if tool.Config != "" {
				hasAnyPackageConfig = true
				break
			}
		}
	}

	// Check NPM package configurations
	if !hasAnyPackageConfig {
		for _, pkg := range config.NPM {
			if pkg.Config != "" {
				hasAnyPackageConfig = true
				break
			}
		}
	}

	if !hasAnyPackageConfig {
		return nil
	}

	fmt.Printf("Package configurations that would be applied:\n")

	// Determine the correct source directory
	sourceDir := getSourceDirectory(plonkDir)

	// Preview Homebrew package configurations
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Config != "" {
			if err := previewConfigPath(sourceDir, pkg.Config); err != nil {
				fmt.Printf("  ⚠️  %s config: %v\n", pkg.Name, err)
			}
		}
	}

	for _, pkg := range config.Homebrew.Casks {
		if pkg.Config != "" {
			if err := previewConfigPath(sourceDir, pkg.Config); err != nil {
				fmt.Printf("  ⚠️  %s config: %v\n", pkg.Name, err)
			}
		}
	}

	// Preview ASDF package configurations
	for _, tool := range config.ASDF {
		if tool.Config != "" {
			if err := previewConfigPath(sourceDir, tool.Config); err != nil {
				fmt.Printf("  ⚠️  %s config: %v\n", tool.Name, err)
			}
		}
	}

	// Preview NPM package configurations
	for _, pkg := range config.NPM {
		if pkg.Config != "" {
			if err := previewConfigPath(sourceDir, pkg.Config); err != nil {
				fmt.Printf("  ⚠️  %s config: %v\n", pkg.Name, err)
			}
		}
	}

	fmt.Println()
	return nil
}

// previewConfigPath shows what would happen for a specific config path
func previewConfigPath(plonkDir, configPath string) error {
	sourcePath := filepath.Join(plonkDir, configPath)
	targetPath := directories.Default.ExpandHomeDir("~/." + configPath)

	// Check if source exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		fmt.Printf("  ⚠️  %s -> %s (source not found)\n", configPath, "~/."+configPath)
		return nil
	}

	// Check if target exists
	if _, err := os.Stat(targetPath); err == nil {
		if sourceInfo.IsDir() {
			fmt.Printf("  📁 %s -> %s (would overwrite existing directory)\n", configPath, "~/."+configPath)
		} else {
			fmt.Printf("  📝 %s -> %s (would overwrite existing file)\n", configPath, "~/."+configPath)
		}
	} else {
		if sourceInfo.IsDir() {
			fmt.Printf("  📁 %s -> %s (would create new directory)\n", configPath, "~/."+configPath)
		} else {
			fmt.Printf("  ✨ %s -> %s (would create new file)\n", configPath, "~/."+configPath)
		}
	}

	return nil
}
