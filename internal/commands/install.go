// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"

	"plonk/internal/directories"
	"plonk/pkg/config"
	"plonk/pkg/managers"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install packages from configuration",
	Long: `Install packages defined in the YAML configuration file.

Reads plonk.yaml (and optionally plonk.local.yaml) from the plonk directory
and installs all defined packages using their respective package managers.

With no arguments, installs all packages from configuration.
With a package name, installs only that specific package.

Examples:
  plonk install                                   # Install all packages from config
  plonk install neovim                            # Install only neovim package`,
	RunE: installCmdRun,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	dryRun := IsDryRun(cmd)
	return runInstallWithOptions(args, dryRun)
}

func runInstall(args []string) error {
	return runInstallWithOptions(args, false)
}

func runInstallWithOptions(args []string, dryRun bool) error {
	plonkDir := directories.Default.PlonkDir()

	// Load configuration.
	config, err := config.LoadConfig(plonkDir)
	if err != nil {
		return WrapConfigError(err)
	}

	// Create package managers.
	executor := &managers.RealCommandExecutor{}
	homebrewMgr := managers.NewHomebrewManager(executor)
	asdfMgr := managers.NewAsdfManager(executor)
	npmMgr := managers.NewNpmManager(executor)

	if dryRun {
		if len(args) == 0 {
			// Preview all packages that would be installed
			return previewAllPackageInstallation(homebrewMgr, asdfMgr, npmMgr, config, plonkDir)
		} else {
			// Preview specific package installation
			packageName := args[0]
			return previewSpecificPackageInstallation(homebrewMgr, asdfMgr, npmMgr, config, plonkDir, packageName)
		}
	}

	if len(args) == 0 {
		// Install all packages
		return installAllPackages(homebrewMgr, asdfMgr, npmMgr, config, plonkDir)
	} else {
		// Install specific package
		packageName := args[0]
		return installSpecificPackage(homebrewMgr, asdfMgr, npmMgr, config, plonkDir, packageName)
	}
}

func installAllPackages(homebrewMgr *managers.HomebrewManager, asdfMgr *managers.AsdfManager, npmMgr *managers.NpmManager, config *config.Config, plonkDir string) error {
	// Track packages that were installed and have configs.
	installedPackages := make(map[string][]string)

	// Install Homebrew packages.
	installedHomebrewPackages, err := installHomebrewPackages(homebrewMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install Homebrew packages: %w", err)
	}
	installedPackages["homebrew"] = installedHomebrewPackages

	// Install ASDF tools.
	installedASDFTools, err := installASDFTools(asdfMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install ASDF tools: %w", err)
	}
	installedPackages["asdf"] = installedASDFTools

	// Install NPM packages.
	installedNPMPackages, err := installNPMPackages(npmMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install NPM packages: %w", err)
	}
	installedPackages["npm"] = installedNPMPackages

	// Extract all packages with configs.
	packagesWithConfigs := extractInstalledPackages(installedPackages)

	// Apply configurations for newly installed packages.
	if len(packagesWithConfigs) > 0 {
		fmt.Println("Applying configurations for installed packages...")
		for _, packageName := range packagesWithConfigs {
			if err := applyPackageConfiguration(plonkDir, config, packageName); err != nil {
				// Don't fail the entire install if config application fails.
				fmt.Printf("Warning: failed to apply configuration for %s: %v\n", packageName, err)
			}
		}
	}

	fmt.Printf("Successfully installed packages from %s\n", filepath.Join(plonkDir, "plonk.yaml"))
	return nil
}

func installSpecificPackage(homebrewMgr *managers.HomebrewManager, asdfMgr *managers.AsdfManager, npmMgr *managers.NpmManager, config *config.Config, plonkDir string, packageName string) error {
	// Check if package exists in configuration
	var packageFound bool
	var installedWithConfig bool

	// Check Homebrew packages
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Name == packageName {
			packageFound = true
			if !homebrewMgr.IsAvailable() {
				return WrapPackageManagerError("homebrew", fmt.Errorf("command not found"))
			}
			if shouldInstallPackage(pkg.Name, homebrewMgr.IsInstalled(pkg.Name)) {
				fmt.Printf("Installing Homebrew package: %s\n", pkg.Name)
				if err := homebrewMgr.Install(pkg.Name); err != nil {
					return WrapInstallError(pkg.Name, err)
				}
				if getPackageConfig(pkg) != "" {
					installedWithConfig = true
				}
			} else {
				fmt.Printf("Homebrew package %s already installed\n", pkg.Name)
			}
			break
		}
	}

	// Check Homebrew casks if not found in brews
	if !packageFound {
		for _, pkg := range config.Homebrew.Casks {
			if pkg.Name == packageName {
				packageFound = true
				if !homebrewMgr.IsAvailable() {
					return WrapPackageManagerError("homebrew", fmt.Errorf("command not found"))
				}
				if shouldInstallPackage(pkg.Name, homebrewMgr.IsInstalled(pkg.Name)) {
					fmt.Printf("Installing Homebrew cask: %s\n", pkg.Name)
					if err := homebrewMgr.InstallCask(pkg.Name); err != nil {
						return WrapInstallError(pkg.Name, err)
					}
					if getPackageConfig(pkg) != "" {
						installedWithConfig = true
					}
				} else {
					fmt.Printf("Homebrew cask %s already installed\n", pkg.Name)
				}
				break
			}
		}
	}

	// Check ASDF tools if not found in Homebrew
	if !packageFound {
		for _, tool := range config.ASDF {
			if tool.Name == packageName {
				packageFound = true
				if !asdfMgr.IsAvailable() {
					return WrapPackageManagerError("asdf", fmt.Errorf("command not found"))
				}
				if shouldInstallPackage(tool.Name, asdfMgr.IsVersionInstalled(tool.Name, tool.Version)) {
					displayName := getPackageDisplayName(tool)
					fmt.Printf("Installing ASDF tool: %s\n", displayName)
					if err := asdfMgr.InstallVersion(tool.Name, tool.Version); err != nil {
						return WrapInstallError(displayName, err)
					}
					if getPackageConfig(tool) != "" {
						installedWithConfig = true
					}
				} else {
					fmt.Printf("ASDF tool %s already installed\n", getPackageDisplayName(tool))
				}
				break
			}
		}
	}

	// Check NPM packages if not found in other managers
	if !packageFound {
		for _, pkg := range config.NPM {
			if pkg.Name == packageName {
				packageFound = true
				if !npmMgr.IsAvailable() {
					return WrapPackageManagerError("npm", fmt.Errorf("command not found"))
				}
				packageDisplayName := getPackageDisplayName(pkg)
				if shouldInstallPackage(packageDisplayName, npmMgr.IsInstalled(packageDisplayName)) {
					fmt.Printf("Installing NPM package: %s\n", packageDisplayName)
					if err := npmMgr.Install(packageDisplayName); err != nil {
						return WrapInstallError(packageDisplayName, err)
					}
					if getPackageConfig(pkg) != "" {
						installedWithConfig = true
					}
				} else {
					fmt.Printf("NPM package %s already installed\n", packageDisplayName)
				}
				break
			}
		}
	}

	if !packageFound {
		return fmt.Errorf("package '%s' not found in configuration", packageName)
	}

	// Apply configuration if package has one and was newly installed
	if installedWithConfig {
		fmt.Printf("Applying configuration for %s...\n", packageName)
		if err := applyPackageConfiguration(plonkDir, config, packageName); err != nil {
			fmt.Printf("Warning: failed to apply configuration for %s: %v\n", packageName, err)
		}
	}

	fmt.Printf("Successfully installed package: %s\n", packageName)
	return nil
}

func installHomebrewPackages(mgr *managers.HomebrewManager, config *config.Config) ([]string, error) {
	// Skip if no packages to install.
	if len(config.Homebrew.Brews) == 0 && len(config.Homebrew.Casks) == 0 {
		return []string{}, nil
	}

	if !mgr.IsAvailable() {
		return nil, WrapPackageManagerError("homebrew", fmt.Errorf("command not found"))
	}

	var installedWithConfigs []string

	// Install brews.
	for _, pkg := range config.Homebrew.Brews {
		if shouldInstallPackage(pkg.Name, mgr.IsInstalled(pkg.Name)) {
			fmt.Printf("Installing Homebrew package: %s\n", pkg.Name)
			if err := mgr.Install(pkg.Name); err != nil {
				return nil, WrapInstallError(pkg.Name, err)
			}
			// If package has config and was installed, add to list.
			if getPackageConfig(pkg) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(pkg))
			}
		} else {
			fmt.Printf("Homebrew package %s already installed\n", pkg.Name)
		}
	}

	// Install casks.
	for _, pkg := range config.Homebrew.Casks {
		if shouldInstallPackage(pkg.Name, mgr.IsInstalled(pkg.Name)) {
			fmt.Printf("Installing Homebrew cask: %s\n", pkg.Name)
			if err := mgr.InstallCask(pkg.Name); err != nil {
				return nil, WrapInstallError(pkg.Name, err)
			}
			// If package has config and was installed, add to list.
			if getPackageConfig(pkg) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(pkg))
			}
		} else {
			fmt.Printf("Homebrew cask %s already installed\n", pkg.Name)
		}
	}

	return installedWithConfigs, nil
}

func installASDFTools(mgr *managers.AsdfManager, config *config.Config) ([]string, error) {
	// Skip if no tools to install.
	if len(config.ASDF) == 0 {
		return []string{}, nil
	}

	if !mgr.IsAvailable() {
		return nil, WrapPackageManagerError("asdf", fmt.Errorf("command not found"))
	}

	var installedWithConfigs []string

	for _, tool := range config.ASDF {
		if shouldInstallPackage(tool.Name, mgr.IsVersionInstalled(tool.Name, tool.Version)) {
			displayName := getPackageDisplayName(tool)
			fmt.Printf("Installing ASDF tool: %s\n", displayName)
			if err := mgr.InstallVersion(tool.Name, tool.Version); err != nil {
				return nil, WrapInstallError(displayName, err)
			}
			// If tool has config and was installed, add to list.
			if getPackageConfig(tool) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(tool))
			}
		} else {
			fmt.Printf("ASDF tool %s already installed\n", getPackageDisplayName(tool))
		}
	}

	return installedWithConfigs, nil
}

func installNPMPackages(mgr *managers.NpmManager, config *config.Config) ([]string, error) {
	// Skip if no packages to install.
	if len(config.NPM) == 0 {
		return []string{}, nil
	}

	if !mgr.IsAvailable() {
		return nil, WrapPackageManagerError("npm", fmt.Errorf("command not found"))
	}

	var installedWithConfigs []string

	for _, pkg := range config.NPM {
		// Use package name if specified, otherwise use name.
		packageName := getPackageDisplayName(pkg)

		if shouldInstallPackage(packageName, mgr.IsInstalled(packageName)) {
			fmt.Printf("Installing NPM package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return nil, WrapInstallError(packageName, err)
			}
			// If package has config and was installed, add to list.
			if getPackageConfig(pkg) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(pkg))
			}
		} else {
			fmt.Printf("NPM package %s already installed\n", packageName)
		}
	}

	return installedWithConfigs, nil
}

// previewAllPackageInstallation shows what packages would be installed without actually installing them
func previewAllPackageInstallation(homebrewMgr *managers.HomebrewManager, asdfMgr *managers.AsdfManager, npmMgr *managers.NpmManager, config *config.Config, plonkDir string) error {
	fmt.Printf("Dry-run mode: Showing what packages would be installed from %s\n\n", filepath.Join(plonkDir, "plonk.yaml"))

	// Preview Homebrew packages
	if err := previewHomebrewPackages(homebrewMgr, config); err != nil {
		return err
	}

	// Preview ASDF tools
	if err := previewASDFTools(asdfMgr, config); err != nil {
		return err
	}

	// Preview NPM packages
	if err := previewNPMPackages(npmMgr, config); err != nil {
		return err
	}

	fmt.Printf("\nDry-run complete. No packages were installed.\n")
	return nil
}

// previewSpecificPackageInstallation shows what would happen when installing a specific package
func previewSpecificPackageInstallation(homebrewMgr *managers.HomebrewManager, asdfMgr *managers.AsdfManager, npmMgr *managers.NpmManager, config *config.Config, plonkDir string, packageName string) error {
	fmt.Printf("Dry-run mode: Showing what would happen when installing package '%s'\n\n", packageName)

	// Check if package exists in configuration and preview its installation
	var packageFound bool

	// Check Homebrew packages
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Name == packageName {
			packageFound = true
			if !homebrewMgr.IsAvailable() {
				fmt.Printf("❌ Homebrew not available - package %s cannot be installed\n", pkg.Name)
			} else if homebrewMgr.IsInstalled(pkg.Name) {
				fmt.Printf("✅ Homebrew package %s is already installed\n", pkg.Name)
			} else {
				fmt.Printf("📦 Would install Homebrew package: %s\n", pkg.Name)
				if getPackageConfig(pkg) != "" {
					fmt.Printf("⚙️  Would apply configuration from: %s\n", getPackageConfig(pkg))
				}
			}
			break
		}
	}

	// Check Homebrew casks if not found in brews
	if !packageFound {
		for _, pkg := range config.Homebrew.Casks {
			if pkg.Name == packageName {
				packageFound = true
				if !homebrewMgr.IsAvailable() {
					fmt.Printf("❌ Homebrew not available - cask %s cannot be installed\n", pkg.Name)
				} else if homebrewMgr.IsInstalled(pkg.Name) {
					fmt.Printf("✅ Homebrew cask %s is already installed\n", pkg.Name)
				} else {
					fmt.Printf("📦 Would install Homebrew cask: %s\n", pkg.Name)
					if getPackageConfig(pkg) != "" {
						fmt.Printf("⚙️  Would apply configuration from: %s\n", getPackageConfig(pkg))
					}
				}
				break
			}
		}
	}

	// Check ASDF tools if not found in Homebrew
	if !packageFound {
		for _, tool := range config.ASDF {
			if tool.Name == packageName {
				packageFound = true
				if !asdfMgr.IsAvailable() {
					fmt.Printf("❌ ASDF not available - tool %s cannot be installed\n", tool.Name)
				} else if asdfMgr.IsVersionInstalled(tool.Name, tool.Version) {
					fmt.Printf("✅ ASDF tool %s is already installed\n", getPackageDisplayName(tool))
				} else {
					fmt.Printf("🔧 Would install ASDF tool: %s\n", getPackageDisplayName(tool))
					if getPackageConfig(tool) != "" {
						fmt.Printf("⚙️  Would apply configuration from: %s\n", getPackageConfig(tool))
					}
				}
				break
			}
		}
	}

	// Check NPM packages if not found in other managers
	if !packageFound {
		for _, pkg := range config.NPM {
			if pkg.Name == packageName {
				packageFound = true
				packageDisplayName := getPackageDisplayName(pkg)
				if !npmMgr.IsAvailable() {
					fmt.Printf("❌ NPM not available - package %s cannot be installed\n", packageDisplayName)
				} else if npmMgr.IsInstalled(packageDisplayName) {
					fmt.Printf("✅ NPM package %s is already installed\n", packageDisplayName)
				} else {
					fmt.Printf("📦 Would install NPM package: %s\n", packageDisplayName)
					if getPackageConfig(pkg) != "" {
						fmt.Printf("⚙️  Would apply configuration from: %s\n", getPackageConfig(pkg))
					}
				}
				break
			}
		}
	}

	if !packageFound {
		return fmt.Errorf("package '%s' not found in configuration", packageName)
	}

	fmt.Printf("\nDry-run complete. No packages were installed.\n")
	return nil
}

// previewHomebrewPackages shows what Homebrew packages would be installed
func previewHomebrewPackages(mgr *managers.HomebrewManager, config *config.Config) error {
	if len(config.Homebrew.Brews) == 0 && len(config.Homebrew.Casks) == 0 {
		return nil
	}

	fmt.Printf("Homebrew packages that would be installed:\n")

	if !mgr.IsAvailable() {
		fmt.Printf("❌ Homebrew not available - packages cannot be installed\n\n")
		return nil
	}

	// Preview brews
	for _, pkg := range config.Homebrew.Brews {
		if mgr.IsInstalled(pkg.Name) {
			fmt.Printf("✅ %s (already installed)\n", pkg.Name)
		} else {
			fmt.Printf("📦 %s (would install)\n", pkg.Name)
			if getPackageConfig(pkg) != "" {
				fmt.Printf("   ⚙️  Configuration: %s\n", getPackageConfig(pkg))
			}
		}
	}

	// Preview casks
	for _, pkg := range config.Homebrew.Casks {
		if mgr.IsInstalled(pkg.Name) {
			fmt.Printf("✅ %s (cask, already installed)\n", pkg.Name)
		} else {
			fmt.Printf("📦 %s (cask, would install)\n", pkg.Name)
			if getPackageConfig(pkg) != "" {
				fmt.Printf("   ⚙️  Configuration: %s\n", getPackageConfig(pkg))
			}
		}
	}

	fmt.Println()
	return nil
}

// previewASDFTools shows what ASDF tools would be installed
func previewASDFTools(mgr *managers.AsdfManager, config *config.Config) error {
	if len(config.ASDF) == 0 {
		return nil
	}

	fmt.Printf("ASDF tools that would be installed:\n")

	if !mgr.IsAvailable() {
		fmt.Printf("❌ ASDF not available - tools cannot be installed\n\n")
		return nil
	}

	for _, tool := range config.ASDF {
		displayName := getPackageDisplayName(tool)
		if mgr.IsVersionInstalled(tool.Name, tool.Version) {
			fmt.Printf("✅ %s (already installed)\n", displayName)
		} else {
			fmt.Printf("🔧 %s (would install)\n", displayName)
			if getPackageConfig(tool) != "" {
				fmt.Printf("   ⚙️  Configuration: %s\n", getPackageConfig(tool))
			}
		}
	}

	fmt.Println()
	return nil
}

// previewNPMPackages shows what NPM packages would be installed
func previewNPMPackages(mgr *managers.NpmManager, config *config.Config) error {
	if len(config.NPM) == 0 {
		return nil
	}

	fmt.Printf("NPM packages that would be installed:\n")

	if !mgr.IsAvailable() {
		fmt.Printf("❌ NPM not available - packages cannot be installed\n\n")
		return nil
	}

	for _, pkg := range config.NPM {
		packageName := getPackageDisplayName(pkg)
		if mgr.IsInstalled(packageName) {
			fmt.Printf("✅ %s (already installed)\n", packageName)
		} else {
			fmt.Printf("📦 %s (would install)\n", packageName)
			if getPackageConfig(pkg) != "" {
				fmt.Printf("   ⚙️  Configuration: %s\n", getPackageConfig(pkg))
			}
		}
	}

	fmt.Println()
	return nil
}
