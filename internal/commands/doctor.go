// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

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
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and diagnose issues",
	Long: `Run comprehensive health checks to diagnose potential issues with plonk.

This command checks:
- System requirements and environment
- Package manager availability and functionality
- Configuration file validity and permissions
- Common path and permission issues
- Suggested fixes for detected problems

The doctor command helps troubleshoot installation and configuration issues,
making it easier to get plonk working correctly on your system.

Examples:
  plonk doctor              # Run all health checks
  plonk doctor -o json      # Output results as JSON for scripting`,
	RunE: runDoctor,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "doctor", "output-format", "invalid output format")
	}

	// Run comprehensive health checks
	healthReport := runHealthChecks()

	return RenderOutput(healthReport, format)
}

// runHealthChecks performs comprehensive system health checks
func runHealthChecks() DoctorOutput {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report := DoctorOutput{
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
	report.Checks = append(report.Checks, checkPackageManagerFunctionality(ctx))

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

	var warnings []string
	var suggestions []string

	// Check essential variables
	home := os.Getenv("HOME")
	if home == "" {
		check.Status = "fail"
		check.Issues = append(check.Issues, "HOME environment variable not set")
		check.Suggestions = append(check.Suggestions, "Set HOME environment variable to your home directory")
	} else {
		check.Details = append(check.Details, fmt.Sprintf("HOME: %s", home))
	}

	// Check PATH
	path := os.Getenv("PATH")
	if path == "" {
		check.Status = "fail"
		check.Issues = append(check.Issues, "PATH environment variable not set")
		check.Suggestions = append(check.Suggestions, "Set PATH environment variable to include system binaries")
	} else {
		check.Details = append(check.Details, fmt.Sprintf("PATH entries: %d", len(strings.Split(path, string(os.PathListSeparator)))))
	}

	// Check optional but useful variables
	if os.Getenv("EDITOR") == "" && os.Getenv("VISUAL") == "" {
		warnings = append(warnings, "No EDITOR or VISUAL environment variable set")
		suggestions = append(suggestions, "Set EDITOR environment variable for better config editing experience")
	}

	if len(warnings) > 0 && check.Status == "pass" {
		check.Status = "warn"
		check.Issues = warnings
		check.Suggestions = suggestions
		check.Message = "Environment variables have warnings"
	}

	return check
}

// checkPermissions checks file and directory permissions
func checkPermissions() HealthCheck {
	check := HealthCheck{
		Name:     "File Permissions",
		Category: "permissions",
		Status:   "pass",
		Message:  "File permissions are correct",
	}

	configDir := config.GetDefaultConfigDirectory()

	// Check if config directory exists and is writable
	if info, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			check.Details = append(check.Details, "Config directory does not exist (will be created when needed)")
		} else {
			check.Status = "warn"
			check.Issues = append(check.Issues, fmt.Sprintf("Cannot access config directory: %v", err))
			check.Suggestions = append(check.Suggestions, "Check permissions on ~/.config directory")
		}
	} else {
		if !info.IsDir() {
			check.Status = "fail"
			check.Issues = append(check.Issues, "Config path exists but is not a directory")
			check.Suggestions = append(check.Suggestions, "Remove ~/.config/plonk file and recreate as directory")
		} else {
			check.Details = append(check.Details, fmt.Sprintf("Config directory: %s", configDir))
		}
	}

	// Test write permissions by creating a temp file
	tempFile := filepath.Join(configDir, ".plonk_doctor_test")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot create config directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Check permissions on ~/.config directory")
	} else {
		if err := os.WriteFile(tempFile, []byte("test"), 0600); err != nil {
			check.Status = "fail"
			check.Issues = append(check.Issues, fmt.Sprintf("Cannot write to config directory: %v", err))
			check.Suggestions = append(check.Suggestions, "Check write permissions on ~/.config/plonk directory")
		} else {
			_ = os.Remove(tempFile) // Clean up
			check.Details = append(check.Details, "Config directory is writable")
		}
	}

	return check
}

// checkConfigurationFile checks if configuration file exists and is accessible
func checkConfigurationFile() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration file is accessible",
	}

	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Check if config file exists
	if info, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			check.Status = "warn"
			check.Issues = append(check.Issues, "Configuration file does not exist")
			check.Suggestions = append(check.Suggestions, "Run 'plonk init' to create a configuration file")
			check.Details = append(check.Details, fmt.Sprintf("Expected location: %s", configPath))
		} else {
			check.Status = "fail"
			check.Issues = append(check.Issues, fmt.Sprintf("Cannot access configuration file: %v", err))
			check.Suggestions = append(check.Suggestions, "Check file permissions and path")
		}
	} else {
		check.Details = append(check.Details,
			fmt.Sprintf("Config file: %s", configPath),
			fmt.Sprintf("Size: %d bytes", info.Size()),
			fmt.Sprintf("Modified: %s", info.ModTime().Format("2006-01-02 15:04:05")),
		)

		// Check if file is readable
		if content, err := os.ReadFile(configPath); err != nil {
			check.Status = "fail"
			check.Issues = append(check.Issues, fmt.Sprintf("Cannot read configuration file: %v", err))
			check.Suggestions = append(check.Suggestions, "Check file permissions")
		} else {
			check.Details = append(check.Details, fmt.Sprintf("Content length: %d characters", len(content)))
		}
	}

	return check
}

// checkConfigurationValidity validates the configuration file
func checkConfigurationValidity() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration Validity",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration is valid",
	}

	configDir := config.GetDefaultConfigDirectory()

	// Try to load configuration
	cfg := config.LoadConfigWithDefaults(configDir)

	// Run validation
	validator := config.NewSimpleValidator()
	result := validator.ValidateConfig(cfg)

	if !result.Valid {
		check.Status = "fail"
		check.Issues = result.Errors
		check.Suggestions = append(check.Suggestions, "Run 'plonk config validate' for detailed error information")
		check.Message = "Configuration has validation errors"
	} else {
		// Count configured items - packages now in lock file
		packageCount := getPackageCountFromLockFile(configDir)

		// Get auto-discovered dotfiles
		adapter := config.NewConfigAdapter(cfg)
		dotfileTargets := adapter.GetDotfileTargets()
		dotfileCount := len(dotfileTargets)

		check.Details = append(check.Details,
			fmt.Sprintf("Default manager: %s", cfg.Resolve().GetDefaultManager()),
			fmt.Sprintf("Configured packages: %d", packageCount),
			fmt.Sprintf("Auto-discovered dotfiles: %d", dotfileCount),
		)

		if len(result.Warnings) > 0 {
			check.Status = "warn"
			check.Issues = result.Warnings
			check.Message = "Configuration is valid but has warnings"
		}
	}

	return check
}

// checkPackageManagerAvailability checks if package managers are available
func checkPackageManagerAvailability(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:     "Package Manager Availability",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Package managers are available",
	}

	registry := managers.NewManagerRegistry()
	managerMap := make(map[string]managers.PackageManager)
	for _, name := range registry.GetAllManagerNames() {
		if manager, err := registry.GetManager(name); err == nil {
			managerMap[name] = manager
		}
	}

	availableCount := 0
	for name, manager := range managerMap {
		available, err := manager.IsAvailable(ctx)
		if err != nil {
			check.Issues = append(check.Issues, fmt.Sprintf("%s: %v", name, err))
			check.Status = "warn"
		} else if available {
			availableCount++
			check.Details = append(check.Details, fmt.Sprintf("%s: ✅ Available", name))
		} else {
			check.Details = append(check.Details, fmt.Sprintf("%s: ❌ Not available", name))
		}
	}

	if availableCount == 0 {
		check.Status = "fail"
		check.Issues = append(check.Issues, "No package managers are available")
		check.Suggestions = append(check.Suggestions, "Install Homebrew or NPM to manage packages")
		check.Message = "No package managers available"
	} else if availableCount < len(managerMap) {
		if check.Status == "pass" {
			check.Status = "warn"
		}
		check.Suggestions = append(check.Suggestions, "Consider installing additional package managers for better compatibility")
		check.Message = "Some package managers are not available"
	}

	return check
}

// checkPackageManagerFunctionality tests basic package manager functionality
func checkPackageManagerFunctionality(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:     "Package Manager Functionality",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Package managers are functional",
	}

	registry := managers.NewManagerRegistry()
	managerMap := make(map[string]managers.PackageManager)
	for _, name := range registry.GetAllManagerNames() {
		if manager, err := registry.GetManager(name); err == nil {
			managerMap[name] = manager
		}
	}

	for name, manager := range managerMap {
		available, err := manager.IsAvailable(ctx)
		if err != nil || !available {
			continue // Skip unavailable managers
		}

		// Test basic functionality
		if packages, err := manager.ListInstalled(ctx); err != nil {
			check.Status = "warn"
			check.Issues = append(check.Issues, fmt.Sprintf("%s: Cannot list installed packages: %v", name, err))
			check.Suggestions = append(check.Suggestions, fmt.Sprintf("Check %s installation and permissions", name))
		} else {
			check.Details = append(check.Details, fmt.Sprintf("%s: Listed %d installed packages", name, len(packages)))
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

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		check.Status = "warn"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot determine executable path: %v", err))
	} else {
		check.Details = append(check.Details, fmt.Sprintf("Executable: %s", execPath))

		// Check if executable is in PATH
		if pathExec, err := exec.LookPath("plonk"); err != nil {
			check.Status = "warn"
			check.Issues = append(check.Issues, "plonk executable not found in PATH")
			check.Suggestions = append(check.Suggestions, "Add plonk to your PATH or use full path to executable")
		} else {
			check.Details = append(check.Details, fmt.Sprintf("Found in PATH: %s", pathExec))
		}
	}

	return check
}

// checkPathConfiguration checks PATH configuration for common issues
func checkPathConfiguration() HealthCheck {
	check := HealthCheck{
		Name:     "PATH Configuration",
		Category: "environment",
		Status:   "pass",
		Message:  "PATH is configured correctly",
	}

	path := os.Getenv("PATH")
	if path == "" {
		check.Status = "fail"
		check.Issues = append(check.Issues, "PATH environment variable is not set")
		return check
	}

	pathDirs := strings.Split(path, string(os.PathListSeparator))
	check.Details = append(check.Details, fmt.Sprintf("PATH contains %d directories", len(pathDirs)))

	// Check for common required directories
	requiredPaths := []string{
		"/usr/bin",
		"/usr/local/bin",
	}

	if runtime.GOOS == "darwin" {
		requiredPaths = append(requiredPaths, "/opt/homebrew/bin")
	}

	for _, reqPath := range requiredPaths {
		found := false
		for _, pathDir := range pathDirs {
			if pathDir == reqPath {
				found = true
				break
			}
		}
		if !found {
			check.Status = "warn"
			check.Issues = append(check.Issues, fmt.Sprintf("Required path not found: %s", reqPath))
			check.Suggestions = append(check.Suggestions, fmt.Sprintf("Add %s to your PATH", reqPath))
		}
	}

	return check
}

// calculateOverallHealth determines overall health based on individual checks
func calculateOverallHealth(checks []HealthCheck) HealthStatus {
	status := HealthStatus{
		Status:  "healthy",
		Message: "All systems operational",
	}

	failCount := 0
	warnCount := 0

	for _, check := range checks {
		switch check.Status {
		case "fail":
			failCount++
		case "warn":
			warnCount++
		}
	}

	if failCount > 0 {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("%d critical issues found", failCount)
	} else if warnCount > 0 {
		status.Status = "warning"
		status.Message = fmt.Sprintf("%d warnings found", warnCount)
	}

	return status
}

// Output structures

type DoctorOutput struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// TableOutput generates human-friendly table output for doctor command
func (d DoctorOutput) TableOutput() string {
	var output strings.Builder

	// Overall status
	output.WriteString("# Plonk Doctor Report\n\n")

	switch d.Overall.Status {
	case "healthy":
		output.WriteString("🟢 Overall Status: HEALTHY\n")
	case "warning":
		output.WriteString("🟡 Overall Status: WARNING\n")
	case "unhealthy":
		output.WriteString("🔴 Overall Status: UNHEALTHY\n")
	}
	output.WriteString(fmt.Sprintf("   %s\n\n", d.Overall.Message))

	// Group checks by category
	categories := make(map[string][]HealthCheck)
	for _, check := range d.Checks {
		categories[check.Category] = append(categories[check.Category], check)
	}

	// Display each category
	categoryOrder := []string{"system", "environment", "permissions", "configuration", "package-managers", "installation"}
	for _, category := range categoryOrder {
		if checks, exists := categories[category]; exists {
			output.WriteString(fmt.Sprintf("## %s\n", strings.Title(strings.ReplaceAll(category, "-", " "))))

			for _, check := range checks {
				// Status icon
				var icon string
				switch check.Status {
				case "pass":
					icon = "✅"
				case "warn":
					icon = "⚠️"
				case "fail":
					icon = "❌"
				}

				output.WriteString(fmt.Sprintf("%s **%s**: %s\n", icon, check.Name, check.Message))

				// Details
				if len(check.Details) > 0 {
					for _, detail := range check.Details {
						output.WriteString(fmt.Sprintf("   • %s\n", detail))
					}
				}

				// Issues
				if len(check.Issues) > 0 {
					output.WriteString("   Issues:\n")
					for _, issue := range check.Issues {
						output.WriteString(fmt.Sprintf("   ⚠️  %s\n", issue))
					}
				}

				// Suggestions
				if len(check.Suggestions) > 0 {
					output.WriteString("   Suggestions:\n")
					for _, suggestion := range check.Suggestions {
						output.WriteString(fmt.Sprintf("   💡 %s\n", suggestion))
					}
				}

				output.WriteString("\n")
			}
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (d DoctorOutput) StructuredData() any {
	return d
}

// getPackageCountFromLockFile counts packages in the lock file
func getPackageCountFromLockFile(configDir string) int {
	lockService := lock.NewYAMLLockService(configDir)

	totalCount := 0
	managers := []string{"homebrew", "npm", "cargo"}

	for _, manager := range managers {
		packages, err := lockService.GetPackages(manager)
		if err == nil {
			totalCount += len(packages)
		}
	}

	return totalCount
}

// checkLockFile checks for the existence and accessibility of the lock file
func checkLockFile() HealthCheck {
	check := HealthCheck{
		Name:     "Lock File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Lock file accessible",
	}

	configDir := config.GetDefaultConfigDirectory()
	lockPath := filepath.Join(configDir, "plonk.lock")

	check.Details = append(check.Details, fmt.Sprintf("Lock file path: %s", lockPath))

	// Check if lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		check.Status = "warn"
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

	configDir := config.GetDefaultConfigDirectory()
	lockService := lock.NewYAMLLockService(configDir)

	// Try to load the lock file
	lockFile, err := lockService.Load()
	if err != nil {
		// If file doesn't exist, that's okay
		if os.IsNotExist(err) {
			check.Status = "info"
			check.Message = "No lock file found (packages will be tracked when added)"
			return check
		}

		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Lock file is invalid: %v", err))
		check.Suggestions = append(check.Suggestions, "Validate lock file format or regenerate by running 'plonk pkg add' commands")
		check.Message = "Lock file has format errors"
		return check
	}

	// Count packages by manager
	totalPackages := 0
	for manager, packages := range lockFile.Packages {
		count := len(packages)
		totalPackages += count
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
