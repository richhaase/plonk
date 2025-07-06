package commands

import (
	"fmt"

	"plonk/pkg/config"
)

// Helper functions to reduce duplication in package installation

// extractInstalledPackages combines package lists from all managers
func extractInstalledPackages(packages map[string][]string) []string {
	var result []string
	for _, pkgs := range packages {
		result = append(result, pkgs...)
	}
	return result
}

// shouldInstallPackage determines if a package needs to be installed
func shouldInstallPackage(packageName string, isInstalled bool) bool {
	return !isInstalled
}

// getPackageDisplayName returns the display name for a package
func getPackageDisplayName(pkg interface{}) string {
	switch p := pkg.(type) {
	case config.HomebrewPackage:
		return p.Name
	case config.ASDFTool:
		return fmt.Sprintf("%s@%s", p.Name, p.Version)
	case config.NPMPackage:
		if p.Package != "" {
			return p.Package
		}
		return p.Name
	default:
		return ""
	}
}

// getPackageConfig returns the config path for a package
func getPackageConfig(pkg interface{}) string {
	switch p := pkg.(type) {
	case config.HomebrewPackage:
		return p.Config
	case config.ASDFTool:
		return p.Config
	case config.NPMPackage:
		return p.Config
	default:
		return ""
	}
}

// getPackageName returns the base name for a package (used for configs)
func getPackageName(pkg interface{}) string {
	switch p := pkg.(type) {
	case config.HomebrewPackage:
		return p.Name
	case config.ASDFTool:
		return p.Name
	case config.NPMPackage:
		return p.Name
	default:
		return ""
	}
}
