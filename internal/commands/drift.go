package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"plonk/pkg/config"
	"plonk/pkg/managers"
)

// DriftReport represents configuration drift between expected and actual state
type DriftReport struct {
	MissingFiles     []string
	ModifiedFiles    []string
	MissingPackages  []string
	ExtraPackages    []string
}

// HasDrift returns true if any drift is detected
func (d *DriftReport) HasDrift() bool {
	return len(d.MissingFiles) > 0 ||
		len(d.ModifiedFiles) > 0 ||
		len(d.MissingPackages) > 0 ||
		len(d.ExtraPackages) > 0
}

// detectConfigDrift compares current system state with plonk configuration
func detectConfigDrift() (*DriftReport, error) {
	plonkDir := getPlonkDir()
	
	// Load configuration
	config, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	drift := &DriftReport{
		MissingFiles:    []string{},
		ModifiedFiles:   []string{},
		MissingPackages: []string{},
		ExtraPackages:   []string{},
	}
	
	// Check dotfile drift
	if err := checkDotfileDrift(plonkDir, config, drift); err != nil {
		return nil, fmt.Errorf("failed to check dotfile drift: %w", err)
	}
	
	// Check package drift
	if err := checkPackageDrift(config, drift); err != nil {
		return nil, fmt.Errorf("failed to check package drift: %w", err)
	}
	
	return drift, nil
}

// checkDotfileDrift compares dotfiles between source and target locations
func checkDotfileDrift(plonkDir string, config *config.YAMLConfig, drift *DriftReport) error {
	dotfileTargets := config.GetDotfileTargets()
	
	for source, target := range dotfileTargets {
		sourcePath := filepath.Join(plonkDir, source)
		targetPath := expandHomeDir(target)
		
		// Check if target file exists
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			drift.MissingFiles = append(drift.MissingFiles, target)
			continue
		}
		
		// Check if files are different
		if different, err := filesAreDifferent(sourcePath, targetPath); err != nil {
			return err
		} else if different {
			drift.ModifiedFiles = append(drift.ModifiedFiles, target)
		}
	}
	
	// Check package configuration files
	for _, pkg := range config.Homebrew.Brews {
		if pkg.Config != "" {
			if err := checkPackageConfigDrift(plonkDir, pkg.Config, drift); err != nil {
				return err
			}
		}
	}
	
	for _, pkg := range config.Homebrew.Casks {
		if pkg.Config != "" {
			if err := checkPackageConfigDrift(plonkDir, pkg.Config, drift); err != nil {
				return err
			}
		}
	}
	
	for _, tool := range config.ASDF {
		if tool.Config != "" {
			if err := checkPackageConfigDrift(plonkDir, tool.Config, drift); err != nil {
				return err
			}
		}
	}
	
	for _, pkg := range config.NPM {
		if pkg.Config != "" {
			if err := checkPackageConfigDrift(plonkDir, pkg.Config, drift); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// checkPackageConfigDrift checks drift for package-specific configuration directories
func checkPackageConfigDrift(plonkDir, configPath string, drift *DriftReport) error {
	sourcePath := filepath.Join(plonkDir, configPath)
	targetPath := expandHomeDir("~/." + configPath)
	
	// Check if source exists
	sourceInfo, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		return nil // No source config, skip
	}
	if err != nil {
		return err
	}
	
	// Check if target exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		drift.MissingFiles = append(drift.MissingFiles, "~/."+configPath)
		return nil
	}
	
	// If source is a directory, check recursively
	if sourceInfo.IsDir() {
		return checkDirectoryDrift(sourcePath, targetPath, "~/."+configPath, drift)
	} else {
		// Single file
		if different, err := filesAreDifferent(sourcePath, targetPath); err != nil {
			return err
		} else if different {
			drift.ModifiedFiles = append(drift.ModifiedFiles, "~/."+configPath)
		}
	}
	
	return nil
}

// checkDirectoryDrift recursively compares directories
func checkDirectoryDrift(sourceDir, targetDir, basePath string, drift *DriftReport) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		sourceFile := filepath.Join(sourceDir, entry.Name())
		targetFile := filepath.Join(targetDir, entry.Name())
		relativePath := filepath.Join(basePath, entry.Name())
		
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			drift.MissingFiles = append(drift.MissingFiles, relativePath)
			continue
		}
		
		if entry.IsDir() {
			if err := checkDirectoryDrift(sourceFile, targetFile, relativePath, drift); err != nil {
				return err
			}
		} else {
			if different, err := filesAreDifferent(sourceFile, targetFile); err != nil {
				return err
			} else if different {
				drift.ModifiedFiles = append(drift.ModifiedFiles, relativePath)
			}
		}
	}
	
	return nil
}

// checkPackageDrift compares installed packages with configuration
func checkPackageDrift(config *config.YAMLConfig, drift *DriftReport) error {
	executor := &managers.RealCommandExecutor{}
	
	// Check Homebrew packages
	if len(config.Homebrew.Brews) > 0 || len(config.Homebrew.Casks) > 0 {
		homebrewMgr := managers.NewHomebrewManager(executor)
		if homebrewMgr.IsAvailable() {
			if err := checkHomebrewDrift(homebrewMgr, config, drift); err != nil {
				return err
			}
		} else {
			// Add all packages as missing if Homebrew not available
			for _, pkg := range config.Homebrew.Brews {
				drift.MissingPackages = append(drift.MissingPackages, "homebrew/"+pkg.Name)
			}
			for _, pkg := range config.Homebrew.Casks {
				drift.MissingPackages = append(drift.MissingPackages, "homebrew-cask/"+pkg.Name)
			}
		}
	}
	
	// Check ASDF tools
	if len(config.ASDF) > 0 {
		asdfMgr := managers.NewAsdfManager(executor)
		if asdfMgr.IsAvailable() {
			if err := checkASDFDrift(asdfMgr, config, drift); err != nil {
				return err
			}
		} else {
			// Add all tools as missing if ASDF not available
			for _, tool := range config.ASDF {
				drift.MissingPackages = append(drift.MissingPackages, "asdf/"+tool.Name+"@"+tool.Version)
			}
		}
	}
	
	// Check NPM packages
	if len(config.NPM) > 0 {
		npmMgr := managers.NewNpmManager(executor)
		if npmMgr.IsAvailable() {
			if err := checkNPMDrift(npmMgr, config, drift); err != nil {
				return err
			}
		} else {
			// Add all packages as missing if NPM not available
			for _, pkg := range config.NPM {
				packageName := pkg.Name
				if pkg.Package != "" {
					packageName = pkg.Package
				}
				drift.MissingPackages = append(drift.MissingPackages, "npm/"+packageName)
			}
		}
	}
	
	return nil
}

// checkHomebrewDrift compares Homebrew packages
func checkHomebrewDrift(mgr *managers.HomebrewManager, config *config.YAMLConfig, drift *DriftReport) error {
	installedPackages, err := mgr.ListInstalled()
	if err != nil {
		return err
	}
	
	installedMap := make(map[string]bool)
	for _, pkg := range installedPackages {
		installedMap[pkg] = true
	}
	
	// Check for missing packages
	for _, pkg := range config.Homebrew.Brews {
		if !installedMap[pkg.Name] {
			drift.MissingPackages = append(drift.MissingPackages, "homebrew/"+pkg.Name)
		}
	}
	
	for _, pkg := range config.Homebrew.Casks {
		if !installedMap[pkg.Name] {
			drift.MissingPackages = append(drift.MissingPackages, "homebrew-cask/"+pkg.Name)
		}
	}
	
	return nil
}

// checkASDFDrift compares ASDF tools and versions
func checkASDFDrift(mgr *managers.AsdfManager, config *config.YAMLConfig, drift *DriftReport) error {
	for _, tool := range config.ASDF {
		if !mgr.IsVersionInstalled(tool.Name, tool.Version) {
			drift.MissingPackages = append(drift.MissingPackages, "asdf/"+tool.Name+"@"+tool.Version)
		}
	}
	
	return nil
}

// checkNPMDrift compares NPM packages
func checkNPMDrift(mgr *managers.NpmManager, config *config.YAMLConfig, drift *DriftReport) error {
	installedPackages, err := mgr.ListInstalled()
	if err != nil {
		return err
	}
	
	installedMap := make(map[string]bool)
	for _, pkg := range installedPackages {
		installedMap[pkg] = true
	}
	
	// Check for missing packages
	for _, pkg := range config.NPM {
		packageName := pkg.Name
		if pkg.Package != "" {
			packageName = pkg.Package
		}
		
		if !installedMap[packageName] {
			drift.MissingPackages = append(drift.MissingPackages, "npm/"+packageName)
		}
	}
	
	return nil
}

// filesAreDifferent compares two files and returns true if they differ
func filesAreDifferent(file1, file2 string) (bool, error) {
	hash1, err := getFileHash(file1)
	if err != nil {
		return false, err
	}
	
	hash2, err := getFileHash(file2)
	if err != nil {
		return false, err
	}
	
	return hash1 != hash2, nil
}

// getFileHash returns SHA256 hash of a file
func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

