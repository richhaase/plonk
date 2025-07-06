package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"plonk/internal/directories"
	"plonk/pkg/config"
	"plonk/pkg/managers"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install packages from configuration",
	Long: `Install packages defined in the YAML configuration file.

Reads plonk.yaml (and optionally plonk.local.yaml) from the plonk directory
and installs all defined packages using their respective package managers.

Examples:
  plonk install                                   # Install all packages from config`,
	RunE: installCmdRun,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	return runInstall(args)
}

func runInstall(args []string) error {
	if err := ValidateNoArgs("install", args); err != nil {
		return err
	}

	plonkDir := directories.Default.PlonkDir()

	// Load configuration
	config, err := config.LoadConfig(plonkDir)
	if err != nil {
		return WrapConfigError(err)
	}

	// Create package managers
	executor := &managers.RealCommandExecutor{}
	homebrewMgr := managers.NewHomebrewManager(executor)
	asdfMgr := managers.NewAsdfManager(executor)
	npmMgr := managers.NewNpmManager(executor)

	// Track packages that were installed and have configs
	installedPackages := make(map[string][]string)

	// Install Homebrew packages
	installedHomebrewPackages, err := installHomebrewPackages(homebrewMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install Homebrew packages: %w", err)
	}
	installedPackages["homebrew"] = installedHomebrewPackages

	// Install ASDF tools
	installedASDFTools, err := installASDFTools(asdfMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install ASDF tools: %w", err)
	}
	installedPackages["asdf"] = installedASDFTools

	// Install NPM packages
	installedNPMPackages, err := installNPMPackages(npmMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install NPM packages: %w", err)
	}
	installedPackages["npm"] = installedNPMPackages

	// Extract all packages with configs
	packagesWithConfigs := extractInstalledPackages(installedPackages)

	// Apply configurations for newly installed packages
	if len(packagesWithConfigs) > 0 {
		fmt.Println("Applying configurations for installed packages...")
		for _, packageName := range packagesWithConfigs {
			if err := applyPackageConfiguration(plonkDir, config, packageName); err != nil {
				// Don't fail the entire install if config application fails
				fmt.Printf("Warning: failed to apply configuration for %s: %v\n", packageName, err)
			}
		}
	}

	fmt.Printf("Successfully installed packages from %s\n", filepath.Join(plonkDir, "plonk.yaml"))
	return nil
}

func installHomebrewPackages(mgr *managers.HomebrewManager, config *config.Config) ([]string, error) {
	// Skip if no packages to install
	if len(config.Homebrew.Brews) == 0 && len(config.Homebrew.Casks) == 0 {
		return []string{}, nil
	}

	if !mgr.IsAvailable() {
		return nil, WrapPackageManagerError("homebrew", fmt.Errorf("command not found"))
	}

	var installedWithConfigs []string

	// Install brews
	for _, pkg := range config.Homebrew.Brews {
		if shouldInstallPackage(pkg.Name, mgr.IsInstalled(pkg.Name)) {
			fmt.Printf("Installing Homebrew package: %s\n", pkg.Name)
			if err := mgr.Install(pkg.Name); err != nil {
				return nil, WrapInstallError(pkg.Name, err)
			}
			// If package has config and was installed, add to list
			if getPackageConfig(pkg) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(pkg))
			}
		} else {
			fmt.Printf("Homebrew package %s already installed\n", pkg.Name)
		}
	}

	// Install casks
	for _, pkg := range config.Homebrew.Casks {
		if shouldInstallPackage(pkg.Name, mgr.IsInstalled(pkg.Name)) {
			fmt.Printf("Installing Homebrew cask: %s\n", pkg.Name)
			if err := mgr.InstallCask(pkg.Name); err != nil {
				return nil, WrapInstallError(pkg.Name, err)
			}
			// If package has config and was installed, add to list
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
	// Skip if no tools to install
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
			// If tool has config and was installed, add to list
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
	// Skip if no packages to install
	if len(config.NPM) == 0 {
		return []string{}, nil
	}

	if !mgr.IsAvailable() {
		return nil, WrapPackageManagerError("npm", fmt.Errorf("command not found"))
	}

	var installedWithConfigs []string

	for _, pkg := range config.NPM {
		// Use package name if specified, otherwise use name
		packageName := getPackageDisplayName(pkg)

		if shouldInstallPackage(packageName, mgr.IsInstalled(packageName)) {
			fmt.Printf("Installing NPM package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return nil, WrapInstallError(packageName, err)
			}
			// If package has config and was installed, add to list
			if getPackageConfig(pkg) != "" {
				installedWithConfigs = append(installedWithConfigs, getPackageName(pkg))
			}
		} else {
			fmt.Printf("NPM package %s already installed\n", packageName)
		}
	}

	return installedWithConfigs, nil
}
