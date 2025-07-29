// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// HealthStatus represents overall system health
type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// HealthReport represents the complete health check report
type HealthReport struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

// RunHealthChecks performs comprehensive system health checks
func RunHealthChecks() HealthReport {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report := HealthReport{
		Overall: HealthStatus{
			Status:  "healthy",
			Message: "All systems operational",
		},
		Checks: []HealthCheck{},
	}

	// System checks
	report.Checks = append(report.Checks, checkSystemRequirements())
	report.Checks = append(report.Checks, checkEnvironmentVariables())
	report.Checks = append(report.Checks, checkPermissions())

	// Configuration checks
	report.Checks = append(report.Checks, checkConfigurationFile())
	report.Checks = append(report.Checks, checkConfigurationValidity())

	// Lock file checks
	report.Checks = append(report.Checks, checkLockFile())
	report.Checks = append(report.Checks, checkLockFileValidity())

	// Package manager checks
	report.Checks = append(report.Checks, checkPackageManagerAvailability(ctx))

	// Path and executable checks
	report.Checks = append(report.Checks, checkExecutablePath())
	report.Checks = append(report.Checks, checkPathConfiguration())

	// Determine overall health
	report.Overall = calculateOverallHealth(report.Checks)

	return report
}

// checkSystemRequirements checks basic system requirements
func checkSystemRequirements() HealthCheck {
	check := HealthCheck{
		Name:     "System Requirements",
		Category: "system",
		Status:   "pass",
		Message:  "System requirements met",
	}

	var issues []string
	var suggestions []string

	// Check OS support
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		issues = append(issues, fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS))
		suggestions = append(suggestions, "plonk is designed for macOS and Linux systems")
		check.Status = "fail"
	}

	// Check Go version (if available)
	if goVersion := runtime.Version(); goVersion != "" {
		check.Details = append(check.Details, fmt.Sprintf("Go version: %s", goVersion))
	}

	check.Details = append(check.Details,
		fmt.Sprintf("OS: %s", runtime.GOOS),
		fmt.Sprintf("Architecture: %s", runtime.GOARCH),
	)

	if len(issues) > 0 {
		check.Issues = issues
		check.Suggestions = suggestions
		check.Message = "System requirements not met"
	}

	return check
}

// checkEnvironmentVariables checks important environment variables
func checkEnvironmentVariables() HealthCheck {
	check := HealthCheck{
		Name:     "Environment Variables",
		Category: "environment",
		Status:   "pass",
		Message:  "Environment variables configured",
	}

	// Check important environment variables
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	check.Details = append(check.Details,
		fmt.Sprintf("HOME: %s", homeDir),
		fmt.Sprintf("PLONK_DIR: %s", configDir),
	)

	// Check PATH environment variable
	path := os.Getenv("PATH")
	if path == "" {
		check.Status = "fail"
		check.Issues = append(check.Issues, "PATH environment variable is not set")
		check.Suggestions = append(check.Suggestions, "Set PATH environment variable in your shell configuration")
		check.Message = "Critical environment variables missing"
	} else {
		pathDirs := strings.Split(path, string(os.PathListSeparator))
		check.Details = append(check.Details, fmt.Sprintf("PATH directories: %d", len(pathDirs)))
	}

	return check
}

// checkPermissions checks file and directory permissions
func checkPermissions() HealthCheck {
	check := HealthCheck{
		Name:     "Permissions",
		Category: "permissions",
		Status:   "pass",
		Message:  "File permissions are correct",
	}

	configDir := config.GetConfigDir()

	// Check if config directory exists and is writable
	if err := os.MkdirAll(configDir, 0755); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot create config directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Check permissions for the config directory")
		check.Message = "Permission issues detected"
		return check
	}

	// Test write access
	testFile := filepath.Join(configDir, ".plonk-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot write to config directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Ensure config directory is writable")
		check.Message = "Config directory is not writable"
	} else {
		os.Remove(testFile) // Clean up test file
		check.Details = append(check.Details, "Config directory is writable")
	}

	return check
}

// checkConfigurationFile checks for the existence and basic properties of the config file
func checkConfigurationFile() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration file exists",
	}

	configDir := config.GetConfigDir()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		check.Status = "info"
		check.Message = "Configuration file does not exist (using defaults)"
		check.Details = append(check.Details, "Will use default configuration")
		return check
	}

	// Check if file is readable
	if content, err := os.ReadFile(configPath); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read config file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Configuration file is not readable"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("Config file size: %d bytes", len(content)))
	}

	return check
}

// checkConfigurationValidity validates the configuration file format and content
func checkConfigurationValidity() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration Validity",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration is valid",
	}

	configDir := config.GetConfigDir()

	// Try to load the configuration
	cfg, err := config.Load(configDir)
	if err != nil {
		// If file doesn't exist, that's okay - we use defaults
		if os.IsNotExist(err) {
			check.Status = "info"
			check.Message = "No config file found (using defaults)"
			return check
		}

		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Configuration is invalid: %v", err))
		check.Suggestions = append(check.Suggestions, "Validate config file format or regenerate with 'plonk init'")
		check.Message = "Configuration has format errors"
		return check
	}

	// Validate configuration content
	if cfg.DefaultManager != "" {
		check.Details = append(check.Details, fmt.Sprintf("Default manager: %s", cfg.DefaultManager))
	}

	if len(cfg.IgnorePatterns) > 0 {
		check.Details = append(check.Details, fmt.Sprintf("Ignore patterns: %d", len(cfg.IgnorePatterns)))
	}

	return check
}

// checkLockFile checks for the existence and basic properties of the lock file
func checkLockFile() HealthCheck {
	check := HealthCheck{
		Name:     "Lock File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Lock file exists",
	}

	configDir := config.GetConfigDir()
	lockPath := filepath.Join(configDir, "plonk.lock")

	// Check if lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		check.Status = "info"
		check.Message = "Lock file does not exist (will be created when packages are added)"
		check.Details = append(check.Details, "Lock file will be automatically created when you add packages")
		return check
	}

	// Check if file is readable
	if content, err := os.ReadFile(lockPath); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read lock file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Lock file is not readable"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("Lock file size: %d bytes", len(content)))

		// Basic file integrity check
		if len(content) == 0 {
			check.Status = "warn"
			check.Message = "Lock file is empty"
			check.Details = append(check.Details, "No packages currently managed")
		}
	}

	return check
}

// checkLockFileValidity validates the lock file format and content
func checkLockFileValidity() HealthCheck {
	check := HealthCheck{
		Name:     "Lock File Validity",
		Category: "configuration",
		Status:   "pass",
		Message:  "Lock file is valid",
	}

	configDir := config.GetConfigDir()
	lockService := lock.NewYAMLLockService(configDir)

	// Try to load the lock file
	lockFile, err := lockService.Read()
	if err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Lock file is invalid: %v", err))
		check.Suggestions = append(check.Suggestions, "Validate lock file format or regenerate by running 'plonk pkg add' commands")
		check.Message = "Lock file has format errors"
		return check
	}

	// Count packages by manager
	totalPackages := 0
	managerCounts := make(map[string]int)
	for _, resource := range lockFile.Resources {
		if resource.Type == "package" {
			if manager, ok := resource.Metadata["manager"].(string); ok {
				managerCounts[manager]++
				totalPackages++
			}
		}
	}

	// Add manager counts to details
	for manager, count := range managerCounts {
		check.Details = append(check.Details, fmt.Sprintf("%s packages: %d", manager, count))
	}

	check.Details = append(check.Details, fmt.Sprintf("Total managed packages: %d", totalPackages))
	check.Details = append(check.Details, fmt.Sprintf("Lock file version: %d", lockFile.Version))

	if totalPackages == 0 {
		check.Status = "info"
		check.Message = "Lock file is valid but contains no packages"
	}

	return check
}

// checkPackageManagerAvailability checks which package managers are available
func checkPackageManagerAvailability(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:     "Package Manager Availability",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Package managers detected",
	}

	registry := packages.NewManagerRegistry()
	availableManagers := []string{}
	unavailableManagers := []string{}

	for _, managerName := range packages.SupportedManagers {
		mgr, err := registry.GetManager(managerName)
		if err != nil {
			unavailableManagers = append(unavailableManagers, managerName)
			continue
		}

		available, err := mgr.IsAvailable(ctx)
		if err != nil || !available {
			unavailableManagers = append(unavailableManagers, managerName)
			check.Details = append(check.Details, fmt.Sprintf("%s: ❌", managerName))
		} else {
			availableManagers = append(availableManagers, managerName)
			check.Details = append(check.Details, fmt.Sprintf("%s: ✅", managerName))
		}
	}

	if len(availableManagers) == 0 {
		check.Status = "fail"
		check.Message = "No package managers available"
		check.Issues = append(check.Issues, "No supported package managers found")
		check.Suggestions = append(check.Suggestions, "Install at least one package manager (brew, npm, pip, cargo, gem, or go)")
	} else {
		check.Message = fmt.Sprintf("%d package managers available", len(availableManagers))
		if len(unavailableManagers) > 0 {
			check.Suggestions = append(check.Suggestions, "Consider installing additional package managers for broader package support")
		}
	}

	return check
}

// checkExecutablePath checks if plonk executable is accessible
func checkExecutablePath() HealthCheck {
	check := HealthCheck{
		Name:     "Executable Path",
		Category: "installation",
		Status:   "pass",
		Message:  "Executable is accessible",
	}

	// Try to find plonk in PATH
	plonkPath, err := exec.LookPath("plonk")
	if err != nil {
		check.Status = "warn"
		check.Issues = append(check.Issues, "plonk executable not found in PATH")
		check.Suggestions = append(check.Suggestions, "Add plonk installation directory to PATH")
		check.Message = "Executable not in PATH"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("plonk found at: %s", plonkPath))
	}

	return check
}

// checkPathConfiguration checks PATH configuration for installed packages
func checkPathConfiguration() HealthCheck {
	check := HealthCheck{
		Name:     "PATH Configuration",
		Category: "installation",
		Status:   "pass",
		Message:  "PATH is properly configured",
	}

	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, string(os.PathListSeparator))
	homeDir := config.GetHomeDir()

	// Check for Python user bin directory
	pythonUserBin := getPythonUserBinDir()

	// Check for Go bin directory (GOBIN or GOPATH/bin)
	goBinDir := getGoBinDir()

	// Define important paths for each package manager
	importantPaths := map[string]string{
		"System":     "/usr/local/bin",
		"Homebrew":   "/opt/homebrew/bin",
		"Cargo":      filepath.Join(homeDir, ".cargo/bin"),
		"Go":         goBinDir,
		"Python/pip": pythonUserBin,
		"Gem":        filepath.Join(homeDir, ".gem/ruby/bin"),
		"NPM":        filepath.Join(homeDir, ".npm-global/bin"),
	}

	missingPaths := []string{}
	for name, importantPath := range importantPaths {
		// Skip empty paths
		if importantPath == "" {
			continue
		}

		found := false
		for _, pathDir := range pathDirs {
			if pathDir == importantPath {
				found = true
				break
			}
		}

		// Check if directory exists
		dirExists := false
		if _, err := os.Stat(importantPath); err == nil {
			dirExists = true
		}

		if found {
			check.Details = append(check.Details, fmt.Sprintf("✅ %s: %s", name, importantPath))
		} else if dirExists {
			check.Details = append(check.Details, fmt.Sprintf("⚠️  %s: %s (exists but not in PATH)", name, importantPath))
			missingPaths = append(missingPaths, importantPath)
			check.Status = "warn"
		} else {
			check.Details = append(check.Details, fmt.Sprintf("ℹ️  %s: %s (directory does not exist)", name, importantPath))
		}
	}

	check.Details = append(check.Details, fmt.Sprintf("\nTotal PATH directories: %d", len(pathDirs)))

	if len(missingPaths) > 0 {
		check.Status = "warn"
		check.Message = "Some package directories are not in PATH"
		check.Issues = append(check.Issues, "The following directories exist but are not in PATH:")
		for _, path := range missingPaths {
			check.Issues = append(check.Issues, fmt.Sprintf("  - %s", path))
		}

		// Detect user's shell and provide exact commands
		shellPath := os.Getenv("SHELL")
		shell := detectShell(shellPath)

		check.Suggestions = append(check.Suggestions, fmt.Sprintf("Detected shell: %s", shell.name))
		check.Suggestions = append(check.Suggestions, "\nCopy and run these commands to fix your PATH:")

		// Generate the PATH export line
		pathExport := generatePathExport(missingPaths)

		// Provide shell-specific commands
		commands := generateShellCommands(shell, pathExport)
		for _, cmd := range commands {
			check.Suggestions = append(check.Suggestions, fmt.Sprintf("  %s", cmd))
		}

		// Add instructions for immediate effect
		check.Suggestions = append(check.Suggestions, "\nOr for immediate effect in this session only:")
		check.Suggestions = append(check.Suggestions, fmt.Sprintf("  %s", pathExport))
	}

	return check
}

// shellInfo represents shell detection information
type shellInfo struct {
	name       string
	configFile string
	reload     string
}

// detectShell detects the user's shell from SHELL environment variable
func detectShell(shellPath string) shellInfo {
	// Default to bash if detection fails
	defaultShell := shellInfo{
		name:       "bash",
		configFile: "~/.bashrc",
		reload:     "source ~/.bashrc",
	}

	if shellPath == "" {
		return defaultShell
	}

	// Extract shell name from path
	shellName := filepath.Base(shellPath)

	switch shellName {
	case "zsh":
		return shellInfo{
			name:       "zsh",
			configFile: "~/.zshrc",
			reload:     "source ~/.zshrc",
		}
	case "bash":
		return defaultShell
	case "fish":
		return shellInfo{
			name:       "fish",
			configFile: "~/.config/fish/config.fish",
			reload:     "source ~/.config/fish/config.fish",
		}
	case "ksh":
		return shellInfo{
			name:       "ksh",
			configFile: "~/.kshrc",
			reload:     ". ~/.kshrc",
		}
	case "tcsh":
		return shellInfo{
			name:       "tcsh",
			configFile: "~/.tcshrc",
			reload:     "source ~/.tcshrc",
		}
	default:
		// Try to infer from common patterns
		if strings.Contains(shellPath, "zsh") {
			return shellInfo{
				name:       "zsh",
				configFile: "~/.zshrc",
				reload:     "source ~/.zshrc",
			}
		}
		return defaultShell
	}
}

// generatePathExport creates the PATH export line for missing paths
func generatePathExport(missingPaths []string) string {
	if len(missingPaths) == 0 {
		return ""
	}

	// Join all paths with colon
	pathString := strings.Join(missingPaths, ":")
	return fmt.Sprintf("export PATH=\"%s:$PATH\"", pathString)
}

// generateShellCommands generates shell-specific commands for PATH configuration
func generateShellCommands(shell shellInfo, pathExport string) []string {
	if shell.name == "fish" {
		// Fish shell has special syntax
		commands := []string{}
		// Extract paths from the export command
		// pathExport looks like: export PATH="/path1:/path2:$PATH"
		if strings.HasPrefix(pathExport, "export PATH=\"") && strings.HasSuffix(pathExport, ":$PATH\"") {
			pathString := strings.TrimPrefix(pathExport, "export PATH=\"")
			pathString = strings.TrimSuffix(pathString, ":$PATH\"")
			paths := strings.Split(pathString, ":")
			for _, path := range paths {
				if path != "" {
					commands = append(commands, fmt.Sprintf("fish_add_path %s", path))
				}
			}
		}
		return commands
	}

	// For most shells, we can append to config file
	commands := []string{
		fmt.Sprintf("echo '%s' >> %s", pathExport, shell.configFile),
		shell.reload,
	}

	return commands
}

// getPythonUserBinDir returns the Python user bin directory
func getPythonUserBinDir() string {
	// Try to get Python user base directory
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-m", "site", "--user-base")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	userBase := strings.TrimSpace(string(output))
	if userBase == "" {
		return ""
	}

	return filepath.Join(userBase, "bin")
}

// getGoBinDir returns the Go bin directory (GOBIN or GOPATH/bin)
func getGoBinDir() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First check GOBIN
	cmd := exec.CommandContext(ctx, "go", "env", "GOBIN")
	output, err := cmd.Output()
	if err == nil {
		gobin := strings.TrimSpace(string(output))
		if gobin != "" {
			return gobin
		}
	}

	// Fall back to GOPATH/bin
	cmd = exec.CommandContext(ctx, "go", "env", "GOPATH")
	output, err = cmd.Output()
	if err == nil {
		gopath := strings.TrimSpace(string(output))
		if gopath != "" {
			return filepath.Join(gopath, "bin")
		}
	}

	// Default to ~/go/bin
	return filepath.Join(config.GetHomeDir(), "go/bin")
}

// calculateOverallHealth determines overall system health from individual checks
func calculateOverallHealth(checks []HealthCheck) HealthStatus {
	hasFailure := false
	hasWarning := false

	for _, check := range checks {
		switch check.Status {
		case "fail":
			hasFailure = true
		case "warn":
			hasWarning = true
		}
	}

	if hasFailure {
		return HealthStatus{
			Status:  "unhealthy",
			Message: "Critical issues detected",
		}
	}

	if hasWarning {
		return HealthStatus{
			Status:  "warning",
			Message: "Some issues detected",
		}
	}

	return HealthStatus{
		Status:  "healthy",
		Message: "All systems operational",
	}
}
