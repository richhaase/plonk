package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
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
	
	plonkDir := getPlonkDir()
	
	// Load configuration
	config, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Create package managers
	executor := &managers.RealCommandExecutor{}
	homebrewMgr := managers.NewHomebrewManager(executor)
	asdfMgr := managers.NewAsdfManager(executor)
	npmMgr := managers.NewNpmManager(executor)
	
	// Install Homebrew packages
	if err := installHomebrewPackages(homebrewMgr, config); err != nil {
		return fmt.Errorf("failed to install Homebrew packages: %w", err)
	}
	
	// Install ASDF tools
	if err := installASDFTools(asdfMgr, config); err != nil {
		return fmt.Errorf("failed to install ASDF tools: %w", err)
	}
	
	// Install NPM packages
	if err := installNPMPackages(npmMgr, config); err != nil {
		return fmt.Errorf("failed to install NPM packages: %w", err)
	}
	
	fmt.Printf("Successfully installed packages from %s\n", filepath.Join(plonkDir, "plonk.yaml"))
	return nil
}

func installHomebrewPackages(mgr *managers.HomebrewManager, config *config.YAMLConfig) error {
	if !mgr.IsAvailable() {
		return fmt.Errorf("Homebrew is not available")
	}
	
	// Install brews
	for _, pkg := range config.Homebrew.Brews {
		packageName := pkg.Name
		if !mgr.IsInstalled(packageName) {
			fmt.Printf("Installing Homebrew package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return fmt.Errorf("failed to install %s: %w", packageName, err)
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
				return fmt.Errorf("failed to install cask %s: %w", packageName, err)
			}
		} else {
			fmt.Printf("Homebrew cask %s already installed\n", packageName)
		}
	}
	
	return nil
}

func installASDFTools(mgr *managers.AsdfManager, config *config.YAMLConfig) error {
	if !mgr.IsAvailable() {
		return fmt.Errorf("ASDF is not available")
	}
	
	for _, tool := range config.ASDF {
		if !mgr.IsVersionInstalled(tool.Name, tool.Version) {
			fmt.Printf("Installing ASDF tool: %s@%s\n", tool.Name, tool.Version)
			if err := mgr.InstallVersion(tool.Name, tool.Version); err != nil {
				return fmt.Errorf("failed to install %s@%s: %w", tool.Name, tool.Version, err)
			}
		} else {
			fmt.Printf("ASDF tool %s@%s already installed\n", tool.Name, tool.Version)
		}
	}
	
	return nil
}

func installNPMPackages(mgr *managers.NpmManager, config *config.YAMLConfig) error {
	if !mgr.IsAvailable() {
		return fmt.Errorf("NPM is not available")
	}
	
	for _, pkg := range config.NPM {
		// Use package name if specified, otherwise use name
		packageName := pkg.Name
		if pkg.Package != "" {
			packageName = pkg.Package
		}
		
		if !mgr.IsInstalled(packageName) {
			fmt.Printf("Installing NPM package: %s\n", packageName)
			if err := mgr.Install(packageName); err != nil {
				return fmt.Errorf("failed to install %s: %w", packageName, err)
			}
		} else {
			fmt.Printf("NPM package %s already installed\n", packageName)
		}
	}
	
	return nil
}