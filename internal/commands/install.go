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
	if len(args) > 0 {
		return fmt.Errorf("install command takes no arguments")
	}
	
	plonkDir := directories.Default.PlonkDir()
	
	// Load configuration
	config, err := config.LoadConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Create package managers
	executor := &managers.RealCommandExecutor{}
	homebrewMgr := managers.NewHomebrewManager(executor)
	asdfMgr := managers.NewAsdfManager(executor)
	npmMgr := managers.NewNpmManager(executor)
	
	// Track packages that were installed and have configs
	var packagesWithConfigs []string
	
	// Install Homebrew packages
	installedHomebrewPackages, err := installHomebrewPackages(homebrewMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install Homebrew packages: %w", err)
	}
	packagesWithConfigs = append(packagesWithConfigs, installedHomebrewPackages...)
	
	// Install ASDF tools
	installedASDFTools, err := installASDFTools(asdfMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install ASDF tools: %w", err)
	}
	packagesWithConfigs = append(packagesWithConfigs, installedASDFTools...)
	
	// Install NPM packages
	installedNPMPackages, err := installNPMPackages(npmMgr, config)
	if err != nil {
		return fmt.Errorf("failed to install NPM packages: %w", err)
	}
	packagesWithConfigs = append(packagesWithConfigs, installedNPMPackages...)
	
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
		return nil, fmt.Errorf("Homebrew is not available")
	}
	
	var installedWithConfigs []string
	
	// Install brews
	for _, pkg := range config.Homebrew.Brews {
		packageName := pkg.Name
		if !mgr.IsInstalled(packageName) {
			fmt.Printf("Installing Homebrew package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return nil, fmt.Errorf("failed to install %s: %w", packageName, err)
			}
			// If package has config and was installed, add to list
			if pkg.Config != "" {
				installedWithConfigs = append(installedWithConfigs, packageName)
			}
		} else {
			fmt.Printf("Homebrew package %s already installed\n", packageName)
		}
	}
	
	// Install casks
	for _, pkg := range config.Homebrew.Casks {
		packageName := pkg.Name
		if !mgr.IsInstalled(packageName) {
			fmt.Printf("Installing Homebrew cask: %s\n", packageName)
			if err := mgr.InstallCask(packageName); err != nil {
				return nil, fmt.Errorf("failed to install cask %s: %w", packageName, err)
			}
			// If package has config and was installed, add to list
			if pkg.Config != "" {
				installedWithConfigs = append(installedWithConfigs, packageName)
			}
		} else {
			fmt.Printf("Homebrew cask %s already installed\n", packageName)
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
		return nil, fmt.Errorf("ASDF is not available")
	}
	
	var installedWithConfigs []string
	
	for _, tool := range config.ASDF {
		if !mgr.IsVersionInstalled(tool.Name, tool.Version) {
			fmt.Printf("Installing ASDF tool: %s@%s\n", tool.Name, tool.Version)
			if err := mgr.InstallVersion(tool.Name, tool.Version); err != nil {
				return nil, fmt.Errorf("failed to install %s@%s: %w", tool.Name, tool.Version, err)
			}
			// If tool has config and was installed, add to list
			if tool.Config != "" {
				installedWithConfigs = append(installedWithConfigs, tool.Name)
			}
		} else {
			fmt.Printf("ASDF tool %s@%s already installed\n", tool.Name, tool.Version)
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
		return nil, fmt.Errorf("NPM is not available")
	}
	
	var installedWithConfigs []string
	
	for _, pkg := range config.NPM {
		// Use package name if specified, otherwise use name
		packageName := pkg.Name
		if pkg.Package != "" {
			packageName = pkg.Package
		}
		
		if !mgr.IsInstalled(packageName) {
			fmt.Printf("Installing NPM package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return nil, fmt.Errorf("failed to install %s: %w", packageName, err)
			}
			// If package has config and was installed, add to list
			if pkg.Config != "" {
				installedWithConfigs = append(installedWithConfigs, pkg.Name)
			}
		} else {
			fmt.Printf("NPM package %s already installed\n", packageName)
		}
	}
	
	return installedWithConfigs, nil
}